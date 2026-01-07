package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/MuhibNayem/Travio/server/api/proto/search/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SearchClient struct {
	client pb.SearchServiceClient
	conn   *grpc.ClientConn
}

func NewSearchClient(url string) (*SearchClient, error) {
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to search service: %w", err)
	}

	return &SearchClient{
		client: pb.NewSearchServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *SearchClient) Close() error {
	return c.conn.Close()
}

func (c *SearchClient) SearchTrips(ctx context.Context, query, fromID, toID, date string, limit, offset int) (*pb.SearchTripsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return c.client.SearchTrips(ctx, &pb.SearchTripsRequest{
		Query:         query,
		FromStationId: fromID,
		ToStationId:   toID,
		Date:          date,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
}

func (c *SearchClient) SearchStations(ctx context.Context, query string, limit int) (*pb.SearchStationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	return c.client.SearchStations(ctx, &pb.SearchStationsRequest{
		Query: query,
		Limit: int32(limit),
	})
}
