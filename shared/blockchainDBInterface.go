package shared

import (
	"crypto/ed25519"
	"crypto/rsa"
)

// BlockchainDBInterface defines a set of operations for interacting with the blockchain's underlying data storage.
// This interface abstracts the database interactions, allowing for flexibility in the implementation of data persistence.
// It includes methods for managing balances, transactions, blocks, and public keys.

type BlockchainDBInterface interface {
	GetBalance(address string, utxos map[string]UTXO) (int, error)
	SendTransaction(fromAddress, toAddress string, amount int, privKey *rsa.PrivateKey) (bool, error)
	SanitizeAndFormatAddress(address string) (string, error)
	InsertBlock(data []byte, blockNumber int) error
	GetLastBlockData() ([]byte, error)
	RetrievePublicKeyFromAddress(address string) (ed25519.PublicKey, error)
	AddTransaction(tx Transaction) error
	UpdateUTXOs(inputs []UTXO, outputs []UTXO) error
	CreateUTXO(id, txID string, index int, address string, amount int) (UTXO, error)
	GetUTXOsForUser(address string, utxos map[string]UTXO) ([]UTXO, error)
	GetAllUTXOs() (map[string]UTXO, error)
	GetUTXOs() (map[string][]UTXO, error)
	CreateAndSignTransaction(txID string, inputs, outputs []UTXO, privKey *rsa.PrivateKey) (Transaction, error)
	InsertOrUpdateEd25519PublicKey(address string, ed25519PublicKey []byte) error
	RetrieveEd25519PublicKey(address string) (ed25519.PublicKey, error)
	RetrievePrivateKey(address string) ([]byte, error)
	InsertOrUpdatePrivateKey(address string, privateKey []byte) error
}
