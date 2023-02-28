package services

import (
	"context"
	"encoding/hex"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/builtin"
	rosettaFilecoinLib "github.com/zondax/rosetta-filecoin-lib"
	"github.com/zondax/rosetta-filecoin-lib/actors"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
	"reflect"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

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
		return BuildError(ErrUnableToRetrieveNetworkName, err, true)
	}

	if networkId.Network != string(validNetwork) {
		return BuildError(ErrInvalidNetwork, nil, true)
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

func GetActorNameFromAddress(address address.Address, lib *rosettaFilecoinLib.RosettaConstructionFilecoin) string {
	var actorCode cid.Cid
	// Search for actor in cache
	var err error
	actorCode, err = tools.ActorsDB.GetActorCode(address)
	if err != nil {
		return actors.UnknownStr
	}

	actorName, err := lib.BuiltinActors.GetActorNameFromCid(actorCode)
	if err != nil {
		return actors.UnknownStr
	}

	return actorName
}

func GetMethodName(msg *filTypes.Message, lib *rosettaFilecoinLib.RosettaConstructionFilecoin) (string, *types.Error) {
	if msg == nil {
		return "", BuildError(ErrMalformedValue, nil, true)
	}

	// Shortcut 1 - Method "0" corresponds to "MethodSend"
	if msg.Method == 0 {
		return "Send", nil
	}

	// Shortcut 2 - Method "1" corresponds to "MethodConstructor"
	if msg.Method == 1 {
		return "Constructor", nil
	}

	actorName := GetActorNameFromAddress(msg.To, lib)

	var method interface{}
	switch actorName {
	case "init":
		method = builtin.MethodsInit
	case "cron":
		method = builtin.MethodsCron
	case "account":
		method = builtin.MethodsAccount
	case "storagepower":
		method = builtin.MethodsPower
	case "storageminer":
		method = builtin.MethodsMiner
	case "storagemarket":
		method = builtin.MethodsMarket
	case "paymentchannel":
		method = builtin.MethodsPaych
	case "multisig":
		method = builtin.MethodsMultisig
	case "reward":
		method = builtin.MethodsReward
	case "verifiedregistry":
		method = builtin.MethodsVerifiedRegistry
	case "evm":
		method = builtin.MethodsEVM
	case "eam":
		method = builtin.MethodsEAM
	case "datacap":
		method = builtin.MethodsDatacap
	case "placeholder":
		method = builtin.MethodsPlaceholder
	case "ethaccount":
		method = builtin.MethodsEthAccount
	default:
		return actors.UnknownStr, nil
	}

	val := reflect.Indirect(reflect.ValueOf(method))

	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Field(i)
		methodNum := field.Uint()
		if methodNum == uint64(msg.Method) {
			methodName := val.Type().Field(i).Name
			return methodName, nil
		}
	}

	return actors.UnknownStr, nil
}

func GetActorPubKey(add address.Address, lib *rosettaFilecoinLib.RosettaConstructionFilecoin) (string, *types.Error) {

	actorCode, err := tools.ActorsDB.GetActorCode(add)
	if err != nil {
		Logger.Error("could not get actor code from address. Err:", err.Error())
		return add.String(), nil
	}

	// Search for actor's pubkey in cache.
	// If cannot get actor's pubkey, GetActorPubKey will return the same address

	// Handler for msig
	if lib.BuiltinActors.IsActor(actorCode, actors.ActorMultisigName) {
		return getPubKeyForMsig(add)
	}

	// Handler for storage miner
	if lib.BuiltinActors.IsActor(actorCode, actors.ActorStorageMinerName) {
		return getPubKeyForStorageMiner(add)
	}

	// For other types, try to return address in "robust" format
	pubKey, err := tools.ActorsDB.GetActorPubKey(add, false)
	if err != nil {
		pubKey = add.String()
	}

	return pubKey, nil
}

func getPubKeyForMsig(add address.Address) (string, *types.Error) {

	var (
		pubKey string
		err    error
	)

	switch add.Protocol() {
	case address.BLS, address.SECP256K1, address.Actor:
		// Use "short" address for msig actors since can be mixed on the blockchain
		// and we need them to be normalized to any of the two formats
		pubKey, err = tools.ActorsDB.GetActorPubKey(add, true)
		if err != nil {
			pubKey = add.String()
		}
	case address.ID:
		pubKey = add.String()
	default:
		// Unknown address type
		pubKey = add.String()
	}

	return pubKey, nil
}

func getPubKeyForStorageMiner(add address.Address) (string, *types.Error) {

	var (
		pubKey string
		err    error
	)

	switch add.Protocol() {
	case address.BLS, address.SECP256K1, address.Actor:
		// Use "short" address for storage miners actors since can be mixed on the blockchain
		// and we need them to be normalized to any of the two formats
		pubKey, err = tools.ActorsDB.GetActorPubKey(add, true)
		if err != nil {
			pubKey = add.String()
		}
	case address.ID:
		pubKey = add.String()
	default:
		// Unknown address type
		pubKey = add.String()
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
