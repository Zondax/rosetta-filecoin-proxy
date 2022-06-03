package actors

import (
	"fmt"
	builtin "github.com/filecoin-project/lotus/chain/actors/builtin"
	"github.com/ipfs/go-cid"
	"strings"
)

const (
	// BuiltinActorsKeys NetworkVersion: 16, ActorsVersion: 8
	// from lotus cli cmd: 'lotus state actor-cids'
	ActorAccountCode          = "bafk2bzacebs3prrp2swegbefkh3hsyuqwvxrnluoiwpmkxzhfh6y4wecdxwv4"
	ActorCronCode             = "bafk2bzacedifvgycuibukwnaesekwdxiqpdt5m25ga7vkh5mgrngskbxtputu"
	ActorInitCode             = "bafk2bzaceaejm2x4jwqownf5jyxjbga4pwim7d7lw6yfxdvgyuetybrwzc7tu"
	ActorStorageMarketCode    = "bafk2bzaceafh7p4wdafrplys6ejimgf66apaaz4f4iuu3rsk3beqfnzzbrras"
	ActorStorageMinerCode     = "bafk2bzaceaq7g4zded65xa5oxwlwx75brh5gxcjthtdu5zl3ei5vtvbnfavzy"
	ActorMultisigCode         = "bafk2bzaceczfz65fvn662qrkdgtmokve7oj3wmdgkbhvucwaigyihues3u6ke"
	ActorPaymentChannelCode   = "bafk2bzaceaguosntcgqhbd5cknw6xe5fa6qxjxarft6osxmuh6ju2y4zrxmsi"
	ActorStoragePowerCode     = "bafk2bzacean7gm2tjoq4hsvsimdx2clvfue6yfls2hcee6sxvm4ld7l4e42jk"
	ActorRewardCode           = "bafk2bzacednljsae765kb6dsgfcg4jebfjza5mnqbauemzpxiizguuyi4i3yi"
	ActorSystemCode           = "bafk2bzacecr3qaggetreqfeurdjek7kjzfeuapiuprwp26hixmjtdvbgb3yfk"
	ActorVerifiedRegistryCode = "bafk2bzaceceekx5csbzck4lq5rlqexq2tu3njbixyglw452hc554aabpxiaai"

	ActorAccountName          = "account"
	ActorCronName             = "cron"
	ActorInitName             = "init"
	ActorStorageMarketName    = "storagemarket"
	ActorStorageMinerName     = "storageminer"
	ActorMultisigName         = "multisig"
	ActorPaymentChannelName   = "paymentchannel"
	ActorStoragePowerName     = "storagepower"
	ActorRewardName           = "reward"
	ActorSystemName           = "system"
	ActorVerifiedRegistryName = "verifiedregistry"
)

var (
	AccountActorCodeID, _          = cid.Parse(ActorAccountCode)
	CronActorCodeID, _             = cid.Parse(ActorCronCode)
	InitActorCodeID, _             = cid.Parse(ActorInitCode)
	StorageMarketActorCodeID, _    = cid.Parse(ActorStorageMarketCode)
	StorageMinerActorCodeID, _     = cid.Parse(ActorStorageMinerCode)
	MultisigActorCodeID, _         = cid.Parse(ActorMultisigCode)
	PaymentChannelActorCodeID, _   = cid.Parse(ActorPaymentChannelCode)
	StoragePowerActorCodeID, _     = cid.Parse(ActorStoragePowerCode)
	RewardActorCodeID, _           = cid.Parse(ActorRewardCode)
	SystemActorCodeID, _           = cid.Parse(ActorSystemCode)
	VerifiedRegistryActorCodeID, _ = cid.Parse(ActorVerifiedRegistryCode)
)

var BuiltinActorsKeys = map[string]string{
	ActorAccountCode:          ActorAccountName,
	ActorCronCode:             ActorCronName,
	ActorInitCode:             ActorInitName,
	ActorStorageMarketCode:    ActorStorageMarketName,
	ActorStorageMinerCode:     ActorStorageMinerName,
	ActorMultisigCode:         ActorMultisigName,
	ActorPaymentChannelCode:   ActorPaymentChannelName,
	ActorStoragePowerCode:     ActorStoragePowerName,
	ActorRewardCode:           ActorRewardName,
	ActorSystemCode:           ActorSystemName,
	ActorVerifiedRegistryCode: ActorVerifiedRegistryName,
}

func IsMultisigActor(actorCode cid.Cid) bool {
	// Valid from V8 and on
	if BuiltinActorsKeys[actorCode.String()] == ActorMultisigName {
		return true
	}

	// Check for older actors versions
	if builtin.IsMultisigActor(actorCode) {
		return true
	}

	return false
}

func GetActorNameFromCid(actorCode cid.Cid) string {
	// Valid from V8 and on
	actorName, ok := BuiltinActorsKeys[actorCode.String()]
	if ok {
		return actorName
	}

	// Check for older actors versions ["fil/<version>/<actorName>"]
	actorName = builtin.ActorNameByCode(actorCode)
	if actorName == "<unknown>" {
		fmt.Println("Warning: invalid actor code CID:", actorCode.String())
		return actorName
	}

	actorNameArr := strings.Split(actorName, "/")
	actorName = actorNameArr[len(actorNameArr)-1]

	return actorName
}
