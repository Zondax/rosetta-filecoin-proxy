// +build rosetta_rpc

package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestCheckSyncStatus(t *testing.T) {
	type args struct {
		ctx  context.Context
		node *api.FullNode
	}
	tests := []struct {
		name  string
		args  args
		want  *SyncStatus
		want1 *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CheckSyncStatus(tt.args.ctx, tt.args.node)
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
