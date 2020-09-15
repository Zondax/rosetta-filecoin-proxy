package services

import (
	"context"
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"strconv"
)

// ChainIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the current chain id
const ChainIDKey = "chainID"

// OptionsSenderIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the sender's account ID
const OptionsSenderIDKey = "idSender"

// OptionsReceiverIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the receiver's account ID
const OptionsReceiverIDKey = "idReceiver"

// OptionsBlockInclKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse determines on how many epochs message should included
// being 0 the fastest (and the most gas expensive one)
const OptionsBlockInclKey = "blockIncl"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "nonce"

// GasPremiumKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies tx's gas premium
const GasPremiumKey = "gasPremium"

// GasLimitKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies tx's gas limit
const GasLimitKey = "gasLimit"

// GasFeeCapKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies gas fee cap
const GasFeeCapKey = "gasFeeCap"

// ConstructionAPIService implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	network *types.NetworkIdentifier
	node    api.FullNode
}

// NewConstructionAPIService creates a new instance of an ConstructionAPIService.
func NewConstructionAPIService(network *types.NetworkIdentifier, node *api.FullNode) server.ConstructionAPIServicer {
	return &ConstructionAPIService{
		network: network,
		node:    *node,
	}
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (c *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	var (
		addressSenderParsed   address.Address
		addressReceiverParsed address.Address
		message               = &filTypes.Message{GasLimit: 0, GasFeeCap: filTypes.NewInt(0), GasPremium: filTypes.NewInt(0)}
		availableFunds        filTypes.BigInt
		err                   error
		nonce                 uint64
		blockInclUint         uint64 = 1
	)

	errNet := ValidateNetworkId(ctx, &c.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	md := make(map[string]interface{})

	if request.Options != nil {
		//Parse block include epochs - this field is optional
		blockIncl, ok := request.Options[OptionsBlockInclKey]
		if ok {
			blockInclUint = uint64(blockIncl.(float64))
		}

		//Parse sender address - this field is optional
		addressSenderRaw, okSender := request.Options[OptionsSenderIDKey]
		//Parse receiver address - this field is optional
		addressReceiverRaw, okReceiver := request.Options[OptionsReceiverIDKey]

		if okSender {
			addressSenderParsed, err = address.NewFromString(addressSenderRaw.(string))
			if err != nil {
				return nil, ErrInvalidAccountAddress
			}
			message.From = addressSenderParsed
		}

		//Parse receiver address - this field is optional
		addressReceiverRaw, okReceiver := request.Options[OptionsReceiverIDKey]
		if okReceiver {
			addressReceiverParsed, err = address.NewFromString(addressReceiverRaw.(string))
			if err != nil {
				return nil, ErrInvalidAccountAddress
			}
			message.To = addressReceiverParsed
		}

		if okSender {
			nonce, err = c.node.MpoolGetNonce(ctx, addressSenderParsed)
			if err != nil {
				return nil, ErrUnableToGetNextNonce
			}
			md[NonceKey] = nonce

			//Get available balance
			actor, errAct := c.node.StateGetActor(context.Background(), addressSenderParsed, filTypes.EmptyTSK)
			if errAct != nil {
				return nil, ErrUnableToGetActor
			}

			if actor.Code == builtin.MultisigActorCodeID {
				//Get the unlocked funds of the multisig account
				availableFunds, err = c.node.MsigGetAvailableBalance(ctx, addressSenderParsed, filTypes.EmptyTSK)
				if err != nil {
					return nil, ErrUnableToGetBalance
				}
			} else {
				availableFunds = actor.Balance
			}

			// GasEstimateMessageGas to get a safely overestimated value for gas limit
			message, err = c.node.GasEstimateMessageGas(ctx, message,
				&api.MessageSendSpec{MaxFee: filTypes.NewInt(build.BlockGasLimit)}, filTypes.TipSetKey{})
			if err != nil {
				return nil, ErrUnableToEstimateGasLimit
			}

			// GasEstimateGasPremium
			gasPremium, gasErr := c.node.GasEstimateGasPremium(ctx, blockInclUint, addressSenderParsed, message.GasLimit, filTypes.TipSetKey{})
			if gasErr != nil {
				return nil, ErrUnableToEstimateGasPremium
			}
			message.GasPremium = gasPremium

			// GasEstimateFeeCap requires gasPremium to be set on message
			gasFeeCap, gasErr := c.node.GasEstimateFeeCap(ctx, message, int64(blockInclUint), filTypes.TipSetKey{})
			if gasErr != nil {
				return nil, ErrUnableToEstimateGasFeeCap
			}
			message.GasFeeCap = gasFeeCap

			// Check if gas is affordable for sender
			// gasCost is the maximum amount of FIL to be paid for the execution of this message
			var gasCost = filTypes.NewInt(0)
			gasLimitBigInt := filTypes.NewInt(uint64(message.GasLimit))
			gasCost.Mul(gasLimitBigInt.Int, gasFeeCap.Int)
			if availableFunds.Cmp(gasCost.Int) < 0 {
				return nil, ErrInsufficientBalanceForGas
			}
		} else {
			// We can only estimate gas premium without a sender address
			gasPremium, gasErr := c.node.GasEstimateGasPremium(ctx, blockInclUint, address.Address{}, message.GasLimit, filTypes.TipSetKey{})
			if gasErr != nil {
				return nil, ErrUnableToEstimateGasPremium
			}
			message.GasPremium = gasPremium
		}
	}

	md[GasLimitKey] = strconv.FormatInt(message.GasLimit, 10)
	md[GasPremiumKey] = message.GasPremium.String()
	md[GasFeeCapKey] = message.GasFeeCap.String()
	md[ChainIDKey] = request.NetworkIdentifier.Network

	resp := &types.ConstructionMetadataResponse{
		Metadata: md,
	}

	return resp, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (c *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {

	if request.SignedTransaction == "" {
		return nil, ErrMalformedValue
	}

	err := ValidateNetworkId(ctx, &c.node, request.NetworkIdentifier)
	if err != nil {
		return nil, err
	}

	rawIn := json.RawMessage(request.SignedTransaction)

	bytes, errJson := rawIn.MarshalJSON()
	if errJson != nil {
		return nil, ErrMalformedValue
	}

	var signedTx filTypes.SignedMessage
	errUnmarshal := json.Unmarshal(bytes, &signedTx)
	if errUnmarshal != nil {
		return nil, ErrMalformedValue
	}

	cid, errTx := c.node.MpoolPush(ctx, &signedTx)
	if errTx != nil {
		return nil, ErrUnableToSubmitTx
	}

	resp := &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: cid.String(),
		},
	}

	return resp, nil
}

func (c *ConstructionAPIService) ConstructionCombine(ctx context.Context, request *types.ConstructionCombineRequest) (*types.ConstructionCombineResponse, *types.Error) {
	return nil, ErrNotImplemented
}

func (c *ConstructionAPIService) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	return nil, ErrNotImplemented
}

func (c *ConstructionAPIService) ConstructionHash(ctx context.Context, request *types.ConstructionHashRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, ErrNotImplemented
}

func (c *ConstructionAPIService) ConstructionParse(ctx context.Context, request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	return nil, ErrNotImplemented
}

func (c *ConstructionAPIService) ConstructionPayloads(ctx context.Context, request *types.ConstructionPayloadsRequest) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, ErrNotImplemented
}

func (c *ConstructionAPIService) ConstructionPreprocess(ctx context.Context, request *types.ConstructionPreprocessRequest) (*types.ConstructionPreprocessResponse, *types.Error) {
	return nil, ErrNotImplemented
}
