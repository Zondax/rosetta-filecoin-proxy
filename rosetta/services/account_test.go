package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"reflect"
	"testing"
)

func TestAccountAPIService_AccountBalance(t *testing.T) {
	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.AccountBalanceRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.AccountBalanceResponse
		want1  *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AccountAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := a.AccountBalance(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalance() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AccountBalance() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewAccountAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		node    *api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.AccountAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountAPIService(tt.args.network, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}
