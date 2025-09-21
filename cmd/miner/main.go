package main

import (
	"blockchain-node/pkg/blockchain"
	"blockchain-node/pkg/wallet"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Miner represents a standalone mining node
type Miner struct {
	nodeURL    string
	wallet     *wallet.Wallet
	mining     bool
	stopChan   chan bool
	difficulty uint32
}

// MiningResult represents the result of mining operation
type MiningResult struct {
	Success   bool   `json:"success"`
	BlockHash string `json:"block_hash,omitempty"`
	Height    int64  `json:"height,omitempty"`
	Reward    int64  `json:"reward,omitempty"`
	Error     string `json:"error,omitempty"`
}

// NodeInfo represents node information from API
type NodeInfo struct {
	Height     int64  `json:"height"`
	Difficulty uint32 `json:"difficulty"`
	LastHash   string `json:"last_hash"`
}

// NewMiner creates a new miner instance
func NewMiner(nodeURL string, walletFile string) (*Miner, error) {
	var minerWallet *wallet.Wallet
	var err error
	
	// Try to load existing wallet or create new one
	if walletFile != "" && fileExists(walletFile) {
		fmt.Printf("Loading miner wallet from: %s\n", walletFile)
		minerWallet, err = wallet.LoadFromFile(walletFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load wallet: %v", err)
		}
	} else {
		fmt.Println("Creating new miner wallet...")
		minerWallet, err = wallet.NewWallet()
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet: %v", err)
		}
		
		// Save wallet if filename provided
		if walletFile != "" {
			err = minerWallet.SaveToFile(walletFile)
			if err != nil {
				log.Printf("Warning: failed to save wallet: %v", err)
			} else {
				fmt.Printf("Miner wallet saved to: %s\n", walletFile)
			}
		}
	}
	
	fmt.Printf("Miner wallet address: %s\n", minerWallet.GetAddress())
	
	return &Miner{
		nodeURL:    nodeURL,
		wallet:     minerWallet,
		mining:     false,
		stopChan:   make(chan bool),
		difficulty: 4, // Default difficulty
	}, nil
}

// Start begins the mining process
func (m *Miner) Start(interval time.Duration) error {
	fmt.Printf("Starting miner with interval: %v\n", interval)
	fmt.Printf("Mining rewards will go to: %s\n", m.wallet.GetAddress())
	
	m.mining = true
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Mining loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if m.mining {
				result := m.mineBlock()
				if result.Success {
					fmt.Printf("✅ Block mined successfully! Hash: %s, Height: %d, Reward: %d\n",
						result.BlockHash, result.Height, result.Reward)
				} else {
					fmt.Printf("❌ Mining failed: %s\n", result.Error)
				}
			}
			
		case <-m.stopChan:
			fmt.Println("Mining stopped by internal signal")
			return nil
			
		case <-sigChan:
			fmt.Println("Mining stopped by user signal")
			m.Stop()
			return nil
		}
	}
}

// Stop stops the mining process
func (m *Miner) Stop() {
	fmt.Println("Stopping miner...")
	m.mining = false
	close(m.stopChan)
}

// mineBlock attempts to mine a single block
func (m *Miner) mineBlock() MiningResult {
	// First, get current node info
	nodeInfo, err := m.getNodeInfo()
	if err != nil {
		return MiningResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get node info: %v", err),
		}
	}
	
	fmt.Printf("Current blockchain height: %d, difficulty: %d\n", nodeInfo.Height, nodeInfo.Difficulty)
	
	// Mine a block via API
	resp, err := http.Post(m.nodeURL+"/api/v1/blocks/mine", "application/json", nil)
	if err != nil {
		return MiningResult{
			Success: false,
			Error:   fmt.Sprintf("mining request failed: %v", err),
		}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return MiningResult{
			Success: false,
			Error:   fmt.Sprintf("mining failed with status %d: %s", resp.StatusCode, string(body)),
		}
	}
	
	// Parse the response to get block info
	var block blockchain.Block
	if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
		return MiningResult{
			Success: false,
			Error:   fmt.Sprintf("failed to decode mining response: %v", err),
		}
	}
	
	// Calculate reward from coinbase transaction
	var reward int64
	if len(block.Transactions) > 0 && block.Transactions[0].IsCoinbase() {
		reward = block.Transactions[0].GetTotalOutput()
	}
	
	return MiningResult{
		Success:   true,
		BlockHash: block.Header.Hash,
		Height:    block.Header.Height,
		Reward:    reward,
	}
}

// getNodeInfo retrieves current node information
func (m *Miner) getNodeInfo() (*NodeInfo, error) {
	resp, err := http.Get(m.nodeURL + "/api/v1/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	var info NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	
	return &info, nil
}

// getBalance retrieves miner wallet balance
func (m *Miner) getBalance() (int64, error) {
	url := fmt.Sprintf("%s/api/v1/wallet/balance/%s", m.nodeURL, m.wallet.GetAddress())
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("balance request failed with status: %d", resp.StatusCode)
	}
	
	var balanceResp struct {
		Address string `json:"address"`
		Balance int64  `json:"balance"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return 0, err
	}
	
	return balanceResp.Balance, nil
}

// ShowStats displays current miner statistics
func (m *Miner) ShowStats() {
	fmt.Println("\n=== Miner Statistics ===")
	
	// Get node info
	nodeInfo, err := m.getNodeInfo()
	if err != nil {
		fmt.Printf("Failed to get node info: %v\n", err)
		return
	}
	
	// Get balance
	balance, err := m.getBalance()
	if err != nil {
		fmt.Printf("Failed to get balance: %v\n", err)
		balance = 0
	}
	
	fmt.Printf("Miner Address: %s\n", m.wallet.GetAddress())
	fmt.Printf("Current Balance: %d satoshis (%.8f coins)\n", balance, float64(balance)/100000000)
	fmt.Printf("Blockchain Height: %d\n", nodeInfo.Height)
	fmt.Printf("Current Difficulty: %d\n", nodeInfo.Difficulty)
	fmt.Printf("Mining Status: %v\n", m.mining)
	fmt.Println("========================")
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// displayHelp shows help information
func displayHelp() {
	fmt.Println("Blockchain Miner")
	fmt.Println("Usage: miner [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -node <url>        Node URL (default: http://localhost:8080)")
	fmt.Println("  -wallet <file>     Wallet file (default: miner.wallet)")
	fmt.Println("  -interval <time>   Mining interval (default: 10s)")
	fmt.Println("  -stats             Show statistics and exit")
	fmt.Println("  -help              Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  miner                                    # Start with defaults")
	fmt.Println("  miner -wallet my-miner.wallet           # Use custom wallet")
	fmt.Println("  miner -interval 30s                     # Mine every 30 seconds")
	fmt.Println("  miner -node http://192.168.1.100:8080   # Connect to remote node")
	fmt.Println("  miner -stats                             # Show current stats")
}

func main() {
	// Default values
	nodeURL := "http://localhost:8080"
	walletFile := "miner.wallet"
	interval := 10 * time.Second
	showStats := false
	
	// Parse command line arguments
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-node":
			if i+1 < len(args) {
				nodeURL = args[i+1]
				i++
			}
		case "-wallet":
			if i+1 < len(args) {
				walletFile = args[i+1]
				i++
			}
		case "-interval":
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					interval = d
				} else {
					fmt.Printf("Invalid interval: %s\n", args[i+1])
					os.Exit(1)
				}
				i++
			}
		case "-stats":
			showStats = true
		case "-help":
			displayHelp()
			return
		default:
			fmt.Printf("Unknown option: %s\n", args[i])
			displayHelp()
			os.Exit(1)
		}
	}
	
	// Create miner
	miner, err := NewMiner(nodeURL, walletFile)
	if err != nil {
		log.Fatalf("Failed to create miner: %v", err)
	}
	
	// Show stats and exit if requested
	if showStats {
		miner.ShowStats()
		return
	}
	
	// Test connection to node
	fmt.Printf("Testing connection to node: %s\n", nodeURL)
	if _, err := miner.getNodeInfo(); err != nil {
		log.Fatalf("Failed to connect to node: %v", err)
	}
	fmt.Println("✅ Connected to blockchain node successfully")
	
	// Show initial stats
	miner.ShowStats()
	
	// Start mining
	fmt.Println("\nStarting mining process... (Press Ctrl+C to stop)")
	if err := miner.Start(interval); err != nil {
		log.Fatalf("Mining failed: %v", err)
	}
	
	fmt.Println("Miner stopped.")
}