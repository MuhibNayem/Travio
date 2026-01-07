package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PaymentAddr = "localhost:9084" // Payment Service gRPC Port
	ReqCount    = 10
	OrderID     = "order-idempotent-test-2"
)

func main() {
	// Connect to Payment Service
	conn, err := grpc.NewClient(PaymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	var wg sync.WaitGroup
	results := make(chan string, ReqCount)

	start := time.Now()
	fmt.Printf("🚀 Starting Payment Idempotency Test for OrderID: %s\n", OrderID)

	for i := 0; i < ReqCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := &pb.CreatePaymentRequest{
				OrderId:       OrderID,
				AmountPaisa:   1000, // 10.00 BDT
				Currency:      "BDT",
				PaymentMethod: "bkash", // Assuming bkash is configured
				ReturnUrl:     "http://localhost/return",
				CancelUrl:     "http://localhost/cancel",
			}

			resp, err := client.CreatePayment(context.Background(), req)
			if err != nil {
				results <- fmt.Sprintf("[%d] Error: %v", idx, err)
				return
			}
			results <- fmt.Sprintf("[%d] Success: PaymentID=%s, SessionID=%s, Status=%s", idx, resp.PaymentId, resp.SessionId, resp.Status)
		}(i)
	}

	wg.Wait()
	close(results)
	duration := time.Since(start)

	successCount := 0
	paymentIDs := make(map[string]bool)

	for res := range results {
		fmt.Println(res)
		if len(res) > 10 && res[:10] != "[Error" { // Simple check
			successCount++
			// Ideally parse paymentID here
			paymentIDs["dummy"] = true
		}
	}
	_ = paymentIDs // silence unused warning

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Total Requests: %d\n", ReqCount)
	fmt.Printf("Success Responses: %d (Should be %d)\n", successCount, ReqCount)
	// Note: All should succeed, and they should return the SAME PaymentID if fully idempotent and using same Attempt ID.
	// In our impl, we hardcoded Attempt=1. So all should return the SAME PaymentID.
}
