package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

type Blockchain struct {
	blocks     []*Block
	difficulty uint32
	mutex      sync.RWMutex
	utxoSet    map[string][]TxOutput 
}

func NewBlockchain() *Blockchain {
	genesis := NewGenesisBlock()
	
	bc := &Blockchain{
		blocks:     []*Block{genesis},
		difficulty: 4, 
		utxoSet:    make(map[string][]TxOutput),
	}
	
	bc.updateUTXOSet(genesis)
	
	return bc
}

func (bc *Blockchain) AddBlock(transactions []Transaction) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	lastBlock := bc.blocks[len(bc.blocks)-1]
	
	newBlock := NewBlock(transactions, lastBlock.Header.Hash, lastBlock.Header.Height+1)
	
	newBlock.Mine(bc.difficulty)
	
	if err := newBlock.Validate(lastBlock); err != nil {
		return fmt.Errorf("block validation failed: %v", err)
	}
	

	if err := bc.validateTransactions(newBlock); err != nil {
		return fmt.Errorf("transaction validation failed: %v", err)
	}
	
	
	bc.blocks = append(bc.blocks, newBlock)
	
	bc.updateUTXOSet(newBlock)
	
	if newBlock.Header.Height%10 == 0 {
		bc.adjustDifficulty()
	}
	
	return nil
}


func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	if len(bc.blocks) == 0 {
		return nil
	}
	return bc.blocks[len(bc.blocks)-1]
}


func (bc *Blockchain) GetBlock(height int64) (*Block, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	if height < 0 || height >= int64(len(bc.blocks)) {
		return nil, errors.New("block height out of range")
	}
	
	return bc.blocks[height], nil
}


func (bc *Blockchain) GetBlockByHash(hash string) (*Block, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	for _, block := range bc.blocks {
		if block.Header.Hash == hash {
			return block, nil
		}
	}
	
	return nil, errors.New("block not found")
}


func (bc *Blockchain) GetHeight() int64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return int64(len(bc.blocks) - 1)
}


func (bc *Blockchain) GetDifficulty() uint32 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	return bc.difficulty
}


func (bc *Blockchain) ValidateChain() error {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	for i := 1; i < len(bc.blocks); i++ {
		currentBlock := bc.blocks[i]
		previousBlock := bc.blocks[i-1]
		
		if err := currentBlock.Validate(previousBlock); err != nil {
			return fmt.Errorf("block %d validation failed: %v", i, err)
		}
	}
	
	return nil
}

func (bc *Blockchain) GetBalance(address string) int64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	var balance int64
	

	if outputs, exists := bc.utxoSet[address]; exists {
		for _, output := range outputs {
			balance += output.Value
		}
	}
	
	return balance
}


func (bc *Blockchain) FindUTXO(address string) []TxOutput {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	if outputs, exists := bc.utxoSet[address]; exists {
		result := make([]TxOutput, len(outputs))
		copy(result, outputs)
		return result
	}
	
	return []TxOutput{}
}

func (bc *Blockchain) CreateTransaction(from, to string, amount int64) (*Transaction, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	utxos := bc.FindUTXO(from)
	var totalInput int64
	var inputs []TxInput
	
	for _, utxo := range utxos {
		if totalInput >= amount {
			break
		}
		
		input := TxInput{
			TxID:        "", 
			OutputIndex: 0, 
			PublicKey:   from,
		}
		inputs = append(inputs, input)
		totalInput += utxo.Value
	}
	
	if totalInput < amount {
		return nil, errors.New("insufficient funds")
	}
	

	outputs := []TxOutput{
		{Value: amount, Address: to},
	}
	
	
	if totalInput > amount {
		change := totalInput - amount
		outputs = append(outputs, TxOutput{Value: change, Address: from})
	}
	
	transaction := NewTransaction(inputs, outputs)
	return transaction, nil
}


func (bc *Blockchain) validateTransactions(block *Block) error {
	for i, tx := range block.Transactions {
		if i == 0 {
			continue
		}
		
		for _, input := range tx.Inputs {
			if !bc.isValidInput(input) {
				return fmt.Errorf("invalid input in transaction %s", tx.ID)
			}
		}
	}
	
	return nil
}
func (bc *Blockchain) isValidInput(input TxInput) bool {
	return true
}

func (bc *Blockchain) updateUTXOSet(block *Block) {
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, input := range tx.Inputs {
				if outputs, exists := bc.utxoSet[input.PublicKey]; exists {
					if len(outputs) > 0 {
						bc.utxoSet[input.PublicKey] = outputs[1:]
						if len(bc.utxoSet[input.PublicKey]) == 0 {
							delete(bc.utxoSet, input.PublicKey)
						}
					}
				}
			}
		}
		
		for _, output := range tx.Outputs {
			bc.utxoSet[output.Address] = append(bc.utxoSet[output.Address], output)
		}
	}
}

func (bc *Blockchain) adjustDifficulty() {
	if len(bc.blocks) < 2 {
		return
	}
	
	lastBlock := bc.blocks[len(bc.blocks)-1]
	prevBlock := bc.blocks[len(bc.blocks)-2]
	
	timeDiff := lastBlock.Header.Timestamp - prevBlock.Header.Timestamp
	
	if timeDiff < 30 { 
		bc.difficulty++
	} else if timeDiff > 60 && bc.difficulty > 1 {
		bc.difficulty--
	}
}


func (bc *Blockchain) GetAllBlocks() []*Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	result := make([]*Block, len(bc.blocks))
	copy(result, bc.blocks)
	return result
}

func (bc *Blockchain) GetTransactionByID(txID string) (*Transaction, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if tx.ID == txID {
				return &tx, nil
			}
		}
	}
	
	return nil, errors.New("transaction not found")
}

func (bc *Blockchain) String() string {
	return fmt.Sprintf("Blockchain{Height: %d, Difficulty: %d, Blocks: %d}",
		bc.GetHeight(), bc.difficulty, len(bc.blocks))
}