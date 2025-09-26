package blockchain

import (
	"encoding/hex"
	"encoding/json"
	"crypto/sha256"
	"time"
	"fmt"
)

type Transaction struct {
	ID        string    `json:"id"`
	Inputs    []TxInput `json:"inputs"`
	Outputs   []TxOutput `json:"outputs"`
	Timestamp int64     `json:"timestamp"`
	Signature string    `json:"signature"`
}

type TxInput struct {
	TxID        string `json:"tx_id"`
	OutputIndex int    `json:"output_index"`
	PublicKey   string `json:"public_key"`
	Signature   string `json:"signature"`
}

type TxOutput struct {
	Value   int64  `json:"value"`
	Address string `json:"address"`
}

func NewTransaction(inputs []TxInput, outputs []TxOutput) *Transaction {
	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.calculateID()
	return tx
}

func NewCoinbaseTransaction(toAddress string, reward int64) *Transaction {
	// Coinbase transaction has no inputs, only outputs
	txOut := TxOutput{
		Value:   reward,
		Address: toAddress,
	}
	
	tx := &Transaction{
		Inputs:    []TxInput{},
		Outputs:   []TxOutput{txOut},
		Timestamp: time.Now().Unix(),
	}
	tx.ID = tx.calculateID()
	return tx
}

func (tx *Transaction) calculateID() string {
	txCopy := Transaction{
		Inputs:    tx.Inputs,
		Outputs:   tx.Outputs,
		Timestamp: tx.Timestamp,
		Signature: tx.Signature,
	}

	data, _ := json.Marshal(txCopy)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 0
}

func (tx *Transaction) GetTotalInput() int64 {
	var total int64
	for _, input := range tx.Inputs {
		if tx.IsCoinbase() {
			return 0
		}
	}
	return total
}

func (tx *Transaction) GetTotalOutput() int64 {
	var total int64
	for _, output := range tx.Outputs {
		total += output.Value
	}
	return total
}

func (tx *Transaction) Validate() error {
	expectedID := tx.calculateID()
	if tx.ID != expectedID {
		return fmt.Errorf("invalid transaction ID")
	}

	if tx.IsCoinbase() {
		if len(tx.Outputs) != 1 {
			return fmt.Errorf("coinbase transaction must have exactly one output")
		}
		return nil
	}
	
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction must have at least one input")
	}
	
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one output")
	}
	
	for _, output := range tx.Outputs {
		if output.Value <= 0 {
			return fmt.Errorf("output value must be positive")
		}
	}
	
	return nil
}

func (tx *Transaction) Serialize() ([]byte, error) {
	return json.Marshal(tx)
}

func DeserializeTransaction(data []byte) (*Transaction, error) {
	var tx Transaction
	err := json.Unmarshal(data, &tx)
	return &tx, err
}

func (tx *Transaction) String() string {
	return fmt.Sprintf("Transaction{ID: %s, Inputs: %d, Outputs: %d, Timestamp: %d}",
		tx.ID, len(tx.Inputs), len(tx.Outputs), tx.Timestamp)
}