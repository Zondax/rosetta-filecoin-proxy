package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
	"time"
)

// TimeOut for RPC Lotus calls
const LotusCallTimeOut = 4 * time.Second

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
		return nil, ErrMalformedValue
	}

	if request.BlockIdentifier == nil && request.BlockIdentifier.Hash == nil {
		return nil, ErrInsufficientQueryInputs
	}

	errNet := ValidateNetworkId(ctx, &s.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	requestedHeight := *request.BlockIdentifier.Index
	if requestedHeight < 0 {
		return nil, ErrMalformedValue
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &s.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if requestedHeight > 0 && !status.IsSynced() {
		return nil, ErrUnableToGetUnsyncedBlock
	}

	if request.BlockIdentifier.Index == nil {
		return nil, ErrInsufficientQueryInputs
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
		return nil, ErrUnableToGetTipset
	}

	if request.BlockIdentifier.Hash != nil {
		tipSetKeyHash, encErr := BuildTipSetKeyHash(tipSet.Key())
		if encErr != nil {
			return nil, ErrUnableToBuildTipSetHash
		}
		if *tipSetKeyHash != *request.BlockIdentifier.Hash {
			return nil, ErrInvalidHash
		}
	}

	//Get parent TipSet
	var parentTipSet *filTypes.TipSet
	if requestedHeight > 0 {
		if tipSet.Parents().IsEmpty() {
			return nil, ErrUnableToGetParentBlk
		}
		impl = func() {
			parentTipSet, err = s.node.ChainGetTipSet(ctx, tipSet.Parents())
		}
		errTimeOut = tools.WrapWithTimeout(impl, LotusCallTimeOut)
		if errTimeOut != nil {
			return nil, ErrLotusCallTimedOut
		}
		if err != nil {
			return nil, ErrUnableToGetParentBlk
		}
	} else {
		// According to rosetta docs, if the requested tipset is
		// the genesis one, set the same tipset as parent
		parentTipSet = tipSet
	}

	//Get executed transactions
	var transactions []*types.Transaction
	block := tipSet.Blocks()[0] // All blocks share the same parent TipSet
	messages, err := s.node.ChainGetParentMessages(ctx, block.Cid())
	if err != nil {
		return nil, ErrUnableToGetTxns
	}
	receipts, err := s.node.ChainGetParentReceipts(ctx, block.Cid())
	if err != nil {
		return nil, ErrUnableToGetTxnReceipt
	}
	if len(messages) != len(receipts) {
		return nil, ErrMsgsAndReceiptsCountMismatch
	}

	for i := range messages {
		var opStatus string
		msg := messages[i]

		if receipts[i].ExitCode.IsSuccess() {
			opStatus = OperationStatusOk
		} else {
			opStatus = OperationStatusFailed
		}

		transactions = append(transactions, &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: msg.Cid.String(),
			},
			Operations: []*types.Operation{},
		})

		opType, err := GetMethodName(&msg, &s.node)
		if err != nil {
			return nil, err
		}

		transactions[i].Operations = appendOp(transactions[i].Operations, opType,
			msg.Message.From.String(), msg.Message.Value.String(), opStatus)
		transactions[i].Operations = appendOp(transactions[i].Operations, opType,
			msg.Message.To.String(), msg.Message.Value.String(), opStatus)
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
		return nil, ErrUnableToBuildTipSetHash
	}
	blockId := &types.BlockIdentifier{
		Index: int64(tipSet.Height()),
		Hash:  *hashTipSet,
	}

	parentBlockId := &types.BlockIdentifier{}
	hashParentTipSet, err := BuildTipSetKeyHash(parentTipSet.Key())
	if err != nil {
		return nil, ErrUnableToBuildTipSetHash
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
		respBlock.Transactions = transactions
	}

	resp := &types.BlockResponse{
		Block: respBlock,
	}

	return resp, nil
}

func appendOp(ops []*types.Operation, opType string, account string, amount string, status string) []*types.Operation {
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
	if opIndex >= 1 {
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
