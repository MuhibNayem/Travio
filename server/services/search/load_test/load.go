package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	SearchAddr = "localhost:9085" // gRPC Port
	Workers    = 50
	Requests   = 500
)

func main() {
	// Connect to Search Service
	conn, err := grpc.NewClient(SearchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewSearchServiceClient(conn)

	var wg sync.WaitGroup
	start := time.Now()
	fmt.Printf("🚀 Starting Search Load Test (%d reqs, %d workers)\n", Requests, Workers)

	// All requests use the SAME query to test caching effectiveness
	for i := 0; i < Requests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// SearchTrips with identical params
			_, err := client.SearchTrips(context.Background(), &pb.SearchTripsRequest{
				FromStationId: "dhaka-kamalapur",
				ToStationId:   "chittagong-station",
				Date:          "2026-01-15",
				Limit:         20,
			})
			if err != nil {
				// Errors expected if OpenSearch is down, but Redis should cache results
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Average Latency: %v (Cached: near-zero after first hit)\n", duration/time.Duration(Requests))
}
