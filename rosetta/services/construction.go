// +build rosetta_rpc

package services

import (
	"context"
	"encoding/hex"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
)


// OptionsIDKey is the name of the key in the Options map inside a
// ConstructionMetadataRequest that specifies the account ID.
const OptionsIDKey = "id"

// NonceKey is the name of the key in the Metadata map inside a
// ConstructionMetadataResponse that specifies the next valid nonce.
const NonceKey = "nonce"

// ConstructionAPIService implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	network *types.NetworkIdentifier
	node api.FullNode
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

	err := ValidateNetworkId(ctx, &c.node, request.NetworkIdentifier)
	if err != nil {
		return nil, err
	}

	addressParsed, adErr := address.NewFromString(addressRaw.(string))
	if adErr != nil {
		return nil, ErrInvalidAccountAddress
	}

	nonce, adErr := c.node.MpoolGetNonce(ctx, addressParsed)
	if adErr != nil {
		return nil, ErrUnableToGetNextNonce
	}

	md := make(map[string]interface{})
	md[NonceKey] = nonce

	resp := &types.ConstructionMetadataResponse{
		Metadata: md,
	}

	return resp, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (c *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.ConstructionSubmitResponse, *types.Error) {

	if request.SignedTransaction == "" {
		return nil, ErrMalformedValue
	}

	err := ValidateNetworkId(ctx, &c.node, request.NetworkIdentifier)
	if err != nil {
		return nil, err
	}

	byteTx, errTx := hex.DecodeString(request.SignedTransaction)
	if errTx != nil {
		return nil, ErrMalformedTx
	}

	signedTx, errTx := filTypes.DecodeSignedMessage(byteTx)
	if errTx != nil {
		return nil, ErrMalformedTx
	}

	cid, errTx := c.node.MpoolPush(ctx, signedTx)
	if errTx != nil {
		return nil, ErrUnableToSubmitTx
	}

	resp := &types.ConstructionSubmitResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: cid.String(),
		},
	}

	return resp, nil
}

