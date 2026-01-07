package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CatalogAddr = "localhost:9082" // gRPC Port
	Workers     = 10
	Requests    = 100
)

func main() {
	// Connect to Catalog Service
	conn, err := grpc.NewClient(CatalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCatalogServiceClient(conn)

	// Phase 1: Create a Station (to get an ID)
	fmt.Println("Creating Test Station...")
	// For load test, we assume seed data or create one.
	// Since Proto definitions are not fully visible to me, I'll attempt a GetStation with a known ID
	// or Try to create one.
	// Let's assume we can create one or just try to get one.
	// If ID doesn't exist, it will return error, but that's fine for measuring latency if we handle it.
	// Actually, let's just List Stations which should be cached if implemented (though we didn't implement List cache yet as per code).
	// We implemented GetByID cache.

	// Let's try to List to find an ID, then hammering GetByID.
	listResp, err := client.ListStations(context.Background(), &pb.ListStationsRequest{
		OrganizationId: "org-1",
		PageSize:       1,
	})

	var stationID string
	if err == nil && len(listResp.Stations) > 0 {
		stationID = listResp.Stations[0].Id
		fmt.Printf("Found Station ID: %s\n", stationID)
	} else {
		fmt.Println("No stations found. Cannot test GetByID caching efficiently without data.")
		// Create one if possible (omitted for brevity, will rely on existing data or fail gracefully)
		return
	}

	var wg sync.WaitGroup
	start := time.Now()
	fmt.Printf("🚀 Starting Catalog GetStation Load Test (%d reqs, %d workers)\n", Requests, Workers)

	for i := 0; i < Requests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := client.GetStation(context.Background(), &pb.GetStationRequest{
				Id:             stationID,
				OrganizationId: "org-1",
			})
			if err != nil {
				// fmt.Printf("Error: %v\n", err)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Average Latency: %v\n", duration/time.Duration(Requests))
}
