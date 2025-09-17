package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
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

// IsForceSafeF3FinalityEnabled checks if safe F3 finality should be forced when V2 APIs are enabled
func IsForceSafeF3FinalityEnabled() bool {
	enabled, err := strconv.ParseBool(ForceSafeF3Finality)
	if err != nil {
		return false // Default to false on parse error
	}
	return enabled
}

// shouldUseV2API determines if V2 API should be used based on configuration and availability
func shouldUseV2API(v2Node v2api.FullNode, finalityTag FinalityTag) bool {
	return IsV2EnabledForService() && v2Node != nil && finalityTag != ""
}

// getEffectiveFinalityTag returns the finality tag to use, applying force safe logic if needed
func getEffectiveFinalityTag(requestedTag FinalityTag, v2Node v2api.FullNode) FinalityTag {
	if requestedTag != "" {
		return requestedTag
	}

	if IsV2EnabledForService() && IsForceSafeF3FinalityEnabled() && v2Node != nil {
		return FinalitySafe
	}

	return ""
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
	effectiveTag := getEffectiveFinalityTag(tag, v2Node)

	if shouldUseV2API(v2Node, effectiveTag) {
		selector := CreateTagSelector(effectiveTag)
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

// StateGetActorWithFallback is a wrapper that uses V2 StateGetActor if enabled,
// otherwise falls back to V1 StateGetActor with EmptyTSK only if no finality tag is specified
func StateGetActorWithFallback(ctx context.Context, v1Node api.FullNode, v2Node v2api.FullNode, addr address.Address, tag FinalityTag) (*filTypes.Actor, error) {
	effectiveTag := getEffectiveFinalityTag(tag, v2Node)

	if shouldUseV2API(v2Node, effectiveTag) {
		selector := CreateTagSelector(effectiveTag)
		actor, err := v2Node.StateGetActor(ctx, addr, selector)
		if err != nil {
			return nil, fmt.Errorf("v2 StateGetActor failed: %w", err)
		}
		return actor, nil
	}

	// If finality tag is specified but V2 is not enabled/available, return error
	if tag != "" {
		return nil, fmt.Errorf("finality_tag '%s' requires V2 APIs to be enabled", tag)
	}

	// Use V1 API when no finality tag is specified
	return v1Node.StateGetActor(ctx, addr, filTypes.EmptyTSK)
}
