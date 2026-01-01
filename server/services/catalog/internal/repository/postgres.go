package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrStationNotFound = errors.New("station not found")
	ErrRouteNotFound   = errors.New("route not found")
	ErrTripNotFound    = errors.New("trip not found")
	ErrDuplicateCode   = errors.New("code already exists")
)

// StationRepository handles station persistence
type StationRepository struct {
	DB *sql.DB
}

func NewStationRepository(db *sql.DB) *StationRepository {
	return &StationRepository{DB: db}
}

func (r *StationRepository) Create(ctx context.Context, station *domain.Station) error {
	station.ID = uuid.New().String()
	station.CreatedAt = time.Now()
	station.UpdatedAt = time.Now()
	station.Status = domain.StationStatusActive

	amenitiesJSON, _ := json.Marshal(station.Amenities)

	query := `INSERT INTO stations (id, organization_id, code, name, city, state, country, 
			  latitude, longitude, timezone, address, amenities, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.DB.ExecContext(ctx, query,
		station.ID, station.OrganizationID, station.Code, station.Name, station.City, station.State,
		station.Country, station.Latitude, station.Longitude, station.Timezone, station.Address,
		amenitiesJSON, station.Status, station.CreatedAt, station.UpdatedAt,
	)
	return err
}

func (r *StationRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Station, error) {
	query := `SELECT id, organization_id, code, name, city, state, country, latitude, longitude, 
			  timezone, address, amenities, status, created_at, updated_at 
			  FROM stations WHERE id = $1 AND organization_id = $2`

	var station domain.Station
	var amenitiesJSON []byte

	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&station.ID, &station.OrganizationID, &station.Code, &station.Name, &station.City,
		&station.State, &station.Country, &station.Latitude, &station.Longitude, &station.Timezone,
		&station.Address, &amenitiesJSON, &station.Status, &station.CreatedAt, &station.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStationNotFound
		}
		return nil, err
	}

	json.Unmarshal(amenitiesJSON, &station.Amenities)
	return &station, nil
}

func (r *StationRepository) List(ctx context.Context, orgID, city string, limit, offset int) ([]*domain.Station, int, error) {
	var args []interface{}
	whereClause := "WHERE organization_id = $1"
	args = append(args, orgID)

	if city != "" {
		whereClause += " AND city = $2"
		args = append(args, city)
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stations %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch data
	query := fmt.Sprintf(`SELECT id, organization_id, code, name, city, state, country, 
			  latitude, longitude, timezone, address, amenities, status, created_at, updated_at 
			  FROM stations %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		whereClause, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var stations []*domain.Station
	for rows.Next() {
		var s domain.Station
		var amenitiesJSON []byte
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.Code, &s.Name, &s.City, &s.State, &s.Country,
			&s.Latitude, &s.Longitude, &s.Timezone, &s.Address, &amenitiesJSON, &s.Status,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(amenitiesJSON, &s.Amenities)
		stations = append(stations, &s)
	}

	return stations, total, nil
}

func (r *StationRepository) Update(ctx context.Context, station *domain.Station) error {
	station.UpdatedAt = time.Now()
	amenitiesJSON, _ := json.Marshal(station.Amenities)

	query := `UPDATE stations SET name = $1, address = $2, amenities = $3, status = $4, updated_at = $5 
			  WHERE id = $6 AND organization_id = $7`

	result, err := r.DB.ExecContext(ctx, query,
		station.Name, station.Address, amenitiesJSON, station.Status, station.UpdatedAt,
		station.ID, station.OrganizationID,
	)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrStationNotFound
	}
	return nil
}

func (r *StationRepository) Delete(ctx context.Context, id, orgID string) error {
	query := `DELETE FROM stations WHERE id = $1 AND organization_id = $2`
	result, err := r.DB.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrStationNotFound
	}
	return nil
}

// RouteRepository handles route persistence
type RouteRepository struct {
	DB *sql.DB
}

func NewRouteRepository(db *sql.DB) *RouteRepository {
	return &RouteRepository{DB: db}
}

func (r *RouteRepository) Create(ctx context.Context, route *domain.Route) error {
	route.ID = uuid.New().String()
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	route.Status = domain.RouteStatusActive

	stopsJSON, _ := json.Marshal(route.IntermediateStops)

	query := `INSERT INTO routes (id, organization_id, code, name, origin_station_id, destination_station_id,
			  intermediate_stops, distance_km, estimated_duration_minutes, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.DB.ExecContext(ctx, query,
		route.ID, route.OrganizationID, route.Code, route.Name, route.OriginStationID,
		route.DestinationStationID, stopsJSON, route.DistanceKm, route.EstimatedDurationMin,
		route.Status, route.CreatedAt, route.UpdatedAt,
	)
	return err
}

func (r *RouteRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Route, error) {
	query := `SELECT id, organization_id, code, name, origin_station_id, destination_station_id,
			  intermediate_stops, distance_km, estimated_duration_minutes, status, created_at, updated_at
			  FROM routes WHERE id = $1 AND organization_id = $2`

	var route domain.Route
	var stopsJSON []byte

	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&route.ID, &route.OrganizationID, &route.Code, &route.Name, &route.OriginStationID,
		&route.DestinationStationID, &stopsJSON, &route.DistanceKm, &route.EstimatedDurationMin,
		&route.Status, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRouteNotFound
		}
		return nil, err
	}

	json.Unmarshal(stopsJSON, &route.IntermediateStops)
	return &route, nil
}

func (r *RouteRepository) List(ctx context.Context, orgID, originID, destID string, limit, offset int) ([]*domain.Route, int, error) {
	var args []interface{}
	whereClause := "WHERE organization_id = $1"
	args = append(args, orgID)
	argIdx := 2

	if originID != "" {
		whereClause += fmt.Sprintf(" AND origin_station_id = $%d", argIdx)
		args = append(args, originID)
		argIdx++
	}
	if destID != "" {
		whereClause += fmt.Sprintf(" AND destination_station_id = $%d", argIdx)
		args = append(args, destID)
		argIdx++
	}

	// Count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM routes %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch
	query := fmt.Sprintf(`SELECT id, organization_id, code, name, origin_station_id, destination_station_id,
			  intermediate_stops, distance_km, estimated_duration_minutes, status, created_at, updated_at
			  FROM routes %s ORDER BY name ASC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var routes []*domain.Route
	for rows.Next() {
		var rt domain.Route
		var stopsJSON []byte
		if err := rows.Scan(
			&rt.ID, &rt.OrganizationID, &rt.Code, &rt.Name, &rt.OriginStationID,
			&rt.DestinationStationID, &stopsJSON, &rt.DistanceKm, &rt.EstimatedDurationMin,
			&rt.Status, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(stopsJSON, &rt.IntermediateStops)
		routes = append(routes, &rt)
	}

	return routes, total, nil
}

// TripRepository handles trip persistence
type TripRepository struct {
	DB *sql.DB
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{DB: db}
}

func (r *TripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	trip.ID = uuid.New().String()
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	trip.Status = domain.TripStatusScheduled
	trip.AvailableSeats = trip.TotalSeats

	pricingJSON, _ := json.Marshal(trip.Pricing)
	segmentsJSON, _ := json.Marshal(trip.Segments)

	query := `INSERT INTO trips (id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status, segments,
			  created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.DB.ExecContext(ctx, query,
		trip.ID, trip.OrganizationID, trip.RouteID, trip.VehicleID, trip.VehicleType, trip.VehicleClass,
		trip.DepartureTime, trip.ArrivalTime, trip.TotalSeats, trip.AvailableSeats, pricingJSON,
		trip.Status, segmentsJSON, trip.CreatedAt, trip.UpdatedAt,
	)
	return err
}

func (r *TripRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	query := `SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status, segments,
			  created_at, updated_at
			  FROM trips WHERE id = $1 AND organization_id = $2`

	var trip domain.Trip
	var pricingJSON, segmentsJSON []byte

	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&trip.ID, &trip.OrganizationID, &trip.RouteID, &trip.VehicleID, &trip.VehicleType,
		&trip.VehicleClass, &trip.DepartureTime, &trip.ArrivalTime, &trip.TotalSeats,
		&trip.AvailableSeats, &pricingJSON, &trip.Status, &segmentsJSON,
		&trip.CreatedAt, &trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	json.Unmarshal(pricingJSON, &trip.Pricing)
	json.Unmarshal(segmentsJSON, &trip.Segments)
	return &trip, nil
}

func (r *TripRepository) Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error) {
	// Complex search joining trips, routes, stations
	query := `
		SELECT t.id, t.organization_id, t.route_id, t.vehicle_id, t.vehicle_type, t.vehicle_class,
			   t.departure_time, t.arrival_time, t.total_seats, t.available_seats, t.pricing, 
			   t.status, t.segments, t.created_at, t.updated_at
		FROM trips t
		JOIN routes r ON t.route_id = r.id
		JOIN stations origin ON r.origin_station_id = origin.id
		JOIN stations dest ON r.destination_station_id = dest.id
		WHERE origin.city = $1 AND dest.city = $2 
		  AND DATE(t.departure_time) = DATE($3)
		  AND t.status = $4
		  AND t.available_seats > 0
		ORDER BY t.departure_time ASC
		LIMIT $5 OFFSET $6`

	args := []interface{}{originCity, destCity, travelDate, domain.TripStatusScheduled, limit, offset}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trips []*domain.Trip
	for rows.Next() {
		var t domain.Trip
		var pricingJSON, segmentsJSON []byte
		if err := rows.Scan(
			&t.ID, &t.OrganizationID, &t.RouteID, &t.VehicleID, &t.VehicleType, &t.VehicleClass,
			&t.DepartureTime, &t.ArrivalTime, &t.TotalSeats, &t.AvailableSeats, &pricingJSON,
			&t.Status, &segmentsJSON, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(pricingJSON, &t.Pricing)
		json.Unmarshal(segmentsJSON, &t.Segments)
		trips = append(trips, &t)
	}

	// Count total (simplified)
	return trips, len(trips), nil
}

func (r *TripRepository) UpdateStatus(ctx context.Context, id, orgID, status string) error {
	query := `UPDATE trips SET status = $1, updated_at = $2 WHERE id = $3 AND organization_id = $4`
	_, err := r.DB.ExecContext(ctx, query, status, time.Now(), id, orgID)
	return err
}

func (r *TripRepository) DecrementSeats(ctx context.Context, id string, count int) error {
	query := `UPDATE trips SET available_seats = available_seats - $1, updated_at = $2 
			  WHERE id = $3 AND available_seats >= $1`
	result, err := r.DB.ExecContext(ctx, query, count, time.Now(), id)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("not enough seats available")
	}
	return nil
}
