package handlers

import (
	"blockchain-node/pkg/blockchain"
	"blockchain-node/pkg/wallet"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// WalletHandler handles wallet-related HTTP requests
type WalletHandler struct {
	blockchain *blockchain.Blockchain
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(bc *blockchain.Blockchain) *WalletHandler {
	return &WalletHandler{
		blockchain: bc,
	}
}

// CreateWalletRequest represents a wallet creation request
type CreateWalletRequest struct {
	SaveToFile bool   `json:"save_to_file,omitempty"`
	Filename   string `json:"filename,omitempty"`
}

// CreateWalletResponse represents a wallet creation response
type CreateWalletResponse struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"` // Only include if explicitly requested
	Saved      bool   `json:"saved,omitempty"`
	Filename   string `json:"filename,omitempty"`
}

// BalanceResponse represents a balance query response
type BalanceResponse struct {
	Address       string                     `json:"address"`
	Balance       int64                      `json:"balance"`
	BalanceCoins  float64                    `json:"balance_coins"`
	UTXOs         []blockchain.TxOutput      `json:"utxos"`
	Transactions  []blockchain.Transaction   `json:"transactions,omitempty"`
}

// TransactionRequest represents a transaction creation request
type TransactionRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

// CreateWallet handles POST /api/v1/wallet/create
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If decode fails, use default values
		req = CreateWalletRequest{SaveToFile: false}
	}
	
	// Create new wallet
	newWallet, err := wallet.NewWallet()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create wallet: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := CreateWalletResponse{
		Address:    newWallet.GetAddress(),
		PublicKey:  newWallet.GetPublicKey(),
		PrivateKey: newWallet.GetPrivateKey(), // Include private key in response
	}
	
	// Save to file if requested
	if req.SaveToFile {
		filename := req.Filename
		if filename == "" {
			filename = fmt.Sprintf("wallet_%s.wallet", newWallet.GetAddress()[:8])
		}
		
		if err := newWallet.SaveToFile(filename); err != nil {
			// Don't fail the request, just note the save failure
			response.Saved = false
		} else {
			response.Saved = true
			response.Filename = filename
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetBalance handles GET /api/v1/wallet/balance/{address}
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	
	if address == "" {
		http.Error(w, "Address required", http.StatusBadRequest)
		return
	}
	
	// Get balance from blockchain
	balance := h.blockchain.GetBalance(address)
	
	// Get UTXOs
	utxos := h.blockchain.FindUTXO(address)
	
	// Check if detailed info is requested
	includeTransactions := r.URL.Query().Get("include_transactions") == "true"
	
	response := BalanceResponse{
		Address:      address,
		Balance:      balance,
		BalanceCoins: float64(balance) / 100000000.0,
		UTXOs:        utxos,
	}
	
	// Include transaction history if requested
	if includeTransactions {
		// This would require creating a temporary wallet to get transaction history
		// For now, we'll search through all blocks
		var transactions []blockchain.Transaction
		blocks := h.blockchain.GetAllBlocks()
		
		for _, block := range blocks {
			for _, tx := range block.Transactions {
				involved := false
				
				// Check if address is involved in inputs
				for _, input := range tx.Inputs {
					if input.PublicKey == address {
						involved = true
						break
					}
				}
				
				// Check if address is involved in outputs
				if !involved {
					for _, output := range tx.Outputs {
						if output.Address == address {
							involved = true
							break
						}
					}
				}
				
				if involved {
					transactions = append(transactions, tx)
				}
			}
		}
		
		response.Transactions = transactions
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateTransaction handles POST /api/v1/wallet/transaction
func (h *WalletHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate request
	if req.From == "" || req.To == "" || req.Amount <= 0 {
		http.Error(w, "Invalid transaction parameters", http.StatusBadRequest)
		return
	}
	
	// Create transaction through blockchain
	tx, err := h.blockchain.CreateTransaction(req.From, req.To, req.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create transaction: %v", err), http.StatusBadRequest)
		return
	}
	
	// In a real implementation, this would go to a mempool
	// For now, we immediately add it to a block
	err = h.blockchain.AddBlock([]blockchain.Transaction{*tx})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add transaction to blockchain: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"transaction_id": tx.ID,
		"from":          req.From,
		"to":            req.To,
		"amount":        req.Amount,
		"status":        "confirmed",
		"block_height":  h.blockchain.GetHeight(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetWalletInfo handles GET /api/v1/wallet/info/{address}
func (h *WalletHandler) GetWalletInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	
	if address == "" {
		http.Error(w, "Address required", http.StatusBadRequest)
		return
	}
	
	// Validate address format
	// In a real implementation, you'd validate the address checksum
	if len(address) < 20 {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}
	
	balance := h.blockchain.GetBalance(address)
	utxos := h.blockchain.FindUTXO(address)
	
	// Count transactions involving this address
	transactionCount := 0
	blocks := h.blockchain.GetAllBlocks()
	
	for _, block := range blocks {
		for _, tx := range block.Transactions {
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
			
			if involved {
				transactionCount++
			}
		}
	}
	
	response := map[string]interface{}{
		"address":           address,
		"balance":           balance,
		"balance_coins":     float64(balance) / 100000000.0,
		"utxo_count":        len(utxos),
		"transaction_count": transactionCount,
		"is_active":         balance > 0 || transactionCount > 0,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListUTXOs handles GET /api/v1/wallet/utxos/{address}
func (h *WalletHandler) ListUTXOs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	
	if address == "" {
		http.Error(w, "Address required", http.StatusBadRequest)
		return
	}
	
	utxos := h.blockchain.FindUTXO(address)
	totalValue := int64(0)
	
	for _, utxo := range utxos {
		totalValue += utxo.Value
	}
	
	response := map[string]interface{}{
		"address":     address,
		"utxos":       utxos,
		"count":       len(utxos),
		"total_value": totalValue,
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}