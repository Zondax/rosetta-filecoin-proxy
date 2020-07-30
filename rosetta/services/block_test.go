package services

import (
	"context"
	"reflect"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/mock"
	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

var NetworkID = &types.NetworkIdentifier{
	Blockchain: "Filecoin",
	Network:    "testnet",
}

func TestBlockAPIService_Block(t *testing.T) {

	nodeMock := mocks.FullNodeMock{}

	// Mock needed input arguments
	var requestedIndex int64 = 0
	requestedHash := "0171a0e40220a63ae9efee6a34c827982013a398f6efcd714d414c2435170efae73669713fe3"
	mockMetadata := make(map[string]interface{})
	mockMetadata[BlockCIDsKey] = []string{"bafy2bzacedecjgc3svxb7itplzqpebioa2g2bdowizz4suv6swg4ctipzgy5o"}
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
		},
	},
	)
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
		node    api.FullNode
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
				node:    &nodeMock,
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
				node:    tt.fields.node,
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
		node    api.FullNode
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
				node:    tt.fields.node,
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
		api     *api.FullNode
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
			if got := NewBlockAPIService(tt.args.network, tt.args.api); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
