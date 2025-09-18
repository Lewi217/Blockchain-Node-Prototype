package main

import (
	"blockchain-node/pkg/blockchain"
	"blockchain-node/pkg/storage"
	"blockchain-node/pkg/wallet"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// Node represents a blockchain node
type Node struct {
	blockchain *blockchain.Blockchain
	storage    storage.Storage
	wallet     *wallet.Wallet
}

// NodeInfo represents node information for API responses
type NodeInfo struct {
	Height     int64  `json:"height"`
	Difficulty uint32 `json:"difficulty"`
	LastHash   string `json:"last_hash"`
	NodeWallet string `json:"node_wallet"`
}

// TransactionRequest represents a transaction request
type TransactionRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

// NewNode creates a new blockchain node
func NewNode() (*Node, error) {
	// Initialize storage (using memory storage for simplicity)
	store, err := storage.NewStorage(storage.StorageTypeMemory, "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}
	
	// Create blockchain
	bc := blockchain.NewBlockchain()
	
	// Create node wallet for mining rewards
	nodeWallet, err := wallet.NewWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create node wallet: %v", err)
	}
	
	// Save genesis block to storage
	genesisBlock := bc.GetAllBlocks()[0]
	err = store.SaveBlock(genesisBlock)
	if err != nil {
		log.Printf("Warning: failed to save genesis block to storage: %v", err)
	}
	
	fmt.Printf("Node wallet address: %s\n", nodeWallet.GetAddress())
	
	return &Node{
		blockchain: bc,
		storage:    store,
		wallet:     nodeWallet,
	}, nil
}

// Start starts the blockchain node server
func (n *Node) Start(port string) error {
	router := mux.NewRouter()
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Node info
	api.HandleFunc("/info", n.handleNodeInfo).Methods("GET")
	
	// Blockchain routes
	api.HandleFunc("/blocks", n.handleGetBlocks).Methods("GET")
	api.HandleFunc("/blocks/{height:[0-9]+}", n.handleGetBlock).Methods("GET")
	api.HandleFunc("/blocks/latest", n.handleGetLatestBlock).Methods("GET")
	api.HandleFunc("/blocks/mine", n.handleMineBlock).Methods("POST")
	
	// Transaction routes
	api.HandleFunc("/transactions", n.handleCreateTransaction).Methods("POST")
	api.HandleFunc("/transactions/{txid}", n.handleGetTransaction).Methods("GET")
	
	// Wallet routes
	api.HandleFunc("/wallet/balance/{address}", n.handleGetBalance).Methods("GET")
	api.HandleFunc("/wallet/new", n.handleCreateWallet).Methods("POST")
	
	// Add CORS middleware
	router.Use(corsMiddleware)
	
	fmt.Printf("Starting blockchain node on port %s\n", port)
	fmt.Printf("API endpoints available at http://localhost:%s/api/v1/\n", port)
	
	return http.ListenAndServe(":"+port, router)
}

// handleNodeInfo returns node information
func (n *Node) handleNodeInfo(w http.ResponseWriter, r *http.Request) {
	latestBlock := n.blockchain.GetLatestBlock()
	
	info := NodeInfo{
		Height:     n.blockchain.GetHeight(),
		Difficulty: n.blockchain.GetDifficulty(),
		LastHash:   latestBlock.Header.Hash,
		NodeWallet: n.wallet.GetAddress(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleGetBlocks returns all blocks
func (n *Node) handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := n.blockchain.GetAllBlocks()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

// handleGetBlock returns a specific block by height
func (n *Node) handleGetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	height, err := strconv.ParseInt(vars["height"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid block height", http.StatusBadRequest)
		return
	}
	
	block, err := n.blockchain.GetBlock(height)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// handleGetLatestBlock returns the latest block
func (n *Node) handleGetLatestBlock(w http.ResponseWriter, r *http.Request) {
	latestBlock := n.blockchain.GetLatestBlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestBlock)
}

// handleMineBlock mines a new block
func (n *Node) handleMineBlock(w http.ResponseWriter, r *http.Request) {
	// Create coinbase transaction (mining reward)
	reward := int64(5000000000) // 50 coins * 100000000 satoshis
	coinbase := blockchain.NewCoinbaseTransaction(n.wallet.GetAddress(), reward)
	
	// Add block to blockchain
	err := n.blockchain.AddBlock([]blockchain.Transaction{*coinbase})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to mine block: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get the newly mined block
	latestBlock := n.blockchain.GetLatestBlock()
	
	// Update wallet balance
	n.wallet.UpdateBalance(n.blockchain)
	
	fmt.Printf("Block mined! Height: %d, Hash: %s\n", latestBlock.Header.Height, latestBlock.Header.Hash)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latestBlock)
}

// handleCreateTransaction creates a new transaction
func (n *Node) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate input
	if req.From == "" || req.To == "" || req.Amount <= 0 {
		http.Error(w, "Invalid transaction parameters", http.StatusBadRequest)
		return
	}
	
	// Create transaction
	tx, err := n.blockchain.CreateTransaction(req.From, req.To, req.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusBadRequest)
		return
	}
	
	// In a real implementation, you'd add this to a mempool
	// For now, we'll immediately add it to a new block
	err = n.blockchain.AddBlock([]blockchain.Transaction{*tx})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add transaction to blockchain: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// handleGetTransaction returns a specific transaction
func (n *Node) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txID := vars["txid"]
	
	tx, err := n.blockchain.GetTransactionByID(txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// handleGetBalance returns the balance for an address
func (n *Node) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	
	balance := n.blockchain.GetBalance(address)
	
	response := map[string]interface{}{
		"address": address,
		"balance": balance,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateWallet creates a new wallet
func (n *Node) handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	newWallet, err := wallet.NewWallet()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create wallet: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"address":    newWallet.GetAddress(),
		"public_key": newWallet.GetPublicKey(),
		"private_key": newWallet.GetPrivateKey(), // In production, don't return this!
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Initializing Blockchain Node...")
	
	node, err := NewNode()
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}
	
	// Start mining in background (simple auto-mining every 30 seconds)
	go func() {
		for {
			time.Sleep(30 * time.Second)
			
			// Create coinbase transaction
			reward := int64(5000000000) // 50 coins
			coinbase := blockchain.NewCoinbaseTransaction(node.wallet.GetAddress(), reward)
			
			// Mine block
			err := node.blockchain.AddBlock([]blockchain.Transaction{*coinbase})
			if err != nil {
				log.Printf("Auto-mining failed: %v", err)
			} else {
				fmt.Printf("Auto-mined block at height: %d\n", node.blockchain.GetHeight())
			}
		}
	}()
	
	// Start the HTTP server
	port := "8080"
	if err := node.Start(port); err != nil {
		log.Fatalf("Failed to start node: %v", err)
	}
}