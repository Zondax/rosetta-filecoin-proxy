// +build rosetta_rpc

package services

import "github.com/coinbase/rosetta-sdk-go/types"

var (
	ErrUnableToGetChainID = &types.Error{
		Code:      1,
		Message:   "unable to get chain ID",
		Retriable: true,
	}

	ErrInvalidBlockchain = &types.Error{
		Code:      2,
		Message:   "invalid blockchain specified in network identifier",
		Retriable: false,
	}

	ErrInvalidSubnetwork = &types.Error{
		Code:      3,
		Message:   "invalid sub-network identifier",
		Retriable: false,
	}

	ErrInvalidNetwork = &types.Error{
		Code:      4,
		Message:   "invalid network specified in network identifier",
		Retriable: false,
	}

	ErrMissingNID = &types.Error{
		Code:      5,
		Message:   "network identifier is missing",
		Retriable: false,
	}

	ErrUnableToGetLatestBlk = &types.Error{
		Code:      6,
		Message:   "unable to get latest block",
		Retriable: true,
	}

	ErrUnableToGetGenesisBlk = &types.Error{
		Code:      7,
		Message:   "unable to get genesis block",
		Retriable: true,
	}

	ErrUnableToGetAccount = &types.Error{
		Code:      8,
		Message:   "unable to get account",
		Retriable: true,
	}

	ErrInsufficientQueryInputs = &types.Error{
		Code:      9,
		Message:   "query inputs insufficient",
		Retriable: false,
	}

	ErrInvalidAccountAddress = &types.Error{
		Code:      10,
		Message:   "invalid account address",
		Retriable: false,
	}

	ErrMustSpecifySubAccount = &types.Error{
		Code:      11,
		Message:   "a valid subaccount must be specified ('general' or 'escrow')",
		Retriable: false,
	}

	ErrUnableToGetBlk = &types.Error{
		Code:      12,
		Message:   "unable to get block",
		Retriable: true,
	}

	ErrNotImplemented = &types.Error{
		Code:      13,
		Message:   "operation not implemented",
		Retriable: false,
	}

	ErrUnableToGetTxns = &types.Error{
		Code:      14,
		Message:   "unable to get transactions",
		Retriable: true,
	}

	ErrUnableToSubmitTx = &types.Error{
		Code:      15,
		Message:   "unable to submit transaction",
		Retriable: false,
	}

	ErrUnableToGetNextNonce = &types.Error{
		Code:      16,
		Message:   "unable to get next nonce",
		Retriable: true,
	}

	ErrMalformedValue = &types.Error{
		Code:      17,
		Message:   "malformed value",
		Retriable: false,
	}

	ErrUnableToGetNodeStatus = &types.Error{
		Code:      18,
		Message:   "unable to get node status",
		Retriable: true,
	}

	ErrUnableToGetTipsetCID = &types.Error{
		Code:      19,
		Message:   "unable to get tipset CID",
		Retriable: true,
	}

	ErrUnableToGetPeers = &types.Error{
		Code:      20,
		Message:   "unable to get peer list",
		Retriable: true,
	}

	ErrUnableToGetWalletBalance = &types.Error{
		Code:      21,
		Message:   "unable to get wallet balance for address",
		Retriable: true,
	}

	ErrUnableToGetTipset = &types.Error{
		Code:      22,
		Message:   "unable to get tipset",
		Retriable: true,
	}

	ErrUnableToGetParentBlk = &types.Error{
		Code:      23,
		Message:   "unable to get parent block",
		Retriable: true,
	}

	ErrUnableToGetNodeInfo = &types.Error{
		Code:      24,
		Message:   "unable to get node information",
		Retriable: true,
	}

	ErrUnableToGetSyncStatus= &types.Error{
		Code:      25,
		Message:   "unable to get sync status",
		Retriable: true,
	}

	ErrUnableToGetUnsyncedBlock= &types.Error{
		Code:      26,
		Message:   "requested block not yet synchronized",
		Retriable: true,
	}

	ErrSyncErrored = &types.Error{
		Code:      27,
		Message:   "error on node sync process",
		Retriable: true,
	}

	ErrUnableToBuildTipSetHash = &types.Error{
		Code:      28,
		Message:   "error on creating TipSetKey hash",
		Retriable: true,
	}

	ErrUnableToRetrieveNetworkName = &types.Error{
		Code:      29,
		Message:   "error when querying network name",
		Retriable: false,
	}

	ErrMalformedTx = &types.Error{
		Code:      30,
		Message:   "malformed transaction",
		Retriable: false,
	}

	ErrInvalidHash = &types.Error{
		Code:      31,
		Message:   "hash does not match with provided block index",
		Retriable: false,
	}

	ErrUnableToGetTxnReceipt = &types.Error{
		Code:      32,
		Message:   "unable to get transaction receipt",
		Retriable: true,
	}

	ErrMsgsAndReceiptsCountMismatch = &types.Error{
		Code:      33,
		Message:   "retrieved Messages count don't match with Receipts count",
		Retriable: false,
	}

	ErrUnableToEstimateGasPrice = &types.Error{
		Code:      34,
		Message:   "unable to estimate gas price",
		Retriable: false,
	}

	ErrInsufficientBalanceForGas = &types.Error{
		Code:      35,
		Message:   "insufficient balance for gas",
		Retriable: false,
	}

	ErrLotusCallTimedOut = &types.Error{
		Code:      36,
		Message:   "Lotus RPC call timed out",
		Retriable: true,
	}

	ErrorList = []*types.Error{
		ErrUnableToGetChainID,
		ErrInvalidBlockchain,
		ErrInvalidSubnetwork,
		ErrInvalidNetwork,
		ErrUnableToRetrieveNetworkName,
		ErrMissingNID,
		ErrUnableToGetLatestBlk,
		ErrUnableToGetGenesisBlk,
		ErrUnableToGetAccount,
		ErrInsufficientQueryInputs,
		ErrInvalidAccountAddress,
		ErrMustSpecifySubAccount,
		ErrUnableToGetBlk,
		ErrNotImplemented,
		ErrUnableToGetTxns,
		ErrUnableToSubmitTx,
		ErrUnableToGetNextNonce,
		ErrMalformedValue,
		ErrUnableToGetNodeStatus,
		ErrUnableToGetTipsetCID,
		ErrUnableToGetPeers,
		ErrUnableToGetWalletBalance,
		ErrUnableToGetTipset,
		ErrUnableToGetParentBlk,
		ErrUnableToGetNodeInfo,
		ErrUnableToGetSyncStatus,
		ErrUnableToGetUnsyncedBlock,
		ErrSyncErrored,
		ErrUnableToBuildTipSetHash,
		ErrMalformedTx,
		ErrInvalidHash,
		ErrUnableToGetTxnReceipt,
		ErrMsgsAndReceiptsCountMismatch,
		ErrUnableToEstimateGasPrice,
		ErrInsufficientBalanceForGas,
		ErrLotusCallTimedOut,
	}
)
