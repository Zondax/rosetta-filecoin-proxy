// +build rosetta_rpc

package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
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
	node api.FullNode
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewBlockAPIService(network *types.NetworkIdentifier, api *api.FullNode) server.BlockAPIServicer {
	return &BlockAPIService{
		network: network,
		node: *api,
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

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &s.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if !status.IsSynced() {
		return nil, ErrUnableToGetUnsyncedBlock
	}

	if request.BlockIdentifier.Index == nil {
		return nil, ErrInsufficientQueryInputs
	}

	requestedHeight := *request.BlockIdentifier.Index
	if requestedHeight < 0 {
		return nil, ErrMalformedValue
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
	if requestedHeight > 0 {
		for _, block := range tipSet.Blocks() {
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
			for i, msg := range messages {
				transactions = append(transactions, BuildTransaction(&msg, receipts[i], &s.node))
			}
		}
	}

	//Add block metadata
	md := make(map[string]interface{})
	var blockCIDs[] string
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
		BlockIdentifier: blockId,
		ParentBlockIdentifier: parentBlockId,
		Timestamp:    int64(tipSet.MinTimestamp()) * FactorSecondToMillisecond, // [ms]
		Metadata: md,
	}
	if transactions != nil {
		respBlock.Transactions = transactions
	}

	resp := &types.BlockResponse{
		Block: respBlock,
	}

	return resp, nil
}

func BuildTransaction(msg *api.Message, receipt *filTypes.MessageReceipt, api *api.FullNode) *types.Transaction {
	var transaction types.Transaction
	transactionId := BuildTransactionIdentifier(msg)
	operations := BuildTransactionOperations(msg, receipt, api)
	if transactionId != nil {
		transaction.TransactionIdentifier = transactionId
	}
	if operations != nil {
		transaction.Operations = operations
	}

	return &transaction
}

func BuildTransactionIdentifier(msg *api.Message) *types.TransactionIdentifier {
	if msg == nil {
		return nil
	}
	return &types.TransactionIdentifier{
		Hash: msg.Cid.String(),
	}
}

func BuildTransactionOperations(msg *api.Message, receipt *filTypes.MessageReceipt, api *api.FullNode)[]*types.Operation{
	if msg == nil || receipt == nil {
		return nil
	}
	var (
		operations []*types.Operation
		operation  types.Operation
	)

	operationId := BuildOperationIdentifier(msg)
	//relatedOperations := BuildRelatedOperations() //TODO
	accountId := BuildAccountIdentifier(msg)
	amount := BuildAmount(msg)
	methodStr, err := GetMethodName(msg, api)
	if err != nil {
		operation.Type = methodStr
	}

	if operationId != nil {
		operation.OperationIdentifier = operationId
	}
	if accountId != nil {
		operation.Account = accountId
	}
	if amount != nil {
		operation.Amount = amount
	}
	operation.Status = receipt.ExitCode.String()

	operations = append(operations, &operation)
	return operations
}

func BuildOperationIdentifier(msg *api.Message) *types.OperationIdentifier {
	if msg == nil {
		return nil
	}

	return &types.OperationIdentifier{
		Index: int64(msg.Message.Nonce),
	}
}

func BuildAccountIdentifier(msg *api.Message) *types.AccountIdentifier {
	if msg == nil {
		return nil
	}

	return &types.AccountIdentifier{
		Address: msg.Message.From.String(),
	}
}

func BuildAmount(msg *api.Message) *types.Amount {
	if msg == nil {
		return nil
	}

	return &types.Amount{
		Currency: GetCurrencyData(),
		Value: msg.Message.Value.String(),
	}
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	return nil, ErrNotImplemented
}
