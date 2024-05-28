package database

// The database package provides functionalities to interact with a relational database
// for storing and retrieving blockchain data, including blocks, transactions, public keys, and UTXOs.

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/thrylos-labs/thrylos/shared"
	"golang.org/x/crypto/blake2b"
)

// BlockchainDB wraps an SQL database connection and provides methods to interact
// with the blockchain data stored within. It supports operations like inserting or updating public keys,
// retrieving balances based on UTXOs, and adding transactions to the database.

type BlockchainDB struct {
	DB            *badger.DB
	utxos         map[string]shared.UTXO
	Blockchain    shared.BlockchainDBInterface // Use the interface here
	encryptionKey []byte                       // The AES-256 key used for encryption and decryption
}

var (
	db   *badger.DB
	once sync.Once
)

var globalUTXOCache *shared.UTXOCache

func init() {
	var err error
	globalUTXOCache, err = shared.NewUTXOCache(1024, 10000, 0.01) // Adjust size and parameters as needed
	if err != nil {
		panic("Failed to create UTXO cache: " + err.Error())
	}
}

// InitializeDatabase sets up the initial database schema including tables for blocks,
// public keys, and transactions. It ensures the database is ready to store blockchain data.
// InitializeDatabase ensures that BadgerDB is only initialized once
func InitializeDatabase(dataDir string) (*badger.DB, error) {
	var err error
	once.Do(func() {
		// Use dataDir for the database directory
		opts := badger.DefaultOptions(dataDir).WithLogger(nil)
		db, err = badger.Open(opts)
	})
	return db, err
}

// NewBlockchainDB creates a new instance of BlockchainDB with the necessary initialization.
// encryptionKey should be securely provided, e.g., from environment variables or a secure vault service.
func NewBlockchainDB(db *badger.DB, encryptionKey []byte) *BlockchainDB {
	return &BlockchainDB{
		DB:            db,
		utxos:         make(map[string]shared.UTXO),
		encryptionKey: encryptionKey,
	}
}

// encryptData encrypts data using AES-256 GCM.
func (bdb *BlockchainDB) encryptData(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(bdb.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptData decrypts data using AES-256 GCM.
func (bdb *BlockchainDB) decryptData(encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(bdb.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < gcm.NonceSize() {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	nonce, ciphertext := encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (bdb *BlockchainDB) SendTransaction(fromAddress, toAddress string, amount int, privKey *rsa.PrivateKey) (bool, error) {
	// Step 1: Create transaction data
	transactionData := map[string]interface{}{
		"from":   fromAddress,
		"to":     toAddress,
		"amount": amount,
	}

	// Step 2: Serialize transaction data to JSON
	jsonData, err := json.Marshal(transactionData)
	if err != nil {
		return false, fmt.Errorf("error serializing transaction data: %v", err)
	}

	// Step 3: Encrypt the transaction data
	encryptedData, err := bdb.encryptData(jsonData)
	if err != nil {
		return false, fmt.Errorf("error encrypting transaction data: %v", err)
	}

	// Step 4: Sign the encrypted data
	signature, err := rsa.SignPSS(rand.Reader, privKey, crypto.SHA256, bdb.hashData(encryptedData), nil)
	if err != nil {
		return false, fmt.Errorf("error signing transaction: %v", err)
	}

	// Step 5: Store the encrypted transaction and signature in the database atomically
	txn := bdb.DB.NewTransaction(true)
	defer txn.Discard()

	err = bdb.storeTransactionInTxn(txn, encryptedData, signature, fromAddress, toAddress)
	if err != nil {
		return false, fmt.Errorf("error storing transaction in the database: %v", err)
	}

	if err := txn.Commit(); err != nil {
		return false, fmt.Errorf("transaction commit failed: %v", err)
	}

	return true, nil
}

func (bdb *BlockchainDB) storeTransactionInTxn(txn *badger.Txn, encryptedData, signature []byte, fromAddress, toAddress string) error {
	err := txn.Set([]byte(fmt.Sprintf("transaction:%s:%s", fromAddress, toAddress)), encryptedData)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(fmt.Sprintf("signature:%s:%s", fromAddress, toAddress)), signature)
	if err != nil {
		return err
	}
	return nil
}

func (bdb *BlockchainDB) hashData(data []byte) []byte {
	hasher, _ := blake2b.New256(nil)
	hasher.Write(data)
	return hasher.Sum(nil)
}

// InsertOrUpdatePrivateKey stores the private key in the database, encrypting it first.
// InsertOrUpdatePrivateKey stores the private key in the database, encrypting it first.
func (bdb *BlockchainDB) InsertOrUpdatePrivateKey(address string, privateKey []byte) error {
	// Encode and encrypt the private key
	encodedKey := base64.StdEncoding.EncodeToString(privateKey)
	encryptedKey, err := bdb.encryptData([]byte(encodedKey))
	if err != nil {
		return fmt.Errorf("error encrypting private key: %v", err)
	}

	// Start a new transaction
	txn := bdb.DB.NewTransaction(true)
	defer txn.Discard() // Ensure the transaction is discarded if not committed

	// Attempt to set the private key in the database
	if err := txn.Set([]byte("privateKey-"+address), encryptedKey); err != nil {
		return fmt.Errorf("error storing encrypted private key: %v", err)
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %v", err)
	}

	return nil
}

// RetrievePrivateKey retrieves the private key for the given address, decrypting it before returning.
func (bdb *BlockchainDB) RetrievePrivateKey(address string) ([]byte, error) {
	var encryptedKey []byte
	err := bdb.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("privateKey-" + address))
		if err != nil {
			return fmt.Errorf("error retrieving private key: %v", err)
		}
		encryptedKey, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	decryptedData, err := bdb.decryptData(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	return decryptedData, nil
}

// fetching of UTXOs from BadgerDB
func (bdb *BlockchainDB) GetUTXOsForAddress(txn *badger.Txn, address string) ([]shared.UTXO, error) {
	var utxos []shared.UTXO

	prefix := []byte(fmt.Sprintf("utxo-%s-", address))
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Rewind(); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		err := item.Value(func(val []byte) error {
			var utxo shared.UTXO
			if err := json.Unmarshal(val, &utxo); err != nil {
				return err
			}
			utxos = append(utxos, utxo)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return utxos, nil
}

func (bdb *BlockchainDB) RetrieveTransaction(txn *badger.Txn, transactionID string) (*shared.Transaction, error) {
	var tx shared.Transaction

	key := []byte("transaction-" + transactionID)

	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}

	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &tx)
	})
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling transaction: %v", err)
	}

	return &tx, nil
}

func (bdb *BlockchainDB) SanitizeAndFormatAddress(address string) (string, error) {
	trimmedAddress := strings.TrimSpace(address)

	if len(trimmedAddress) == 0 {
		return "", fmt.Errorf("invalid address: empty or only whitespace")
	}

	formattedAddress := strings.ToLower(trimmedAddress)

	if !regexp.MustCompile(`^[a-z0-9]+$`).MatchString(formattedAddress) {
		return "", fmt.Errorf("invalid address: contains invalid characters")
	}

	return formattedAddress, nil
}

func (bdb *BlockchainDB) InsertOrUpdateEd25519PublicKey(address string, ed25519PublicKey []byte) error {
	formattedAddress, err := bdb.SanitizeAndFormatAddress(address)
	if err != nil {
		return err
	}

	// Prepare the data to be inserted or updated
	data, err := json.Marshal(map[string][]byte{"ed25519PublicKey": ed25519PublicKey})
	if err != nil {
		return fmt.Errorf("Failed to marshal public key: %v", err)
	}

	// Start a new transaction for the database operation
	txn := bdb.DB.NewTransaction(true)
	defer txn.Discard() // Ensure that the transaction is discarded if not committed

	// Attempt to set the public key in the database
	if err := txn.Set([]byte("publicKey-"+formattedAddress), data); err != nil {
		return fmt.Errorf("Failed to insert public key for address %s: %v", formattedAddress, err)
	}

	// Commit the transaction
	if err := txn.Commit(); err != nil {
		return fmt.Errorf("Transaction commit failed for public key update for address %s: %v", formattedAddress, err)
	}

	return nil
}

type publicKeyData struct {
	Ed25519PublicKey []byte `json:"ed25519PublicKey"`
}

func (bdb *BlockchainDB) RetrieveEd25519PublicKey(address string) (ed25519.PublicKey, error) {
	formattedAddress, err := bdb.SanitizeAndFormatAddress(address)
	if err != nil {
		return nil, err
	}

	var publicKeyData []byte
	err = bdb.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("publicKey-" + formattedAddress))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			var data map[string][]byte
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			publicKeyData = data["ed25519PublicKey"]
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return ed25519.PublicKey(publicKeyData), nil
}

func (bdb *BlockchainDB) InsertOrUpdatePublicKey(address string, ed25519PublicKey []byte) error {
	data, err := json.Marshal(map[string][]byte{
		"ed25519PublicKey": ed25519PublicKey,
	})
	if err != nil {
		log.Printf("Error marshalling public key data for address %s: %v", address, err)
		return err
	}

	err = bdb.DB.Update(func(txn *badger.Txn) error {
		log.Printf("Attempting to store public key for address %s", address)
		return txn.Set([]byte("publicKey-"+address), data)
	})

	if err != nil {
		log.Printf("Error updating public key in the database for address %s: %v", address, err)
	} else {
		log.Printf("Successfully updated public key in the database for address %s", address)
	}

	return err
}

// RetrievePublicKeyFromAddress fetches the public key for a given blockchain address from the database.
// It is essential for verifying transaction signatures and ensuring the integrity of transactions.
func (bdb *BlockchainDB) RetrievePublicKeyFromAddress(address string) (ed25519.PublicKey, error) {
	log.Printf("Attempting to retrieve public key for address: %s", address)
	var publicKeyData []byte
	err := bdb.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("publicKey-" + address))
		if err != nil {
			log.Printf("Error retrieving key from DB for address %s: %v", address, err)
			return err
		}
		return item.Value(func(val []byte) error {
			publicKeyData = append([]byte{}, val...) // Make a copy of the data
			log.Printf("Retrieved public key data for address %s", address)
			return nil
		})
	})
	if err != nil {
		log.Printf("Failed to retrieve or decode public key for address %s: %v", address, err)
		return nil, err
	}

	var keys map[string][]byte
	if err := json.Unmarshal(publicKeyData, &keys); err != nil {
		log.Printf("Error unmarshalling public key data for address %s: %v", address, err)
		return nil, err
	}

	ed25519Key, ok := keys["ed25519PublicKey"]
	if !ok || len(ed25519Key) == 0 {
		log.Printf("No Ed25519 public key found for address %s", address)
		return nil, fmt.Errorf("no Ed25519 public key found for address %s", address)
	}

	log.Printf("Successfully retrieved and parsed Ed25519 public key for address %s", address)
	return ed25519.PublicKey(ed25519Key), nil
}

// GetBalance calculates the total balance for a given address based on its UTXOs.
// This function is useful for determining the spendable balance of a blockchain account.
func (bdb *BlockchainDB) GetBalance(address string, utxos map[string]shared.UTXO) (int, error) {
	userUTXOs, err := bdb.Blockchain.GetUTXOsForUser(address, utxos)
	if err != nil {
		return 0, err
	}
	var balance int
	for _, utxo := range userUTXOs {
		balance += utxo.Amount
	}
	return balance, nil
}

func (db *BlockchainDB) BeginTransaction() (*shared.TransactionContext, error) {
	txn := db.DB.NewTransaction(true)
	return shared.NewTransactionContext(txn), nil
}

func (db *BlockchainDB) CommitTransaction(txn *shared.TransactionContext) error {
	return txn.Txn.Commit()
}

func (db *BlockchainDB) RollbackTransaction(txn *shared.TransactionContext) error {
	txn.Txn.Discard()
	return nil
}

func (db *BlockchainDB) SetTransaction(txn *shared.TransactionContext, key []byte, value []byte) error {
	return txn.Txn.Set(key, value)
}

// AddTransaction stores a new transaction in the database. It serializes transaction inputs,
// outputs, and the signature for persistent storage.
func (bdb *BlockchainDB) AddTransaction(tx shared.Transaction) error {
	txn := bdb.DB.NewTransaction(true)
	defer txn.Discard()

	txJSON, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("error marshaling transaction: %v", err)
	}

	key := []byte("transaction-" + tx.ID)

	if err := txn.Set(key, txJSON); err != nil {
		return fmt.Errorf("error storing transaction in BadgerDB: %v", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %v", err)
	}

	return nil
}

func (bdb *BlockchainDB) GetUTXOsByAddress(address string) (map[string][]shared.UTXO, error) {
	utxos := make(map[string][]shared.UTXO)

	err := bdb.DB.View(func(txn *badger.Txn) error {
		prefix := []byte("utxo-" + address + "-") // Assuming keys are prefixed with utxo-<address>-
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var utxo shared.UTXO
				if err := json.Unmarshal(val, &utxo); err != nil {
					return fmt.Errorf("error unmarshalling UTXO: %v", err)
				}

				// Extract the UTXO index from the key, format is "utxo-<address>-<index>"
				keyParts := strings.Split(string(item.Key()), "-")
				if len(keyParts) >= 3 {
					index := keyParts[2]
					utxos[index] = append(utxos[index], utxo)
				}

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error retrieving UTXOs for address %s: %v", address, err)
	}

	return utxos, nil
}

func (bdb *BlockchainDB) GetAllUTXOs() (map[string][]shared.UTXO, error) {
	utxos := make(map[string][]shared.UTXO)

	err := bdb.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("utxo-")
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var utxo shared.UTXO
				if err := json.Unmarshal(val, &utxo); err != nil {
					return fmt.Errorf("error unmarshalling UTXO: %v", err)
				}

				key := string(item.Key())
				// Assuming the UTXO ID is part of the key, and the key format is "utxo-<address>-<index>"
				// Here we just use the full key to categorize UTXOs under unique keys
				utxos[key] = append(utxos[key], utxo)

				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error retrieving UTXOs: %v", err)
	}

	return utxos, nil
}

func (bdb *BlockchainDB) GetTransactionByID(txID string, recipientPrivateKey *rsa.PrivateKey) (*shared.Transaction, error) {
	var encryptedTx shared.Transaction // Use your actual transaction structure here

	err := bdb.DB.View(func(txn *badger.Txn) error {
		key := []byte("transaction-" + txID)
		item, err := txn.Get(key)
		if err != nil {
			return fmt.Errorf("error retrieving transaction: %v", err)
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &encryptedTx)
		})
	})

	if err != nil {
		return nil, err
	}

	// encryptedTx.EncryptedAESKey contains the RSA-encrypted AES key
	encryptedKey := encryptedTx.EncryptedAESKey // This field should exist in your encrypted transaction structure

	// Decrypt the encrypted inputs and outputs using the AES key
	decryptedInputsData, err := shared.DecryptTransactionData(encryptedTx.EncryptedInputs, encryptedKey, recipientPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt inputs: %v", err)
	}

	decryptedOutputsData, err := shared.DecryptTransactionData(encryptedTx.EncryptedOutputs, encryptedKey, recipientPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt outputs: %v", err)
	}

	// Deserialize the decrypted data into your actual data structures
	var inputs []shared.UTXO
	var outputs []shared.UTXO
	if err := json.Unmarshal(decryptedInputsData, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %v", err)
	}
	if err := json.Unmarshal(decryptedOutputsData, &outputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %v", err)
	}

	// Construct the decrypted transaction object
	tx := &shared.Transaction{
		ID:        encryptedTx.ID,
		Timestamp: encryptedTx.Timestamp,
		Inputs:    inputs,
		Outputs:   outputs,
		// You can continue populating this struct with the necessary fields...
	}

	return tx, nil
}

func (bdb *BlockchainDB) GetLatestBlockData() ([]byte, error) {
	var latestBlockData []byte

	err := bdb.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true // Iterate in reverse order
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if strings.HasPrefix(string(key), "block-") {
				// We've found the latest block
				err := item.Value(func(val []byte) error {
					// Make a copy of the block data
					latestBlockData = append([]byte(nil), val...)
					return nil
				})
				return err // Return from the View function after finding the latest block
			}
		}

		return fmt.Errorf("no blocks found in the database")
	})

	if err != nil {
		return nil, err
	}

	return latestBlockData, nil
}

func (bdb *BlockchainDB) ProcessTransaction(tx *shared.Transaction) error {
	return bdb.DB.Update(func(txn *badger.Txn) error {
		if err := bdb.updateUTXOsInTxn(txn, tx.Inputs, tx.Outputs); err != nil {
			return err
		}
		if err := bdb.addTransactionInTxn(txn, tx); err != nil {
			return err
		}
		return nil
	})
}

func (bdb *BlockchainDB) updateUTXOsInTxn(txn *badger.Txn, inputs, outputs []shared.UTXO) error {
	for _, input := range inputs {
		key := []byte(fmt.Sprintf("utxo-%s-%d", input.TransactionID, input.Index))
		input.IsSpent = true
		utxoData, err := json.Marshal(input)
		if err != nil {
			return err
		}
		if err := txn.Set(key, utxoData); err != nil {
			return err
		}
		globalUTXOCache.Remove(fmt.Sprintf("%s-%d", input.TransactionID, input.Index))
	}

	for _, output := range outputs {
		key := []byte(fmt.Sprintf("utxo-%s-%d", output.TransactionID, output.Index))
		utxoData, err := json.Marshal(output)
		if err != nil {
			return err
		}
		if err := txn.Set(key, utxoData); err != nil {
			return err
		}
		globalUTXOCache.Add(fmt.Sprintf("%s-%d", output.TransactionID, output.Index), &output)
	}

	return nil
}

func (bdb *BlockchainDB) addTransactionInTxn(txn *badger.Txn, tx *shared.Transaction) error {
	key := []byte("transaction-" + tx.ID)
	value, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return txn.Set(key, value)
}

func (bdb *BlockchainDB) CreateAndStoreUTXO(id, txID string, index int, owner string, amount int) error {
	utxo := shared.CreateUTXO(id, txID, index, owner, amount)

	// Marshal the UTXO object into JSON for storage.
	utxoJSON, err := json.Marshal(utxo)
	if err != nil {
		return fmt.Errorf("error marshalling UTXO: %v", err)
	}

	// Prepare the key for this UTXO entry in the database.
	key := []byte("utxo-" + id)

	// Use BadgerDB transaction to put the UTXO data into the database.
	err = bdb.DB.Update(func(txn *badger.Txn) error {
		return txn.Set(key, utxoJSON)
	})
	if err != nil {
		return fmt.Errorf("error inserting UTXO into BadgerDB: %v", err)
	}

	return nil
}

// UpdateUTXOs updates the UTXOs in the database, marking the inputs as spent and adding new outputs.
func (bdb *BlockchainDB) UpdateUTXOs(inputs []shared.UTXO, outputs []shared.UTXO) error {
	txn := bdb.DB.NewTransaction(true)
	defer txn.Discard()

	for _, input := range inputs {
		err := bdb.MarkUTXOAsSpent(txn, input)
		if err != nil {
			return fmt.Errorf("error marking UTXO as spent: %w", err)
		}
	}

	for _, output := range outputs {
		err := bdb.addNewUTXO(txn, output)
		if err != nil {
			return fmt.Errorf("error adding new UTXO: %w", err)
		}
	}

	return txn.Commit()
}

// MarkUTXOAsSpent marks a UTXO as spent in the database.
func (bdb *BlockchainDB) MarkUTXOAsSpent(txn *badger.Txn, utxo shared.UTXO) error {
	key := []byte(fmt.Sprintf("utxo-%s-%d", utxo.TransactionID, utxo.Index))
	utxo.IsSpent = true
	utxoData, err := json.Marshal(utxo)
	if err != nil {
		return err
	}
	return txn.Set(key, utxoData)
}

// addNewUTXO adds a new UTXO to the database.
func (bdb *BlockchainDB) addNewUTXO(txn *badger.Txn, utxo shared.UTXO) error {
	key := []byte(fmt.Sprintf("utxo-%s-%d", utxo.TransactionID, utxo.Index))
	utxoData, err := json.Marshal(utxo)
	if err != nil {
		return err
	}
	return txn.Set(key, utxoData)
}

// GetUTXOs retrieves all UTXOs for a specific address.
func (bdb *BlockchainDB) GetUTXOs(address string) (map[string][]shared.UTXO, error) {
	utxos := make(map[string][]shared.UTXO)
	err := bdb.DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte("utxo-" + address)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var utxo shared.UTXO
				if err := json.Unmarshal(val, &utxo); err != nil {
					return err
				}
				if !utxo.IsSpent {
					utxos[address] = append(utxos[address], utxo)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return utxos, err
}

func (bdb *BlockchainDB) InsertBlock(blockData []byte, blockNumber int) error {
	key := fmt.Sprintf("block-%d", blockNumber)
	log.Printf("Inserting block %d into database", blockNumber)

	err := bdb.DB.Update(func(txn *badger.Txn) error {
		log.Printf("Storing data at key: %s", key)
		return txn.Set([]byte(key), blockData)
	})

	if err != nil {
		log.Printf("Error inserting block %d: %v", blockNumber, err)
		return fmt.Errorf("error inserting block into BadgerDB: %v", err)
	}

	log.Printf("Block %d inserted successfully", blockNumber)
	return nil
}

// StoreBlock stores serialized block data.
func (bdb *BlockchainDB) StoreBlock(blockData []byte, blockNumber int) error {
	key := fmt.Sprintf("block-%d", blockNumber)
	log.Printf("Storing block %d in the database", blockNumber)

	return bdb.DB.Update(func(txn *badger.Txn) error {
		log.Printf("Storing data at key: %s", key)
		return txn.Set([]byte(key), blockData)
	})
}

// RetrieveBlock retrieves serialized block data by block number.
func (bdb *BlockchainDB) RetrieveBlock(blockNumber int) ([]byte, error) {
	key := fmt.Sprintf("block-%d", blockNumber)
	log.Printf("Retrieving block %d from the database", blockNumber)
	var blockData []byte

	err := bdb.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		blockData, err = item.ValueCopy(nil)
		if err != nil {
			log.Printf("Error retrieving block data from key %s: %v", key, err)
		}
		return err
	})

	if err != nil {
		log.Printf("Failed to retrieve block %d: %v", blockNumber, err)
		return nil, fmt.Errorf("failed to retrieve block data: %v", err)
	}
	log.Printf("Block %d retrieved successfully", blockNumber)
	return blockData, nil
}

func (bdb *BlockchainDB) GetLastBlockData() ([]byte, int, error) {
	var blockData []byte
	var lastIndex int = -1

	err := bdb.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			if strings.HasPrefix(string(key), "block-") {
				blockNumberStr := strings.TrimPrefix(string(key), "block-")
				var parseErr error
				lastIndex, parseErr = strconv.Atoi(blockNumberStr)
				if parseErr != nil {
					return fmt.Errorf("error parsing block number: %v", parseErr)
				}
				blockData, parseErr = item.ValueCopy(nil)
				if parseErr != nil {
					return fmt.Errorf("error retrieving block data: %v", parseErr)
				}
				return nil
			}
		}
		return fmt.Errorf("no blocks found in the database")
	})

	if err != nil {
		return nil, -1, err
	}

	if lastIndex == -1 {
		return nil, -1, fmt.Errorf("no blocks found in the database")
	}

	return blockData, lastIndex, nil
}

func (bdb *BlockchainDB) GetLastBlockIndex() (int, error) {
	var lastIndex int = -1 // Default to -1 to indicate no blocks if none found

	err := bdb.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true // Iterate in reverse order to get the latest block first
		it := txn.NewIterator(opts)
		defer it.Close()

		if it.Rewind(); it.Valid() {
			item := it.Item()
			key := item.Key()
			if strings.HasPrefix(string(key), "block-") {
				blockNumberStr := strings.TrimPrefix(string(key), "block-")
				var parseErr error
				lastIndex, parseErr = strconv.Atoi(blockNumberStr)
				if parseErr != nil {
					log.Printf("Error parsing block number from key %s: %v", key, parseErr)
					return parseErr
				}
				return nil // Stop after the first (latest) block
			}
		}
		return fmt.Errorf("no blocks found in the database")
	})

	if err != nil {
		log.Printf("Failed to retrieve the last block index: %v", err)
		return -1, err // Return -1 when no block is found
	}

	return lastIndex, nil
}

func (bdb *BlockchainDB) CreateAndSignTransaction(txID string, inputs, outputs []shared.UTXO, privKey *rsa.PrivateKey) (shared.Transaction, error) {
	tx := shared.NewTransaction(txID, inputs, outputs)

	// Serialize the transaction without the signature
	txBytes, err := tx.SerializeWithoutSignature()
	if err != nil {
		return tx, fmt.Errorf("error serializing transaction: %v", err) // returning tx, error
	}

	// Hash the serialized transaction using BLAKE2b
	hasher, _ := blake2b.New256(nil)
	hasher.Write(txBytes)
	hashedTx := hasher.Sum(nil)

	// Sign the hashed transaction
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashedTx[:])
	if err != nil {
		return tx, fmt.Errorf("error signing transaction: %v", err) // returning tx, error
	}

	// Encode the signature to base64
	base64Signature := base64.StdEncoding.EncodeToString(signature)

	// Set the encoded signature on the transaction
	tx.Signature = []byte(base64Signature)
	return tx, nil // returning tx, nil
}

func (bdb *BlockchainDB) CreateUTXO(id, txID string, index int, address string, amount int) (shared.UTXO, error) {
	// Use the existing CreateUTXO method to create a UTXO object
	utxo := shared.CreateUTXO(id, txID, index, address, amount)

	// Check if the UTXO ID already exists to avoid duplicates
	if _, exists := bdb.utxos[id]; exists {
		return shared.UTXO{}, fmt.Errorf("UTXO with ID %s already exists", id)
	}

	// Add the created UTXO to the map
	bdb.utxos[id] = utxo

	return utxo, nil
}

func (bdb *BlockchainDB) GetUTXOsForUser(address string, utxos map[string]shared.UTXO) ([]shared.UTXO, error) {
	// I am using provided utxos map as it is one of the parameters in your interface
	// If utxos should be obtained from the BlockchainDB's utxos, replace utxos with bdb.utxos
	userUTXOs := []shared.UTXO{}
	for _, utxo := range utxos {
		if utxo.OwnerAddress == address {
			userUTXOs = append(userUTXOs, utxo)
		}
	}

	return userUTXOs, nil
}
