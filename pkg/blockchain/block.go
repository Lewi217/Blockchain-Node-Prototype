package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Header       BlockHeader   `json:"header"`
	Transactions []Transaction `json:"transactions"`
}

type BlockHeader struct {
	Version      int32  `json:"version"`       // Block version
	PreviousHash string `json:"previous_hash"` // Hash of previous block
	MerkleRoot   string `json:"merkle_root"`   // Merkle root of transactions
	Timestamp    int64  `json:"timestamp"`     // Block creation timestamp
	Difficulty   uint32 `json:"difficulty"`    // Mining difficulty target
	Nonce        uint64 `json:"nonce"`         // Proof of work nonce
	Hash         string `json:"hash"`          // Block hash
	Height       int64  `json:"height"`        // Block height/index
}


func NewBlock(transactions []Transaction, previousHash string, height int64) *Block {
	block := &Block{
		Header: BlockHeader{
			Version:      1,
			PreviousHash: previousHash,
			Timestamp:    time.Now().Unix(),
			Difficulty:   0, // Will be set during mining
			Nonce:        0,
			Height:       height,
		},
		Transactions: transactions,
	}
	
	block.Header.MerkleRoot = block.calculateMerkleRoot()
	
	return block
}


func NewGenesisBlock() *Block {
	coinbase := NewCoinbaseTransaction("genesis", 5000000000) 
	
	genesis := &Block{
		Header: BlockHeader{
			Version:      1,
			PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
			Timestamp:    time.Now().Unix(),
			Difficulty:   1,
			Nonce:        0,
			Height:       0,
		},
		Transactions: []Transaction{*coinbase},
	}
	
	genesis.Header.MerkleRoot = genesis.calculateMerkleRoot()
	genesis.Header.Hash = genesis.calculateHash()
	
	return genesis
}

func (b *Block) calculateHash() string {
	headerData := fmt.Sprintf("%d%s%s%d%d%d%d",
		b.Header.Version,
		b.Header.PreviousHash,
		b.Header.MerkleRoot,
		b.Header.Timestamp,
		b.Header.Difficulty,
		b.Header.Nonce,
		b.Header.Height,
	)
	
	hash := sha256.Sum256([]byte(headerData))
	return hex.EncodeToString(hash[:])
}

func (b *Block) calculateMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return ""
	}
	
	var txHashes []string
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	
	return b.buildMerkleTree(txHashes)
}


func (b *Block) buildMerkleTree(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	
	if len(hashes) == 1 {
		return hashes[0]
	}
	
	var newLevel []string
	for i := 0; i < len(hashes); i += 2 {
		var combined string
		if i+1 < len(hashes) {
			combined = hashes[i] + hashes[i+1]
		} else {
			combined = hashes[i] + hashes[i]
		}
		
		hash := sha256.Sum256([]byte(combined))
		newLevel = append(newLevel, hex.EncodeToString(hash[:]))
	}
	
	return b.buildMerkleTree(newLevel)
}

func (b *Block) SetHash(hash string) {
	b.Header.Hash = hash
}

func (b *Block) Mine(difficulty uint32) {
	b.Header.Difficulty = difficulty
	target := getTarget(difficulty)
	
	for {
		b.Header.Nonce++
		hash := b.calculateHash()
		
		if isHashValid(hash, target) {
			b.Header.Hash = hash
			fmt.Printf("Block mined: %s\n", hash)
			break
		}
	}
}


func getTarget(difficulty uint32) string {
	target := ""
	for i := uint32(0); i < difficulty; i++ {
		target += "0"
	}
	return target
}

func isHashValid(hash, target string) bool {
	return hash[:len(target)] == target
}

func (b *Block) Validate(previousBlock *Block) error {
	expectedHash := b.calculateHash()
	if b.Header.Hash != expectedHash {
		return fmt.Errorf("invalid block hash")
	}
	
	if previousBlock != nil {
		if b.Header.PreviousHash != previousBlock.Header.Hash {
			return fmt.Errorf("invalid previous hash")
		}
	
		if b.Header.Height != previousBlock.Header.Height+1 {
			return fmt.Errorf("invalid block height")
		}
	}
	
	expectedMerkleRoot := b.calculateMerkleRoot()
	if b.Header.MerkleRoot != expectedMerkleRoot {
		return fmt.Errorf("invalid merkle root")
	}
	

	if len(b.Transactions) == 0 {
		return fmt.Errorf("block must contain at least one transaction")
	}
	
	
	if !b.Transactions[0].IsCoinbase() {
		return fmt.Errorf("first transaction must be coinbase")
	}
	

	coinbaseCount := 0
	for _, tx := range b.Transactions {
		if tx.IsCoinbase() {
			coinbaseCount++
		}
		
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid transaction: %v", err)
		}
	}
	
	if coinbaseCount != 1 {
		return fmt.Errorf("block must contain exactly one coinbase transaction")
	}
	
	return nil
}


func (b *Block) GetSize() int {
	data, _ := b.Serialize()
	return len(data)
}

func (b *Block) Serialize() ([]byte, error) {
	return json.Marshal(b)
}


func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	err := json.Unmarshal(data, &block)
	return &block, err
}

func (b *Block) String() string {
	return fmt.Sprintf("Block{Hash: %s, Height: %d, Transactions: %d, Timestamp: %d}",
		b.Header.Hash, b.Header.Height, len(b.Transactions), b.Header.Timestamp)
}