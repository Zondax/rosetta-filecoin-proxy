package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestBlockAPIService_Block(t *testing.T) {
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
		// TODO: Add test cases.
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
