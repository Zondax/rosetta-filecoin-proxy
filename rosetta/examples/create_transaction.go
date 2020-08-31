package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	rosettaFilecoinLib "github.com/zondax/rosetta-filecoin-lib"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
	"net/http"
	"strconv"
	"time"
	//filtypes "github.com/filecoin-project/lotus/chain/types"
)

const ServerURL = "http://localhost:8080"

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

func main() {

	var (
		ctx = context.Background()

		Network = &types.NetworkIdentifier{
			Blockchain: "Filecoin",
			Network:    "testnet",
		}
	)

	rosettaClient := setupRosettaClient()

	var options = make(map[string]interface{})
	options[services.OptionsSenderIDKey] = "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
	options[services.OptionsBlockInclKey] = 2

	requestMetadata := &types.ConstructionMetadataRequest{
		NetworkIdentifier: Network,
		Options:           options,
	}

	respMetadata, err1, err2 := rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, requestMetadata)
	if err1 != nil {
		panic(err1.Message)
	}

	if err2 != nil {
		panic(err2.Error())
	}

	if respMetadata == nil {
		panic("Panicking")
	}

	r := &rosettaFilecoinLib.RosettaConstructionFilecoin{false}

	gasLimit, err := strconv.ParseInt(respMetadata.Metadata[services.GasLimitKey].(string), 10, 64)
	if err != nil {
		panic(err)
	}

	gasPremium, err := strconv.ParseInt(respMetadata.Metadata[services.GasPremiumKey].(string), 10, 64)
	if err != nil {
		panic(err)
	}
	gasFeeCap, err := strconv.ParseInt(respMetadata.Metadata[services.GasFeeCapKey].(string), 10, 64)
	if err != nil {
		panic(err)
	}

	mtx := rosettaFilecoinLib.TxMetadata{
		Nonce:      uint64(respMetadata.Metadata[services.NonceKey].(float64)),
		GasPremium: gasPremium,
		GasFeeCap:  gasFeeCap,
		GasLimit:   gasLimit,
	}
	pr := &rosettaFilecoinLib.PaymentRequest{
		From:     "t1d2xrzcslx7xlbbylc5c3d5lvandqw4iwl6epxba",
		To:       "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy",
		Quantity: 100000,
		Metadata: mtx,
	}

	txBase64, err := r.ConstructPayment(pr)
	if err != nil {
		panic(err)
	}

	sk, err := base64.StdEncoding.DecodeString("8VcW07ADswS4BV2cxi5rnIadVsyTDDhY1NfDH19T8Uo=")
	if err != nil {
		panic(err)
	}

	sig, err := r.SignTx(txBase64, sk)
	if err != nil {
		panic(err)
	}

	fmt.Println(sig)
	requestSubmit := &types.ConstructionSubmitRequest{
		NetworkIdentifier: Network,
		SignedTransaction: sig,
	}

	respSubmit, err1, err2 := rosettaClient.ConstructionAPI.ConstructionSubmit(ctx, requestSubmit)
	if err1 != nil {
		panic(err1.Message)
	}

	if err2 != nil {
		panic(err2.Error())
	}

	if respSubmit == nil {
		panic("Panicking")
	}

	fmt.Println(respSubmit.TransactionIdentifier.Hash)

	hash, err := r.Hash(sig)
	if err != nil {
		panic(err)
	}

	fmt.Println(hash)
	if hash != respSubmit.TransactionIdentifier.Hash {
		panic("NOT MATCHING")
	}

}
