package main

import (
	"blockchain-node/pkg/wallet"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// WalletCLI represents the wallet command line interface
type WalletCLI struct {
	nodeURL string
}

// APIResponse represents a generic API response
type APIResponse struct {
	Address    string `json:"address,omitempty"`
	Balance    int64  `json:"balance,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

// TransactionRequest represents a transaction request to the node
type TransactionRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

// NewWalletCLI creates a new wallet CLI instance
func NewWalletCLI(nodeURL string) *WalletCLI {
	return &WalletCLI{
		nodeURL: nodeURL,
	}
}

// createWallet creates a new wallet
func (cli *WalletCLI) createWallet(filename string) {
	fmt.Println("Creating new wallet...")
	
	// Generate new wallet
	w, err := wallet.NewWallet()
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	
	// Save to file if filename provided
	if filename != "" {
		err = w.SaveToFile(filename)
		if err != nil {
			log.Fatalf("Failed to save wallet to file: %v", err)
		}
		fmt.Printf("Wallet saved to: %s\n", filename)
	}
	
	// Display wallet info
	fmt.Printf("Address: %s\n", w.GetAddress())
	fmt.Printf("Public Key: %s\n", w.GetPublicKey())
	fmt.Printf("Private Key: %s\n", w.GetPrivateKey())
	fmt.Println("\nWARNING: Keep your private key secure and never share it!")
}

// loadWallet loads a wallet from file
func (cli *WalletCLI) loadWallet(filename string) *wallet.Wallet {
	w, err := wallet.LoadFromFile(filename)
	if err != nil {
		log.Fatalf("Failed to load wallet: %v", err)
	}
	return w
}

// getBalance gets the balance for an address
func (cli *WalletCLI) getBalance(address string) {
	url := fmt.Sprintf("%s/api/v1/wallet/balance/%s", cli.nodeURL, address)
	
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API request failed with status: %d", resp.StatusCode)
	}
	
	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}
	
	fmt.Printf("Address: %s\n", response.Address)
	fmt.Printf("Balance: %d satoshis (%.8f coins)\n", response.Balance, float64(response.Balance)/100000000)
}

// sendTransaction sends a transaction
func (cli *WalletCLI) sendTransaction(fromAddress, toAddress string, amount int64) {
	// Create transaction request
	txReq := TransactionRequest{
		From:   fromAddress,
		To:     toAddress,
		Amount: amount,
	}
	
	jsonData, err := json.Marshal(txReq)
	if err != nil {
		log.Fatalf("Failed to marshal transaction: %v", err)
	}
	
	// Send to node
	url := fmt.Sprintf("%s/api/v1/transactions", cli.nodeURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("Transaction failed: %s", string(body))
	}
	
	fmt.Println("Transaction sent successfully!")
	
	// Display response
	var response map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	
	if txID, ok := response["id"]; ok {
		fmt.Printf("Transaction ID: %s\n", txID)
	}
}

// listWallets lists all wallet files in current directory
func (cli *WalletCLI) listWallets() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}
	
	fmt.Println("Wallet files found:")
	found := false
	
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".wallet") {
			fmt.Printf("- %s\n", file.Name())
			found = true
		}
	}
	
	if !found {
		fmt.Println("No wallet files found in current directory.")
	}
}

// displayHelp displays help information
func (cli *WalletCLI) displayHelp() {
	fmt.Println("Blockchain Wallet CLI")
	fmt.Println("Usage: wallet [command] [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create [filename]           - Create a new wallet")
	fmt.Println("  balance <address>           - Get balance for an address")
	fmt.Println("  send <from> <to> <amount>   - Send coins (amount in satoshis)")
	fmt.Println("  list                        - List all wallet files")
	fmt.Println("  info <filename>             - Display wallet information")
	fmt.Println("  help                        - Display this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  wallet create my-wallet.wallet")
	fmt.Println("  wallet balance 1A2B3C4D5E...")
	fmt.Println("  wallet send 1A2B3C... 1X2Y3Z... 1000000")
	fmt.Println("  wallet list")
	fmt.Println()
	fmt.Println("Note: 1 coin = 100,000,000 satoshis")
}

// displayWalletInfo displays information about a wallet file
func (cli *WalletCLI) displayWalletInfo(filename string) {
	w := cli.loadWallet(filename)
	
	fmt.Printf("Wallet Information (%s)\n", filename)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Address: %s\n", w.GetAddress())
	fmt.Printf("Public Key: %s\n", w.GetPublicKey())
	
	// Get balance from node
	cli.getBalance(w.GetAddress())
}

func main() {
	// Default node URL
	nodeURL := "http://localhost:8080"
	
	// Check for custom node URL in environment
	if customURL := os.Getenv("BLOCKCHAIN_NODE_URL"); customURL != "" {
		nodeURL = customURL
	}
	
	cli := NewWalletCLI(nodeURL)
	
	// Parse command line arguments
	if len(os.Args) < 2 {
		cli.displayHelp()
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "create":
		filename := ""
		if len(os.Args) > 2 {
			filename = os.Args[2]
		}
		cli.createWallet(filename)
		
	case "balance":
		if len(os.Args) < 3 {
			fmt.Println("Error: Address required")
			fmt.Println("Usage: wallet balance <address>")
			return
		}
		address := os.Args[2]
		cli.getBalance(address)
		
	case "send":
		if len(os.Args) < 5 {
			fmt.Println("Error: Missing arguments")
			fmt.Println("Usage: wallet send <from_address> <to_address> <amount>")
			return
		}
		
		fromAddress := os.Args[2]
		toAddress := os.Args[3]
		amount, err := strconv.ParseInt(os.Args[4], 10, 64)
		if err != nil {
			log.Fatalf("Invalid amount: %v", err)
		}
		
		cli.sendTransaction(fromAddress, toAddress, amount)
		
	case "list":
		cli.listWallets()
		
	case "info":
		if len(os.Args) < 3 {
			fmt.Println("Error: Wallet filename required")
			fmt.Println("Usage: wallet info <filename>")
			return
		}
		filename := os.Args[2]
		cli.displayWalletInfo(filename)
		
	case "help", "--help", "-h":
		cli.displayHelp()
		
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Use 'wallet help' for available commands.")
	}
}