package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestConstructionAPIService_ConstructionMetadata(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.ConstructionMetadataRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.ConstructionMetadataResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConstructionAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := c.ConstructionMetadata(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConstructionMetadata() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ConstructionMetadata() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConstructionAPIService_ConstructionSubmit(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.ConstructionSubmitRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.TransactionIdentifierResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConstructionAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := c.ConstructionSubmit(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConstructionSubmit() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ConstructionSubmit() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewConstructionAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		node    *api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.ConstructionAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConstructionAPIService(tt.args.network, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConstructionAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
