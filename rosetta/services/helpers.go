// +build rosetta_rpc

package services

import (
	"context"
	"encoding/hex"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"reflect"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func BuildTipSetKeyHash(key filTypes.TipSetKey) (*string, error) {

	cidBuilder := cid.V1Builder{Codec: cid.DagCBOR, MhType: multihash.BLAKE2B_MIN + 31}
	tipSetKeyHash, err := cidBuilder.Sum(key.Bytes())
	if err != nil {
		return nil, err
	}

	outStr := hex.EncodeToString(tipSetKeyHash.Bytes())

	return &outStr, nil
}

func ValidateNetworkId(ctx context.Context, node *api.FullNode, networkId *types.NetworkIdentifier) *types.Error {

	if networkId == nil {
		return ErrMalformedValue
	}

	fullAPI:= *node
	validNetwork, err := fullAPI.StateNetworkName(ctx)
	if err != nil {
		return ErrUnableToRetrieveNetworkName
	}

	if networkId.Network != string(validNetwork) {
		return ErrInvalidNetwork
	}

	return nil
}

func GetCurrencyData() *types.Currency {
	//TODO get this from external config file
	return &types.Currency{
		Symbol:   "FIL",
		Decimals: 18,
		Metadata: nil,
	}
}

func GetMethodName(msg *api.Message, api *api.FullNode) (string, error) {
	actor, err := (*api).StateGetActor(context.Background(), msg.Message.From, filTypes.EmptyTSK)
	if err != nil {
		return "", err
	}
	var method interface{}
	switch actor.Code {
	case builtin.InitActorCodeID:
		method = builtin.MethodsInit
	case builtin.CronActorCodeID:
		method = builtin.MethodsCron
	case builtin.AccountActorCodeID:
		method = builtin.MethodsAccount
	case builtin.StoragePowerActorCodeID:
		method = builtin.MethodsPower
	case builtin.StorageMinerActorCodeID:
		method = builtin.MethodsMiner
	case builtin.StorageMarketActorCodeID:
		method = builtin.MethodsMarket
	case builtin.PaymentChannelActorCodeID:
		method = builtin.MethodsPaych
	case builtin.MultisigActorCodeID:
		method = builtin.MethodsMultisig
	case builtin.RewardActorCodeID:
		method = builtin.MethodsReward
	case builtin.VerifiedRegistryActorCodeID:
		method = builtin.MethodsVerifiedRegistry
	default:
		return "", nil
	}

	val := reflect.Indirect(reflect.ValueOf(method))
	methodName := val.Type().Field(int(msg.Message.Method)).Name
	return methodName, nil
}