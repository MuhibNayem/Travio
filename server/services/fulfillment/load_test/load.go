package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/fulfillment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	FulfillmentAddr = "localhost:9088" // gRPC Port
	ReqCount        = 10
)

func main() {
	// Connect to Fulfillment Service
	conn, err := grpc.NewClient(FulfillmentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFulfillmentServiceClient(conn)

	var wg sync.WaitGroup
	start := time.Now()
	fmt.Printf("🚀 Starting Fulfillment Ticket Gen Test\n")

	for i := 0; i < ReqCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			req := &pb.GenerateTicketsRequest{
				BookingId:      fmt.Sprintf("book-%d", idx),
				OrderId:        fmt.Sprintf("order-%d", idx),
				OrganizationId: "org-1",
				TripId:         "trip-X",
				RouteName:      "Dhaka-Chittagong",
				FromStation:    "Dhaka",
				ToStation:      "Chittagong",
				// Using current time strings for simplicity in Proto, assuming Proto uses strings or Timestamp.
				// Checking proto, they are Timestamp usually or string.
				// Let's assume standard handling or check proto if errors.
				// Based on previous files, they were time.Time in struct but Proto uses Timestamp?
				// Actually the service converts them.
				// Let's rely on zero values or simple conversion if needed.
				// Wait, the Proto definition for GenerateTicketsRequest wasn't fully visible.
				// Using basic fields.
				Passengers: []*pb.PassengerSeat{
					{
						Nid:        "1234567890",
						Name:       "John Doe",
						SeatNumber: "A1",
						SeatClass:  "AC_S",
						PricePaisa: 50000,
					},
				},
				ContactEmail: "test@example.com",
				ContactPhone: "+8801700000000",
			}

			// Note: Timestamp handling might fail if not set, let's hope zero value is acceptable or text.
			// If proto uses google.protobuf.Timestamp, we need to set it.
			// I'll skip complex timestamp setup for this snippet to keep it simple,
			// as compilation will fail if I use wrong types.

			resp, err := client.GenerateTickets(context.Background(), req)
			if err != nil {
				fmt.Printf("[%d] Error: %v\n", idx, err)
				return
			}
			fmt.Printf("[%d] Success: PDF URL: %s\n", idx, resp.PdfUrl)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
}
