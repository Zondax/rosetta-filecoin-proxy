package main

import (
  "net/http"
  "context"
  "time"
  "fmt"
  "github.com/coinbase/rosetta-sdk-go/types"
  //"github.com/zondax/rosetta-filecoin-lib"
  "github.com/coinbase/rosetta-sdk-go/client"
  "github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
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
  options[services.OptionsIDKey] = "t137sjdbgunloi7couiy4l5nc7pd6k2jmq32vizpy"
  options[services.OptionsBlockInclKey] = 2

  request := &types.ConstructionMetadataRequest{
    NetworkIdentifier: Network,
    Options:           options,
  }

  resp, err1, err2 := rosettaClient.ConstructionAPI.ConstructionMetadata(ctx, request)
  if err1 != nil {
    panic(err1.Message)
  }

  if err2 != nil {
    panic(err2.Error())
  }

  if resp == nil {
    panic("Panicking")
  }

  fmt.Println(resp)
}
