package services

var (
	// Versions info to be injected on build time
	RosettaSDKVersion = "Unknown"
	LotusVersion      = "Unknown"
	GitRevision       = "Unknown"

	// ServerPort to be injected on build time
	RosettaServerPort = "8083"

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

// BuiltinActorsKeys NetworkVersion: 16, ActorsVersion: 8
// from cli cmd: 'lotus state actor-cids'
var BuiltinActorsKeys = map[string]string{
	"bafk2bzacebs3prrp2swegbefkh3hsyuqwvxrnluoiwpmkxzhfh6y4wecdxwv4": "account",
	"bafk2bzacedifvgycuibukwnaesekwdxiqpdt5m25ga7vkh5mgrngskbxtputu": "cron",
	"bafk2bzaceaejm2x4jwqownf5jyxjbga4pwim7d7lw6yfxdvgyuetybrwzc7tu": "init",
	"bafk2bzaceafh7p4wdafrplys6ejimgf66apaaz4f4iuu3rsk3beqfnzzbrras": "storagemarket",
	"bafk2bzaceaq7g4zded65xa5oxwlwx75brh5gxcjthtdu5zl3ei5vtvbnfavzy": "storageminer",
	"bafk2bzaceczfz65fvn662qrkdgtmokve7oj3wmdgkbhvucwaigyihues3u6ke": "multisig",
	"bafk2bzaceaguosntcgqhbd5cknw6xe5fa6qxjxarft6osxmuh6ju2y4zrxmsi": "paymentchannel",
	"bafk2bzacean7gm2tjoq4hsvsimdx2clvfue6yfls2hcee6sxvm4ld7l4e42jk": "storagepower",
	"bafk2bzacednljsae765kb6dsgfcg4jebfjza5mnqbauemzpxiizguuyi4i3yi": "reward",
	"bafk2bzacecr3qaggetreqfeurdjek7kjzfeuapiuprwp26hixmjtdvbgb3yfk": "system",
	"bafk2bzaceceekx5csbzck4lq5rlqexq2tu3njbixyglw452hc554aabpxiaai": "verifiedregistry",
}
