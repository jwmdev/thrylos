package main

import (
	"context"
	"log"
	"time"

	pb "github.com/thrylos-labs/thrylos" // ensure this import path is correct
	"google.golang.org/grpc"
)

func main() {
	// Connect to the gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBlockchainServiceClient(conn) // Use the correct client constructor from generated protobuf code

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Correctly prepare Inputs and Outputs according to the UTXO structure
	inputs := []*pb.UTXO{{
		TransactionId: "prev-tx-id",
		Index:         0,
		OwnerAddress:  "owner-address-example",
		Amount:        50, // Example amount, ensure this matches the expected logic
	}}
	outputs := []*pb.UTXO{{
		TransactionId: "new-tx-id",
		Index:         0,
		OwnerAddress:  "recipient-address-example",
		Amount:        100, // As an example
	}}

	// Build the transaction
	transaction := &pb.Transaction{
		Id:               "transaction-id",
		Timestamp:        time.Now().Unix(), // Set the current Unix time
		Inputs:           inputs,
		Outputs:          outputs,
		Signature:        "transaction-signature",
		PreviousTxIds:    []string{"prev-tx-id1", "prev-tx-id2"}, // Example previous transaction IDs
		EncryptedAesKey:  []byte("example-encrypted-key"),        // Example encrypted AES key
		EncryptedInputs:  []byte("encrypted-inputs-data"),        // Example encrypted inputs
		EncryptedOutputs: []byte("encrypted-outputs-data"),       // Example encrypted outputs
		Sender:           "sender-address",
	}

	// Create TransactionRequest with the transaction
	r, err := c.SubmitTransaction(ctx, &pb.TransactionRequest{Transaction: transaction})
	if err != nil {
		log.Fatalf("Could not submit transaction: %v", err)
	}
	log.Printf("Transaction Status: %s", r.Status)
}

func getLastBlock(client pb.BlockchainServiceClient) (*pb.BlockResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	resp, err := client.GetLastBlock(ctx, &pb.EmptyRequest{})
	if err != nil {
		return nil, err
	}
	return resp, nil // Adjusted to return the correct type
}

func waitForBlockConfirmation(client pb.BlockchainServiceClient, expectedBlockIndex int) bool {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return false
		case <-ticker.C:
			res, err := getLastBlock(client) // Use the correct handling for getting the last block
			if err != nil {
				continue
			}
			if res.BlockIndex >= int32(expectedBlockIndex) {
				return true
			}
		}
	}
}
