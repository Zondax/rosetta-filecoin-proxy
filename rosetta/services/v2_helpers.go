package services

import (
	"context"
	"fmt"

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
	FinalityLatest    FinalityTag = "latest"
	FinalitySafe      FinalityTag = "safe"
	FinalityFinalized FinalityTag = "finalized"
)

func shouldUseV2API(v2Node v2api.FullNode, finalityTag FinalityTag) bool {
	return EnableLotusV2APIs && v2Node != nil && finalityTag != ""
}

func GetFinalityTagFromMetadata(metadata map[string]interface{}) (FinalityTag, error) {
	if metadata == nil {
		return "", nil
	}

	if tagValue, exists := metadata[MetadataFinalityTag]; exists {
		if tagStr, ok := tagValue.(string); ok {
			switch tagStr {
			case "safe":
				return FinalitySafe, nil
			case "finalized":
				return FinalityFinalized, nil
			case "latest":
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
	if EnableFinalityAnchor {
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

// MaxNullTipSetStretch is the largest consecutive null-tipset stretch
// that ResolveSuccessorTipSet will walk before giving up. 50 is
// generous: even calibration testnet's longest historical null runs
// were single-digit; bound exists primarily to guarantee loop
// termination if a misconfigured upstream node never returns a
// non-null tipset above a given height.
const MaxNullTipSetStretch = 50

// ResolveSuccessorTipSet returns the next non-null tipset strictly
// above `resolvedHeight`, walking forward through any null epochs.
// The returned tipset's PARENT state reflects end-of-`resolvedHeight`,
// which is what callers want to pass to StateGetActor / MsigGet* to
// read balance "as of the end of block `resolvedHeight`".
//
// Returns (nil, nil) when no successor is found within
// MaxNullTipSetStretch epochs OR when the walk would go past chain
// head. Callers should treat this as "no observable successor" and
// fall back to reading state at parent(resolvedHeight) with a shifted
// response identifier (matching the pre-PR #310 useHeadTipSet branch).
//
// Returns (nil, err) on any RPC error during the walk.
//
// Mode dispatch:
//
//   - Default (no finality tag): walk via v1Node.ChainGetTipSetByHeight
//     on the head chain.
//   - V2 Anchor Mode (EnableFinalityAnchor && shouldUseV2API): walk via
//     v2Node.ChainGetTipSet with a finality-chain TipSetSelector. This
//     is essential because resolution.TipSet in anchor mode lives on
//     the finality chain — querying successors via v1 would land on
//     the head chain, potentially a different fork after a reorg
//     between finality anchor and head, and StateGetActor on a
//     head-chain successor would read state on the wrong chain.
//   - V2 Height-Comparison Mode (default V2 behavior): walks via v1
//     since resolution.TipSet there is the v1-resolved height when
//     requestedHeight >= finalityHeight, or the finality tipset
//     otherwise — both are reachable through v1's chain.
func ResolveSuccessorTipSet(
	ctx context.Context,
	v1Node api.FullNode,
	v2Node v2api.FullNode,
	resolvedHeight int64,
	headHeight int64,
	finalityTag FinalityTag,
) (*filTypes.TipSet, error) {
	useV2Anchor := EnableFinalityAnchor && shouldUseV2API(v2Node, finalityTag)

	for offset := int64(1); offset <= MaxNullTipSetStretch; offset++ {
		target := resolvedHeight + offset
		if target > headHeight {
			// Walked past head — no successor exists yet.
			return nil, nil
		}

		var candidate *filTypes.TipSet
		var candidateErr error
		if useV2Anchor {
			epoch := abi.ChainEpoch(target)
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
			candidate, candidateErr = v2Node.ChainGetTipSet(ctx, selector)
		} else {
			candidate, candidateErr = v1Node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(target), filTypes.EmptyTSK)
		}
		if candidateErr != nil {
			return nil, candidateErr
		}
		if int64(candidate.Height()) > resolvedHeight {
			return candidate, nil
		}
		// Otherwise the lookup returned a tipset at a lower height,
		// meaning `target` was null. Try the next offset.
	}
	// Exhausted maxNullStretch without finding a successor. Caller
	// should fall back as if at head.
	return nil, nil
}
