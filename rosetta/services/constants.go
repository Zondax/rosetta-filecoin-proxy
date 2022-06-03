package services

var (
	// Versions info to be injected on build time
	RosettaSDKVersion = "Unknown"
	LotusVersion      = "Unknown"
	GitRevision       = "Unknown"

	// ServerPort to be injected on build time
	RosettaServerPort = "8080"

	// Other configs
	RetryConnectAttempts = "1000000"
)

const (
	// Network
	BlockChainName = "Filecoin"

	// Currency
	CurrencySymbol   = "FIL"
	CurrencyDecimals = 18

	// Operation status
	OperationStatusOk     = "Ok"
	OperationStatusFailed = "Fail"

	// Account
	LockedBalanceStr         = "LockedBalance"
	SpendableBalanceStr      = "SpendableBalance"
	VestingScheduleStr       = "VestingSchedule"
	LockedFundsKey           = "LockedFunds"
	VestingStartEpochKey     = "StartEpoch"
	VestingUnlockDurationKey = "UnlockDuration"
	VestingInitialBalanceKey = "InitialBalance"

	// Misc
	ProxyLoggerName = "rosetta-filecoin-proxy"

	// Actors names
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

// Supported operations
var SupportedOperations = map[string]bool{
	"Send":                true, // Common
	"Fee":                 true, // Common
	"Exec":                true, // MethodsInit
	"SwapSigner":          true, // MethodsMultisig
	"Propose":             true, // MethodsMultisig
	"AwardBlockReward":    true, // MethodsReward
	"OnDeferredCronEvent": true, // MethodsMiner
	"PreCommitSector":     true, // MethodsMiner
	"ProveCommitSector":   true, // MethodsMiner
	"SubmitWindowedPoSt":  true, // MethodsMiner
	"ApplyRewards":        true, // MethodsMiner
	"AddBalance":          true, // MethodsMarket
	"RepayDebt":           true, // MethodsMiner
}
