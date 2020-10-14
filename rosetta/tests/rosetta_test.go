package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	rosettaFilecoinLib "github.com/zondax/rosetta-filecoin-lib"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
	"net/http"
	"reflect"
	"testing"
	"time"
)

const ServerURL = "http://localhost:8080"

var (
	ctx = context.Background()

	NetworkID = &types.NetworkIdentifier{
		Blockchain: services.BlockChainName,
		Network:    services.NetworkName,
	}
)

func setupRosettaClient() *client.APIClient {
	clientCfg := client.NewConfiguration(
		ServerURL,
		"rosetta-test",
		&http.Client{
			Timeout: 120 * time.Second,
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
		t.Error()
	}
}

func TestGetGenesisBlock(t *testing.T) {

	rosettaClient := setupRosettaClient()

	var requestHeight int64 = 0
	var request = types.BlockRequest{
		NetworkIdentifier: NetworkID,
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
	options[services.OptionsSenderIDKey] = "t1itpqzzcx6yf52oc35dgsoxfqkoxpy6kdmygbaja"
	options[services.OptionsReceiverIDKey] = "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
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

func TestSendTransaction(t *testing.T) {
	rosettaClient := setupRosettaClient()
	addressA := "t1d2xrzcslx7xlbbylc5c3d5lvandqw4iwl6epxba"
	addressB := "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
	pkA := "8VcW07ADswS4BV2cxi5rnIadVsyTDDhY1NfDH19T8Uo="
	pkB := "YbDPh1vq3fBClzbiwDt6WjniAdZn8tNcCwcBO2hDwyk="
	var options = make(map[string]interface{})
	var amount = "1"

	// Send from A to B
	options[services.OptionsSenderIDKey] = addressA
	options[services.OptionsReceiverIDKey] = addressB
	options[services.OptionsBlockInclKey] = 1

	requestMetadata := &types.ConstructionMetadataRequest{
		NetworkIdentifier: NetworkID,
		Options:           options,
	}

	respMetadata, err1, err2 := rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, requestMetadata)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if respMetadata == nil {
		t.Fatal("Panicking")
	}

	r := &rosettaFilecoinLib.RosettaConstructionFilecoin{}

	gasLimit := respMetadata.Metadata[services.GasLimitKey].(int64)
	gasPremium := respMetadata.Metadata[services.GasPremiumKey].(string)
	gasFeeCap := respMetadata.Metadata[services.GasFeeCapKey].(string)

	mtx := rosettaFilecoinLib.TxMetadata{
		Nonce:      uint64(respMetadata.Metadata[services.NonceKey].(float64)),
		GasPremium: gasPremium,
		GasFeeCap:  gasFeeCap,
		GasLimit:   gasLimit,
	}
	pr := &rosettaFilecoinLib.PaymentRequest{
		From:     addressA,
		To:       addressB,
		Quantity: amount,
		Metadata: mtx,
	}

	txJSON, err := r.ConstructPayment(pr)
	if err != nil {
		t.Fatal(err)
	}

	sk, err := base64.StdEncoding.DecodeString(pkA)
	if err != nil {
		t.Fatal(err)
	}

	signedTxJSON, err := r.SignTxJSON(txJSON, sk)
	if err != nil {
		t.Fatal(err)
	}

	requestSubmit := &types.ConstructionSubmitRequest{
		NetworkIdentifier: NetworkID,
		SignedTransaction: signedTxJSON,
	}

	respSubmit, err1, err2 := rosettaClient.ConstructionAPI.ConstructionSubmit(ctx, requestSubmit)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if respSubmit == nil {
		t.Fatal("Panicking")
	}

	hash, err := r.Hash(signedTxJSON)
	if err != nil {
		t.Fatal(err)
	}

	if hash != respSubmit.TransactionIdentifier.Hash {
		t.Fatal("NOT MATCHING")
	}

	// Send tokens back to A
	options[services.OptionsSenderIDKey] = addressB
	options[services.OptionsReceiverIDKey] = addressA
	options[services.OptionsBlockInclKey] = 2

	requestMetadata = &types.ConstructionMetadataRequest{
		NetworkIdentifier: NetworkID,
		Options:           options,
	}

	respMetadata, err1, err2 = rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, requestMetadata)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if respMetadata == nil {
		t.Fatal("Panicking")
	}

	r = &rosettaFilecoinLib.RosettaConstructionFilecoin{}

	gasLimit = respMetadata.Metadata[services.GasLimitKey].(int64)
	gasPremium = respMetadata.Metadata[services.GasPremiumKey].(string)
	gasFeeCap = respMetadata.Metadata[services.GasFeeCapKey].(string)

	mtx = rosettaFilecoinLib.TxMetadata{
		Nonce:      uint64(respMetadata.Metadata[services.NonceKey].(float64)),
		GasPremium: gasPremium,
		GasFeeCap:  gasFeeCap,
		GasLimit:   gasLimit,
	}
	pr = &rosettaFilecoinLib.PaymentRequest{
		From:     addressB,
		To:       addressA,
		Quantity: amount,
		Metadata: mtx,
	}

	txJSON, err = r.ConstructPayment(pr)
	if err != nil {
		t.Fatal(err)
	}

	sk, err = base64.StdEncoding.DecodeString(pkB)
	if err != nil {
		t.Fatal(err)
	}

	signedTxJSON, err = r.SignTxJSON(txJSON, sk)
	if err != nil {
		t.Fatal(err)
	}

	requestSubmit = &types.ConstructionSubmitRequest{
		NetworkIdentifier: NetworkID,
		SignedTransaction: signedTxJSON,
	}

	respSubmit, err1, err2 = rosettaClient.ConstructionAPI.ConstructionSubmit(ctx, requestSubmit)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if respSubmit == nil {
		t.Fatal("Panicking")
	}

	hash, err = r.Hash(signedTxJSON)
	if err != nil {
		t.Fatal(err)
	}

	if hash != respSubmit.TransactionIdentifier.Hash {
		t.Fatal("NOT MATCHING")
	}
}

func TestGetBalanceOfMultiSig(t *testing.T) {
	rosettaClient := setupRosettaClient()
	testAddress := "t020406"
	fmt.Println("Testing on address:", testAddress)

	// Get full balance (locked + spendable)
	req := &types.AccountBalanceRequest{
		NetworkIdentifier: NetworkID,
		AccountIdentifier: &types.AccountIdentifier{
			Address:    testAddress,
			SubAccount: nil,
		},
	}

	resp, err1, err2 := rosettaClient.AccountAPI.AccountBalance(ctx, req)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil {
		t.Fatal()
	}

	full := resp.Balances[0].Value
	fmt.Println("Total balance is:", full)

	// Get locked balance
	req = &types.AccountBalanceRequest{
		NetworkIdentifier: NetworkID,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAddress,
			SubAccount: &types.SubAccountIdentifier{
				Address: services.LockedBalanceStr,
			},
		},
	}

	resp, err1, err2 = rosettaClient.AccountAPI.AccountBalance(ctx, req)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil {
		t.Fatal()
	}

	locked := resp.Balances[0].Value
	fmt.Println("Locked balance is:", locked)

	// Get spendable balance
	req = &types.AccountBalanceRequest{
		NetworkIdentifier: NetworkID,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAddress,
			SubAccount: &types.SubAccountIdentifier{
				Address: services.SpendableBalanceStr,
			},
		},
	}

	resp, err1, err2 = rosettaClient.AccountAPI.AccountBalance(ctx, req)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil {
		t.Fatal()
	}

	spendable := resp.Balances[0].Value
	fmt.Println("Spendable balance is:", spendable)

	// Get vesting schedule
	req = &types.AccountBalanceRequest{
		NetworkIdentifier: NetworkID,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAddress,
			SubAccount: &types.SubAccountIdentifier{
				Address: services.VestingScheduleStr,
			},
		},
	}

	resp, err1, err2 = rosettaClient.AccountAPI.AccountBalance(ctx, req)
	if err1 != nil {
		t.Fatal(err1.Message)
	}

	if err2 != nil {
		t.Fatal(err2.Error())
	}

	if resp == nil || len(resp.Metadata) == 0 {
		t.Fatal()
	}

	fmt.Println("Vesting schedule is:", resp.Metadata)
}
