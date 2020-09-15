package services

var (
	//Versions info to be injected on build time
	RosettaSDKVersion = "Unknown"
	LotusVersion      = "Unknown"
	GitRevision       = "Unknown"
	///

	//Other configs
	RetryConnectAttempts = "1000000"
	///
)

const (
	//Network
	BlockChainName    = "Filecoin"
	RosettaServerPort = 8080
	///

	//Currency
	CurrencySymbol   = "FIL"
	CurrencyDecimals = 18
	///

	//Operation status
	OperationStatusOk     = "Ok"
	OperationStatusFailed = "Fail"
	///
)

//Supported operations
var SupportedOperations = map[string]bool{"Propose": true, "Approve": true}
