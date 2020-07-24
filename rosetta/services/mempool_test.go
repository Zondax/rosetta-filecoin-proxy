// +build rosetta_rpc

package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestMemPoolAPIService_Mempool(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.MempoolRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.MempoolResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemPoolAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := m.Mempool(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Mempool() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Mempool() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMemPoolAPIService_MempoolTransaction(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.MempoolTransactionRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.MempoolTransactionResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MemPoolAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := m.MempoolTransaction(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MempoolTransaction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MempoolTransaction() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewMemPoolAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		api     *api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.MempoolAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemPoolAPIService(tt.args.network, tt.args.api); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemPoolAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
