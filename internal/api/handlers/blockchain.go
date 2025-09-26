package handlers

import (
	"blockchain-node/pkg/blockchain"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// BlockchainHandler handles blockchain-related HTTP requests
type BlockchainHandler struct {
	blockchain *blockchain.Blockchain
}

// NewBlockchainHandler creates a new blockchain handler
func NewBlockchainHandler(bc *blockchain.Blockchain) *BlockchainHandler {
	return &BlockchainHandler{
		blockchain: bc,
	}
}

// GetBlockchainInfo handles GET /api/v1/blockchain/info
func (h *BlockchainHandler) GetBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	latestBlock := h.blockchain.GetLatestBlock()
	
	info := map[string]interface{}{
		"height":       h.blockchain.GetHeight(),
		"difficulty":   h.blockchain.GetDifficulty(),
		"latest_hash":  latestBlock.Header.Hash,
		"total_blocks": len(h.blockchain.GetAllBlocks()),
		"network":      "testnet", // Could be configurable
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetAllBlocks handles GET /api/v1/blockchain/blocks
func (h *BlockchainHandler) GetAllBlocks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	blocks := h.blockchain.GetAllBlocks()
	
	// Apply pagination if specified
	if limitStr != "" || offsetStr != "" {
		limit := len(blocks)
		offset := 0
		
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}
		
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}
		
		// Apply bounds checking
		if offset >= len(blocks) {
			blocks = []*blockchain.Block{}
		} else {
			end := offset + limit
			if end > len(blocks) {
				end = len(blocks)
			}
			blocks = blocks[offset:end]
		}
	}
	
	response := map[string]interface{}{
		"blocks": blocks,
		"count":  len(blocks),
		"total":  len(h.blockchain.GetAllBlocks()),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetBlock handles GET /api/v1/blockchain/blocks/{height}
func (h *BlockchainHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr := vars["height"]
	
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid block height", http.StatusBadRequest)
		return
	}
	
	block, err := h.blockchain.GetBlock(height)
	if err != nil {
		http.Error(w, fmt.Sprintf("Block not found: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(block); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetBlockByHash handles GET /api/v1/blockchain/blocks/hash/{hash}
func (h *BlockchainHandler) GetBlockByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]
	
	if hash == "" {
		http.Error(w, "Block hash required", http.StatusBadRequest)
		return
	}
	
	block, err := h.blockchain.GetBlockByHash(hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Block not found: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(block); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetLatestBlock handles GET /api/v1/blockchain/blocks/latest
func (h *BlockchainHandler) GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	latestBlock := h.blockchain.GetLatestBlock()
	if latestBlock == nil {
		http.Error(w, "No blocks found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(latestBlock); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ValidateChain handles GET /api/v1/blockchain/validate
func (h *BlockchainHandler) ValidateChain(w http.ResponseWriter, r *http.Request) {
	err := h.blockchain.ValidateChain()
	
	response := map[string]interface{}{
		"valid": err == nil,
	}
	
	if err != nil {
		response["error"] = err.Error()
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetTransaction handles GET /api/v1/blockchain/transactions/{txid}
func (h *BlockchainHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txID := vars["txid"]
	
	if txID == "" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}
	
	tx, err := h.blockchain.GetTransactionByID(txID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transaction not found: %v", err), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tx); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SearchTransactions handles GET /api/v1/blockchain/transactions with query parameters
func (h *BlockchainHandler) SearchTransactions(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 100 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	var transactions []blockchain.Transaction
	blocks := h.blockchain.GetAllBlocks()
	
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			// If address filter is specified, check if transaction involves the address
			if address != "" {
				involved := false
				
				// Check inputs
				for _, input := range tx.Inputs {
					if input.PublicKey == address {
						involved = true
						break
					}
				}
				
				// Check outputs
				if !involved {
					for _, output := range tx.Outputs {
						if output.Address == address {
							involved = true
							break
						}
					}
				}
				
				if !involved {
					continue
				}
			}
			
			transactions = append(transactions, tx)
			
			// Apply limit
			if len(transactions) >= limit {
				break
			}
		}
		
		if len(transactions) >= limit {
			break
		}
	}
	
	response := map[string]interface{}{
		"transactions": transactions,
		"count":        len(transactions),
		"limit":        limit,
	}
	
	if address != "" {
		response["filtered_by_address"] = address
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}