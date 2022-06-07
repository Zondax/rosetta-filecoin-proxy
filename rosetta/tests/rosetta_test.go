package tests

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
)

const ServerURL = "http://localhost:8081"

const NetworkName = "mainnet"

var (
	ctx = context.Background()

	NetworkID = &types.NetworkIdentifier{
		Blockchain: services.BlockChainName,
		Network:    NetworkName,
	}
)

func setupRosettaClient() *client.APIClient {
	clientCfg := client.NewConfiguration(
		ServerURL,
		"rosetta-test",
		&http.Client{
			Timeout: 60 * 5 * time.Second,
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

	if resp.NetworkIdentifiers[0].Blockchain != NetworkID.Blockchain {
		t.Error()
	}

	if resp.NetworkIdentifiers[0].Network != NetworkID.Network {
		t.Errorf("Networks don't match %s != %s", resp.NetworkIdentifiers[0].Network, NetworkID.Network)
	}
}

func TestGetBlock(t *testing.T) {
	rosettaClient := setupRosettaClient()
	var requestHeight int64 = 790000
	var request = types.BlockRequest{
		NetworkIdentifier: NetworkID,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &requestHeight,
		},
	}
	blockResponseA, _, err := rosettaClient.BlockAPI.Block(ctx, &request)
	if err != nil {
		t.Fatal(err)
	}
	if blockResponseA.Block.ParentBlockIdentifier == nil {
		t.Error("Block parent is null")
	}
	requestHeight++
	blockResponseB, _, err := rosettaClient.BlockAPI.Block(ctx, &request)
	if err != nil {
		t.Fatal(err)
	}
	if blockResponseB.Block.ParentBlockIdentifier == nil {
		t.Error("Block parent is null")
	}
	if !reflect.DeepEqual(blockResponseA.Block.BlockIdentifier.Hash,
		blockResponseB.Block.ParentBlockIdentifier.Hash) {
		t.Fatalf("Invalid parent for block")
	}
}

func TestConstructionMetadata(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsSenderIDKey] = "f1abjxfbp274xpdqcpuaykwkfb43omjotacm2p3za"
	options[services.OptionsReceiverIDKey] = "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
	options[services.OptionsBlockInclKey] = 1
	options[services.OptionsValueKey] = "5"

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: NetworkID,
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

	fmt.Println("gasPremium", resp.Metadata[services.GasPremiumKey])
	fmt.Println("gasLimit", resp.Metadata[services.GasLimitKey])
	fmt.Println("gasFeeCap", resp.Metadata[services.GasFeeCapKey])
	fmt.Println("nonce", resp.Metadata[services.NonceKey])
	fmt.Println("Receivers actor id", resp.Metadata[services.DestinationActorIdKey])
}

func TestConstructionMetadataNonexistentReceiverActor(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsSenderIDKey] = "f1abjxfbp274xpdqcpuaykwkfb43omjotacm2p3za"
	options[services.OptionsReceiverIDKey] = "f1pfmrkoipk2byrdz33usb3m25s56kyrvhchypfai" // This address doesn't exist on chain
	options[services.OptionsBlockInclKey] = 1
	options[services.OptionsValueKey] = "5"

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: NetworkID,
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

	fmt.Println("gasPremium", resp.Metadata[services.GasPremiumKey])
	fmt.Println("gasLimit", resp.Metadata[services.GasLimitKey])
	fmt.Println("gasFeeCap", resp.Metadata[services.GasFeeCapKey])
	fmt.Println("nonce", resp.Metadata[services.NonceKey])
	fmt.Println("Receivers actor id", resp.Metadata[services.DestinationActorIdKey])
}

func TestConstructionMetadataForGasPremiumTrack(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsBlockInclKey] = 1

	request := &types.ConstructionMetadataRequest{
		NetworkIdentifier: NetworkID,
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

	fmt.Println("gasPremium", resp.Metadata[services.GasPremiumKey])
	fmt.Println("gasLimit", resp.Metadata[services.GasLimitKey])
	fmt.Println("gasFeeCap", resp.Metadata[services.GasFeeCapKey])
}

func TestMempool(t *testing.T) {

	rosettaClient := setupRosettaClient()
	req := &types.NetworkRequest{
		NetworkIdentifier: NetworkID,
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
		NetworkIdentifier:     NetworkID,
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
