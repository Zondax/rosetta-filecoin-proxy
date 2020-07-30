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

	//Get latest TipSet
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

	//Get latest TipSet
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

		txId := &types.TransactionIdentifier{
			Hash: msg.Cid().String(),
		}

		var operations []*types.Operation
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index:        int64(msg.Message.Nonce),
				NetworkIndex: nil,
			},
			RelatedOperations: nil,
			Type:              "", //TODO https://github.com/Zondax/rosetta-filecoin/issues/11
			Status:            "", //TODO https://github.com/Zondax/rosetta-filecoin/issues/11
			Account: &types.AccountIdentifier{
				Address: msg.Message.From.String(),
			},
			Amount: &types.Amount{
				Value: msg.Message.ValueReceived().String(),
				Currency: &types.Currency{
					Symbol:   "FIL", //TODO https://github.com/Zondax/rosetta-filecoin/issues/6
					Decimals: 18,    //TODO https://github.com/Zondax/rosetta-filecoin/issues/6
				},
				Metadata: nil,
			},
			Metadata: nil,
		}
		operations = append(operations, op)

		transaction = &types.Transaction{
			TransactionIdentifier: txId,
			Operations:            operations,
			Metadata:              nil,
		}
	}

	if !found {
		return nil, ErrUnableToGetTxns
	}

	resp := &types.MempoolTransactionResponse{
		Transaction: transaction,
		Metadata:    nil,
	}

	return resp, nil
}
