package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
)

// AccountAPIService implements the server.BlockAPIServicer interface.
type AccountAPIService struct {
	network *types.NetworkIdentifier
	node    api.FullNode
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewAccountAPIService(network *types.NetworkIdentifier, node *api.FullNode) server.AccountAPIServicer {
	return &AccountAPIService{
		network: network,
		node:    *node,
	}
}

// AccountBalance implements the /account/balance endpoint.
func (a AccountAPIService) AccountBalance(ctx context.Context,
	request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {

	errNet := ValidateNetworkId(ctx, &a.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	addr, err := address.NewFromString(request.AccountIdentifier.Address)
	if err != nil {
		return nil, ErrInvalidAccountAddress
	}

	balance, err := a.node.WalletBalance(ctx, addr)
	if err != nil {
		return nil, ErrUnableToGetWalletBalance
	}

	resp := &types.AccountBalanceResponse{
		BlockIdentifier: nil,
		Balances: []*types.Amount{
			{
				Value:    balance.String(),
				Currency: GetCurrencyData(),
			},
		},
	}

	return resp, nil
}
