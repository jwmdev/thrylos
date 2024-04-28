package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	pb "github.com/thrylos-labs/thrylos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	// Initialize the listener for the in-memory connection
	lis = bufconn.Listen(bufSize)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func startMockServer() *grpc.Server {
	server := grpc.NewServer()
	pb.RegisterBlockchainServiceServer(server, &mockBlockchainServer{})
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return server
}

func TestSubmitTransaction(t *testing.T) {
	server := startMockServer()
	defer server.Stop()

	// Set up the client connection to the server
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)

	// Prepare and send the request
	transaction := &pb.Transaction{
		Id:        "transaction-id",
		Timestamp: time.Now().Unix(),
		Inputs: []*pb.UTXO{
			{
				TransactionId: "prev-tx-id",
				Index:         0,
				OwnerAddress:  "owner-address-example",
				Amount:        50,
			},
		},
		Outputs: []*pb.UTXO{
			{
				TransactionId: "new-tx-id",
				Index:         0,
				OwnerAddress:  "recipient-address-example",
				Amount:        50,
			},
		},
		Signature: []byte("transaction-signature"),
		Sender:    "sender-address",
	}
	r, err := client.SubmitTransaction(ctx, &pb.TransactionRequest{Transaction: transaction})
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, "Transaction added successfully", r.Status)
}

type mockBlockchainServer struct {
	pb.UnimplementedBlockchainServiceServer
}

func (s *mockBlockchainServer) SubmitTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	return &pb.TransactionResponse{Status: "Transaction added successfully"}, nil
}

// // go test -v -timeout 30s -run ^TestBlockTimeWithGRPC$ github.com/thrylos-labs/thrylos/cmd/thrylosnode

func TestBlockTimeWithGRPC(t *testing.T) {
	server := startMockServer() // Start your in-memory gRPC server
	defer server.Stop()

	conn, err := grpc.Dial("", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewBlockchainServiceClient(conn)

	numTransactions := 1000
	transactionsPerBlock := 100
	var wg sync.WaitGroup
	var blockFinalizeTimes []time.Duration

	start := time.Now()

	for i := 0; i < numTransactions; i += transactionsPerBlock {
		wg.Add(1)
		go func(startIndex int) {
			defer wg.Done()
			blockStartTime := time.Now()

			var blockTransactions []*pb.Transaction
			for j := startIndex; j < startIndex+transactionsPerBlock && j < numTransactions; j++ {
				txID := fmt.Sprintf("tx%d", j)
				transaction := &pb.Transaction{
					Id:        txID,
					Timestamp: time.Now().Unix(),
					Inputs:    []*pb.UTXO{{TransactionId: "tx0", Index: 0, OwnerAddress: "Alice", Amount: 100}},
					Outputs:   []*pb.UTXO{{TransactionId: txID, Index: 0, OwnerAddress: "Bob", Amount: 100}},
				}
				_, edPrivateKey, _ := ed25519.GenerateKey(rand.Reader)
				txBytes, _ := json.Marshal(transaction)
				edSignature := ed25519.Sign(edPrivateKey, txBytes)
				transaction.Signature = edSignature

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				_, err := client.SubmitTransaction(ctx, &pb.TransactionRequest{Transaction: transaction})
				cancel()
				if err != nil {
					t.Errorf("Error submitting transaction %d: %v", j, err)
					continue
				}
				blockTransactions = append(blockTransactions, transaction)
			}

			time.Sleep(time.Millisecond * 500) // Simulate block finalization
			blockEndTime := time.Now()
			blockFinalizeTimes = append(blockFinalizeTimes, blockEndTime.Sub(blockStartTime))
		}(i)
	}

	wg.Wait()

	var totalBlockTime time.Duration
	for _, bt := range blockFinalizeTimes {
		totalBlockTime += bt
	}
	averageBlockTime := totalBlockTime / time.Duration(len(blockFinalizeTimes))

	elapsedOverall := time.Since(start)
	t.Logf("Processed %d transactions into blocks with average block time of %s. Total elapsed time: %s", numTransactions, averageBlockTime, elapsedOverall)
}