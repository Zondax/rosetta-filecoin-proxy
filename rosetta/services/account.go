package services

import (
	"context"
	"strconv"

	rosettaFilecoinLib "github.com/zondax/rosetta-filecoin-lib"
	"github.com/zondax/rosetta-filecoin-lib/actors"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
)

// AccountAPIService implements the server.BlockAPIServicer interface.
type AccountAPIService struct {
	network    *types.NetworkIdentifier
	v1Node     api.FullNode
	v2Node     v2api.FullNode
	rosettaLib *rosettaFilecoinLib.RosettaConstructionFilecoin
}

func NewAccountAPIService(network *types.NetworkIdentifier, v1API *api.FullNode, v2API *v2api.FullNode, r *rosettaFilecoinLib.RosettaConstructionFilecoin) server.AccountAPIServicer {
	return &AccountAPIService{
		network:    network,
		v1Node:     *v1API,
		v2Node:     *v2API,
		rosettaLib: r,
	}
}

// AccountBalance implements the /account/balance endpoint.
func (a AccountAPIService) AccountBalance(ctx context.Context,
	request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {

	errNet := ValidateNetworkId(ctx, &a.v1Node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	addr, filErr := address.NewFromString(request.AccountIdentifier.Address)
	if filErr != nil {
		return nil, BuildError(ErrInvalidAccountAddress, nil, true)
	}

	status, syncErr := CheckSyncStatus(ctx, &a.v1Node)
	if syncErr != nil {
		return nil, syncErr
	}

	if !status.IsSynced() {
		return nil, BuildError(ErrNodeNotSynced, nil, true)
	}

	var requestedHeight int64 = -1
	if request.BlockIdentifier != nil && request.BlockIdentifier.Index != nil {
		requestedHeight = *request.BlockIdentifier.Index
	}

	finalityTag, err := GetFinalityTagFromNetworkIdentifier(request.NetworkIdentifier)
	if err != nil {
		return nil, BuildError(ErrUnableToGetLatestBlk, err, true)
	}

	resolution, err := ResolveTipSetForFinality(ctx, a.v1Node, a.v2Node, requestedHeight, finalityTag)
	if err != nil {
		return nil, BuildError(ErrUnableToGetTipset, err, true)
	}

	tipSet := resolution.TipSet
	requestedHeight = resolution.Height

	var queryTipSet *filTypes.TipSet
	var responseTipSet *filTypes.TipSet

	// Now we need to get the appropriate query tipset for StateGetActor
	// StateGetActor computes the state at parent's tipSet, so we need to query at (height + 1)
	queryTipSet, filErr = a.v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(requestedHeight+1), filTypes.EmptyTSK)
	if filErr != nil {
		// If we can't get the +1 tipset, use the current tipset
		queryTipSet = tipSet
	}
	responseTipSet = tipSet

	var balanceStr = "0"
	queryTipSetHeight := int64(responseTipSet.Height())
	queryTipSetHash, filErr := BuildTipSetKeyHash(responseTipSet.Key())
	if filErr != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, filErr, true)
	}

	actor, err := a.v1Node.StateGetActor(ctx, addr, queryTipSet.Key())
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
		if !a.rosettaLib.BuiltinActors.IsActor(actor.Code, actors.ActorMultisigName) {
			return nil, BuildError(ErrAddNotMSig, nil, true)
		}

		switch request.AccountIdentifier.SubAccount.Address {
		case LockedBalanceStr:
			lockedBalance := actor.Balance
			spendableBalance, err := a.v1Node.MsigGetAvailableBalance(ctx, addr, queryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetBalance, err, true)
			}
			lockedBalance.Sub(lockedBalance.Int, spendableBalance.Int)
			balanceStr = lockedBalance.String()
		case SpendableBalanceStr:
			spendableBalance, err := a.v1Node.MsigGetAvailableBalance(ctx, addr, queryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetBalance, err, true)
			}
			balanceStr = spendableBalance.String()
		case VestingScheduleStr:
			vestingSch, err := a.v1Node.MsigGetVestingSchedule(ctx, addr, queryTipSet.Key())
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
