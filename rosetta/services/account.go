package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	filBuiltin "github.com/filecoin-project/lotus/chain/actors/builtin"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"strconv"
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

	addr, filErr := address.NewFromString(request.AccountIdentifier.Address)
	if filErr != nil {
		return nil, BuildError(ErrInvalidAccountAddress, nil, true)
	}

	// Check sync status
	status, syncErr := CheckSyncStatus(ctx, &a.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if !status.IsSynced() {
		return nil, BuildError(ErrNodeNotSynced, nil, true)
	}

	var fixedQueryTipSet *filTypes.TipSet
	var originalQueryTipSet *filTypes.TipSet

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index == nil {
			return nil, BuildError(ErrInsufficientQueryInputs, nil, true)
		}

		originalQueryHeight := *request.BlockIdentifier.Index
		// From lotus v1.5 and on, StateGetActor computes the state at parent's tipset.
		// To get the state on the requested height, we need to query the block at (height + 1).
		fixedQueryHeight := (*request.BlockIdentifier.Index) + 1

		fixedQueryTipSet, filErr = a.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(fixedQueryHeight), filTypes.EmptyTSK)
		if filErr != nil {
			return nil, BuildError(ErrUnableToGetBlk, filErr, true)
		}
		originalQueryTipSet, filErr = a.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(originalQueryHeight), filTypes.EmptyTSK)
		if filErr != nil {
			return nil, BuildError(ErrUnableToGetBlk, filErr, true)
		}
	} else {
		originalQueryTipSet, filErr = a.node.ChainHead(ctx)
		fixedQueryTipSet = originalQueryTipSet
		if filErr != nil {
			return nil, BuildError(ErrUnableToGetLatestBlk, filErr, true)
		}
	}

	var balanceStr = "0"
	queryTipSetHeight := int64(originalQueryTipSet.Height()) // To return in response
	queryTipSetHash, err := BuildTipSetKeyHash(originalQueryTipSet.Key()) // To return in response
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, err, true)
	}

	actor, err := a.node.StateGetActor(ctx, addr, fixedQueryTipSet.Key())
	if err != nil {
		// If actor is not found on chain, return 0 balance
		return &types.AccountBalanceResponse{
			BlockIdentifier: &types.BlockIdentifier{
				Index: queryTipSetHeight,
				Hash:  *queryTipSetHash,
			},
			Balances: []*types.Amount{
				{
					Value:    "0",
					Currency: GetCurrencyData(),
				},
			},
		}, nil
	}

	md := make(map[string]interface{})

	if request.AccountIdentifier.SubAccount != nil {
		// First, check if account is a multisig
		if !filBuiltin.IsMultisigActor(actor.Code) {
			return nil, BuildError(ErrAddNotMSig, nil, true)
		}

		switch request.AccountIdentifier.SubAccount.Address {
		case LockedBalanceStr:
			lockedBalance := actor.Balance
			spendableBalance, err := a.node.MsigGetAvailableBalance(ctx, addr, fixedQueryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetBalance, err, true)
			}
			lockedBalance.Sub(lockedBalance.Int, spendableBalance.Int)
			balanceStr = lockedBalance.String()
		case SpendableBalanceStr:
			spendableBalance, err := a.node.MsigGetAvailableBalance(ctx, addr, fixedQueryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetBalance, err, true)
			}
			balanceStr = spendableBalance.String()
		case VestingScheduleStr:
			vestingSch, err := a.node.MsigGetVestingSchedule(ctx, addr, fixedQueryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetVesting, err, true)
			}
			vestingMap := map[string]string{}
			vestingMap[VestingStartEpochKey] = vestingSch.StartEpoch.String()
			vestingMap[VestingUnlockDurationKey] = vestingSch.UnlockDuration.String()
			vestingMap[VestingInitialBalanceKey] = vestingSch.InitialBalance.String()
			md[VestingScheduleStr] = vestingMap
		default:
			return nil, BuildError(ErrMustSpecifySubAccount, nil, true)
		}
	} else {
		// Get available balance (spendable + locked if multisig)
		balanceStr = actor.Balance.String()
	}

	// Fill nonce
	md[NonceKey] = strconv.FormatUint(actor.Nonce, 10)

	resp := &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: queryTipSetHeight,
			Hash:  *queryTipSetHash,
		},
		Balances: []*types.Amount{
			{
				Value:    balanceStr,
				Currency: GetCurrencyData(),
			},
		},
		Metadata: md,
	}

	return resp, nil
}

func (a AccountAPIService) AccountCoins(ctx context.Context, request *types.AccountCoinsRequest) (*types.AccountCoinsResponse, *types.Error) {
	return nil, ErrNotImplemented
}
