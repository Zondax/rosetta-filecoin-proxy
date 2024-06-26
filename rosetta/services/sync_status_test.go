package services

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	filTypes "github.com/filecoin-project/lotus/chain/types"
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

func TestCheckSyncStatus(t *testing.T) {

	type args struct {
		ctx  context.Context
		node api.FullNode
	}
	tests := []struct {
		name     string
		args     args
		scenario int
		want     *SyncStatus
		want1    *types.Error
	}{
		{
			name: "SyncWithError",
			args: args{
				ctx:  context.Background(),
				node: setupMockForScenario(ScnSyncError),
			},
			want: &SyncStatus{
				currentHeight:   []int64{0},
				targetIndex:     []int64{100},
				globalSyncState: api.StageSyncErrored,
			},
			want1: nil,
		},
		{
			name: "SyncCompleted",
			args: args{
				ctx:  context.Background(),
				node: setupMockForScenario(ScnAllComplete),
			},
			want: &SyncStatus{
				currentHeight:   []int64{100, 200},
				targetIndex:     []int64{100, 200},
				globalSyncState: api.StageSyncComplete,
			},
			want1: nil,
		},
		{
			name: "SyncIdle - Workers Idle",
			args: args{
				ctx:  context.Background(),
				node: setupMockForScenario(ScnAllIdle),
			},
			want: &SyncStatus{
				currentHeight:   []int64{0, 0},
				targetIndex:     []int64{100, 200},
				globalSyncState: api.StageIdle,
			},
			want1: nil,
		},
		{
			name: "SyncIdle - Targets nil",
			args: args{
				ctx:  context.Background(),
				node: setupMockForScenario(ScnAllIdle),
			},
			want: &SyncStatus{
				currentHeight:   []int64{0, 0},
				targetIndex:     []int64{100, 200},
				globalSyncState: api.StageIdle,
			},
			want1: nil,
		},
		{
			name: "WorkerWithError",
			args: args{
				ctx:  context.Background(),
				node: setupMockForScenario(ScnWorkerWithError),
			},
			want: &SyncStatus{
				currentHeight:   []int64{0, 0},
				targetIndex:     []int64{100, 200},
				globalSyncState: api.StageSyncErrored,
			},
			want1: nil,
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
			fields: fields{currentHeight: []int64{},
				globalSyncState: api.StageSyncComplete},
			want: 0,
		},
		{
			name: "ErrorForSyncError",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageSyncErrored},
			want: -1,
		},
		{
			name: "MaxZeroForSyncingHeadersStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageHeaders},
			want: 0,
		},
		{
			name: "MaxForSyncingMessagesStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageMessages},
			want: 300,
		},
		{
			name: "MaxForSyncCompleteStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageSyncComplete},
			want: 300,
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

func TestSyncStatus_GetTargetIndex(t *testing.T) {
	type fields struct {
		targetIndex     []int64
		globalSyncState api.SyncStateStage
	}

	var want1 int64 = 200

	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "ReturnMaxValue",
			fields: fields{
				targetIndex:     []int64{100, want1},
				globalSyncState: api.StageSyncComplete},
			want: want1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := SyncStatus{
				targetIndex:     tt.fields.targetIndex,
				globalSyncState: tt.fields.globalSyncState,
			}
			got := status.GetTargetIndex()
			if *got != tt.want {
				t.Errorf("GetTargetIndex() = %v, want %v", got, tt.want)
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
			fields: fields{currentHeight: []int64{},
				globalSyncState: api.StageSyncComplete},
			want: 0,
		},
		{
			name: "ErrorForSyncedError",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageSyncErrored},
			want: -1,
		},
		{
			name: "MinZeroForSyncingHeadersStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageHeaders},
			want: 0,
		},
		{
			name: "MinZeroForPersistingHeadersStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StagePersistHeaders},
			want: 0,
		},
		{
			name: "MinValueForSyncingMessagesStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
				globalSyncState: api.StageMessages},
			want: 100,
		},
		{
			name: "MinValueForSyncCompleteStage",
			fields: fields{currentHeight: []int64{100, 200, 300},
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

func buildMockTargetTipSet(epoch int64) *filTypes.TipSet {
	mockCid, _ := cid.Parse("bafkqaaa")
	mockMiner, _ := address.NewFromString("t00")
	mockTargetTipSet, _ := filTypes.NewTipSet([]*filTypes.BlockHeader{
		{
			Miner:                 mockMiner,
			Height:                abi.ChainEpoch(epoch),
			ParentStateRoot:       mockCid,
			Messages:              mockCid,
			ParentMessageReceipts: mockCid,
			BlockSig:              &crypto.Signature{Type: crypto.SigTypeBLS},
			BLSAggregate:          &crypto.Signature{Type: crypto.SigTypeBLS},
		},
	},
	)
	return mockTargetTipSet
}

func setupMockForScenario(scn int) *mocks.FullNode {

	nodeMock := mocks.FullNode{}

	mockTargetTipSet1 := buildMockTargetTipSet(100)
	mockTargetTipSet2 := buildMockTargetTipSet(200)

	switch scn {
	case ScnSyncError:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage:  api.StageSyncErrored,
						Height: 0,
						Target: mockTargetTipSet1,
					},
				},
			}, nil)
	case ScnAllComplete:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage:  api.StageSyncComplete,
						Height: 100,
						Target: mockTargetTipSet1,
					},
					{
						Stage:  api.StageSyncComplete,
						Height: 200,
						Target: mockTargetTipSet2,
					},
				},
			}, nil)
	case ScnAllIdle:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage:  api.StageIdle,
						Height: 0,
						Target: mockTargetTipSet1,
					},
					{
						Stage:  api.StageIdle,
						Height: 0,
						Target: mockTargetTipSet2,
					},
				},
			}, nil)
	case ScnWorkerWithNoTarget:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage:  api.StageHeaders,
						Height: 0,
					},
					{
						Stage:  api.StageHeaders,
						Height: 0,
					},
				},
			}, nil)
	case ScnWorkerWithError:
		nodeMock.On("SyncState", mock.Anything).
			Return(&api.SyncState{
				ActiveSyncs: []api.ActiveSync{
					{
						Stage:  api.StageSyncComplete,
						Height: 0,
						Target: mockTargetTipSet1,
					},
					{
						Stage:  api.StageSyncErrored,
						Height: 0,
						Target: mockTargetTipSet2,
					},
				},
			}, nil)
	}

	return &nodeMock
}
