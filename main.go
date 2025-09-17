package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	rosettaFilecoinLib "github.com/zondax/rosetta-filecoin-lib"

	rosettaAsserter "github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/api/v2api"
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

func getFullNodeAPI(addr string, token string) (api.FullNode, v2api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{}
	if len(token) > 0 {
		headers.Add("Authorization", "Bearer "+token)
	}

	// Determine the V1 and V2 endpoints based on the provided address
	var v1Addr, v2Addr string

	if strings.HasSuffix(addr, "/rpc") {
		// New format: base URL ends with /rpc
		// Append /v1 and /v2 to create specific endpoints
		v1Addr = addr + "/v1"
		v2Addr = addr + "/v2"
		srv.Logger.Infof("Using base /rpc endpoint - V1: %s, V2: %s", v1Addr, v2Addr)
	} else if strings.Contains(addr, "/rpc/v1") {
		// Legacy format: already has /rpc/v1
		v1Addr = addr
		v2Addr = strings.Replace(addr, "/rpc/v1", "/rpc/v2", 1)
		srv.Logger.Infof("Using legacy /rpc/v1 endpoint - V1: %s, V2: %s", v1Addr, v2Addr)
	} else if strings.Contains(addr, "/rpc/v2") {
		// If someone provides v2 endpoint directly, derive v1 from it
		v1Addr = strings.Replace(addr, "/rpc/v2", "/rpc/v1", 1)
		v2Addr = addr
		srv.Logger.Infof("Using /rpc/v2 endpoint - V1: %s, V2: %s", v1Addr, v2Addr)
	} else {
		// Unrecognized format - return error
		return nil, nil, nil, fmt.Errorf("unrecognized RPC endpoint format: %s. Expected format ending with /rpc, /rpc/v1, or /rpc/v2", addr)
	}

	// Always create V1 client
	v1Client, v1Closer, err := client.NewFullNodeRPCV1(context.Background(), v1Addr, headers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create V1 client: %w", err)
	}

	// Check if V2 APIs should be created
	useV2, _ := strconv.ParseBool(srv.EnableLotusV2APIs)
	if useV2 {
		// Try to create V2 client
		v2Client, v2Closer, err := client.NewFullNodeRPCV2(context.Background(), v2Addr, headers)
		if err != nil {
			srv.Logger.Warnf("V2 APIs enabled but failed to create V2 client: %v. Will use V1 only", err)
			return v1Client, nil, v1Closer, nil
		}
		// Return both clients, but use combined closer
		combinedCloser := func() {
			v1Closer()
			v2Closer()
		}
		return v1Client, v2Client, combinedCloser, nil
	}

	return v1Client, nil, v1Closer, nil
}

// newBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func newBlockchainRouter(
	network *types.NetworkIdentifier,
	asserter *rosettaAsserter.Asserter,
	v1API api.FullNode,
	v2API v2api.FullNode,
	rosettaLib *rosettaFilecoinLib.RosettaConstructionFilecoin,
) http.Handler {
	accountAPIService := srv.NewAccountAPIService(network, &v1API, v2API, rosettaLib)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	networkAPIService := srv.NewNetworkAPIService(network, &v1API, v2API, srv.GetSupportedOpList())
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := srv.NewBlockAPIService(network, &v1API, v2API, rosettaLib)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	mempoolAPIService := srv.NewMemPoolAPIService(network, &v1API, v2API, rosettaLib)
	mempoolAPIController := server.NewMempoolAPIController(
		mempoolAPIService,
		asserter,
	)

	constructionAPIService := srv.NewConstructionAPIService(network, &v1API, rosettaLib)
	constructionAPIController := server.NewConstructionAPIController(
		constructionAPIService,
		asserter,
	)

	return server.NewRouter(accountAPIController, networkAPIController,
		blockAPIController, mempoolAPIController, constructionAPIController)
}

func startRosettaRPC(ctx context.Context, v1API api.FullNode, v2API v2api.FullNode) error {
	netName, _ := v1API.StateNetworkName(ctx)
	network := &types.NetworkIdentifier{
		Blockchain: BlockchainName,
		Network:    string(netName),
	}

	// Create network identifier with f3 sub-network for finality support
	networkWithF3 := &types.NetworkIdentifier{
		Blockchain: BlockchainName,
		Network:    string(netName),
		SubNetworkIdentifier: &types.SubNetworkIdentifier{
			Network: srv.SubNetworkF3,
		},
	}

	// The asserter automatically rejects incorrectly formatted
	// requests.
	asserter, err := rosettaAsserter.NewServer(
		srv.GetSupportedOpList(),
		true,
		[]*types.NetworkIdentifier{network, networkWithF3},
		nil,
		false,
		"",
	)
	if err != nil {
		srv.Logger.Fatal(err)
	}

	// Create instance of RosettaFilecoinLib for current network
	r := rosettaFilecoinLib.NewRosettaConstructionFilecoin(v1API)

	router := newBlockchainRouter(network, asserter, v1API, v2API, r)
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

func connectAPI(addr string, token string) (api.FullNode, v2api.FullNode, jsonrpc.ClientCloser, error) {
	v1API, v2API, clientCloser, err := getFullNodeAPI(addr, token)
	if err != nil {
		srv.Logger.Errorf("Error %s\n", err)
		return nil, nil, nil, err
	}

	networkName, err := v1API.StateNetworkName(context.Background())
	if err != nil {
		srv.Logger.Warn("Could not get Lotus network name!")
	}

	srv.NetworkName = string(networkName)

	version, err := v1API.Version(context.Background())
	if err != nil {
		srv.Logger.Warn("Could not get Lotus api version!")
	}

	if v2API != nil {
		srv.Logger.Infof("Connected to Lotus node version: %s | Network: %s | V2 APIs: enabled", version.String(), srv.NetworkName)
	} else {
		srv.Logger.Infof("Connected to Lotus node version: %s | Network: %s | V2 APIs: disabled", version.String(), srv.NetworkName)
	}

	return v1API, v2API, clientCloser, nil
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

	// Configure V2 API usage
	if enableV2 := os.Getenv("ENABLE_LOTUS_V2_APIS"); enableV2 != "" {
		srv.EnableLotusV2APIs = enableV2
	}

	srv.Logger.Info("Starting Rosetta Proxy")
	srv.Logger.Infof("LOTUS_RPC_URL: %s", addr)

	var lotusV1API api.FullNode
	var lotusV2API v2api.FullNode
	var clientCloser jsonrpc.ClientCloser
	var err error

	retryAttempts, _ := strconv.Atoi(srv.RetryConnectAttempts)

	for i := 1; i <= retryAttempts; i++ {
		lotusV1API, lotusV2API, clientCloser, err = connectAPI(addr, token)
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

	setupActorsDatabase(&lotusV1API)

	ctx := context.Background()
	err = startRosettaRPC(ctx, lotusV1API, lotusV2API)
	if err != nil {
		srv.Logger.Infof("Exit Rosetta rpc: %s", err.Error())
	}
}
