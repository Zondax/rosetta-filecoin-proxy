package services

import (
	"context"
	"github.com/filecoin-project/go-address"
	types2 "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/ipfs/go-cid"
	"reflect"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"github.com/stretchr/testify/mock"
	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

const (
	ScnSyncError = iota
	ScnAllComplete
	ScnAllIdle
	ScnWorkerWithNoTarget
	ScnWorkerWithError
)

var nodeMock = mocks.FullNodeMock{}


func TestCheckSyncStatus(t *testing.T) {

	type args struct {
		ctx  context.Context
		node api.FullNode
	}
	tests := []struct {
		name  string
		args  args
		scenario int
		want  *SyncStatus
		want1 *types.Error
	}{
		{
			name: "SyncWithError",
			args: args{
						ctx: context.Background(),
						node: setupMockForScenario(ScnSyncError),
					},
			want:  nil,
			want1: ErrSyncErrored,
		},
		{
			name: "SyncCompleted",
			args: args{
				ctx: context.Background(),
				node: setupMockForScenario(ScnAllComplete),
			},
			want:  &SyncStatus{
				currentHeight:   []int64{0,0},
				globalSyncState: api.StageSyncComplete,
			},
			want1: nil,
		},
		{
			name: "SyncIdle - Workers Idle",
			args: args{
				ctx: context.Background(),
				node: setupMockForScenario(ScnAllIdle),
			},
			want:  &SyncStatus{
				currentHeight:   []int64{0,0},
				globalSyncState: api.StageIdle,
			},
			want1: nil,
		},
		{
			name: "SyncIdle - Targets nil",
			args: args{
				ctx: context.Background(),
				node: setupMockForScenario(ScnAllIdle),
			},
			want:  &SyncStatus{
				currentHeight:   []int64{0,0},
				globalSyncState: api.StageIdle,
			},
			want1: nil,
		},
		{
			name: "WorkerWithError",
			args: args{
				ctx: context.Background(),
				node: setupMockForScenario(ScnWorkerWithError),
			},
			want:  nil,
			want1: ErrSyncErrored,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckSyncStatus(tt.args.ctx, &tt.args.node)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckSyncStatus() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CheckSyncStatus() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestSyncStatus_GetMaxHeight(t *testing.T) {
	type fields struct {
		currentHeight   []int64
		globalSyncState api.SyncStateStage
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "MaxZeroForEmptyHeights",
			fields: fields { currentHeight:   []int64{},
				globalSyncState: api.StageSyncComplete},
			want: 0,
		},
		{
			name: "ErrorForSyncError",
			fields: fields { currentHeight:   []int64{100, 200, 300},
							 globalSyncState: api.StageSyncErrored},
			want: -1,
		},
		{
			name: "MaxZeroForSyncingHeadersStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageHeaders},
			want: 0,
		},
		{
			name: "MaxForSyncingMessagesStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageMessages},
			want: 299,
		},
		{
			name: "MaxForSyncCompleteStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageSyncComplete},
			want: 299,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := SyncStatus{
				currentHeight:   tt.fields.currentHeight,
				globalSyncState: tt.fields.globalSyncState,
			}
			if got := status.GetMaxHeight(); got != tt.want {
				t.Errorf("GetMaxHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncStatus_GetMinHeight(t *testing.T) {
	type fields struct {
		currentHeight   []int64
		globalSyncState api.SyncStateStage
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "MinZeroForEmptyHeights",
			fields: fields { currentHeight:   []int64{},
				globalSyncState: api.StageSyncComplete},
			want: 0,
		},
		{
			name: "ErrorForSyncedError",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageSyncErrored},
			want: -1,
		},
		{
			name: "MinZeroForSyncingHeadersStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageHeaders},
			want: 0,
		},
		{
			name: "MinZeroForPersistingHeadersStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StagePersistHeaders},
			want: 0,
		},
		{
			name: "MinValueForSyncingMessagesStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageMessages},
			want: 100,
		},
		{
			name: "MinValueForSyncCompleteStage",
			fields: fields { currentHeight:   []int64{100, 200, 300},
				globalSyncState: api.StageSyncComplete},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := SyncStatus{
				currentHeight:   tt.fields.currentHeight,
				globalSyncState: tt.fields.globalSyncState,
			}
			if got := status.GetMinHeight(); got != tt.want {
				t.Errorf("GetMinHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncStatus_IsSynced(t *testing.T) {
	type fields struct {
		currentHeight   []int64
		globalSyncState api.SyncStateStage
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "NotSyncedOnStageIdle",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StageIdle,
			},
			want: false,
		},
		{
			name: "NotSyncedOnStageHeaders",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StageHeaders,
			},
			want: false,
		},
		{
			name: "NotSyncedOnStagePersistHeaders",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StagePersistHeaders,
			},
			want: false,
		},
		{
			name: "NotSyncedOnStageMessages",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StageMessages,
			},
			want: false,
		},
		{
			name: "SyncedOnStageSyncComplete",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StageSyncComplete,
			},
			want: true,
		},
		{
			name: "NotSyncedOnStageSyncErrored",
			fields: fields{
				currentHeight:   nil,
				globalSyncState: api.StageSyncErrored,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := SyncStatus{
				currentHeight:   tt.fields.currentHeight,
				globalSyncState: tt.fields.globalSyncState,
			}
			if got := status.IsSynced(); got != tt.want {
				t.Errorf("IsSynced() = %v, want %v", got, tt.want)
			}
		})
	}
}

func setupMockForScenario(scn int) *mocks.FullNodeMock {

	nodeMock := mocks.FullNodeMock{}

	mockCid, _ := cid.Parse("bafkqaaa")
	mockMiner, _ := address.NewFromString("t00")
	mockTargetTipSet, _ := types2.NewTipSet([]*types2.BlockHeader{
		{
			Miner:                 mockMiner,
			Height:                abi.ChainEpoch(0),
			ParentStateRoot:       mockCid,
			Messages:              mockCid,
			ParentMessageReceipts: mockCid,
			BlockSig:              &crypto.Signature{Type: crypto.SigTypeBLS},
			BLSAggregate:          &crypto.Signature{Type: crypto.SigTypeBLS},
		},
	},
	)

	switch scn {
	case ScnSyncError:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage: api.StageSyncErrored,
						Height: 0,
						Target: mockTargetTipSet,
					},
				},
			}, nil)
	case ScnAllComplete:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage: api.StageSyncComplete,
						Height: 0,
						Target: mockTargetTipSet,
					},
					{
						Stage: api.StageSyncComplete,
						Height: 0,
						Target: mockTargetTipSet,
					},
				},
			}, nil)
	case ScnAllIdle:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage: api.StageIdle,
						Height: 0,
						Target: mockTargetTipSet,
					},
					{
						Stage: api.StageIdle,
						Height: 0,
						Target: mockTargetTipSet,
					},
				},
			}, nil)
	case ScnWorkerWithNoTarget:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage: api.StageHeaders,
						Height: 0,
					},
					{
						Stage: api.StageHeaders,
						Height: 0,
					},
				},
			}, nil)
	case ScnWorkerWithError:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage: api.StageSyncComplete,
						Height: 0,
						Target: mockTargetTipSet,
					},
					{
						Stage: api.StageSyncErrored,
						Height: 0,
						Target: mockTargetTipSet,
					},
				},
			}, nil)
	}

	return &nodeMock
}