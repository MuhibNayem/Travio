package handler

import (
	"context"
	"encoding/json"
	"time"

	fleetv1 "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	fleetv1.UnimplementedFleetServiceServer
	service *service.FleetService
}

func NewGRPCHandler(svc *service.FleetService) *GRPCHandler {
	return &GRPCHandler{service: svc}
}

// --- Assets ---

func (h *GRPCHandler) RegisterAsset(ctx context.Context, req *fleetv1.RegisterAssetRequest) (*fleetv1.Asset, error) {
	// Basic validation
	if req.Name == "" || req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "Name and OrganizationID are required")
	}

	// Serialize Config
	configJSON := "{}"
	if req.Config != nil {
		cfgMap := map[string]string{
			"layout_type": req.Config.LayoutType,
			"features":    req.Config.Features,
		}
		bytes, err := json.Marshal(cfgMap)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "Invalid config format")
		}
		configJSON = string(bytes)
	}

	asset, err := h.service.RegisterAsset(ctx, req.Name, req.OrganizationId, req.LicensePlate, req.Vin, req.Make, req.Model, req.Year, req.Type.String(), req.Status.String(), configJSON)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapAssetToProto(asset), nil
}

func (h *GRPCHandler) GetAsset(ctx context.Context, req *fleetv1.GetAssetRequest) (*fleetv1.Asset, error) {
	asset, err := h.service.GetAsset(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return mapAssetToProto(asset), nil
}

func (h *GRPCHandler) UpdateAssetStatus(ctx context.Context, req *fleetv1.UpdateAssetStatusRequest) (*fleetv1.Asset, error) {
	err := h.service.UpdateAssetStatus(ctx, req.Id, req.Status.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Fetch updated asset
	asset, err := h.service.GetAsset(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to refetch asset")
	}
	return mapAssetToProto(asset), nil
}

func (h *GRPCHandler) ListAssets(ctx context.Context, req *fleetv1.ListAssetsRequest) (*fleetv1.ListAssetsResponse, error) {
	assets, err := h.service.ListAssets(ctx, req.OrganizationId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to list assets")
	}

	var protoAssets []*fleetv1.Asset
	for _, a := range assets {
		protoAssets = append(protoAssets, mapAssetToProto(a))
	}

	return &fleetv1.ListAssetsResponse{
		Assets:     protoAssets,
		TotalCount: int32(len(protoAssets)),
	}, nil
}

// --- Tracking ---

func (h *GRPCHandler) UpdateLocation(ctx context.Context, req *fleetv1.UpdateLocationRequest) (*fleetv1.UpdateLocationResponse, error) {
	ts := time.Now()
	if req.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, req.Timestamp)
		if err == nil {
			ts = parsed
		}
	}

	err := h.service.UpdateLocation(ctx, req.AssetId, req.OrganizationId, req.Latitude, req.Longitude, req.Speed, req.Heading, ts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &fleetv1.UpdateLocationResponse{Success: true}, nil
}

func (h *GRPCHandler) GetLocation(ctx context.Context, req *fleetv1.GetLocationRequest) (*fleetv1.AssetLocation, error) {
	loc, err := h.service.GetLatestLocation(ctx, req.AssetId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Location not found")
	}

	return &fleetv1.AssetLocation{
		AssetId:        loc.AssetID,
		OrganizationId: loc.OrganizationID,
		Latitude:       loc.Latitude,
		Longitude:      loc.Longitude,
		Speed:          loc.Speed,
		Heading:        loc.Heading,
		Timestamp:      loc.Timestamp.Format(time.RFC3339),
	}, nil
}

func (h *GRPCHandler) StreamLocations(req *fleetv1.StreamLocationsRequest, stream fleetv1.FleetService_StreamLocationsServer) error {
	return status.Error(codes.Unimplemented, "StreamLocations not implemented yet")
}

// --- Helpers ---

func mapAssetToProto(a *domain.Asset) *fleetv1.Asset {
	// Robust mapping
	var aType fleetv1.AssetType
	if val, ok := fleetv1.AssetType_value[a.Type]; ok {
		aType = fleetv1.AssetType(val)
	} else {
		aType = fleetv1.AssetType_ASSET_TYPE_UNSPECIFIED
	}

	var aStatus fleetv1.AssetStatus
	if val, ok := fleetv1.AssetStatus_value[a.Status]; ok {
		aStatus = fleetv1.AssetStatus(val)
	} else {
		aStatus = fleetv1.AssetStatus_ASSET_STATUS_ACTIVE
	}

	// Config Parsing
	var config *fleetv1.Config
	if a.Config != "" {
		var cfgMap map[string]string
		if err := json.Unmarshal([]byte(a.Config), &cfgMap); err == nil {
			config = &fleetv1.Config{
				LayoutType: cfgMap["layout_type"],
				Features:   cfgMap["features"],
			}
		}
	}

	return &fleetv1.Asset{
		Id:             a.ID,
		OrganizationId: a.OrganizationID,
		Name:           a.Name,
		LicensePlate:   a.LicensePlate,
		Vin:            a.VIN,
		Make:           a.Make,
		Model:          a.Model,
		Year:           a.Year,
		Type:           aType,
		Status:         aStatus,
		Config:         config,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      a.UpdatedAt.Format(time.RFC3339),
	}
}
