package services

import (
	"context"

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

	resp := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: BlockChainName,
				Network:    string(networkName),
			},
			{
				Blockchain: BlockChainName,
				Network:    string(networkName),
				SubNetworkIdentifier: &types.SubNetworkIdentifier{
					Network: SubNetworkF3,
				},
			},
		},
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
	syncStatus := &types.SyncStatus{
		Stage:        &stage,
		CurrentIndex: &currentIndex,
		TargetIndex:  targetIndex,
	}
	if !status.IsSynced() {
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
	s.response.SyncStatus = syncStatus

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

	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: RosettaSDKVersion,
			NodeVersion:    version.Version,
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
