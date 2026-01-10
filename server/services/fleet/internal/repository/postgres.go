package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/fleet/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
)

type AssetRepository struct {
	DB *sql.DB
}

func NewAssetRepository(db *sql.DB) *AssetRepository {
	return &AssetRepository{DB: db}
}

func (r *AssetRepository) CreateAsset(asset *domain.Asset) error {
	asset.ID = uuid.New().String()
	asset.CreatedAt = time.Now()
	asset.UpdatedAt = time.Now()

	query := `INSERT INTO assets (id, organization_id, name, license_plate, vin, make, model, year, type, status, config, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.DB.Exec(query, asset.ID, asset.OrganizationID, asset.Name, asset.LicensePlate, asset.VIN, asset.Make, asset.Model, asset.Year, asset.Type, asset.Status, asset.Config, asset.CreatedAt, asset.UpdatedAt)
	return err
}

func (r *AssetRepository) GetAsset(id string) (*domain.Asset, error) {
	query := `SELECT id, organization_id, name, license_plate, vin, make, model, year, type, status, config, created_at, updated_at FROM assets WHERE id = $1`
	row := r.DB.QueryRow(query, id)

	var a domain.Asset
	err := row.Scan(&a.ID, &a.OrganizationID, &a.Name, &a.LicensePlate, &a.VIN, &a.Make, &a.Model, &a.Year, &a.Type, &a.Status, &a.Config, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *AssetRepository) UpdateStatus(id, status string) error {
	query := `UPDATE assets SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.DB.Exec(query, status, time.Now(), id)
	return err
}

func (r *AssetRepository) ListAssets(orgID string) ([]*domain.Asset, error) {
	query := `SELECT id, organization_id, name, license_plate, vin, make, model, year, type, status, config, created_at, updated_at FROM assets WHERE organization_id = $1`
	rows, err := r.DB.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*domain.Asset
	for rows.Next() {
		var a domain.Asset
		if err := rows.Scan(&a.ID, &a.OrganizationID, &a.Name, &a.LicensePlate, &a.VIN, &a.Make, &a.Model, &a.Year, &a.Type, &a.Status, &a.Config, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, &a)
	}
	return assets, nil
}
