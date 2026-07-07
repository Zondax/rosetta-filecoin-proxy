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
			name:      "EVM actor InvokeContractDelegate resolves by name",
			actorName: "evm",
			method:    abi.MethodNum(6),
			want:      "InvokeContractDelegate",
		},
		{
			name:      "multisig known method resolves by name",
			actorName: "multisig",
			method:    abi.MethodNum(2),
			want:      "Propose",
		},
		{
			name:      "init known method resolves by name",
			actorName: "init",
			method:    abi.MethodNum(2),
			want:      "Exec",
		},
		{
			name:      "account exported method at threshold falls back to Send",
			actorName: ACCOUNT_ACTOR_NAME,
			method:    abi.MethodNum(FIRST_EXPORTED_METHOD_NUMBER),
			want:      METHOD_FALLBACK,
		},
		{
			name:      "account method just below exported threshold stays unknown",
			actorName: ACCOUNT_ACTOR_NAME,
			method:    abi.MethodNum(FIRST_EXPORTED_METHOD_NUMBER - 1),
			want:      actors.UnknownStr,
		},
		{
			name:      "account fallback matching is case-insensitive",
			actorName: "Account",
			method:    invokeEVMMethod,
			want:      METHOD_FALLBACK,
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

func TestGetMethodNameShortcuts(t *testing.T) {
	// These branches short-circuit before any actor/lib resolution, so they can
	// be exercised with a nil lib.
	t.Run("nil message returns an error", func(t *testing.T) {
		if name, err := GetMethodName(nil, nil); err == nil {
			t.Fatalf("expected an error for a nil message, got name %q", name)
		}
	})

	t.Run("method 0 resolves to Send", func(t *testing.T) {
		name, err := GetMethodName(&filTypes.MessageTrace{Method: abi.MethodNum(0)}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != METHOD_SEND {
			t.Errorf("GetMethodName(method=0) = %q, want %q", name, METHOD_SEND)
		}
	})

	t.Run("method 1 resolves to Constructor", func(t *testing.T) {
		name, err := GetMethodName(&filTypes.MessageTrace{Method: abi.MethodNum(1)}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "Constructor" {
			t.Errorf("GetMethodName(method=1) = %q, want %q", name, "Constructor")
		}
	})
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
