package services

var (
	//Versions info to be injected on build time
	RosettaSDKVersion = "Unknown"
	LotusVersion      = "Unknown"
	GitRevision       = "Unknown"
	///
)

const (
	//Network related
	BlockChainName    = "Filecoin"
	RosettaServerPort = 8080
	///

	//Currency related
	CurrencySymbol   = "FIL"
	CurrencyDecimals = 18
	///

	//Operation related
	OperationStatusOk     = "Ok"
	OperationStatusFailed = "Fail"
	///
)
