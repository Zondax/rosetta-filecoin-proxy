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
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/tools"
)

const (
	BlockchainName = services.BlockChainName
	ServerPort     = services.RosettaServerPort
)

var log = logging.Logger("rosetta-filecoin-proxy")

func logVersionsInfo() {
	log.Info("****************************************************")
	log.Infof("Rosetta SDK version: %s", services.RosettaSDKVersion)
	log.Infof("Lotus version: %s", services.LotusVersion)
	log.Infof("Git revision: %s", services.GitRevision)
	log.Info("****************************************************")
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

	return client.NewFullNodeRPC(context.Background(), addr, headers)
}

// newBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func newBlockchainRouter(
	network *types.NetworkIdentifier,
	asserter *rosettaAsserter.Asserter,
	api api.FullNode,
) http.Handler {
	accountAPIService := services.NewAccountAPIService(network, &api)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	networkAPIService := services.NewNetworkAPIService(network, &api)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := services.NewBlockAPIService(network, &api)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	mempoolAPIService := services.NewMemPoolAPIService(network, &api)
	mempoolAPIController := server.NewMempoolAPIController(
		mempoolAPIService,
		asserter,
	)

	constructionAPIService := services.NewConstructionAPIService(network, &api)
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
		services.GetSupportedOpList(),
		false,
		[]*types.NetworkIdentifier{network},
	)
	if err != nil {
		log.Fatal(err)
	}

	router := newBlockchainRouter(network, asserter, api)
	loggedRouter := server.LoggerMiddleware(router)
	corsRouter := server.CorsMiddleware(loggedRouter)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", ServerPort), Handler: corsRouter}

	sigCh := make(chan os.Signal, 2)

	go func() {
		<-sigCh
		log.Warn("Shutting down rosetta...")

		err = srv.Shutdown(context.TODO())
		if err != nil {
			log.Error(err)
		} else {
			log.Warn("Graceful shutdown of rosetta successful")
		}
	}()

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	log.Infof("Rosetta listening on port %d\n", ServerPort)
	return srv.ListenAndServe()
}

func connectAPI(addr string, token string) (api.FullNode, jsonrpc.ClientCloser, error) {
	lotusAPI, clientCloser, err := getFullNodeAPI(addr, token)
	if err != nil {
		log.Errorf("Error %s\n", err)
		return nil, nil, err
	}

	version, err := lotusAPI.Version(context.Background())
	if err != nil {
		log.Warn("Could not get Lotus api version!")
	}

	log.Info("Connected to Lotus version: ", version.String())

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

	log.Info("Starting Rosetta Proxy")
	log.Infof("LOTUS_RPC_URL: %s", addr)

	var lotusAPI api.FullNode
	var clientCloser jsonrpc.ClientCloser
	var err error

	retryAttempts, _ := strconv.Atoi(services.RetryConnectAttempts)

	for i := 1; i <= retryAttempts; i++ {
		lotusAPI, clientCloser, err = connectAPI(addr, token)
		if err == nil {
			break
		}
		log.Errorf("Could not connect to api. Retrying attempt %d", i)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Connect to Lotus api gave up after %d attempts", retryAttempts)
		return
	}
	defer clientCloser()

	setupActorsDatabase(&lotusAPI)

	ctx := context.Background()
	err = startRosettaRPC(ctx, lotusAPI)
	if err != nil {
		log.Info("Exit Rosetta rpc")
	}
}
