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
	resolvedHeight := resolution.Height

	var queryTipSet *filTypes.TipSet
	var responseTipSet *filTypes.TipSet

	// StateGetActor returns state at the PARENT of the tipset key it's
	// given, so to read state at the end of `resolvedHeight` we need a
	// tipset whose parent is at `resolvedHeight` — naturally a tipset
	// at `resolvedHeight + 1`. Two failure modes the pre-PR #310 code
	// handled but the post-PR #310 ad-hoc lookup did not:
	//
	//   1. `resolvedHeight + 1` is null. Lotus's ChainGetTipSetByHeight
	//      falls back to the latest non-null tipset whose height is
	//      <= the requested epoch, which can be tipset-at-`resolvedHeight`
	//      (or even lower). Using that fallback's Key() makes
	//      StateGetActor read state at the parent of `resolvedHeight` —
	//      i.e., state at the end of `resolvedHeight - 1`, silently
	//      dropping every message that landed at `resolvedHeight`.
	//      Walk forward until we find a tipset strictly above
	//      `resolvedHeight`.
	//
	//   2. `resolvedHeight` is already at chain head. No successor
	//      exists yet. We can't observe state at the end of head;
	//      best behavior is to report state at the end of head-1 AND
	//      shift the response's block_identifier down to head-1 so
	//      (balance, block_identifier) remain self-consistent.
	queryTipSet = nil
	headForLoop, headErr := a.v1Node.ChainHead(ctx)
	if headErr != nil {
		return nil, BuildError(ErrUnableToGetLatestBlk, headErr, true)
	}
	headHeight := int64(headForLoop.Height())
	const maxNullStretch = 50 // safety bound; consecutive null epochs are rare
	for offset := int64(1); offset <= maxNullStretch; offset++ {
		target := resolvedHeight + offset
		if target > headHeight {
			// Walked past head — by definition, no successor tipset
			// exists yet. Fall through to the at-head branch below.
			break
		}
		candidate, candidateErr := a.v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(target), filTypes.EmptyTSK)
		if candidateErr != nil {
			// Surface RPC errors instead of silently falling through to
			// the at-head branch — that branch's response is only
			// correct when there's genuinely no successor tipset (we're
			// at chain head). A transient ChainGetTipSetByHeight error
			// is a different failure mode; masking it as a
			// height-shifted "200 OK" hides real upstream issues from
			// the caller. Pre-PR #310 code returned ErrUnableToGetBlk
			// for this case; restore that behavior.
			return nil, BuildError(ErrUnableToGetTipset, candidateErr, true)
		}
		if int64(candidate.Height()) > resolvedHeight {
			queryTipSet = candidate
			break
		}
		// Otherwise Lotus bumped us backward — (resolvedHeight + offset)
		// is null; try the next offset.
	}
	if queryTipSet == nil {
		// No successor available (either at head or no non-null tipset
		// within the safety bound). Use the resolution tipset and
		// self-consistently label the response with its parent so
		// (balance, identifier) remain aligned: balance reflects state
		// at parent(resolvedHeight) = end of resolvedHeight - 1, and
		// the identifier reports the height whose state is actually
		// returned.
		parentTipSet, parentErr := a.v1Node.ChainGetTipSet(ctx, tipSet.Parents())
		if parentErr != nil {
			return nil, BuildError(ErrUnableToGetParentBlk, parentErr, true)
		}
		queryTipSet = tipSet
		responseTipSet = parentTipSet
	} else {
		responseTipSet = tipSet
	}

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
