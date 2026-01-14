package clients

import (
	"context"

	fleetpb "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	inventorypb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FleetClient implements client for Fleet Service
type FleetClient struct {
	client fleetpb.FleetServiceClient
}

func NewFleetClient(addr string) (*FleetClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &FleetClient{client: fleetpb.NewFleetServiceClient(conn)}, nil
}

func (c *FleetClient) GetAsset(ctx context.Context, id, orgID string) (*fleetpb.Asset, error) {
	resp, err := c.client.GetAsset(ctx, &fleetpb.GetAssetRequest{
		Id:             id,
		OrganizationId: orgID,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// InventoryClient implements client for Inventory Service
type InventoryClient struct {
	client inventorypb.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &InventoryClient{client: inventorypb.NewInventoryServiceClient(conn)}, nil
}

func (c *InventoryClient) InitializeTripInventory(ctx context.Context, req *inventorypb.InitializeTripInventoryRequest) (*inventorypb.InitializeTripInventoryResponse, error) {
	resp, err := c.client.InitializeTripInventory(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
