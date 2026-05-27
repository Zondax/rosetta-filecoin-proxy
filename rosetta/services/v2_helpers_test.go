package services

import (
	"context"
	"errors"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// dummyV2Node is used only for non-nil pointer checks in shouldUseV2API
// It embeds the v2api.FullNode interface but is never actually called in tests
type dummyV2Node struct {
	v2api.FullNode
}

// mockV2FullNode embeds the full v2api.FullNode interface and adds mock capabilities
// We only mock the methods we actually use in tests
type mockV2FullNode struct {
	mock.Mock
	v2api.FullNode // Embed to satisfy interface
}

func (m *mockV2FullNode) ChainGetTipSet(ctx context.Context, selector filTypes.TipSetSelector) (*filTypes.TipSet, error) {
	args := m.Called(ctx, selector)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filTypes.TipSet), args.Error(1)
}

// mockV1FullNode embeds the api.FullNode interface and adds mock capabilities
type mockV1FullNode struct {
	mock.Mock
	api.FullNode // Embed to satisfy interface
}

func (m *mockV1FullNode) ChainGetTipSetByHeight(ctx context.Context, height abi.ChainEpoch, tsk filTypes.TipSetKey) (*filTypes.TipSet, error) {
	args := m.Called(ctx, height, tsk)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filTypes.TipSet), args.Error(1)
}

func TestShouldUseV2API(t *testing.T) {
	tests := []struct {
		name           string
		enableV2APIs   bool
		hasV2Node      bool
		finalityTag    FinalityTag
		expectedResult bool
	}{
		{
			name:           "V2 APIs enabled with valid node and tag",
			enableV2APIs:   true,
			hasV2Node:      true,
			finalityTag:    FinalitySafe,
			expectedResult: true,
		},
		{
			name:           "V2 APIs disabled",
			enableV2APIs:   false,
			hasV2Node:      true,
			finalityTag:    FinalitySafe,
			expectedResult: false,
		},
		{
			name:           "V2 node is nil",
			enableV2APIs:   true,
			hasV2Node:      false,
			finalityTag:    FinalitySafe,
			expectedResult: false,
		},
		{
			name:           "Empty finality tag",
			enableV2APIs:   true,
			hasV2Node:      true,
			finalityTag:    "",
			expectedResult: false,
		},
		{
			name:           "All conditions false",
			enableV2APIs:   false,
			hasV2Node:      false,
			finalityTag:    "",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalValue := EnableLotusV2APIs
			EnableLotusV2APIs = tt.enableV2APIs
			defer func() { EnableLotusV2APIs = originalValue }()

			var v2Node v2api.FullNode
			if tt.hasV2Node {
				v2Node = &dummyV2Node{}
			}

			result := shouldUseV2API(v2Node, tt.finalityTag)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetFinalityTagFromMetadata(t *testing.T) {
	tests := []struct {
		name        string
		metadata    map[string]interface{}
		expected    FinalityTag
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil metadata",
			metadata:    nil,
			expected:    "",
			expectError: false,
		},
		{
			name:        "Empty metadata",
			metadata:    map[string]interface{}{},
			expected:    "",
			expectError: false,
		},
		{
			name: "Valid safe tag",
			metadata: map[string]interface{}{
				MetadataFinalityTag: "safe",
			},
			expected:    FinalitySafe,
			expectError: false,
		},
		{
			name: "Valid finalized tag",
			metadata: map[string]interface{}{
				MetadataFinalityTag: "finalized",
			},
			expected:    FinalityFinalized,
			expectError: false,
		},
		{
			name: "Valid latest tag",
			metadata: map[string]interface{}{
				MetadataFinalityTag: "latest",
			},
			expected:    FinalityLatest,
			expectError: false,
		},
		{
			name: "Empty string tag",
			metadata: map[string]interface{}{
				MetadataFinalityTag: "",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "empty finality tag not allowed",
		},
		{
			name: "Unknown finality tag",
			metadata: map[string]interface{}{
				MetadataFinalityTag: "invalid",
			},
			expected:    "",
			expectError: true,
			errorMsg:    "unknown finality tag: invalid",
		},
		{
			name: "Non-string tag value",
			metadata: map[string]interface{}{
				MetadataFinalityTag: 123,
			},
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFinalityTagFromMetadata(tt.metadata)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFinalityTagFromNetworkIdentifier(t *testing.T) {
	tests := []struct {
		name              string
		networkIdentifier *types.NetworkIdentifier
		expected          FinalityTag
		expectError       bool
		errorMsg          string
	}{
		{
			name:              "Nil network identifier",
			networkIdentifier: nil,
			expected:          "",
			expectError:       false,
		},
		{
			name: "Nil sub network identifier",
			networkIdentifier: &types.NetworkIdentifier{
				Blockchain: "Filecoin",
				Network:    "mainnet",
			},
			expected:    "",
			expectError: false,
		},
		{
			name: "Sub network identifier without metadata",
			networkIdentifier: &types.NetworkIdentifier{
				Blockchain: "Filecoin",
				Network:    "mainnet",
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
				},
			},
			expected:    "",
			expectError: true,
			errorMsg:    "sub_network_identifier requires metadata with finality_tag",
		},
		{
			name: "Valid safe tag in sub network",
			networkIdentifier: &types.NetworkIdentifier{
				Blockchain: "Filecoin",
				Network:    "mainnet",
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
					Metadata: map[string]interface{}{
						MetadataFinalityTag: "safe",
					},
				},
			},
			expected:    FinalitySafe,
			expectError: false,
		},
		{
			name: "Valid finalized tag in sub network",
			networkIdentifier: &types.NetworkIdentifier{
				Blockchain: "Filecoin",
				Network:    "mainnet",
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
					Metadata: map[string]interface{}{
						MetadataFinalityTag: "finalized",
					},
				},
			},
			expected:    FinalityFinalized,
			expectError: false,
		},
		{
			name: "Invalid tag in sub network metadata",
			networkIdentifier: &types.NetworkIdentifier{
				Blockchain: "Filecoin",
				Network:    "mainnet",
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
					Metadata: map[string]interface{}{
						MetadataFinalityTag: "unknown",
					},
				},
			},
			expected:    "",
			expectError: true,
			errorMsg:    "unknown finality tag: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFinalityTagFromNetworkIdentifier(tt.networkIdentifier)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveTipSetForFinality(t *testing.T) {
	ctx := context.Background()

	t.Run("Null tipset detection when heights don't match", func(t *testing.T) {
		// When requesting height 55 but getting back 54 (null tipset)
		originalV2 := EnableLotusV2APIs
		EnableLotusV2APIs = false
		defer func() { EnableLotusV2APIs = originalV2 }()

		v1Mock := &mockV1FullNode{}
		returnedTS := createMockTipSet(54) // Returns height 54 when 55 was requested
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(55), filTypes.EmptyTSK).Return(returnedTS, nil)

		result, err := ResolveTipSetForFinality(ctx, v1Mock, nil, 55, "")

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(54), result.Height)
		assert.True(t, result.IsNullTipSet, "Requested 55 but got 54, should be null")
		v1Mock.AssertExpectations(t)
	})

	t.Run("Anchor mode queries exact height from finality chain", func(t *testing.T) {
		originalV2 := EnableLotusV2APIs
		originalAnchor := EnableFinalityAnchor
		EnableLotusV2APIs = true
		EnableFinalityAnchor = true
		defer func() {
			EnableLotusV2APIs = originalV2
			EnableFinalityAnchor = originalAnchor
		}()

		v2Mock := &mockV2FullNode{}
		v1Mock := &mockV1FullNode{}

		expectedTS := createMockTipSet(50)
		// Verify the selector has the correct anchor structure
		v2Mock.On("ChainGetTipSet", ctx, mock.MatchedBy(func(selector filTypes.TipSetSelector) bool {
			if selector.Height == nil {
				return false
			}
			if selector.Height.Anchor == nil {
				return false
			}
			if selector.Height.At == nil {
				return false
			}
			return *selector.Height.At == abi.ChainEpoch(50)
		})).Return(expectedTS, nil)

		result, err := ResolveTipSetForFinality(ctx, v1Mock, v2Mock, 50, FinalitySafe)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(50), result.Height)
		assert.False(t, result.IsNullTipSet)
		v2Mock.AssertExpectations(t)
	})

	t.Run("Anchor mode detects null tipset", func(t *testing.T) {
		originalV2 := EnableLotusV2APIs
		originalAnchor := EnableFinalityAnchor
		EnableLotusV2APIs = true
		EnableFinalityAnchor = true
		defer func() {
			EnableLotusV2APIs = originalV2
			EnableFinalityAnchor = originalAnchor
		}()

		v2Mock := &mockV2FullNode{}
		v1Mock := &mockV1FullNode{}

		// Request height 55 but get back 54 (null tipset in finality chain)
		returnedTS := createMockTipSet(54)
		v2Mock.On("ChainGetTipSet", ctx, mock.Anything).Return(returnedTS, nil)

		result, err := ResolveTipSetForFinality(ctx, v1Mock, v2Mock, 55, FinalitySafe)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(54), result.Height)
		assert.True(t, result.IsNullTipSet, "Requested 55 but got 54 in anchor mode")
		v2Mock.AssertExpectations(t)
	})

	t.Run("Height comparison mode - requested >= finality", func(t *testing.T) {
		// When requested height (100) >= finality height (80), use requested height
		originalV2 := EnableLotusV2APIs
		originalAnchor := EnableFinalityAnchor
		EnableLotusV2APIs = true
		EnableFinalityAnchor = false
		defer func() {
			EnableLotusV2APIs = originalV2
			EnableFinalityAnchor = originalAnchor
		}()

		v2Mock := &mockV2FullNode{}
		v1Mock := &mockV1FullNode{}

		finalityTS := createMockTipSet(80)
		requestedTS := createMockTipSet(100)

		v2Mock.On("ChainGetTipSet", ctx, mock.Anything).Return(finalityTS, nil)
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(100), filTypes.EmptyTSK).Return(requestedTS, nil)

		result, err := ResolveTipSetForFinality(ctx, v1Mock, v2Mock, 100, FinalitySafe)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(100), result.Height, "Should use requested height (100)")
		assert.Equal(t, requestedTS, result.TipSet)
		assert.False(t, result.IsNullTipSet)
		v2Mock.AssertExpectations(t)
		v1Mock.AssertExpectations(t)
	})

	t.Run("Height comparison mode - requested < finality", func(t *testing.T) {
		// When requested height (50) < finality height (80), use finality height
		originalV2 := EnableLotusV2APIs
		originalAnchor := EnableFinalityAnchor
		EnableLotusV2APIs = true
		EnableFinalityAnchor = false
		defer func() {
			EnableLotusV2APIs = originalV2
			EnableFinalityAnchor = originalAnchor
		}()

		v2Mock := &mockV2FullNode{}
		v1Mock := &mockV1FullNode{}

		finalityTS := createMockTipSet(80)
		v2Mock.On("ChainGetTipSet", ctx, mock.Anything).Return(finalityTS, nil)

		result, err := ResolveTipSetForFinality(ctx, v1Mock, v2Mock, 50, FinalitySafe)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(80), result.Height, "Should use finality height (80) instead of requested (50)")
		assert.Equal(t, finalityTS, result.TipSet)
		assert.False(t, result.IsNullTipSet)
		v2Mock.AssertExpectations(t)
	})
}

// Helper function to create mock tipset
func createMockTipSet(height int64) *filTypes.TipSet {
	mockCid, _ := cid.Parse("bafkqaaa")
	mockMiner, _ := address.NewFromString("t01000")

	ts, _ := filTypes.NewTipSet([]*filTypes.BlockHeader{
		{
			Miner:                 mockMiner,
			Height:                abi.ChainEpoch(height),
			ParentStateRoot:       mockCid,
			Messages:              mockCid,
			ParentMessageReceipts: mockCid,
			BlockSig:              &crypto.Signature{Type: crypto.SigTypeBLS},
			BLSAggregate:          &crypto.Signature{Type: crypto.SigTypeBLS},
			// See buildMockTargetTipSet in sync_status_test.go — Lotus v1.36
			// requires non-nil Ticket on every BlockHeader passed to NewTipSet.
			Ticket: &filTypes.Ticket{VRFProof: []byte{0}},
		},
	})
	return ts
}

// TestResolveSuccessorTipSet covers the helper that finds the next
// non-null tipset above `resolvedHeight`. The helper is mode-aware:
// in V2 anchor mode it dispatches the walk to v2Node.ChainGetTipSet
// (finality chain), and in all other modes it uses v1Node's
// ChainGetTipSetByHeight (head chain). The mode dispatch is the
// fix for the chain-mismatch class of bug — without it the walk
// would land on the v1 head chain even when resolution.TipSet came
// from the v2 finality chain, and StateGetActor on that successor
// would read state on the wrong fork after a finality / head reorg.
func TestResolveSuccessorTipSet(t *testing.T) {
	ctx := context.Background()

	t.Run("v1 happy path: successor at +1 exists", func(t *testing.T) {
		v1Mock := &mockV1FullNode{}
		tsAt101 := createMockTipSet(101)
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(101), filTypes.EmptyTSK).
			Return(tsAt101, nil)

		got, err := ResolveSuccessorTipSet(ctx, v1Mock, nil, 100, 200, "")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(101), int64(got.Height()))
	})

	t.Run("v1 null +1: walks to +2", func(t *testing.T) {
		v1Mock := &mockV1FullNode{}
		tsAt100 := createMockTipSet(100)
		tsAt102 := createMockTipSet(102)
		// Lotus's null-tipset fallback: requesting epoch 101 returns the
		// tipset at 100 (latest non-null ≤ 101).
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(101), filTypes.EmptyTSK).
			Return(tsAt100, nil)
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(102), filTypes.EmptyTSK).
			Return(tsAt102, nil)

		got, err := ResolveSuccessorTipSet(ctx, v1Mock, nil, 100, 200, "")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(102), int64(got.Height()),
			"walk must skip the null +1 and land on the next non-null tipset above resolvedHeight")
	})

	t.Run("walked past head returns nil-nil", func(t *testing.T) {
		v1Mock := &mockV1FullNode{}
		// No mocks registered — should never be called because target>head
		// short-circuits the loop before any RPC.
		got, err := ResolveSuccessorTipSet(ctx, v1Mock, nil, 200, 200, "")

		require.NoError(t, err)
		assert.Nil(t, got, "no successor when resolvedHeight is at head")
	})

	t.Run("v1 RPC error propagates", func(t *testing.T) {
		v1Mock := &mockV1FullNode{}
		v1Mock.On("ChainGetTipSetByHeight", ctx, abi.ChainEpoch(101), filTypes.EmptyTSK).
			Return((*filTypes.TipSet)(nil), errors.New("upstream lotus: rpc closed"))

		got, err := ResolveSuccessorTipSet(ctx, v1Mock, nil, 100, 200, "")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rpc closed")
		assert.Nil(t, got)
	})

	t.Run("V2 anchor mode dispatches to v2Node, NOT v1Node", func(t *testing.T) {
		// This test is the Bug 2 regression: under anchor mode, the
		// successor walk must use v2Node.ChainGetTipSet against the
		// finality chain, not v1Node.ChainGetTipSetByHeight against the
		// head chain. If the dispatch is wrong, the walk would land on
		// a potentially-different fork and StateGetActor would read
		// state on the wrong chain.
		originalV2 := EnableLotusV2APIs
		originalAnchor := EnableFinalityAnchor
		EnableLotusV2APIs = true
		EnableFinalityAnchor = true
		defer func() {
			EnableLotusV2APIs = originalV2
			EnableFinalityAnchor = originalAnchor
		}()

		v1Mock := &mockV1FullNode{}
		// CRITICAL: v1Mock has NO ChainGetTipSetByHeight expectation.
		// If the helper dispatches to v1 in anchor mode (the bug), the
		// mock will panic with "unexpected method call", failing the
		// test. The fix dispatches to v2 instead — v1 is untouched.

		v2Mock := &mockV2FullNode{}
		tsAt101 := createMockTipSet(101)
		v2Mock.On("ChainGetTipSet", ctx,
			mock.MatchedBy(func(s filTypes.TipSetSelector) bool {
				return s.Height != nil && s.Height.At != nil && int64(*s.Height.At) == 101 &&
					s.Height.Anchor != nil && s.Height.Anchor.Tag != nil &&
					string(*s.Height.Anchor.Tag) == "finalized"
			}),
		).Return(tsAt101, nil)

		got, err := ResolveSuccessorTipSet(ctx, v1Mock, v2Mock, 100, 200, FinalityFinalized)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(101), int64(got.Height()))
		v2Mock.AssertExpectations(t)
	})
}
