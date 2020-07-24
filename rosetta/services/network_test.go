package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestNetworkAPIService_NetworkList(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.MetadataRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.NetworkListResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NetworkAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := s.NetworkList(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NetworkList() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NetworkList() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNetworkAPIService_NetworkOptions(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.NetworkRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.NetworkOptionsResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NetworkAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := s.NetworkOptions(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NetworkOptions() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NetworkOptions() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNetworkAPIService_NetworkStatus(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.NetworkRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.NetworkStatusResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &NetworkAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := s.NetworkStatus(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NetworkStatus() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("NetworkStatus() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewNetworkAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		node    *api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.NetworkAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNetworkAPIService(tt.args.network, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNetworkAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
