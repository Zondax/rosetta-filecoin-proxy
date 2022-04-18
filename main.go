package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	rosettaAsserter "github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	logging "github.com/ipfs/go-log"
	srv "github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
)

var (
	BlockchainName = srv.BlockChainName
	ServerPort, _  = strconv.Atoi(srv.RosettaServerPort)
)

func logVersionsInfo() {
	srv.Logger.Info("****************************************************")
	srv.Logger.Infof("Rosetta SDK version: %s", srv.RosettaSDKVersion)
	srv.Logger.Infof("Lotus version: %s", srv.LotusVersion)
	srv.Logger.Infof("Git revision: %s", srv.GitRevision)
	srv.Logger.Info("****************************************************")
}

func startLogger(level string) {
	lvl, err := logging.LevelFromString(level)
	if err != nil {
		panic(err)
	}
	logging.SetAllLoggers(lvl)
}

func getFullNodeAPI(addr string, token string) (api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(token) > 0 {
		headers.Add("Authorization", "Bearer "+token)
	}

	return client.NewFullNodeRPCV1(context.Background(), addr, headers)
}

// newBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func newBlockchainRouter(
	network *types.NetworkIdentifier,
	asserter *rosettaAsserter.Asserter,
	api api.FullNode,
) http.Handler {
	accountAPIService := srv.NewAccountAPIService(network, &api)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	networkAPIService := srv.NewNetworkAPIService(network, &api)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := srv.NewBlockAPIService(network, &api)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	mempoolAPIService := srv.NewMemPoolAPIService(network, &api)
	mempoolAPIController := server.NewMempoolAPIController(
		mempoolAPIService,
		asserter,
	)

	constructionAPIService := srv.NewConstructionAPIService(network, &api)
	constructionAPIController := server.NewConstructionAPIController(
		constructionAPIService,
		asserter,
	)

	return server.NewRouter(accountAPIController, networkAPIController,
		blockAPIController, mempoolAPIController, constructionAPIController)
}

func startRosettaRPC(ctx context.Context, api api.FullNode) error {
	netName, _ := api.StateNetworkName(ctx)
	network := &types.NetworkIdentifier{
		Blockchain: BlockchainName,
		Network:    string(netName),
	}

	// The asserter automatically rejects incorrectly formatted
	// requests.
	asserter, err := rosettaAsserter.NewServer(
		srv.GetSupportedOpList(),
		true,
		[]*types.NetworkIdentifier{network},
		nil,
		false,
		"",
	)
	if err != nil {
		srv.Logger.Fatal(err)
	}

	router := newBlockchainRouter(network, asserter, api)
	loggedRouter := server.LoggerMiddleware(router)
	corsRouter := server.CorsMiddleware(loggedRouter)
	server := &http.Server{Addr: fmt.Sprintf(":%d", ServerPort), Handler: corsRouter}

	sigCh := make(chan os.Signal, 2)

	go func() {
		<-sigCh
		srv.Logger.Warn("Shutting down rosetta...")

		err = server.Shutdown(context.TODO())
		if err != nil {
			srv.Logger.Error(err)
		} else {
			srv.Logger.Warn("Graceful shutdown of rosetta successful")
		}
	}()

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	srv.Logger.Infof("Rosetta listening on port %d\n", ServerPort)
	return server.ListenAndServe()
}

func connectAPI(addr string, token string) (api.FullNode, jsonrpc.ClientCloser, error) {
	lotusAPI, clientCloser, err := getFullNodeAPI(addr, token)
	if err != nil {
		srv.Logger.Errorf("Error %s\n", err)
		return nil, nil, err
	}

	version, err := lotusAPI.Version(context.Background())
	if err != nil {
		srv.Logger.Warn("Could not get Lotus api version!")
	}

	srv.Logger.Info("Connected to Lotus version: ", version.String())

	return lotusAPI, clientCloser, nil
}

func setupActorsDatabase(api *api.FullNode) {
	var db tools.Database = &tools.Cache{}
	db.NewImpl(api)
	tools.ActorsDB = db
}

func main() {
	startLogger("info")
	logVersionsInfo()

	addr := os.Getenv("LOTUS_RPC_URL")
	token := os.Getenv("LOTUS_RPC_TOKEN")

	srv.Logger.Info("Starting Rosetta Proxy")
	srv.Logger.Infof("LOTUS_RPC_URL: %s", addr)

	var lotusAPI api.FullNode
	var clientCloser jsonrpc.ClientCloser
	var err error

	retryAttempts, _ := strconv.Atoi(srv.RetryConnectAttempts)

	for i := 1; i <= retryAttempts; i++ {
		lotusAPI, clientCloser, err = connectAPI(addr, token)
		if err == nil {
			break
		}
		srv.Logger.Errorf("Could not connect to api. Retrying attempt %d", i)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		srv.Logger.Fatalf("Connect to Lotus api gave up after %d attempts", retryAttempts)
		return
	}
	defer clientCloser()

	setupActorsDatabase(&lotusAPI)

	ctx := context.Background()
	err = startRosettaRPC(ctx, lotusAPI)
	if err != nil {
		srv.Logger.Infof("Exit Rosetta rpc: %s", err.Error())
	}
}
