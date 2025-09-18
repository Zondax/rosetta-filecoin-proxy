package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
)

// NetworkAPIService implements the server.NetworkAPIServicer interface.
type NetworkAPIService struct {
	response     *types.NetworkStatusResponse
	network      *types.NetworkIdentifier
	v1Node       api.FullNode
	v2Node       v2api.FullNode
	supportedOps []string
}

// NewNetworkAPIService creates a new instance of a NetworkAPIService.
func NewNetworkAPIService(network *types.NetworkIdentifier, v1API *api.FullNode, v2API v2api.FullNode, supportedOps []string) server.NetworkAPIServicer {
	return &NetworkAPIService{
		network:      network,
		v1Node:       *v1API,
		v2Node:       v2API,
		supportedOps: supportedOps,
	}
}

// NetworkList implements the /network/list endpoint
func (s *NetworkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	networkName, err := s.v1Node.StateNetworkName(ctx)
	if err != nil {
		return nil, ErrUnableToGetChainID
	}

	f3NetworkIdentifier := []*types.NetworkIdentifier{}
	if IsV2EnabledForService() {
		f3NetworkIdentifier = []*types.NetworkIdentifier{
			{
				Blockchain: BlockChainName,
				Network:    string(networkName),
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
					Metadata: map[string]interface{}{
						MetadataFinalityTag: fmt.Sprintf("%s/%s/%s", FinalityTagLatest, FinalityTagSafe, FinalityTagFinalized),
					},
				},
			},
		}
	}

	resp := &types.NetworkListResponse{
		NetworkIdentifiers: append([]*types.NetworkIdentifier{
			{
				Blockchain: BlockChainName,
				Network:    string(networkName),
			},
		}, f3NetworkIdentifier...),
	}

	return resp, nil
}

// NetworkStatus implements the /network/status endpoint.
func (s *NetworkAPIService) NetworkStatus(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {

	var (
		headTipSet       *filTypes.TipSet
		err              error
		useGenesisTipSet = false
	)

	// Extract finality tag from network identifier if F3 sub-network is requested
	finalityTag, err := GetFinalityTagFromNetworkIdentifier(request.NetworkIdentifier)
	if err != nil {
		return nil, BuildError(ErrMalformedValue, err, false)
	}

	// Check sync status
	status, syncErr := CheckSyncStatus(ctx, &s.v1Node)
	if syncErr != nil {
		return nil, syncErr
	}

	currentIndex := status.GetMaxHeight()
	targetIndex := status.GetTargetIndex()

	stage := status.globalSyncState.String()
	synced := status.IsSynced()
	syncStatus := &types.SyncStatus{
		Stage:        &stage,
		CurrentIndex: &currentIndex,
		TargetIndex:  targetIndex,
		Synced:       &synced,
	}

	if !synced {
		// Cannot retrieve any TipSet while node is syncing
		// use Genesis TipSet instead
		useGenesisTipSet = true
	}

	// Get head TipSet using v2 API with finality tag if requested
	if !useGenesisTipSet {
		headTipSet, err = ChainGetTipSetWithFallback(ctx, s.v1Node, s.v2Node, finalityTag)
		if err != nil {
			return nil, BuildError(ErrUnableToGetLatestBlk, err, true)
		}
	} else {
		// When syncing, always use v1 ChainHead
		headTipSet, err = s.v1Node.ChainHead(ctx)
		if err != nil || headTipSet == nil {
			return nil, BuildError(ErrUnableToGetLatestBlk, err, true)
		}
	}

	hashHeadTipSet, err := BuildTipSetKeyHash(headTipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, err, true)
	}

	// Get genesis TipSet
	genesisTipSet, err := s.v1Node.ChainGetGenesis(ctx)
	if err != nil || genesisTipSet == nil {
		return nil, BuildError(ErrUnableToGetGenesisBlk, err, true)
	}

	hashGenesisTipSet, err := BuildTipSetKeyHash(genesisTipSet.Key())
	if err != nil {
		return nil, BuildError(ErrUnableToBuildTipSetHash, err, true)
	}

	// Get peers data
	peersFil, err := s.v1Node.NetPeers(ctx)
	if err != nil {
		return nil, BuildError(ErrUnableToGetPeers, err, true)
	}

	var peers []*types.Peer
	for _, peerFil := range peersFil {
		peers = append(peers, &types.Peer{
			PeerID: peerFil.ID.String(),
		})
	}

	if s.response == nil {
		// We should only enter this codepath once
		// Initialize the very first response
		s.response = &types.NetworkStatusResponse{
			CurrentBlockIdentifier: &types.BlockIdentifier{
				Index: 0,
				Hash:  *hashGenesisTipSet,
			},
			CurrentBlockTimestamp: int64(genesisTipSet.MinTimestamp()) * FactorSecondToMillisecond, // [ms]
			GenesisBlockIdentifier: &types.BlockIdentifier{
				Index: int64(genesisTipSet.Height()),
				Hash:  *hashGenesisTipSet,
			},
		}
	}

	if !useGenesisTipSet {
		// Update block height, hash and timestamp
		s.response.CurrentBlockIdentifier = &types.BlockIdentifier{
			Index: int64(headTipSet.Height()),
			Hash:  *hashHeadTipSet,
		}

		s.response.CurrentBlockTimestamp = int64(headTipSet.MinTimestamp()) * FactorSecondToMillisecond // [ms]
	}

	// Always update Peers and SyncStatus
	s.response.Peers = peers

	// Enhance SyncStatus with F3 finality information
	enhancedSyncStatus := &types.SyncStatus{
		Stage:        syncStatus.Stage,
		CurrentIndex: syncStatus.CurrentIndex,
		TargetIndex:  syncStatus.TargetIndex,
		Synced:       syncStatus.Synced,
	}

	// Add F3 metadata to sync status
	if IsV2EnabledForService() {
		// Create stage info that includes F3 status
		f3Stage := ""
		if syncStatus.Stage != nil {
			f3Stage = *syncStatus.Stage
		}

		f3Info := fmt.Sprintf("%s (F3: enabled", f3Stage)
		if finalityTag != "" {
			f3Info += fmt.Sprintf(", finality: %s", finalityTag)
		}
		if IsForceSafeF3FinalityEnabled() {
			f3Info += fmt.Sprintf(", force_f3: true, default_f3_finality: %s)", FinalityTagSafe)
		}
		enhancedSyncStatus.Stage = &f3Info
	} else {
		// Indicate F3 is disabled
		f3Stage := ""
		if syncStatus.Stage != nil {
			f3Stage = *syncStatus.Stage
		}
		f3Info := fmt.Sprintf("%s (F3: disabled)", f3Stage)
		enhancedSyncStatus.Stage = &f3Info
	}

	s.response.SyncStatus = enhancedSyncStatus

	return s.response, nil
}

// NetworkOptions implements the /network/options endpoint.
func (s *NetworkAPIService) NetworkOptions(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {

	version, err := s.v1Node.Version(ctx)
	if err != nil {
		return nil, BuildError(ErrUnableToGetNodeInfo, err, false)
	}

	// Create version metadata with F3 information
	versionMetadata := map[string]interface{}{}

	if IsV2EnabledForService() {
		versionMetadata["f3_enabled"] = true
		versionMetadata["f3_supported_finality_tags"] = []string{
			FinalityTagLatest,
			FinalityTagSafe,
			FinalityTagFinalized,
		}
		versionMetadata["f3_sub_network"] = SubNetworkF3
		if IsForceSafeF3FinalityEnabled() {
			versionMetadata["force_f3"] = true
			versionMetadata["default_f3_finality_tag"] = FinalityTagSafe
		}
	} else {
		versionMetadata["f3_enabled"] = false
		versionMetadata["f3_reason"] = "V2 APIs disabled"
	}

	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: RosettaSDKVersion,
			NodeVersion:    version.Version,
			Metadata:       versionMetadata,
		},
		Allow: &types.Allow{
			HistoricalBalanceLookup: true,
			OperationStatuses: []*types.OperationStatus{
				{
					Status:     OperationStatusOk,
					Successful: true,
				},
				{
					Status:     OperationStatusFailed,
					Successful: false,
				},
			},
			OperationTypes: s.supportedOps,
			Errors:         ErrorList,
		},
	}, nil
}
