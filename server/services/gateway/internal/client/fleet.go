package client

import (
	"context"
	"time"

	fleetv1 "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

type FleetClient struct {
	conn   *grpc.ClientConn
	client fleetv1.FleetServiceClient
}

func NewFleetClient(address string, tlsCfg TLSConfig) (*FleetClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to fleet service", "address", address, "tls", tlsCfg.CertFile != "")
	return &FleetClient{
		conn:   conn,
		client: fleetv1.NewFleetServiceClient(conn),
	}, nil
}

func (c *FleetClient) Close() error {
	return c.conn.Close()
}

// --- Assets ---

func (c *FleetClient) RegisterAsset(ctx context.Context, req *fleetv1.RegisterAssetRequest) (*fleetv1.Asset, error) {
	return c.client.RegisterAsset(ctx, req)
}

func (c *FleetClient) GetAsset(ctx context.Context, req *fleetv1.GetAssetRequest) (*fleetv1.Asset, error) {
	return c.client.GetAsset(ctx, req)
}

func (c *FleetClient) UpdateAssetStatus(ctx context.Context, req *fleetv1.UpdateAssetStatusRequest) (*fleetv1.Asset, error) {
	return c.client.UpdateAssetStatus(ctx, req)
}

// --- Location ---

func (c *FleetClient) UpdateLocation(ctx context.Context, req *fleetv1.UpdateLocationRequest) (*fleetv1.UpdateLocationResponse, error) {
	return c.client.UpdateLocation(ctx, req)
}

func (c *FleetClient) GetLocation(ctx context.Context, req *fleetv1.GetLocationRequest) (*fleetv1.AssetLocation, error) {
	return c.client.GetLocation(ctx, req)
}
