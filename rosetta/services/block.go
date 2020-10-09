package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
	"time"
)

// TimeOut for RPC Lotus calls
const LotusCallTimeOut = 40 * time.Second

// BlockCIDsKey is the name of the key in the Metadata map inside a
// BlockResponse that specifies blocks' CIDs inside a TipSet.
const BlockCIDsKey = "blockCIDs"

// BlockAPIService implements the server.BlockAPIServicer interface.
type BlockAPIService struct {
	network *types.NetworkIdentifier
	node    api.FullNode
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewBlockAPIService(network *types.NetworkIdentifier, api *api.FullNode) server.BlockAPIServicer {
	return &BlockAPIService{
		network: network,
		node:    *api,
	}
}

// Block implements the /block endpoint.
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {

	if request.BlockIdentifier == nil {
		return nil, BuildError(ErrMalformedValue, nil)
	}

	if request.BlockIdentifier == nil && request.BlockIdentifier.Hash == nil {
		return nil, BuildError(ErrInsufficientQueryInputs, nil)
	}

	errNet := ValidateNetworkId(ctx, &s.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	requestedHeight := *request.BlockIdentifier.Index
	if requestedHeight < 0 {
		return nil, BuildError(ErrMalformedValue, nil)
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &s.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if requestedHeight > 0 && !status.IsSynced() {
		return nil, BuildError(ErrUnableToGetUnsyncedBlock, nil)
	}

	if request.BlockIdentifier.Index == nil {
		return nil, BuildError(ErrInsufficientQueryInputs, nil)
	}

	var tipSet *filTypes.TipSet
	var err error
	impl := func() {
		tipSet, err = s.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(requestedHeight), filTypes.EmptyTSK)
	}

	errTimeOut := tools.WrapWithTimeout(impl, LotusCallTimeOut)
	if errTimeOut != nil {
		return nil, ErrLotusCallTimedOut
	}

	if err != nil {
		return nil, BuildError(ErrUnableToGetTipset, err)
	}

	//If a TipSet has empty blocks, lotus api will return a TipSet at a different epoch
	//Check if the retrieved TipSet is actually the requested one
	//details on: https://github.com/filecoin-project/lotus/blob/49d64f7f7e22973ca0cfbaaf337fcfb3c2d47707/api/api_full.go#L65-L67
	if int64(tipSet.Height()) != requestedHeight {
		return &types.BlockResponse{}, nil
	}

	if request.BlockIdentifier.Hash != nil {
		tipSetKeyHash, encErr := BuildTipSetKeyHash(tipSet.Key())
		if encErr != nil {
			return nil, BuildError(ErrUnableToBuildTipSetHash, encErr)
		}
		if *tipSetKeyHash != *request.BlockIdentifier.Hash {
			return nil, BuildError(ErrInvalidHash, nil)
		}
	}

	//Get parent TipSet
	var parentTipSet *filTypes.TipSet
	if requestedHeight > 0 {
		if tipSet.Parents().IsEmpty() {
			return nil, BuildError(ErrUnableToGetParentBlk, nil)
		}
		impl = func() {
			parentTipSet, err = s.node.ChainGetTipSet(ctx, tipSet.Parents())
		}
		errTimeOut = tools.WrapWithTimeout(impl, LotusCallTimeOut)
		if errTimeOut != nil {
			return nil, ErrLotusCallTimedOut
		}
		if err != nil {
			return nil, BuildError(ErrUnableToGetParentBlk, err)
		}
	} else {
		// According to rosetta docs, if the requested tipset is
		// the genesis one, set the same tipset as parent
		parentTipSet = tipSet
	}

	//Build transactions data
	var transactions *[]*types.Transaction
	if requestedHeight > 1 {
		states, err := getLotusStateCompute(ctx, &s.node, tipSet)
		if err != nil {
			return nil, err
		}
		transactions = buildTransactions(states)
	}

	//Add block metadata
	md := make(map[string]interface{})
	var blockCIDs []string
	for _, cid := range tipSet.Cids() {
		blockCIDs = append(blockCIDs, cid.String())
	}
	md[BlockCIDsKey] = blockCIDs

	hashTipSet, err := BuildTipSetKeyHash(tipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, nil)
	}
	blockId := &types.BlockIdentifier{
		Index: int64(tipSet.Height()),
		Hash:  *hashTipSet,
	}

	parentBlockId := &types.BlockIdentifier{}
	hashParentTipSet, err := BuildTipSetKeyHash(parentTipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, nil)
	}
	parentBlockId.Index = int64(parentTipSet.Height())
	parentBlockId.Hash = *hashParentTipSet

	respBlock := &types.Block{
		BlockIdentifier:       blockId,
		ParentBlockIdentifier: parentBlockId,
		Timestamp:             int64(tipSet.MinTimestamp()) * FactorSecondToMillisecond, // [ms]
		Metadata:              md,
	}
	if transactions != nil {
		respBlock.Transactions = *transactions
	}

	resp := &types.BlockResponse{
		Block: respBlock,
	}

	return resp, nil
}

func buildTransactions(states *api.ComputeStateOutput) *[]*types.Transaction {
	defer TimeTrack(time.Now(), "[Proxy]TraceAnalysis")

	var transactions []*types.Transaction
	for i := range states.Trace {
		trace := states.Trace[i]
		var operations []*types.Operation

		// Analyze full trace recursively
		processTrace(&trace.ExecutionTrace, &operations)

		if len(operations) > 0 {
			//Add the corresponding "Fee" operation
			if trace.MsgRct.GasUsed > 0 {
				fee := abi.NewTokenAmount(trace.MsgRct.GasUsed)
				opStatus := OperationStatusFailed
				if trace.MsgRct.ExitCode.IsSuccess() {
					opStatus = OperationStatusOk
				}
				operations = appendOp(operations, "Fee", trace.Msg.From.String(),
					fee.Neg().String(), opStatus, false)
			}

			transactions = append(transactions, &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: trace.Msg.Cid().String(),
				},
				Operations: operations,
			})
		}
	}
	return &transactions
}

func getLotusStateCompute(ctx context.Context, node *api.FullNode, tipSet *filTypes.TipSet) (*api.ComputeStateOutput, *types.Error) {
	defer TimeTrack(time.Now(), "[Lotus]StateCompute")

	//StateCompute includes the messages at height N-1.
	//So, we're getting the traces of the messages created at N-1, executed at N
	states, err := (*node).StateCompute(ctx, tipSet.Height(), nil, tipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToGetTrace, err)
	}
	return states, nil
}

func processTrace(trace *filTypes.ExecutionTrace, operations *[]*types.Operation) {
	baseMethod, err := GetMethodName(trace.Msg)
	if err != nil {
		return
	}

	if IsOpSupported(baseMethod) {
		fromPk, err1 := GetActorPubKey(trace.Msg.From)
		toPk, err2 := GetActorPubKey(trace.Msg.To)
		if err1 != nil || err2 != nil {
			Logger.Error("could not retrieve one or both pubkeys for addresses:",
				trace.Msg.From.String(), trace.Msg.To.String())
			return
		}

		opStatus := OperationStatusFailed
		if trace.MsgRct.ExitCode.IsSuccess() {
			opStatus = OperationStatusOk
		}

		switch baseMethod {
		case "Send":
			{
				*operations = appendOp(*operations, baseMethod, fromPk,
					trace.Msg.Value.Neg().String(), opStatus, true)
				*operations = appendOp(*operations, baseMethod, toPk,
					trace.Msg.Value.String(), opStatus, true)
			}
		case "Propose":
			{
				*operations = appendOp(*operations, baseMethod, fromPk,
					"0", opStatus, true)
			}
		case "SwapSigner":
			{
				*operations = appendOp(*operations, baseMethod, fromPk,
					"0", opStatus, true)
				*operations = appendOp(*operations, baseMethod, toPk,
					"0", opStatus, true)
			}
		case "AwardBlockReward", "OnDeferredCronEvent":
			{
				*operations = appendOp(*operations, baseMethod, toPk,
					trace.Msg.Value.String(), opStatus, true)
			}
		}
	}

	for i := range trace.Subcalls {
		subTrace := trace.Subcalls[i]
		processTrace(&subTrace, operations)
	}
}

func appendOp(ops []*types.Operation, opType string, account string, amount string, status string, relateOp bool) []*types.Operation {
	opIndex := int64(len(ops))
	op := &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index: opIndex,
		},
		Type:   opType,
		Status: status,
		Account: &types.AccountIdentifier{
			Address: account,
		},
		Amount: &types.Amount{
			Value:    amount,
			Currency: GetCurrencyData(),
		},
	}

	// Add related operation
	if relateOp && opIndex > 0 {
		op.RelatedOperations = []*types.OperationIdentifier{
			{
				Index: opIndex - 1,
			},
		}
	}

	return append(ops, op)
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	return nil, ErrNotImplemented
}
