package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	InventoryAddr = "localhost:9083" // Default gRPC port
	TotalReqs     = 1000
	Workers       = 50
	TripID        = "trip-123" // Improve: Ensure this trip exists or mocking handling
	SeatID        = "seat-A1"  // Contention Target
)

func main() {
	fmt.Println("🚀 Starting Inventory Scalability Load Test...")
	fmt.Printf("Target: %s (gRPC)\n", InventoryAddr)

	// Connect to gRPC
	conn, err := grpc.NewClient(InventoryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewInventoryServiceClient(conn)

	var (
		successCount  int32
		failCount     int32
		contentionErr int32
	)

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < TotalReqs/Workers; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

				// Simulate Hold Request for the SAME seat
				_, err := client.HoldSeats(ctx, &pb.HoldSeatsRequest{
					TripId:        TripID,
					FromStationId: "Dhaka",
					ToStationId:   "Chittagong",
					SeatIds:       []string{SeatID},
					UserId:        fmt.Sprintf("user-%d-%d", id, j),
					SessionId:     "session-test",
				})
				cancel()

				if err == nil {
					atomic.AddInt32(&successCount, 1)
				} else {
					atomic.AddInt32(&failCount, 1)
					// Check for contention message (from our code)
					// "seat is currently being processed" or "seat query contention"
					errStr := err.Error()
					if isContentionError(errStr) {
						atomic.AddInt32(&contentionErr, 1)
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Load Test Completed in %v\n", duration)
	fmt.Printf("Total Requests: %d\n", TotalReqs)
	fmt.Printf("Success (Held): %d (Should be 1 if clean run/db reset)\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)
	fmt.Printf("  -> Contention (Redis Pre-Lock): %d\n", contentionErr)
	fmt.Printf("  -> Other Errors: %d\n", failCount-contentionErr)
}

func isContentionError(errStr string) bool {
	return strings.Contains(errStr, "seat query contention") || strings.Contains(errStr, "currently being processed")
}
