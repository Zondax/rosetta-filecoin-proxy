// +build rosetta_rpc

package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"reflect"
	"testing"
)

func TestBuildTipSetKeyHash(t *testing.T) {
	type args struct {
		key filTypes.TipSetKey
	}
	tests := []struct {
		name    string
		args    args
		want    *string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildTipSetKeyHash(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTipSetKeyHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildTipSetKeyHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateNetworkId(t *testing.T) {
	type args struct {
		ctx       context.Context
		node      *api.FullNode
		networkId *types.NetworkIdentifier
	}
	tests := []struct {
		name string
		args args
		want *types.Error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateNetworkId(tt.args.ctx, tt.args.node, tt.args.networkId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateNetworkId() = %v, want %v", got, tt.want)
			}
		})
	}
}
