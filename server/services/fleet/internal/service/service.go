package service

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/fleet/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/fleet/internal/repository"
)

type FleetService struct {
	assetRepo    *repository.AssetRepository
	locationRepo *repository.LocationRepository
}

func NewFleetService(assetRepo *repository.AssetRepository, locRepo *repository.LocationRepository) *FleetService {
	return &FleetService{
		assetRepo:    assetRepo,
		locationRepo: locRepo,
	}
}

// --- Asset Management ---

func (s *FleetService) RegisterAsset(ctx context.Context, name, orgID, plate, vin, make, model string, year int32, aType, status, config string) (*domain.Asset, error) {
	asset := &domain.Asset{
		OrganizationID: orgID,
		Name:           name,
		LicensePlate:   plate,
		VIN:            vin,
		Make:           make,
		Model:          model,
		Year:           year,
		Type:           aType,
		Status:         status,
		Config:         config,
	}

	if err := s.assetRepo.CreateAsset(asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *FleetService) GetAsset(ctx context.Context, id string) (*domain.Asset, error) {
	return s.assetRepo.GetAsset(id)
}

func (s *FleetService) ListAssets(ctx context.Context, orgID string) ([]*domain.Asset, error) {
	return s.assetRepo.ListAssets(orgID)
}

func (s *FleetService) UpdateAssetStatus(ctx context.Context, id, status string) error {
	return s.assetRepo.UpdateStatus(id, status)
}

// --- Location Tracking ---

func (s *FleetService) UpdateLocation(ctx context.Context, assetID, orgID string, lat, lon, speed, heading float64, timestamp time.Time) error {
	loc := &domain.AssetLocation{
		AssetID:        assetID,
		OrganizationID: orgID,
		Latitude:       lat,
		Longitude:      lon,
		Speed:          speed,
		Heading:        heading,
		Timestamp:      timestamp,
	}
	return s.locationRepo.UpdateLocation(loc)
}

func (s *FleetService) GetLatestLocation(ctx context.Context, assetID string) (*domain.AssetLocation, error) {
	return s.locationRepo.GetLatestLocation(assetID)
}

func (s *FleetService) GetLocationHistory(ctx context.Context, assetID string, start, end time.Time) ([]*domain.AssetLocation, error) {
	return s.locationRepo.GetLocationHistory(assetID, start, end)
}
