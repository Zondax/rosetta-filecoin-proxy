package services

// Tests in this file pin down how AccountBalance resolves the tipset it
// reads state from (the "+1 trick") across the canonical paths.
//
// The +1 trick: Lotus's StateGetActor(ctx, addr, tipset.Key()) returns
// the actor state at the PARENT of `tipset`. So to read state at the
// end of height N (after N's messages have been applied) we need a
// tipset whose parent is at height N — naturally, that's the tipset at
// height N+1. The interesting cases arise when N+1 is null, doesn't
// exist yet (head), or N itself is null.
//
// The existing TestAccountAPIService_AccountBalance test uses one
// catch-all mock for ChainGetTipSetByHeight that returns the same
// tipset for every epoch. That broad-brush approach can't see the +1
// trick going wrong because both the response lookup (at N) and the
// query lookup (at N+1) get the same value back. The helpers below let
// each test return different tipsets per epoch and different actors
// per tipset, so the assertion can verify which tipset's state was
// actually read.

import (
	"context"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

// matchEpoch returns a mock matcher that fires when the mocked
// ChainGetTipSetByHeight is called with the given epoch. Use it to
// register different return tipsets for different epochs in one mock.
func matchEpoch(want int64) interface{} {
	return mock.MatchedBy(func(e abi.ChainEpoch) bool { return int64(e) == want })
}

// matchTipSetKey returns a mock matcher for StateGetActor calls that
// fires only when the third argument's TipSetKey equals the given
// tipset's key. Use it to register different actor returns depending
// on which tipset's state is being read, so the assertion can verify
// which tipset the code actually consulted.
func matchTipSetKey(ts *filTypes.TipSet) interface{} {
	wantKey := ts.Key().String()
	return mock.MatchedBy(func(tsk filTypes.TipSetKey) bool { return tsk.String() == wantKey })
}

// commonAccountBalanceMocks registers the boilerplate that every
// AccountBalance test in this file needs: network name (for
// ValidateNetworkId), sync state (for CheckSyncStatus → returns
// synced), and chain head. Returns the head tipset so each caller can
// build per-test assertions around it.
func commonAccountBalanceMocks(nodeMock *mocks.FullNode, headHeight int64) *filTypes.TipSet {
	headTipSet := buildMockTargetTipSet(headHeight)
	nodeMock.On("StateNetworkName", mock.Anything).
		Return(dtypes.NetworkName(NetworkID.Network), nil)
	nodeMock.On("SyncState", mock.Anything).
		Return(&api.SyncState{
			ActiveSyncs: []api.ActiveSync{
				{Stage: api.StageSyncComplete, Target: &filTypes.TipSet{}},
			},
		}, nil)
	nodeMock.On("ChainHead", mock.Anything).Return(headTipSet, nil)
	return headTipSet
}

// TestAccountBalance_HistoricalHeight_HappyPath establishes the baseline:
// a request for a historical block whose +1 tipset exists and is
// non-null. AccountBalance should return the actor balance from the +1
// tipset's parent state (i.e., state at end of requestedHeight) and the
// response's block_identifier should honestly reflect the requested
// height with the tipset-at-requestedHeight's hash.
//
// This test PASSES on origin/master and after any future refactor of
// the tipset resolution logic.
func TestAccountBalance_HistoricalHeight_HappyPath(t *testing.T) {
	nodeMock := &mocks.FullNode{}

	var requestedHeight int64 = 100
	const headHeight int64 = 200

	tipsetAt100 := buildMockTargetTipSet(100)
	tipsetAt101 := buildMockTargetTipSet(101)
	tipsetAt100Hash, err := BuildTipSetKeyHash(tipsetAt100.Key())
	require.NoError(t, err)

	// The balance the test will assert on. This is the actor state at
	// the END of height 100, i.e., visible in tipsetAt101's parent state.
	actorEndOf100 := buildActorMock(cid.Cid{}, "200000000000")

	commonAccountBalanceMocks(nodeMock, headHeight)

	nodeMock.On("ChainGetTipSetByHeight", mock.Anything, matchEpoch(100), mock.Anything).
		Return(tipsetAt100, nil)
	nodeMock.On("ChainGetTipSetByHeight", mock.Anything, matchEpoch(101), mock.Anything).
		Return(tipsetAt101, nil)

	// Only the +1 tipset's key should reach StateGetActor in the happy
	// path. If the code mistakenly reads at tipsetAt100.Key() the mock
	// will return mock.AnythingOfType missing, surfacing as a panic
	// (or, worse, a different actor — the test would assert on it).
	nodeMock.On("StateGetActor", mock.Anything, mock.Anything, matchTipSetKey(tipsetAt101)).
		Return(actorEndOf100, nil)

	a := AccountAPIService{network: NetworkID, v1Node: nodeMock, v2Node: nil}

	got, gotErr := a.AccountBalance(context.Background(), &types.AccountBalanceRequest{
		NetworkIdentifier: NetworkID,
		BlockIdentifier:   &types.PartialBlockIdentifier{Index: &requestedHeight},
		AccountIdentifier: &types.AccountIdentifier{Address: "t0128015"},
	})

	require.Nil(t, gotErr, "AccountBalance should succeed for a healthy historical query")
	require.NotNil(t, got)
	require.Len(t, got.Balances, 1)

	assert.Equal(t, actorEndOf100.Balance.String(), got.Balances[0].Value,
		"happy path: balance should reflect state at end of requestedHeight (read via the +1 tipset)")
	assert.Equal(t, requestedHeight, got.BlockIdentifier.Index)
	assert.Equal(t, *tipsetAt100Hash, got.BlockIdentifier.Hash)
}

// Silence unused-import warning for abi when no other test in this file
// uses it yet; subsequent commits add tests that do.
var _ abi.ChainEpoch
