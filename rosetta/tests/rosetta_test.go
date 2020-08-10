package tests

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const ServerURL = "http://localhost:8080"

var (
	ctx = context.Background()

	Network = &types.NetworkIdentifier{
		Blockchain: "Filecoin",
		Network:    "testnet",
	}
)

func setupRosettaClient() *client.APIClient {
	clientCfg := client.NewConfiguration(
		ServerURL,
		"rosetta-test",
		&http.Client{
			Timeout: 4 * time.Second,
		},
	)

	rosettaClient := client.NewAPIClient(clientCfg)
	return rosettaClient
}

func TestNetworkList(t *testing.T) {

	rosettaClient := setupRosettaClient()

	resp, _, err := rosettaClient.NetworkAPI.NetworkList(ctx, &types.MetadataRequest{})

	if err != nil {
		t.Fatalf("Failed to get NetworkList: %s", err)
	}

	if len(resp.NetworkIdentifiers) == 0 {
		t.Fatal("NetworkIdentifiers is empty")
	}

	if resp.NetworkIdentifiers[0].Blockchain != "Filecoin" {
		t.Error()
	}

	if resp.NetworkIdentifiers[0].Network != "testnet" {
		t.Error()
	}
}

func TestGetGenesisBlock(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var requestHeight int64 = 0
	var request = types.BlockRequest{
		NetworkIdentifier: Network,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &requestHeight,
		},
	}

	blockResponse, _, err := rosettaClient.BlockAPI.Block(ctx, &request)
	if err != nil {
		t.Fatal(err)
	}

	if blockResponse.Block.ParentBlockIdentifier == nil {
		t.Error("Block parent is null")
	}

	if !reflect.DeepEqual(blockResponse.Block.BlockIdentifier,
		blockResponse.Block.ParentBlockIdentifier) {

		t.Fatalf("Invalid parent for genesis block")
	}
}

func TestConstructionMetadata(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsIDKey] = "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
	options[services.OptionsBlockInclKey] = 1

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: Network,
		Options:           options,
	}

	resp, err1, err2 := rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, request)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil {
		t.Fatal()
	}
}

func TestConstructionMetadataForGasPriceTrack(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsBlockInclKey] = 1

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: Network,
		Options:           options,
	}

	resp, err1, err2 := rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, request)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil {
		t.Fatal()
	}
}

func TestMempool(t *testing.T) {

	rosettaClient := setupRosettaClient()
	req := &types.NetworkRequest{
		NetworkIdentifier: Network,
		Metadata:          nil,
	}

	resp, err1, err2 := rosettaClient.MempoolAPI.Mempool(ctx, req)

	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil || len(resp.TransactionIdentifiers) == 0 {
		t.Fatal()
	}

	txReq := &types.MempoolTransactionRequest{
		NetworkIdentifier:     Network,
		TransactionIdentifier: resp.TransactionIdentifiers[0],
	}
	txResp, err1, err2 := rosettaClient.MempoolAPI.MempoolTransaction(ctx, txReq)

	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if txResp == nil {
		t.Fatal()
	}
}
