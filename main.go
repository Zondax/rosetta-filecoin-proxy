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

	endpoints, err := srv.ResolveRPCEndpoints(addr)
	if err != nil {
		return nil, nil, nil, err
	}
	srv.Logger.Infof("Resolved RPC endpoints - V1: %s, V2: %s", endpoints.V1, endpoints.V2)

	v1Client, v1Closer, err := client.NewFullNodeRPCV1(context.Background(), endpoints.V1, headers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create V1 client: %w", err)
	}

	useV2, _ := strconv.ParseBool(srv.EnableLotusV2APIs)
	if useV2 {
		v2Client, v2Closer, err := client.NewFullNodeRPCV2(context.Background(), endpoints.V2, headers)
		if err != nil {
			v1Closer()
			return nil, nil, nil, fmt.Errorf("V2 APIs enabled but failed to create V2 client: %w", err)
		}
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
	accountAPIService := srv.NewAccountAPIService(network, &v1API, &v2API, rosettaLib)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	networkAPIService := srv.NewNetworkAPIService(network, &v1API, srv.GetSupportedOpList())
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	blockAPIService := srv.NewBlockAPIService(network, &v1API, &v2API, rosettaLib)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	mempoolAPIService := srv.NewMemPoolAPIService(network, &v1API, rosettaLib)
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
	createNetworkIdentifierWithF3 := func(tag srv.FinalityTag) *types.NetworkIdentifier {
		return &types.NetworkIdentifier{
			Blockchain: BlockchainName,
			Network:    string(netName),
			SubNetworkIdentifier: &types.SubNetworkIdentifier{
				Network: srv.SubNetworkF3,
				Metadata: map[string]interface{}{
					srv.MetadataFinalityTag: string(tag),
				},
			},
		}
	}

	f3NetworkIdentifiers := []*types.NetworkIdentifier{}
	if srv.IsV2EnabledForService() {
		f3NetworkIdentifiers = []*types.NetworkIdentifier{
			createNetworkIdentifierWithF3(srv.FinalityTagLatest),
			createNetworkIdentifierWithF3(srv.FinalityTagSafe),
			createNetworkIdentifierWithF3(srv.FinalityTagFinalized),
		}
	}

	// The asserter automatically rejects incorrectly formatted
	// requests.
	asserter, err := rosettaAsserter.NewServer(
		srv.GetSupportedOpList(),
		true,
		append([]*types.NetworkIdentifier{network}, f3NetworkIdentifiers...),
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

	srv.Logger.Infof("Connected to Lotus node version: %s | Network: %s | V2 APIs: %v | Finality Anchor Mode: %v", version.String(), srv.NetworkName, v2API != nil, srv.EnableFinalityAnchor != "false")

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

	if enableFinalityAnchor := os.Getenv("ENABLE_FINALITY_ANCHOR"); enableFinalityAnchor != "" {
		srv.EnableFinalityAnchor = enableFinalityAnchor
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
