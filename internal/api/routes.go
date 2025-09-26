package api

import (
	"blockchain-node/internal/api/handlers"
	"blockchain-node/internal/api/middleware"
	"blockchain-node/pkg/blockchain"
	"blockchain-node/pkg/wallet"

	"github.com/gorilla/mux"
)

// Router creates and configures the API routes
type Router struct {
	blockchainHandler *handlers.BlockchainHandler
	walletHandler     *handlers.WalletHandler
	miningHandler     *handlers.MiningHandler
}

// NewRouter creates a new API router
func NewRouter(bc *blockchain.Blockchain, minerWallet *wallet.Wallet) *Router {
	return &Router{
		blockchainHandler: handlers.NewBlockchainHandler(bc),
		walletHandler:     handlers.NewWalletHandler(bc),
		miningHandler:     handlers.NewMiningHandler(bc, minerWallet),
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes() *mux.Router {
	router := mux.NewRouter()
	
	// Add middleware
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	
	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Root info endpoint
	api.HandleFunc("/info", r.handleNodeInfo).Methods("GET")
	api.HandleFunc("/health", r.handleHealthCheck).Methods("GET")
	
	// Blockchain routes
	r.setupBlockchainRoutes(api)
	
	// Wallet routes
	r.setupWalletRoutes(api)
	
	// Mining routes
	r.setupMiningRoutes(api)
	
	// Legacy routes (for backward compatibility)
	r.setupLegacyRoutes(api)
	
	return router
}

// setupBlockchainRoutes configures blockchain-related routes
func (r *Router) setupBlockchainRoutes(api *mux.Router) {
	blockchain := api.PathPrefix("/blockchain").Subrouter()
	
	// Blockchain info
	blockchain.HandleFunc("/info", r.blockchainHandler.GetBlockchainInfo).Methods("GET")
	blockchain.HandleFunc("/validate", r.blockchainHandler.ValidateChain).Methods("GET")
	
	// Blocks
	blockchain.HandleFunc("/blocks", r.blockchainHandler.GetAllBlocks).Methods("GET")
	blockchain.HandleFunc("/blocks/{height:[0-9]+}", r.blockchainHandler.GetBlock).Methods("GET")
	blockchain.HandleFunc("/blocks/hash/{hash}", r.blockchainHandler.GetBlockByHash).Methods("GET")
	blockchain.HandleFunc("/blocks/latest", r.blockchainHandler.GetLatestBlock).Methods("GET")
	
	// Transactions
	blockchain.HandleFunc("/transactions", r.blockchainHandler.SearchTransactions).Methods("GET")
	blockchain.HandleFunc("/transactions/{txid}", r.blockchainHandler.GetTransaction).Methods("GET")
}

// setupWalletRoutes configures wallet-related routes
func (r *Router) setupWalletRoutes(api *mux.Router) {
	wallet := api.PathPrefix("/wallet").Subrouter()
	
	// Wallet operations
	wallet.HandleFunc("/create", r.walletHandler.CreateWallet).Methods("POST")
	wallet.HandleFunc("/new", r.walletHandler.CreateWallet).Methods("POST") // Alias
	
	// Balance and info
	wallet.HandleFunc("/balance/{address}", r.walletHandler.GetBalance).Methods("GET")
	wallet.HandleFunc("/info/{address}", r.walletHandler.GetWalletInfo).Methods("GET")
	wallet.HandleFunc("/utxos/{address}", r.walletHandler.ListUTXOs).Methods("GET")
	
	// Transactions
	wallet.HandleFunc("/transaction", r.walletHandler.CreateTransaction).Methods("POST")
}

// setupMiningRoutes configures mining-related routes
func (r *Router) setupMiningRoutes(api *mux.Router) {
	mining := api.PathPrefix("/mining").Subrouter()
	
	// Mining operations
	mining.HandleFunc("/mine", r.miningHandler.MineBlock).Methods("POST")
	mining.HandleFunc("/start", r.miningHandler.StartMining).Methods("POST")
	mining.HandleFunc("/stop", r.miningHandler.StopMining).Methods("POST")
	
	// Mining info
	mining.HandleFunc("/info", r.miningHandler.GetMiningInfo).Methods("GET")
	mining.HandleFunc("/stats", r.miningHandler.GetMiningStats).Methods("GET")
	mining.HandleFunc("/difficulty", r.miningHandler.SetDifficulty).Methods("POST")
}

// setupLegacyRoutes configures legacy routes for backward compatibility
func (r *Router) setupLegacyRoutes(api *mux.Router) {
	// Legacy routes from the original node implementation
	api.HandleFunc("/blocks", r.blockchainHandler.GetAllBlocks).Methods("GET")
	api.HandleFunc("/blocks/{height:[0-9]+}", r.blockchainHandler.GetBlock).Methods("GET")
	api.HandleFunc("/blocks/latest", r.blockchainHandler.GetLatestBlock).Methods("GET")
	api.HandleFunc("/blocks/mine", r.miningHandler.MineBlock).Methods("POST")
	
	api.HandleFunc("/transactions", r.walletHandler.CreateTransaction).Methods("POST")
	api.HandleFunc("/transactions/{txid}", r.blockchainHandler.GetTransaction).Methods("GET")
	
	api.HandleFunc("/wallet/balance/{address}", r.walletHandler.GetBalance).Methods("GET")
	api.HandleFunc("/wallet/new", r.walletHandler.CreateWallet).Methods("POST")
}

// handleNodeInfo provides general node information
func (r *Router) handleNodeInfo(w http.ResponseWriter, req *http.Request) {
	// This delegates to blockchain handler but could include additional node-specific info
	r.blockchainHandler.GetBlockchainInfo(w, req)
}

// handleHealthCheck provides a health check endpoint
func (r *Router) handleHealthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"status":"healthy","service":"blockchain-node"}`))
}