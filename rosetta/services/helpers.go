package services

import (
	"context"
	"encoding/hex"
	"github.com/filecoin-project/go-address"
	builtin2 "github.com/filecoin-project/lotus/chain/actors/builtin"
	methods "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
	"reflect"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

const unknownStr = "Unknown"

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Logger.Info(name, " took ", elapsed)
}

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
		return BuildError(ErrUnableToRetrieveNetworkName, err)
	}

	if networkId.Network != string(validNetwork) {
		return BuildError(ErrInvalidNetwork, nil)
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

	if msg == nil {
		return "", BuildError(ErrMalformedValue, nil)
	}

	// Shortcut 1 - Method "0" corresponds to "MethodSend"
	if msg.Method == 0 {
		return "Send", nil
	}

	// Shortcut 2 - Method "1" corresponds to "MethodConstructor"
	if msg.Method == 1 {
		return "Constructor", nil
	}

	var actorCode cid.Cid
	// Search for actor in cache
	var err error
	actorCode, err = tools.ActorsDB.GetActorCode(msg.To)
	if err != nil {
		return unknownStr, nil
	}

	actorNameArr := strings.Split(builtin2.ActorNameByCode(actorCode), "/")
	actorName := actorNameArr[len(actorNameArr)-1]

	var method interface{}
	switch actorName {
	case "init":
		method = methods.MethodsInit
	case "cron":
		method = methods.MethodsCron
	case "account":
		method = methods.MethodsAccount
	case "storagepower":
		method = methods.MethodsPower
	case "storageminer":
		method = methods.MethodsMiner
	case "storagemarket":
		method = methods.MethodsMarket
	case "paymentchannel":
		method = methods.MethodsPaych
	case "multisig":
		method = methods.MethodsMultisig
	case "reward":
		method = methods.MethodsReward
	case "verifiedregistry":
		method = methods.MethodsVerifiedRegistry
	default:
		return unknownStr, nil
	}

	val := reflect.Indirect(reflect.ValueOf(method))
	idx := int(msg.Method)
	if idx > 0 {
		idx--
	}

	if val.Type().NumField() < idx {
		return unknownStr, nil
	}

	methodName := val.Type().Field(idx).Name
	return methodName, nil
}

func GetActorPubKey(add address.Address) (string, *types.Error) {
	var pubKey string
	switch add.Protocol() {
	case address.BLS, address.SECP256K1, address.Actor:
		pubKey = add.String()
	default:
		// Search for actor's pubkey in cache.
		// If cannot get actor's pubkey, GetActorPubKey will return the same address
		var err error
		pubKey, err = tools.ActorsDB.GetActorPubKey(add)
		if err != nil {
			return add.String(), nil
		}
	}

	return pubKey, nil
}

func GetSupportedOpList() []string {
	operations := make([]string, 0, len(SupportedOperations))
	for op := range SupportedOperations {
		operations = append(operations, op)
	}

	return operations
}

func IsOpSupported(op string) bool {
	supported, ok := SupportedOperations[op]
	if ok && supported {
		return true
	}

	return false
}
