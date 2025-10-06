package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log"
)

var v2Logger = logging.Logger("v2-helpers")

// FinalityTag represents the different finality levels for V2 API
type FinalityTag string

const (
	FinalityLatest    FinalityTag = FinalityTagLatest
	FinalitySafe      FinalityTag = FinalityTagSafe
	FinalityFinalized FinalityTag = FinalityTagFinalized
)

// IsV2EnabledForService checks if V2 APIs should be used
func IsV2EnabledForService() bool {
	enabled, err := strconv.ParseBool(EnableLotusV2APIs)
	if err != nil {
		return false // Default to V1 on parse error
	}
	return enabled
}

// IsFinalityAnchorEnabled checks if finality anchor mode is enabled
func IsFinalityAnchorEnabled() bool {
	enabled, err := strconv.ParseBool(EnableFinalityAnchor)
	if err != nil {
		return false // Default to false on parse error
	}
	return enabled
}

// shouldUseV2API determines if V2 API should be used based on configuration and availability
func shouldUseV2API(v2Node v2api.FullNode, finalityTag FinalityTag) bool {
	return IsV2EnabledForService() && v2Node != nil && finalityTag != ""
}

// GetFinalityTagFromMetadata extracts finality tag from Rosetta request metadata
func GetFinalityTagFromMetadata(metadata map[string]interface{}) (FinalityTag, error) {
	if metadata == nil {
		// No metadata means no finality tag specified - return empty to use V1
		return "", nil
	}

	if tagValue, exists := metadata[MetadataFinalityTag]; exists {
		if tagStr, ok := tagValue.(string); ok {
			switch tagStr {
			case FinalityTagSafe:
				return FinalitySafe, nil
			case FinalityTagFinalized:
				return FinalityFinalized, nil
			case FinalityTagLatest:
				return FinalityLatest, nil
			case "":
				return "", fmt.Errorf("empty finality tag not allowed")
			default:
				return "", fmt.Errorf("unknown finality tag: %s", tagStr)
			}
		}
	}

	// This will force fallback to v1 since finality tag is not specified
	return "", nil
}

// GetFinalityTagFromNetworkIdentifier extracts finality tag from NetworkIdentifier's sub_network_identifier metadata
func GetFinalityTagFromNetworkIdentifier(networkIdentifier *types.NetworkIdentifier) (FinalityTag, error) {
	if networkIdentifier == nil {
		return "", nil
	}

	// Check if sub_network_identifier exists
	if networkIdentifier.SubNetworkIdentifier == nil {
		return "", nil
	}

	// If sub_network_identifier exists, metadata is required
	if networkIdentifier.SubNetworkIdentifier.Metadata == nil {
		return "", fmt.Errorf("sub_network_identifier requires metadata with finality_tag")
	}

	// Extract finality_tag from sub_network_identifier metadata
	return GetFinalityTagFromMetadata(networkIdentifier.SubNetworkIdentifier.Metadata)
}

// CreateTagSelector creates a V2 TipSetSelector for the given finality tag
func CreateTagSelector(tag FinalityTag) filTypes.TipSetSelector {
	var tipsetTag filTypes.TipSetTag

	switch tag {
	case FinalityLatest:
		tipsetTag = "latest"
	case FinalitySafe:
		tipsetTag = "safe"
	case FinalityFinalized:
		tipsetTag = "finalized"
	default:
		tipsetTag = "latest"
	}

	return filTypes.TipSetSelector{
		Tag: &tipsetTag,
	}
}

// ChainGetTipSetWithFallback is a wrapper that uses V2 ChainGetTipSet if enabled,
// otherwise falls back to V1 ChainHead
func ChainGetTipSetWithFallback(ctx context.Context, v1Node api.FullNode, v2Node v2api.FullNode, tag FinalityTag) (*filTypes.TipSet, error) {
	if shouldUseV2API(v2Node, tag) {
		selector := CreateTagSelector(tag)
		tipSet, err := v2Node.ChainGetTipSet(ctx, selector)
		if err != nil {
			v2Logger.Errorf("failed to get tipset with v2: %v", err)
			return nil, fmt.Errorf("v2 ChainGetTipSet failed: %w", err)
		}
		return tipSet, nil
	}

	// If finality tag is specified but V2 is not enabled/available, return error
	if tag != "" {
		return nil, fmt.Errorf("finality_tag '%s' requires V2 APIs to be enabled", tag)
	}

	// Use V1 API when no finality tag is specified
	return v1Node.ChainHead(ctx)
}

// TipSetResolution contains the result of tipset resolution
type TipSetResolution struct {
	TipSet       *filTypes.TipSet
	Height       int64
	IsNullTipSet bool
}

// ResolveTipSetForFinality resolves which tipset to use based on requested height, finality tag, and anchor mode.
// This is the central function that implements both Height Comparison Mode and Finality Anchor Mode.
//
// Parameters:
//   - ctx: context
//   - v1Node: Lotus V1 API node
//   - v2Node: Lotus V2 API node
//   - requestedHeight: the requested height (-1 means not specified, 0 means current finality height)
//   - finalityTag: the finality tag (empty string means no finality requested)
//
// Returns:
//   - TipSetResolution with the resolved tipset, actual height, and null tipset flag
//   - error if resolution fails
//
// Behavior:
//
// When requestedHeight == -1 (no block_identifier):
//   - Returns finality tipset if finalityTag is set
//   - Returns chain head if no finalityTag
//
// When requestedHeight == 0 with finalityTag:
//   - Returns current finality height (special case per design)
//
// When requestedHeight > 0 with finalityTag:
//   - Height Comparison Mode (EnableFinalityAnchor=false): Returns max(requestedHeight, finality_height)
//   - Anchor Mode (EnableFinalityAnchor=true): Returns tipset at requestedHeight from finality chain
//   - Error if requestedHeight > finality_height (height not yet on finality chain)
//   - Sets IsNullTipSet=true if returned height != requested (null tipset on chain)
//
// When requestedHeight >= 0 with no finalityTag:
//   - Returns tipset at requestedHeight using V1 API
func ResolveTipSetForFinality(
	ctx context.Context,
	v1Node api.FullNode,
	v2Node v2api.FullNode,
	requestedHeight int64,
	finalityTag FinalityTag,
) (*TipSetResolution, error) {
	// Case 1: No block_identifier specified (requestedHeight == -1)
	// TODO: What happens when no finality tag is specified?
	if requestedHeight == -1 {
		tipSet, err := ChainGetTipSetWithFallback(ctx, v1Node, v2Node, finalityTag)
		if err != nil {
			return nil, err
		}
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: false,
		}, nil
	}

	// Case 2: No finality tag - return requested block using V1
	if finalityTag == "" {
		tipSet, err := v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(requestedHeight), filTypes.EmptyTSK)
		if err != nil {
			return nil, err
		}
		// Check if this is a null tipset (returned height doesn't match requested)
		isNull := int64(tipSet.Height()) != requestedHeight
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: isNull,
		}, nil
	}

	// Case 3: Special case - height=0 with finality_tag returns current finality height
	if requestedHeight == 0 {
		tipSet, err := ChainGetTipSetWithFallback(ctx, v1Node, v2Node, finalityTag)
		if err != nil {
			return nil, err
		}
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: false,
		}, nil
	}

	// Case 4: Both requestedHeight > 0 and finalityTag are set
	// Get the finality-based tipset
	finalityTipSet, err := ChainGetTipSetWithFallback(ctx, v1Node, v2Node, finalityTag)
	if err != nil {
		return nil, err
	}
	finalityHeight := int64(finalityTipSet.Height())

	if IsFinalityAnchorEnabled() {
		// Finality Anchor Mode: Query finality chain at exact height using V2 API
		if requestedHeight > finalityHeight {
			return nil, fmt.Errorf("height %d not yet on finality chain (current finality: %d)", requestedHeight, finalityHeight)
		}

		// Use V2 API with TipSetSelector that anchors to the finality chain
		// Build selector with height at requested epoch, anchored to finality tag
		epoch := abi.ChainEpoch(requestedHeight)
		tipsetTag := filTypes.TipSetTag(finalityTag)
		selector := filTypes.TipSetSelector{
			Height: &filTypes.TipSetHeight{
				At:       &epoch,
				Previous: false,
				Anchor: &filTypes.TipSetAnchor{
					Tag: &tipsetTag,
				},
			},
		}

		tipSet, err := v2Node.ChainGetTipSet(ctx, selector)
		if err != nil {
			return nil, fmt.Errorf("failed to get tipset at height %d from finality chain: %w", requestedHeight, err)
		}

		// Check if this is a null tipset on the finality chain
		isNull := int64(tipSet.Height()) != requestedHeight

		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: isNull,
		}, nil
	}

	// Height Comparison Mode: Return max(requestedHeight, finalityHeight)
	if requestedHeight >= finalityHeight {
		// Requested height is at or beyond finality - return requested block
		tipSet, err := v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(requestedHeight), filTypes.EmptyTSK)
		if err != nil {
			return nil, err
		}
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: false,
		}, nil
	}

	// Requested height is before finality - return finality block
	return &TipSetResolution{
		TipSet:       finalityTipSet,
		Height:       finalityHeight,
		IsNullTipSet: false,
	}, nil
}
