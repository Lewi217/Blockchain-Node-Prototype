package storage

import (
	"blockchain-node/pkg/blockchain"
	"fmt"
	"sync"
)

// MemoryStorage implements in-memory storage for testing and development
type MemoryStorage struct {
	blocks       map[string]*blockchain.Block         // hash -> block
	blocksByHeight map[int64]*blockchain.Block        // height -> block
	transactions map[string]*blockchain.Transaction   // txID -> transaction
	utxos        map[string][]blockchain.TxOutput     // address -> UTXOs
	metadata     map[string][]byte                    // key -> value
	mutex        sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		blocks:         make(map[string]*blockchain.Block),
		blocksByHeight: make(map[int64]*blockchain.Block),
		transactions:   make(map[string]*blockchain.Transaction),
		utxos:          make(map[string][]blockchain.TxOutput),
		metadata:       make(map[string][]byte),
	}
}

// SaveBlock saves a block to memory storage
func (ms *MemoryStorage) SaveBlock(block *blockchain.Block) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	if block == nil {
		return fmt.Errorf("block cannot be nil")
	}
	
	// Save by hash
	ms.blocks[block.Header.Hash] = block
	
	// Save by height
	ms.blocksByHeight[block.Header.Height] = block
	
	// Save all transactions in the block
	for _, tx := range block.Transactions {
		ms.transactions[tx.ID] = &tx
	}
	
	return nil
}

// GetBlock retrieves a block by hash
func (ms *MemoryStorage) GetBlock(hash string) (*blockchain.Block, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	block, exists := ms.blocks[hash]
	if !exists {
		return nil, fmt.Errorf("block not found: %s", hash)
	}
	
	return block, nil
}

// GetBlockByHeight retrieves a block by height
func (ms *MemoryStorage) GetBlockByHeight(height int64) (*blockchain.Block, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	block, exists := ms.blocksByHeight[height]
	if !exists {
		return nil, fmt.Errorf("block not found at height: %d", height)
	}
	
	return block, nil
}

// DeleteBlock removes a block from storage
func (ms *MemoryStorage) DeleteBlock(hash string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	block, exists := ms.blocks[hash]
	if !exists {
		return fmt.Errorf("block not found: %s", hash)
	}
	
	// Remove from both maps
	delete(ms.blocks, hash)
	delete(ms.blocksByHeight, block.Header.Height)
	
	// Remove transactions
	for _, tx := range block.Transactions {
		delete(ms.transactions, tx.ID)
	}
	
	return nil
}

// SaveTransaction saves a transaction to memory storage
func (ms *MemoryStorage) SaveTransaction(tx *blockchain.Transaction) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	
	ms.transactions[tx.ID] = tx
	return nil
}

// GetTransaction retrieves a transaction by ID
func (ms *MemoryStorage) GetTransaction(txID string) (*blockchain.Transaction, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	tx, exists := ms.transactions[txID]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", txID)
	}
	
	return tx, nil
}

// DeleteTransaction removes a transaction from storage
func (ms *MemoryStorage) DeleteTransaction(txID string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	_, exists := ms.transactions[txID]
	if !exists {
		return fmt.Errorf("transaction not found: %s", txID)
	}
	
	delete(ms.transactions, txID)
	return nil
}

// SaveUTXO saves UTXO data for an address
func (ms *MemoryStorage) SaveUTXO(address string, outputs []blockchain.TxOutput) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	if outputs == nil {
		outputs = []blockchain.TxOutput{}
	}
	
	// Create a copy to prevent external modification
	utxoCopy := make([]blockchain.TxOutput, len(outputs))
	copy(utxoCopy, outputs)
	
	ms.utxos[address] = utxoCopy
	return nil
}

// GetUTXO retrieves UTXO data for an address
func (ms *MemoryStorage) GetUTXO(address string) ([]blockchain.TxOutput, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	utxos, exists := ms.utxos[address]
	if !exists {
		return []blockchain.TxOutput{}, nil // Return empty slice, not error
	}
	
	// Return a copy to prevent external modification
	result := make([]blockchain.TxOutput, len(utxos))
	copy(result, utxos)
	
	return result, nil
}

// DeleteUTXO removes UTXO data for an address
func (ms *MemoryStorage) DeleteUTXO(address string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	delete(ms.utxos, address)
	return nil
}

// SaveMetadata saves metadata
func (ms *MemoryStorage) SaveMetadata(key string, value []byte) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	if value == nil {
		value = []byte{}
	}
	
	// Create a copy to prevent external modification
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	
	ms.metadata[key] = valueCopy
	return nil
}

// GetMetadata retrieves metadata
func (ms *MemoryStorage) GetMetadata(key string) ([]byte, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	value, exists := ms.metadata[key]
	if !exists {
		return nil, fmt.Errorf("metadata not found: %s", key)
	}
	
	// Return a copy to prevent external modification
	result := make([]byte, len(value))
	copy(result, value)
	
	return result, nil
}

// DeleteMetadata removes metadata
func (ms *MemoryStorage) DeleteMetadata(key string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	delete(ms.metadata, key)
	return nil
}

// Close closes the storage (no-op for memory storage)
func (ms *MemoryStorage) Close() error {
	return nil
}

// Clear removes all data from storage
func (ms *MemoryStorage) Clear() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	ms.blocks = make(map[string]*blockchain.Block)
	ms.blocksByHeight = make(map[int64]*blockchain.Block)
	ms.transactions = make(map[string]*blockchain.Transaction)
	ms.utxos = make(map[string][]blockchain.TxOutput)
	ms.metadata = make(map[string][]byte)
	
	return nil
}

// GetAllBlocks returns all blocks (for debugging/testing)
func (ms *MemoryStorage) GetAllBlocks() []*blockchain.Block {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	blocks := make([]*blockchain.Block, 0, len(ms.blocks))
	for _, block := range ms.blocks {
		blocks = append(blocks, block)
	}
	
	return blocks
}

// GetBlockCount returns the number of blocks stored
func (ms *MemoryStorage) GetBlockCount() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	return len(ms.blocks)
}

// GetTransactionCount returns the number of transactions stored
func (ms *MemoryStorage) GetTransactionCount() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	
	return len(ms.transactions)
}