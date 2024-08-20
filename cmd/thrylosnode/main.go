package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // This is important as it registers pprof handlers with the default mux.
	"os"
	"path/filepath"
	"strings"

	"github.com/supabase-community/supabase-go"
	"github.com/thrylos-labs/thrylos"
	"github.com/thrylos-labs/thrylos/core"
	"github.com/thrylos-labs/thrylos/database"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func loadEnv() (map[string]string, error) {
	env := os.Getenv("ENV")
	var envPath string
	if env == "production" {
		envPath = "../../.env.prod" // The Cert is managed through the droplet
	} else {
		envPath = "../../.env.dev" // Managed through local host
	}
	envFile, err := godotenv.Read(envPath)

	return envFile, err
}

func main() {
	// Load environment variables
	envFile, err := loadEnv()
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Setup Supabase client
	supabaseURL := envFile["SUPABASE_URL"]
	supabasePublicKey := envFile["SUPABASE_PUBLIC_KEY"]
	supabaseClient, err := supabase.NewClient(supabaseURL, supabasePublicKey, nil)

	if err != nil {
		log.Fatalf("Error creating Supabase client: %v", err)
	}

	// Environment variables
	grpcAddress := envFile["GRPC_NODE_ADDRESS"]
	knownPeers := envFile["PEERS"]
	nodeDataDir := envFile["DATA"]
	testnet := envFile["TESTNET"] == "true" // Convert to boolea]
	wasmPath := envFile["WASM_PATH"]
	dataDir := envFile["DATA_DIR"]
	chainID := "0x539" // Default local chain ID (1337 in decimal)
	// domainName := envFile["DOMAIN_NAME")

	if dataDir == "" {
		log.Fatal("DATA_DIR environment variable is not set")
	}

	if testnet {
		fmt.Println("Running in Testnet Mode")
		chainID = "0x5" // Goerli Testnet chain ID
	}

	if wasmPath == "" {
		log.Fatal("WASM_PATH environment variable not set")
	}

	// Fetch and load WebAssembly binary
	response, err := http.Get(wasmPath)
	if err != nil {
		log.Fatalf("Failed to fetch wasm file from %s: %v", wasmPath, err)
	}
	defer response.Body.Close()

	// Load WebAssembly binary

	wasmBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read wasm file: %v", err)
	}

	// Execute the WebAssembly module
	result := thrylos.ExecuteWasm(wasmBytes)
	fmt.Printf("Result from wasm: %d\n", result)

	// Fetch the Base64-encoded AES key from the environment variable
	base64Key := envFile["AES_KEY_ENV_VAR"]
	if base64Key == "" {
		log.Fatal("AES key is not set in environment variables")
	}

	aesKey, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		log.Fatalf("Error decoding AES key: %v", err)
	}

	// Genesis account
	genesisAccount := envFile["GENESIS_ACCOUNT"]
	if genesisAccount == "" {
		log.Fatal("Genesis account is not set in environment variables. Please configure a genesis account before starting.")
	}

	// Get the absolute path of the node data directory
	absPath, err := filepath.Abs(nodeDataDir)
	if err != nil {
		log.Fatalf("Error resolving the absolute path of the blockchain data directory: %v", err)
	}
	log.Printf("Using blockchain data directory: %s", absPath)

	// Initialize the blockchain and database with the AES key

	// Remember to set TestMode to false in your production environment to ensure that the fallback mechanism is never used with real transactions.
	blockchain, _, err := core.NewBlockchain(absPath, aesKey, genesisAccount, true, supabaseClient)
	if err != nil {
		log.Fatalf("Failed to initialize the blockchain at %s: %v", absPath, err)
	}

	// Perform an integrity check on the blockchain
	if !blockchain.CheckChainIntegrity() {
		log.Fatal("Blockchain integrity check failed.")
	} else {
		fmt.Println("Blockchain integrity check passed.")
	}

	// Initialize the database
	blockchainDB, err := database.InitializeDatabase(dataDir)
	if err != nil {
		log.Fatalf("Failed to create blockchain database: %v", err)
	}

	// Initialize a new node with the specified address and known peers
	peersList := []string{}
	if knownPeers != "" {
		peersList = strings.Split(knownPeers, ",")
	}

	node := core.NewNode(grpcAddress, peersList, nodeDataDir, nil)

	node.SetChainID(chainID)

	// Set up routes
	mux := node.SetupRoutes()

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Blockchain status: %s", blockchain.Status())
	})

	// Start background tasks
	node.StartBackgroundTasks()

	// Create a sample HTTP handler
	// mux := http.NewServeMux()

	// Setup and start servers
	setupServers(mux, envFile)

	// Create BlockchainDB instance
	encryptionKey := []byte(aesKey) // This should ideally come from a secure source
	blockchainDatabase := database.NewBlockchainDB(blockchainDB, encryptionKey)

	// Setup and start gRPC server
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddress, err)
	}

	var s *grpc.Server

	if envFile["ENV"] == "development" {
		// Development mode: No TLS
		log.Println("Starting gRPC server in development mode (no TLS)")
		s = grpc.NewServer()
	} else {
		// Production mode: Use TLS
		log.Println("Starting gRPC server in production mode (with TLS)")
		creds := loadTLSCredentials(envFile)
		if err != nil {
			log.Fatalf("Failed to load TLS credentials: %v", err)
		}
		s = grpc.NewServer(grpc.Creds(creds))
	}

	// Setup and start gRPC server
	// lis, err := net.Listen("tcp", grpcAddress)
	// if err != nil {
	// 	log.Fatalf("Failed to listen on %s: %v", grpcAddress, err)
	// }

	log.Printf("Starting gRPC server on %s\n", grpcAddress)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC on %s: %v", grpcAddress, err)
	}
	thrylos.RegisterBlockchainServiceServer(s, &server{db: blockchainDatabase})

	log.Printf("Starting gRPC server on %s\n", grpcAddress)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC on %s: %v", grpcAddress, err)
	}
}

func setupServers(r http.Handler, envFile map[string]string) {
	wsAddress := envFile["WS_ADDRESS"]
	httpAddress := envFile["HTTP_NODE_ADDRESS"]
	isDevelopment := envFile["ENV"] == "development"

	var tlsConfig *tls.Config = nil
	if !isDevelopment {
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{loadCertificate(envFile)},
		}
	}

	// WebSocket server
	wsServer := &http.Server{
		Addr:      wsAddress,
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	// HTTP(S) server
	httpServer := &http.Server{
		Addr:      httpAddress,
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	// Start servers
	go startServer(wsServer, "WebSocket", isDevelopment)
	go startServer(httpServer, "HTTP(S)", isDevelopment)
}

func startServer(server *http.Server, serverType string, isDevelopment bool) {
	var err error
	protocol := "HTTP"
	if !isDevelopment {
		protocol = "HTTPS"
		log.Printf("Starting %s server in production mode (with TLS) on %s\n", serverType, server.Addr)
		err = server.ListenAndServeTLS("", "")
	} else {
		log.Printf("Starting %s server in development mode (no TLS) on %s\n", serverType, server.Addr)
		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start %s %s server: %v", protocol, serverType, err)
	}
}

func loadTLSCredentials(envFile map[string]string) credentials.TransportCredentials {
	var certPath, keyPath string

	// Determine paths based on the environment
	if os.Getenv("ENV") == "production" {
		certPath = envFile["TLS_CERT_PATH"]
		keyPath = envFile["TLS_KEY_PATH"]
	} else { // Default to development paths
		certPath = "../../localhost.pem"
		keyPath = "../../localhost-key.pem"
	}

	// Load the server's certificate and its private key
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("could not load TLS keys: %v", err)
	}

	// Create the credentials and return them
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Optionally set ClientCAs and ClientAuth if you need client certificates for mutual TLS
	}

	return credentials.NewTLS(config)
}

func loadCertificate(envFile map[string]string) tls.Certificate {
	cert, err := tls.LoadX509KeyPair(envFile["CERT_FILE"], envFile["KEY_FILE"])
	if err != nil {
		log.Fatalf("Failed to load TLS certificate: %v", err)
	}
	return cert
}

// Get the blockchain stats: curl http://localhost:50051/get-stats
// Retrieve the genesis block: curl "http://localhost:50051/get-block?id=0"
// Retrieve pending transactions: curl http://localhost:50051/pending-transactions
// Retrive a balance from a specific address: curl "http://localhost:50051/get-balance?address=your_address_here"

// Server-Side Steps
// Blockchain Initialization:
// Initialize the blockchain database and genesis block upon starting the server.
// Load or create stakeholders, UTXOs, and transactions for the genesis block.
// Transaction Handling and Block Management:
// Receive transactions from clients, add to the pending transaction pool, and process them periodically.
// Create new blocks from pending transactions, ensuring transactions are valid, updating the UTXO set, and managing block links.
// Fork Resolution and Integrity Checks:
// Check for forks in the blockchain and resolve by selecting the longest chain.
// Perform regular integrity checks on the blockchain to ensure no tampering or inconsistencies.
