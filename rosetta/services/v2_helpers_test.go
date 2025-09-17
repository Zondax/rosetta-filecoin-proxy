package services

import (
	"context"
	"os"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

func TestIsV2EnabledForService(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"true", true},
		{"false", false},
		{"invalid", false},
		{"", false},
	}

	original := os.Getenv("ENABLE_LOTUS_V2_APIS")
	defer os.Setenv("ENABLE_LOTUS_V2_APIS", original)

	for _, tt := range tests {
		os.Setenv("ENABLE_LOTUS_V2_APIS", tt.envValue)
		EnableLotusV2APIs = tt.envValue
		assert.Equal(t, tt.want, IsV2EnabledForService())
	}
}

func TestIsForceSafeF3FinalityEnabled(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"true", true},
		{"false", false},
		{"invalid", false},
		{"", false},
	}

	originalV2 := os.Getenv("ENABLE_LOTUS_V2_APIS")
	originalForce := os.Getenv("FORCE_SAFE_F3_FINALITY")
	defer func() {
		os.Setenv("ENABLE_LOTUS_V2_APIS", originalV2)
		os.Setenv("FORCE_SAFE_F3_FINALITY", originalForce)
	}()

	for _, tt := range tests {
		os.Setenv("FORCE_SAFE_F3_FINALITY", tt.envValue)
		ForceSafeF3Finality = tt.envValue
		assert.Equal(t, tt.want, IsForceSafeF3FinalityEnabled())
	}
}

func TestGetFinalityTagFromMetadata(t *testing.T) {
	// Test valid tags
	validTags := map[string]FinalityTag{
		"safe":      FinalitySafe,
		"finalized": FinalityFinalized,
		"latest":    FinalityLatest,
	}

	for tagStr, expected := range validTags {
		metadata := map[string]any{MetadataFinalityTag: tagStr}
		got, err := GetFinalityTagFromMetadata(metadata)
		require.NoError(t, err)
		assert.Equal(t, expected, got)
	}

	// Test error cases
	_, err := GetFinalityTagFromMetadata(map[string]any{MetadataFinalityTag: ""})
	assert.Error(t, err)

	_, err = GetFinalityTagFromMetadata(map[string]any{MetadataFinalityTag: "invalid"})
	assert.Error(t, err)

	// Test nil/empty/missing
	got, err := GetFinalityTagFromMetadata(nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetFinalityTagFromNetworkIdentifier(t *testing.T) {
	// Test valid case
	netID := &types.NetworkIdentifier{
		Blockchain: "Filecoin",
		Network:    "mainnet",
		SubNetworkIdentifier: &types.SubNetworkIdentifier{
			Network:  "f3",
			Metadata: map[string]any{MetadataFinalityTag: "safe"},
		},
	}
	got, err := GetFinalityTagFromNetworkIdentifier(netID)
	require.NoError(t, err)
	assert.Equal(t, FinalitySafe, got)

	// Test error case - missing metadata
	netIDNoMeta := &types.NetworkIdentifier{
		SubNetworkIdentifier: &types.SubNetworkIdentifier{Network: "f3"},
	}
	_, err = GetFinalityTagFromNetworkIdentifier(netIDNoMeta)
	assert.Error(t, err)

	// Test nil case
	got, err = GetFinalityTagFromNetworkIdentifier(nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCreateTagSelector(t *testing.T) {
	// Test valid tags
	tags := map[FinalityTag]string{
		FinalityLatest:    "latest",
		FinalitySafe:      "safe",
		FinalityFinalized: "finalized",
		"":                "latest", // defaults to latest
	}

	for tag, expected := range tags {
		selector := CreateTagSelector(tag)
		require.NotNil(t, selector.Tag)
		assert.Equal(t, filTypes.TipSetTag(expected), *selector.Tag)
	}
}

func TestChainGetTipSetWithFallback(t *testing.T) {
	ctx := context.Background()
	mockCid, _ := cid.Parse("bafkqaaa")
	mockMiner, _ := address.NewFromString("t00")
	mockTipSet, _ := filTypes.NewTipSet([]*filTypes.BlockHeader{{
		Miner: mockMiner, Height: abi.ChainEpoch(100), ParentStateRoot: mockCid,
		Messages: mockCid, ParentMessageReceipts: mockCid,
		BlockSig:     &crypto.Signature{Type: crypto.SigTypeBLS},
		BLSAggregate: &crypto.Signature{Type: crypto.SigTypeBLS},
	}})

	originalV2 := os.Getenv("ENABLE_LOTUS_V2_APIS")
	originalForce := os.Getenv("FORCE_SAFE_F3_FINALITY")
	defer func() {
		os.Setenv("ENABLE_LOTUS_V2_APIS", originalV2)
		os.Setenv("FORCE_SAFE_F3_FINALITY", originalForce)
	}()

	// Test V1 fallback when V2 disabled
	os.Setenv("ENABLE_LOTUS_V2_APIS", "false")
	os.Setenv("FORCE_SAFE_F3_FINALITY", "false")
	EnableLotusV2APIs = "false"
	ForceSafeF3Finality = "false"

	v1Mock := &mocks.FullNode{}
	v1Mock.On("ChainHead", ctx).Return(mockTipSet, nil)

	got, err := ChainGetTipSetWithFallback(ctx, v1Mock, nil, "")
	require.NoError(t, err)
	assert.Equal(t, mockTipSet, got)

	// Test error when tag specified but V2 disabled
	_, err = ChainGetTipSetWithFallback(ctx, v1Mock, nil, FinalitySafe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires V2 APIs")

	v1Mock.AssertExpectations(t)
}

func TestStateGetActorWithFallback(t *testing.T) {
	ctx := context.Background()
	testAddr, _ := address.NewFromString("t01234")
	mockCid, _ := cid.Parse("bafkqaaa")
	mockActor := &filTypes.Actor{
		Code: mockCid, Head: mockCid, Nonce: 10, Balance: filTypes.NewInt(1000),
	}

	originalV2 := os.Getenv("ENABLE_LOTUS_V2_APIS")
	originalForce := os.Getenv("FORCE_SAFE_F3_FINALITY")
	defer func() {
		os.Setenv("ENABLE_LOTUS_V2_APIS", originalV2)
		os.Setenv("FORCE_SAFE_F3_FINALITY", originalForce)
	}()

	// Test V1 fallback when V2 disabled
	os.Setenv("ENABLE_LOTUS_V2_APIS", "false")
	os.Setenv("FORCE_SAFE_F3_FINALITY", "false")
	EnableLotusV2APIs = "false"
	ForceSafeF3Finality = "false"

	v1Mock := &mocks.FullNode{}
	v1Mock.On("StateGetActor", ctx, testAddr, filTypes.EmptyTSK).Return(mockActor, nil)

	got, err := StateGetActorWithFallback(ctx, v1Mock, nil, testAddr, "")
	require.NoError(t, err)
	assert.Equal(t, mockActor, got)

	// Test error when tag specified but V2 disabled
	_, err = StateGetActorWithFallback(ctx, v1Mock, nil, testAddr, FinalitySafe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires V2 APIs")

	v1Mock.AssertExpectations(t)
}

func TestForceSafeF3FinalityBehavior(t *testing.T) {
	originalV2 := os.Getenv("ENABLE_LOTUS_V2_APIS")
	originalForce := os.Getenv("FORCE_SAFE_F3_FINALITY")
	defer func() {
		os.Setenv("ENABLE_LOTUS_V2_APIS", originalV2)
		os.Setenv("FORCE_SAFE_F3_FINALITY", originalForce)
	}()

	// Test that when ForceSafeF3Finality is enabled, the flag works correctly
	os.Setenv("ENABLE_LOTUS_V2_APIS", "true")
	os.Setenv("FORCE_SAFE_F3_FINALITY", "true")
	EnableLotusV2APIs = "true"
	ForceSafeF3Finality = "true"

	assert.True(t, IsV2EnabledForService())
	assert.True(t, IsForceSafeF3FinalityEnabled())

	// Test when ForceSafeF3Finality is disabled
	os.Setenv("FORCE_SAFE_F3_FINALITY", "false")
	ForceSafeF3Finality = "false"

	assert.True(t, IsV2EnabledForService())
	assert.False(t, IsForceSafeF3FinalityEnabled())
}
