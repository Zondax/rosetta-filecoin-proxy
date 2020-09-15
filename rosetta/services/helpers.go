package services

import (
	"context"
	"encoding/hex"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
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

	fullAPI := *node
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
	return &types.Currency{
		Symbol:   CurrencySymbol,
		Decimals: CurrencyDecimals,
		Metadata: nil,
	}
}

func GetMethodName(msg *filTypes.Message) (string, *types.Error) {

	var (
		actorCode cid.Cid
		skipDB    bool
	)

	//Shortcut 1 - t1 and t3 address are always account actors
	if len(msg.To.String()) > 2 {
		addPrefix := msg.To.String()[0:2]
		if addPrefix == "t1" || addPrefix == "t3" {
			actorCode = builtin.AccountActorCodeID
			skipDB = true
		}
	}

	// Search for actor in cache
	if !skipDB {
		var err error
		actorCode, err = tools.ActorsDB.GetActorCode(msg.To)
		if err != nil {
			return "Unknown", nil
		}
	}

	//Method "0" corresponds to "MethodSend"
	if msg.Method == 0 {
		return "Send", nil
	}

	var method interface{}
	switch actorCode {
	case builtin.InitActorCodeID:
		method = builtin.MethodsInit
	case builtin.CronActorCodeID:
		method = builtin.MethodsCron
	case builtin.AccountActorCodeID:
		method = builtin.MethodsMultisig
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
		return "Unknown", nil
	}

	val := reflect.Indirect(reflect.ValueOf(method))
	idx := int(msg.Method)
	if idx > 0 {
		idx--
	}

	methodName := val.Type().Field(idx).Name
	return methodName, nil
}
