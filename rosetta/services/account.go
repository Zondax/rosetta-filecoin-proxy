package services

import (
	"context"
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
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
		return nil, BuildError(ErrInvalidAccountAddress, nil)
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &a.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if !status.IsSynced() {
		return nil, BuildError(ErrNodeNotSynced, nil)
	}

	var queryTipSet *filTypes.TipSet

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index == nil {
			return nil, BuildError(ErrInsufficientQueryInputs, nil)
		}

		queryTipSet, filErr = a.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(*request.BlockIdentifier.Index), filTypes.EmptyTSK)
		if filErr != nil {
			return nil, BuildError(ErrUnableToGetBlk, filErr)
		}
		if request.BlockIdentifier.Hash != nil {
			tipSetKeyHash, encErr := BuildTipSetKeyHash(queryTipSet.Key())
			if encErr != nil {
				return nil, BuildError(ErrUnableToBuildTipSetHash, encErr)
			}
			if *tipSetKeyHash != *request.BlockIdentifier.Hash {
				return nil, BuildError(ErrInvalidHash, nil)
			}
		}
	} else {
		queryTipSet, filErr = a.node.ChainHead(ctx)
		if filErr != nil {
			return nil, BuildError(ErrUnableToGetLatestBlk, filErr)
		}
	}

	var balanceStr = "0"
	queryTipSetHeight := int64(queryTipSet.Height())
	queryTipSetHash, err := BuildTipSetKeyHash(queryTipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, err)
	}

	actor, err := a.node.StateGetActor(ctx, addr, queryTipSet.Key())
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
		// First, check if account is multisig
		if !actor.IsMultisigActor() {
			return nil, BuildError(ErrAddNotMSig, nil)
		}

		actorState, err := a.node.StateReadState(ctx, addr, queryTipSet.Key())
		if err != nil || actorState == nil {
			return nil, BuildError(ErrUnableToGetActorState, err)
		}

		tmpMap, ok := actorState.State.(map[string]interface{})
		if !ok {
			return nil, BuildError(ErrMalformedValue, nil)
		}
		stateMultisig, err := getMultisigState(tmpMap)
		if err != nil {
			return nil, BuildError(ErrMalformedValue, err)
		}

		switch request.AccountIdentifier.SubAccount.Address {
		case LockedBalanceStr:
			lockedFunds := stateMultisig.AmountLocked(queryTipSet.Height() - stateMultisig.StartEpoch)
			balanceStr = lockedFunds.String()
		case SpendableBalanceStr:
			available, err := a.node.MsigGetAvailableBalance(ctx, addr, queryTipSet.Key())
			if err != nil {
				return nil, BuildError(ErrUnableToGetBalance, err)
			}
			balanceStr = available.String()
		case VestingScheduleStr:
			stEpoch := stateMultisig.StartEpoch.String()
			unlockDuration := stateMultisig.UnlockDuration.String()
			vestingMap := map[string]string{}
			vestingMap[VestingStartEpochKey] = stEpoch
			vestingMap[VestingUnlockDurationKey] = unlockDuration
			md[VestingScheduleStr] = vestingMap
		default:
			return nil, BuildError(ErrMustSpecifySubAccount, nil)
		}
	} else {
		//Get available balance (spendable + locked)
		balanceStr = actor.Balance.String()
	}

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
	}

	if len(md) > 0 {
		resp.Metadata = md
	}

	return resp, nil
}

func getMultisigState(m map[string]interface{}) (multisig.State, error) {
	data, _ := json.Marshal(m)
	var result multisig.State
	err := json.Unmarshal(data, &result)
	return result, err
}
