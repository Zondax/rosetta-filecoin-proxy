package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/v2api"
	"reflect"
	"testing"
)

func TestMemPoolAPIService_Mempool(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		v1Node  api.FullNode
		v2Node  v2api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.NetworkRequest
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
				v1Node:  tt.fields.v1Node,
				v2Node:  tt.fields.v2Node,
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
		v1Node  api.FullNode
		v2Node  v2api.FullNode
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
				v1Node:  tt.fields.v1Node,
				v2Node:  tt.fields.v2Node,
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
		v1API   *api.FullNode
		v2API   v2api.FullNode
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
			if got := NewMemPoolAPIService(tt.args.network, tt.args.v1API, tt.args.v2API, rosettaLib); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemPoolAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
