package services

import (
	"context"
	"github.com/ipfs/go-cid"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
)

// BlockAPIService implements the server.BlockAPIServicer interface.
type MemPoolAPIService struct {
	network *types.NetworkIdentifier
	node    api.FullNode
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewMemPoolAPIService(network *types.NetworkIdentifier, api *api.FullNode) server.MempoolAPIServicer {
	return &MemPoolAPIService{
		network: network,
		node:    *api,
	}
}

// Mempool implements the /mempool endpoint.
func (m *MemPoolAPIService) Mempool(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {

	errNet := ValidateNetworkId(ctx, &m.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &m.node)
	if syncErr != nil {
		return nil, syncErr
	}

	if !status.IsSynced() {
		return nil, ErrUnableToGetUnsyncedBlock
	}

	//Get head TipSet
	headTipSet, err := m.node.ChainHead(ctx)
	if err != nil || headTipSet == nil {
		return nil, ErrUnableToGetLatestBlk
	}

	pendingMsg, err := m.node.MpoolPending(ctx, headTipSet.Key())
	if err != nil {
		return nil, ErrUnableToGetTxns
	}

	var transactions []*types.TransactionIdentifier
	for _, msg := range pendingMsg {
		transactions = append(transactions, &types.TransactionIdentifier{
			Hash: msg.Cid().String(),
		})
	}

	resp := &types.MempoolResponse{
		TransactionIdentifiers: transactions,
	}

	return resp, nil
}

// MempoolTransaction implements the /mempool/transaction endpoint.
func (m MemPoolAPIService) MempoolTransaction(
	ctx context.Context,
	request *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {

	errNet := ValidateNetworkId(ctx, &m.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &m.node)
	if syncErr != nil {
		return nil, syncErr
	}

	if !status.IsSynced() {
		return nil, ErrUnableToGetUnsyncedBlock
	}

	if request.TransactionIdentifier == nil {
		return nil, ErrMalformedValue
	}

	requestedCid, err := cid.Parse(request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, ErrMalformedValue
	}

	//Get head TipSet
	headTipSet, err := m.node.ChainHead(ctx)
	if err != nil || headTipSet == nil {
		return nil, ErrUnableToGetLatestBlk
	}

	pendingMsg, err := m.node.MpoolPending(ctx, headTipSet.Key())
	if err != nil {
		return nil, ErrUnableToGetTxns
	}

	var found = false
	var transaction *types.Transaction
	for _, msg := range pendingMsg {
		if msg.Cid() != requestedCid {
			continue
		}
		found = true
		transaction = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: msg.Cid().String(),
			},
			Operations: []*types.Operation{},
		}

		opType, err := GetMethodName(&msg.Message)
		if err != nil {
			return nil, err
		}

		opStatus := "Pending" //TODO get status from receipt?

		transaction.Operations = appendOp(transaction.Operations, opType,
			msg.Message.From.String(), msg.Message.Value.String(), opStatus)
		transaction.Operations = appendOp(transaction.Operations, opType,
			msg.Message.To.String(), msg.Message.Value.String(), opStatus)

		break
	}

	if !found {
		return nil, ErrUnableToGetTxns
	}

	resp := &types.MempoolTransactionResponse{
		Transaction: transaction,
	}

	return resp, nil
}
