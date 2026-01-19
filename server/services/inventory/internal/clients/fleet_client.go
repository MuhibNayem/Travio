package clients

import (
	"context"

	fleetpb "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
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
