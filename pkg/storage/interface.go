package storage

import "blockchain-node/pkg/blockchain"

// Storage interface defines methods for blockchain data persistence
type Storage interface {
	// Block operations
	SaveBlock(block *blockchain.Block) error
	GetBlock(hash string) (*blockchain.Block, error)
	GetBlockByHeight(height int64) (*blockchain.Block, error)
	DeleteBlock(hash string) error
	
	// Transaction operations
	SaveTransaction(tx *blockchain.Transaction) error
	GetTransaction(txID string) (*blockchain.Transaction, error)
	DeleteTransaction(txID string) error
	
	// UTXO operations
	SaveUTXO(address string, outputs []blockchain.TxOutput) error
	GetUTXO(address string) ([]blockchain.TxOutput, error)
	DeleteUTXO(address string) error
	
	// Metadata operations
	SaveMetadata(key string, value []byte) error
	GetMetadata(key string) ([]byte, error)
	DeleteMetadata(key string) error
	
	// General operations
	Close() error
	Clear() error
}

// StorageType represents different storage implementations
type StorageType int

const (
	StorageTypeMemory StorageType = iota
	StorageTypeLevelDB
)

// NewStorage creates a new storage instance based on type
func NewStorage(storageType StorageType, path string) (Storage, error) {
	switch storageType {
	case StorageTypeMemory:
		return NewMemoryStorage(), nil
	case StorageTypeLevelDB:
		return NewLevelDBStorage(path)
	default:
		return NewMemoryStorage(), nil
	}
}

// NewLevelDBStorage placeholder - implement if needed
func NewLevelDBStorage(path string) (Storage, error) {
	// For now, return memory storage as fallback
	// You can implement actual LevelDB storage later
	return NewMemoryStorage(), nil
}