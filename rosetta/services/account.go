package services

import (
	"context"
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
)

const (
	LockedBalanceStr   = "LockedBalance"
	VestingScheduleStr = "VestingSchedule"

	LockedFundsKey           = "LockedFunds"
	VestingStartEpochKey     = "StartEpoch"
	VestingUnlockDurationKey = "UnlockDuration"
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
		return nil, ErrInvalidAccountAddress
	}

	//Check sync status
	status, syncErr := CheckSyncStatus(ctx, &a.node)
	if syncErr != nil {
		return nil, syncErr
	}
	if !status.IsSynced() {
		return nil, ErrNodeNotSynced
	}

	var queryTipSet *filTypes.TipSet

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index == nil {
			return nil, ErrInsufficientQueryInputs
		}

		queryTipSet, filErr = a.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(*request.BlockIdentifier.Index), filTypes.EmptyTSK)
		if filErr != nil {
			return nil, ErrUnableToGetBlk
		}
		if request.BlockIdentifier.Hash != nil {
			tipSetKeyHash, encErr := BuildTipSetKeyHash(queryTipSet.Key())
			if encErr != nil {
				return nil, ErrUnableToBuildTipSetHash
			}
			if *tipSetKeyHash != *request.BlockIdentifier.Hash {
				return nil, ErrInvalidHash
			}
		}
	} else {
		queryTipSet, filErr = a.node.ChainHead(ctx)
		if filErr != nil {
			return nil, ErrUnableToGetLatestBlk
		}
	}

	actor, err := a.node.StateGetActor(context.Background(), addr, queryTipSet.Key())
	if err != nil {
		return nil, ErrUnableToGetActor
	}

	md := make(map[string]interface{})
	var balanceStr string
	isMultiSig := actor.Code == builtin.MultisigActorCodeID

	if request.AccountIdentifier.SubAccount != nil {
		// First, check if account is multisig
		if !isMultiSig {
			return nil, ErrAddNotMSig
		}

		actorState, err := a.node.StateReadState(ctx, addr, queryTipSet.Key())
		if err != nil || actorState == nil {
			return nil, ErrUnableToGetActorState
		}

		tmpMap, ok := actorState.State.(map[string]interface{})
		if !ok {
			return nil, ErrMalformedValue
		}
		stateMultisig, err := getMultisigState(tmpMap)
		if err != nil {
			return nil, ErrMalformedValue
		}

		switch request.AccountIdentifier.SubAccount.Address {
		case LockedBalanceStr:
			lockedFunds := stateMultisig.AmountLocked(queryTipSet.Height())
			balanceStr = lockedFunds.String()
		case VestingScheduleStr:
			stEpoch := stateMultisig.StartEpoch.String()
			unlockDuration := stateMultisig.UnlockDuration.String()
			vestingMap := map[string]string{}
			vestingMap[VestingStartEpochKey] = stEpoch
			vestingMap[VestingUnlockDurationKey] = unlockDuration
			md[VestingScheduleStr] = vestingMap
		default:
			return nil, ErrMustSpecifySubAccount
		}
	} else {
		//Get available balance
		if isMultiSig {
			balance, err := a.node.MsigGetAvailableBalance(ctx, addr, queryTipSet.Key())
			if err != nil {
				return nil, ErrUnableToGetBalance
			}
			balanceStr = balance.String()
		} else {
			balanceStr = actor.Balance.String()
		}
	}

	queryTipSetHeight := int64(queryTipSet.Height())
	queryTipSetHash, err := BuildTipSetKeyHash(queryTipSet.Key())
	if err != nil {
		return nil, ErrUnableToBuildTipSetHash
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