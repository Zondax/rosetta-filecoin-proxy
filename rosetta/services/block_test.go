package services

import (
	"context"
	"reflect"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/mock"
	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

var NetworkID = &types.NetworkIdentifier{
	Blockchain: "Filecoin",
	Network:    "testnet",
}

func TestBlockAPIService_Block(t *testing.T) {

	nodeMock := mocks.FullNode{}

	// Mock needed input arguments
	var requestedIndex int64 = 0
	mockCid, _ := cid.Parse("bafkqaaa")
	mockMiner, _ := address.NewFromString("t00")
	mockTipSet, _ := filTypes.NewTipSet([]*filTypes.BlockHeader{
		{
			Miner:                 mockMiner,
			Height:                abi.ChainEpoch(requestedIndex),
			ParentStateRoot:       mockCid,
			Messages:              mockCid,
			ParentMessageReceipts: mockCid,
			BlockSig:              &crypto.Signature{Type: crypto.SigTypeBLS},
			BLSAggregate:          &crypto.Signature{Type: crypto.SigTypeBLS},
			// Lotus v1.36's filTypes.NewTipSet requires non-nil Ticket.
			Ticket: &filTypes.Ticket{VRFProof: []byte{0}},
		},
	},
	)
	// Derive identifiers from the tipset rather than hardcoding them.
	// The previous fixture hardcoded `requestedHash` and the BlockCIDs
	// entry as the literal strings produced before Lotus v1.36 forced
	// the Ticket field to be non-nil — adding a Ticket changes the
	// block's content-addressed CID and therefore the TipSetKey hash,
	// invalidating the hardcoded values. Deriving them keeps the test
	// resilient to BlockHeader-shape changes.
	requestedHashPtr, _ := BuildTipSetKeyHash(mockTipSet.Key())
	requestedHash := *requestedHashPtr
	mockMetadata := make(map[string]interface{})
	blockCIDs := make([]string, 0, len(mockTipSet.Cids()))
	for _, c := range mockTipSet.Cids() {
		blockCIDs = append(blockCIDs, c.String())
	}
	mockMetadata[BlockCIDsKey] = blockCIDs
	///

	// Mock functions
	nodeMock.On("StateNetworkName", mock.Anything).
		Return(dtypes.NetworkName(NetworkID.Network), nil)
	nodeMock.On("SyncState", mock.Anything).
		Return(&api.SyncState{
			ActiveSyncs: []api.ActiveSync{
				{
					Stage:  api.StageSyncComplete,
					Target: &filTypes.TipSet{},
				},
			},
		},
			nil)
	nodeMock.On("ChainGetTipSetByHeight", mock.Anything, mock.Anything, mock.Anything).
		Return(mockTipSet, nil)
	nodeMock.On("ChainGetParentMessages", mock.Anything, mock.Anything).
		Return([]api.Message{}, nil)

	nodeMock.On("ChainGetParentReceipts", mock.Anything, mock.Anything).
		Return([]*filTypes.MessageReceipt{}, nil)
	///
	// Output
	var responseTest1 = &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: requestedIndex,
				Hash:  requestedHash,
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: requestedIndex,
				Hash:  requestedHash,
			},
			Timestamp: 0,
			Metadata:  mockMetadata,
		},
	}

	///

	type fields struct {
		network *types.NetworkIdentifier
		v1Node  api.FullNode
		v2Node  v2api.FullNode
	}

	type args struct {
		ctx     context.Context
		request *types.BlockRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.BlockResponse
		want1  *types.Error
	}{
		{
			name: "RetrieveGenesisTipSet",
			fields: fields{
				network: NetworkID,
				v1Node:  &nodeMock,
				v2Node:  nil,
			},
			args: args{
				ctx: context.Background(),
				request: &types.BlockRequest{
					NetworkIdentifier: NetworkID,
					BlockIdentifier: &types.PartialBlockIdentifier{
						Index: &requestedIndex,
					},
				},
			},
			want:  responseTest1,
			want1: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BlockAPIService{
				network: tt.fields.network,
				v1Node:  tt.fields.v1Node,
				v2Node:  tt.fields.v2Node,
			}
			got, got1 := s.Block(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Block() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Block() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBlockAPIService_BlockTransaction(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		v1Node  api.FullNode
		v2Node  v2api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.BlockTransactionRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.BlockTransactionResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BlockAPIService{
				network: tt.fields.network,
				v1Node:  tt.fields.v1Node,
				v2Node:  tt.fields.v2Node,
			}
			got, got1 := s.BlockTransaction(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BlockTransaction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BlockTransaction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewBlockAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		v1API   *api.FullNode
		v2API   *v2api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.BlockAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBlockAPIService(tt.args.network, tt.args.v1API, tt.args.v2API, rosettaLib); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
