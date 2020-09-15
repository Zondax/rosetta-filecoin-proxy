package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
)

const DummyHash = "0000000000000000000000000000000000000000"

// NetworkAPIService implements the server.NetworkAPIServicer interface.
type NetworkAPIService struct {
	network *types.NetworkIdentifier
	node    api.FullNode
}

// NewNetworkAPIService creates a new instance of a NetworkAPIService.
func NewNetworkAPIService(network *types.NetworkIdentifier, node *api.FullNode) server.NetworkAPIServicer {
	return &NetworkAPIService{
		network: network,
		node:    *node,
	}
}

// NetworkList implements the /network/list endpoint
func (s *NetworkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	networkName, err := s.node.StateNetworkName(ctx)
	if err != nil {
		return nil, ErrUnableToGetChainID
	}

	resp := &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			{
				Blockchain: BlockChainName,
				Network:    string(networkName),
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
		headTipSet            *filTypes.TipSet
		err                   error
		useDummyHead          = false
		blockIndex, timeStamp int64
		blockHashedTipSet     string
	)

	//Check sync status

	status, syncErr := CheckSyncStatus(ctx, &s.node)
	if syncErr != nil {
		return nil, syncErr
	}
	stage := status.globalSyncState.String()
	syncStatus := &types.SyncStatus{
		Stage:        &stage,
		CurrentIndex: status.GetMaxHeight(),
		TargetIndex:  status.GetTargetIndex(),
	}
	if !status.IsSynced() {
		//Cannot retrieve any TipSet while node is syncing
		//use a dummy TipSet instead
		useDummyHead = true
	}

	//Get head TipSet
	headTipSet, err = s.node.ChainHead(ctx)

	if err != nil || headTipSet == nil {
		return nil, ErrUnableToGetLatestBlk
	}

	hashHeadTipSet, err := BuildTipSetKeyHash(headTipSet.Key())
	if err != nil {
		return nil, ErrUnableToBuildTipSetHash
	}

	//Get genesis TipSet
	genesisTipSet, err := s.node.ChainGetGenesis(ctx)
	if err != nil || genesisTipSet == nil {
		return nil, ErrUnableToGetGenesisBlk
	}

	hashGenesisTipSet, err := BuildTipSetKeyHash(genesisTipSet.Key())
	if err != nil {
		return nil, ErrUnableToBuildTipSetHash
	}

	//Get peers data
	peersFil, err := s.node.NetPeers(ctx)
	if err != nil {
		return nil, ErrUnableToGetPeers
	}

	var peers []*types.Peer
	for _, peerFil := range peersFil {
		peers = append(peers, &types.Peer{
			PeerID: peerFil.ID.String(),
		})
	}

	if !useDummyHead {
		blockIndex = int64(headTipSet.Height())
		timeStamp = int64(headTipSet.MinTimestamp()) * FactorSecondToMillisecond
		blockHashedTipSet = *hashHeadTipSet
	} else {
		blockIndex = 0
		timeStamp = 0
		blockHashedTipSet = DummyHash
	}

	resp := &types.NetworkStatusResponse{
		CurrentBlockIdentifier: &types.BlockIdentifier{
			Index: blockIndex,
			Hash:  blockHashedTipSet,
		},
		CurrentBlockTimestamp: timeStamp, // [ms]
		GenesisBlockIdentifier: &types.BlockIdentifier{
			Index: int64(genesisTipSet.Height()),
			Hash:  *hashGenesisTipSet,
		},
		Peers:      peers,
		SyncStatus: syncStatus,
	}

	return resp, nil
}

// NetworkOptions implements the /network/options endpoint.
func (s *NetworkAPIService) NetworkOptions(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {

	version, err := s.node.Version(ctx)
	if err != nil {
		return nil, ErrUnableToGetNodeInfo
	}

	operations := make([]string, 0, len(SupportedOperations))
	for op := range SupportedOperations {
		operations = append(operations, op)
	}

	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: RosettaSDKVersion,
			NodeVersion:    version.Version,
		},
		Allow: &types.Allow{
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
			OperationTypes: operations,
			Errors:         ErrorList,
		},
	}, nil
}
