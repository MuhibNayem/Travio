package consumer

import (
	"context"

	catalogpb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	orderpb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CatalogClient interface {
	GetTrip(ctx context.Context, orgID, tripID string) (*catalogpb.Trip, error)
	GetRoute(ctx context.Context, orgID, routeID string) (*catalogpb.Route, error)
	GetStation(ctx context.Context, orgID, stationID string) (*catalogpb.Station, error)
	Close() error
}

type OrderClient interface {
	GetOrder(ctx context.Context, orderID, userID string) (*orderpb.Order, error)
	Close() error
}

type grpcCatalogClient struct {
	conn   *grpc.ClientConn
	client catalogpb.CatalogServiceClient
}

func NewCatalogClient(addr string) (CatalogClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcCatalogClient{
		conn:   conn,
		client: catalogpb.NewCatalogServiceClient(conn),
	}, nil
}

func (c *grpcCatalogClient) GetTrip(ctx context.Context, orgID, tripID string) (*catalogpb.Trip, error) {
	return c.client.GetTrip(ctx, &catalogpb.GetTripRequest{
		Id:             tripID,
		OrganizationId: orgID,
	})
}

func (c *grpcCatalogClient) GetRoute(ctx context.Context, orgID, routeID string) (*catalogpb.Route, error) {
	return c.client.GetRoute(ctx, &catalogpb.GetRouteRequest{
		Id:             routeID,
		OrganizationId: orgID,
	})
}

func (c *grpcCatalogClient) GetStation(ctx context.Context, orgID, stationID string) (*catalogpb.Station, error) {
	return c.client.GetStation(ctx, &catalogpb.GetStationRequest{
		Id:             stationID,
		OrganizationId: orgID,
	})
}

func (c *grpcCatalogClient) Close() error {
	return c.conn.Close()
}

type grpcOrderClient struct {
	conn   *grpc.ClientConn
	client orderpb.OrderServiceClient
}

func NewOrderClient(addr string) (OrderClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcOrderClient{
		conn:   conn,
		client: orderpb.NewOrderServiceClient(conn),
	}, nil
}

func (c *grpcOrderClient) GetOrder(ctx context.Context, orderID, userID string) (*orderpb.Order, error) {
	return c.client.GetOrder(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
}

func (c *grpcOrderClient) Close() error {
	return c.conn.Close()
}
