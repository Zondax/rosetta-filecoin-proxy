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
)

// ChainIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the current chain id
const ChainIDKey = "ChainID"

// OptionsIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the account ID
const OptionsIDKey = "Id"

// OptionsBlockInclKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse determines on how many epochs message should included
// being 0 the fastest (and the most gas expensive one)
const OptionsBlockInclKey = "BlockIncl"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "Nonce"

// GasPriceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies tx's gas price
const GasPriceKey = "GasPrice"

// GasLimitKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies tx's gas limit
const GasLimitKey = "GasLimit"

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

	if request.Options == nil {
		return nil, ErrInvalidAccountAddress
	}

	addressRaw, ok := request.Options[OptionsIDKey]
	if !ok {
		return nil, ErrInvalidAccountAddress
	}

	var blockInclUint uint64
	blockIncl, ok := request.Options[OptionsBlockInclKey]
	if !ok {
		blockInclUint = 1
	} else {
		blockInclUint = uint64(blockIncl.(float64))
	}

	errNet := ValidateNetworkId(ctx, &c.node, request.NetworkIdentifier)
	if errNet != nil {
		return nil, errNet
	}

	addressParsed, adErr := address.NewFromString(addressRaw.(string))
	if adErr != nil {
		return nil, ErrInvalidAccountAddress
	}

	nonce, err := c.node.MpoolGetNonce(ctx, addressParsed)
	if err != nil {
		return nil, ErrUnableToGetNextNonce
	}

	gasLimit := filTypes.NewInt(uint64(build.BlockGasLimit))
	gasPrice, gasErr := c.node.MpoolEstimateGasPrice(ctx, blockInclUint, addressParsed,
		gasLimit.Int64(), filTypes.TipSetKey{})
	if gasErr != nil {
		return nil, ErrUnableToEstimateGasPrice
	}

	//Check if gas is affordable for address
	actor, errAct := c.node.StateGetActor(context.Background(), addressParsed, filTypes.EmptyTSK)
	if errAct != nil {
		return nil, ErrUnableToGetActor
	}

	var availableFunds filTypes.BigInt
	if actor.Code == builtin.MultisigActorCodeID {
		//Get the unlocked funds of the multisig account
		availableFunds, err = c.node.MsigGetAvailableBalance(ctx, addressParsed, filTypes.EmptyTSK)
		if err != nil {
			return nil, ErrUnableToGetBalance
		}
	} else {
		availableFunds = actor.Balance
	}

	var gasCost = filTypes.NewInt(0)
	gasCost.Mul(gasLimit.Int, gasPrice.Int)
	if availableFunds.Cmp(gasCost.Int) < 0 {
		return nil, ErrInsufficientBalanceForGas
	}

	md := make(map[string]interface{})
	md[NonceKey] = nonce
	md[GasLimitKey] = gasLimit
	md[GasPriceKey] = gasPrice.Int64()
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
