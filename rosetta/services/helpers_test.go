package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/zondax/rosetta-filecoin-lib/actors"
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

func TestResolveMethodName(t *testing.T) {
	// invokeEVMMethod is the FRC-42 method number for "InvokeEVM"
	// (builtin.MethodsEVM.InvokeContract). Ethereum-originated messages carry
	// this method number regardless of the recipient actor type.
	const invokeEVMMethod = abi.MethodNum(3844450837)

	tests := []struct {
		name      string
		actorName string
		method    abi.MethodNum
		want      string
	}{
		{
			name:      "EVM actor InvokeEVM resolves to InvokeContract",
			actorName: "evm",
			method:    invokeEVMMethod,
			want:      "InvokeContract",
		},
		{
			name:      "account actor InvokeEVM falls back to Send",
			actorName: ACCOUNT_ACTOR_NAME,
			method:    invokeEVMMethod,
			want:      METHOD_FALLBACK,
		},
		{
			name:      "ethaccount actor InvokeEVM falls back to Send",
			actorName: ETHACCOUNT_ACTOR_NAME,
			method:    invokeEVMMethod,
			want:      METHOD_FALLBACK,
		},
		{
			name:      "account actor known method resolves by name",
			actorName: ACCOUNT_ACTOR_NAME,
			method:    abi.MethodNum(2), // PubkeyAddress
			want:      "PubkeyAddress",
		},
		{
			name:      "account actor non-exported unknown method stays unknown",
			actorName: ACCOUNT_ACTOR_NAME,
			method:    abi.MethodNum(99),
			want:      actors.UnknownStr,
		},
		{
			name:      "placeholder actor InvokeEVM is not covered by fallback (documents current gap)",
			actorName: "placeholder",
			method:    invokeEVMMethod,
			want:      actors.UnknownStr,
		},
		{
			name:      "unknown actor stays unknown",
			actorName: "not-a-real-actor",
			method:    invokeEVMMethod,
			want:      actors.UnknownStr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveMethodName(tt.actorName, tt.method); got != tt.want {
				t.Errorf("resolveMethodName(%q, %d) = %q, want %q", tt.actorName, tt.method, got, tt.want)
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
