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

func IsV2EnabledForService() bool {
	enabled, err := strconv.ParseBool(EnableLotusV2APIs)
	if err != nil {
		return false
	}
	return enabled
}

func IsFinalityAnchorEnabled() bool {
	enabled, err := strconv.ParseBool(EnableFinalityAnchor)
	if err != nil {
		return false
	}
	return enabled
}

func shouldUseV2API(v2Node v2api.FullNode, finalityTag FinalityTag) bool {
	return IsV2EnabledForService() && v2Node != nil && finalityTag != ""
}

func GetFinalityTagFromMetadata(metadata map[string]interface{}) (FinalityTag, error) {
	if metadata == nil {
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

	return "", nil
}

func GetFinalityTagFromNetworkIdentifier(networkIdentifier *types.NetworkIdentifier) (FinalityTag, error) {
	if networkIdentifier == nil {
		return "", nil
	}

	if networkIdentifier.SubNetworkIdentifier == nil {
		return "", nil
	}

	if networkIdentifier.SubNetworkIdentifier.Metadata == nil {
		return "", fmt.Errorf("sub_network_identifier requires metadata with finality_tag")
	}

	return GetFinalityTagFromMetadata(networkIdentifier.SubNetworkIdentifier.Metadata)
}

func CreateTagSelector(tag FinalityTag) filTypes.TipSetSelector {
	var tipsetTag filTypes.TipSetTag

	switch tag {
	case FinalityLatest, FinalitySafe, FinalityFinalized:
		tipsetTag = filTypes.TipSetTag(tag)
	default:
		tipsetTag = "latest"
	}

	return filTypes.TipSetSelector{
		Tag: &tipsetTag,
	}
}

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

	if tag != "" {
		return nil, fmt.Errorf("finality_tag '%s' requires V2 APIs to be enabled", tag)
	}

	return v1Node.ChainHead(ctx)
}

type TipSetResolution struct {
	TipSet       *filTypes.TipSet
	Height       int64
	IsNullTipSet bool
}

// ResolveTipSetForFinality resolves tipset based on height, finality tag, and anchor mode.
// Implements both Height Comparison Mode (max logic) and Finality Anchor Mode (chain anchoring).
//
// requestedHeight: -1 (unspecified), 0 (current finality), or specific height
// finalityTag: "latest", "safe", "finalized", or "" (no finality)
//
// Modes:
//   - Anchor Mode (EnableFinalityAnchor=true): Returns exact height from finality chain
//   - Height Comparison Mode (default): Returns max(requestedHeight, finality_height)
func ResolveTipSetForFinality(
	ctx context.Context,
	v1Node api.FullNode,
	v2Node v2api.FullNode,
	requestedHeight int64,
	finalityTag FinalityTag,
) (*TipSetResolution, error) {
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

	if finalityTag == "" {
		tipSet, err := v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(requestedHeight), filTypes.EmptyTSK)
		if err != nil {
			return nil, err
		}
		isNull := int64(tipSet.Height()) != requestedHeight
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: isNull,
		}, nil
	}

	// Special case: height=0 returns current finality height
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

	// Anchor Mode: Query exact height from finality chain using V2 API
	if IsFinalityAnchorEnabled() {
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

		isNull := int64(tipSet.Height()) != requestedHeight
		return &TipSetResolution{
			TipSet:       tipSet,
			Height:       int64(tipSet.Height()),
			IsNullTipSet: isNull,
		}, nil
	}

	// Height Comparison Mode: Return max(requestedHeight, finality_height)
	finalityTipSet, err := ChainGetTipSetWithFallback(ctx, v1Node, v2Node, finalityTag)
	if err != nil {
		return nil, err
	}
	finalityHeight := int64(finalityTipSet.Height())

	if requestedHeight >= finalityHeight {
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

	return &TipSetResolution{
		TipSet:       finalityTipSet,
		Height:       finalityHeight,
		IsNullTipSet: false,
	}, nil
}
