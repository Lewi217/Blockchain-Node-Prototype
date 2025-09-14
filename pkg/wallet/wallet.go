package wallet

import (
	"blockchain-node/pkg/blockchain"
	"blockchain-node/pkg/crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)


type Wallet struct {
	KeyPair *crypto.KeyPair
	Address string
	Balance int64
	mutex   sync.RWMutex
}

type WalletData struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}


func NewWallet() (*Wallet, error) {
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	address := crypto.GenerateAddressFromKeyPair(keyPair, crypto.AddressVersionMainNet)
	
	return &Wallet{
		KeyPair: keyPair,
		Address: address,
		Balance: 0,
	}, nil
}


func NewWalletFromPrivateKey(privateKeyHex string) (*Wallet, error) {
	keyPair, err := crypto.PrivateKeyFromHex(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to create key pair from private key: %v", err)
	}
	
	address := crypto.GenerateAddressFromKeyPair(keyPair, crypto.AddressVersionMainNet)
	
	return &Wallet{
		KeyPair: keyPair,
		Address: address,
		Balance: 0,
	}, nil
}

func (w *Wallet) GetAddress() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.Address
}

func (w *Wallet) GetPublicKey() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.KeyPair.GetPublicKeyHex()
}

func (w *Wallet) GetPrivateKey() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.KeyPair.GetPrivateKeyHex()
}

func (w *Wallet) UpdateBalance(bc *blockchain.Blockchain) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	w.Balance = bc.GetBalance(w.Address)
}

func (w *Wallet) GetBalance() int64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.Balance
}

func (w *Wallet) CreateTransaction(to string, amount int64, bc *blockchain.Blockchain) (*blockchain.Transaction, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	w.Balance = bc.GetBalance(w.Address)
	
	if w.Balance < amount {
		return nil, fmt.Errorf("insufficient funds: have %d, need %d", w.Balance, amount)
	}
	
	tx, err := bc.CreateTransaction(w.Address, to, amount)
	if err != nil {
		return nil, err
	}
	
	err = w.SignTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}
	
	return tx, nil
}

func (w *Wallet) SignTransaction(tx *blockchain.Transaction) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	txCopy := blockchain.Transaction{
		ID:        tx.ID,
		Inputs:    tx.Inputs,
		Outputs:   tx.Outputs,
		Timestamp: tx.Timestamp,
	}
	
	txData, err := json.Marshal(txCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %v", err)
	}
	
	signature, err := crypto.SignHex(txData, w.KeyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create signature: %v", err)
	}
	
	
	tx.Signature = signature
	
	for i := range tx.Inputs {
		tx.Inputs[i].PublicKey = w.GetPublicKey()
		tx.Inputs[i].Signature = signature
	}
	
	return nil
}

func (w *Wallet) VerifyTransaction(tx *blockchain.Transaction) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	txCopy := blockchain.Transaction{
		ID:        tx.ID,
		Inputs:    tx.Inputs,
		Outputs:   tx.Outputs,
		Timestamp: tx.Timestamp,
	}
	
	txData, err := json.Marshal(txCopy)
	if err != nil {
		return false
	}
	

	return crypto.VerifyHex(txData, tx.Signature, w.KeyPair.PublicKey)
}

func (w *Wallet) SaveToFile(filename string) error {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	walletData := WalletData{
		PrivateKey: w.GetPrivateKey(),
		PublicKey:  w.GetPublicKey(),
		Address:    w.Address,
	}
	
	data, err := json.MarshalIndent(walletData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %v", err)
	}
	
	err = ioutil.WriteFile(filename, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write wallet file: %v", err)
	}
	
	return nil
}


func LoadFromFile(filename string) (*Wallet, error) {
	
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("wallet file does not exist: %s", filename)
	}
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %v", err)
	}
	
	var walletData WalletData
	err = json.Unmarshal(data, &walletData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %v", err)
	}
	
	
	wallet, err := NewWalletFromPrivateKey(walletData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet from private key: %v", err)
	}
	
	
	if wallet.Address != walletData.Address {
		return nil, fmt.Errorf("address mismatch in wallet file")
	}
	
	return wallet, nil
}

func (w *Wallet) GetTransactionHistory(bc *blockchain.Blockchain) []blockchain.Transaction {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	var transactions []blockchain.Transaction
	blocks := bc.GetAllBlocks()
	
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			
			involved := false
			
			for _, input := range tx.Inputs {
				if input.PublicKey == w.GetPublicKey() {
					involved = true
					break
				}
			}
			
			
			if !involved {
				for _, output := range tx.Outputs {
					if output.Address == w.Address {
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
	
	return transactions
}


func (w *Wallet) String() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	
	return fmt.Sprintf("Wallet{Address: %s, Balance: %d}",
		w.Address, w.Balance)
}