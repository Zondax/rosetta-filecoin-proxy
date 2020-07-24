package main

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"log"
	"net/http"
	"os"
)

func getFullNodeAPI(addr string, token string) (api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(token) > 0 {
		headers.Add("Authorization", "Bearer "+token)
	}

	return client.NewFullNodeRPC(addr, headers)
}

func main() {
	addr := os.Getenv("LOTUS_RPC_ENDPOINT")
	token := os.Getenv("LOTUS_RPC_TOKEN")

	log.Printf("Starting Rosetta Proxy")
	log.Printf("LOTUS_RPC_ENDPOINT: %s", addr)

	// TODO: We need to abstract this away and reconnect to the api if there are problems

	api, clientCloser, err := getFullNodeAPI(addr, token)
	if err != nil {
		log.Printf("Error %s\n", err)
		return
	}
	defer clientCloser()

	log.Printf("Connected to Lotus")

	// TODO: instantiate Rosetta and pass the API
	ctx := context.Background()
	state, err := api.SyncState(ctx)

	log.Printf("---- Active syncs %d", len(state.ActiveSyncs))
	for _, as := range state.ActiveSyncs {
		log.Printf("%v", as)
	}
}
