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

// PostgresStationRepository handles station persistence
type PostgresStationRepository struct {
	DB *sql.DB
}

func NewStationRepository(db *sql.DB) *PostgresStationRepository {
	return &PostgresStationRepository{DB: db}
}

func (r *PostgresStationRepository) Create(ctx context.Context, station *domain.Station) error {
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

func (r *PostgresStationRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Station, error) {
	var (
		query string
		args  []interface{}
	)

	if orgID == "" {
		query = `SELECT id, organization_id, code, name, city, state, country, latitude, longitude, 
			  timezone, address, amenities, status, created_at, updated_at 
			  FROM stations WHERE id = $1`
		args = append(args, id)
	} else {
		// Allow finding the station if it belongs to the org OR if it is a public station
		query = `SELECT id, organization_id, code, name, city, state, country, latitude, longitude, 
			  timezone, address, amenities, status, created_at, updated_at 
			  FROM stations WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL)`
		args = append(args, id, orgID)
	}

	var station domain.Station
	var amenitiesJSON []byte
	var (
		nullOrgID    sql.NullString
		nullState    sql.NullString
		nullAddress  sql.NullString
		nullTimezone sql.NullString
		nullLat      sql.NullFloat64
		nullLong     sql.NullFloat64
	)

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(
		&station.ID, &nullOrgID, &station.Code, &station.Name, &station.City,
		&nullState, &station.Country, &nullLat, &nullLong, &nullTimezone,
		&nullAddress, &amenitiesJSON, &station.Status, &station.CreatedAt, &station.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStationNotFound
		}
		return nil, err
	}

	if nullOrgID.Valid {
		station.OrganizationID = nullOrgID.String
	}
	if nullState.Valid {
		station.State = nullState.String
	}
	if nullAddress.Valid {
		station.Address = nullAddress.String
	}
	if nullTimezone.Valid {
		station.Timezone = nullTimezone.String
	}
	if nullLat.Valid {
		station.Latitude = nullLat.Float64
	}
	if nullLong.Valid {
		station.Longitude = nullLong.Float64
	}

	json.Unmarshal(amenitiesJSON, &station.Amenities)
	return &station, nil
}

func (r *PostgresStationRepository) List(ctx context.Context, orgID, city string, limit, offset int) ([]*domain.Station, int, error) {
	var args []interface{}
	var whereClause string
	argIdx := 1

	// Build WHERE clause dynamically
	if orgID != "" {
		whereClause = fmt.Sprintf("WHERE organization_id = $%d", argIdx)
		args = append(args, orgID)
		argIdx++
	} else {
		// If no orgID specified, list all stations
		whereClause = "WHERE 1=1"
	}

	if city != "" {
		whereClause += fmt.Sprintf(" AND city = $%d", argIdx)
		args = append(args, city)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stations %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch data
	query := fmt.Sprintf(`SELECT id, organization_id, code, name, city, state, country, 
		  latitude, longitude, timezone, address, amenities, status, created_at, updated_at 
		  FROM stations %s ORDER BY name ASC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
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
		var orgID, state, address, timezone sql.NullString
		var latitude, longitude sql.NullFloat64
		if err := rows.Scan(
			&s.ID, &orgID, &s.Code, &s.Name, &s.City, &state, &s.Country,
			&latitude, &longitude, &timezone, &address, &amenitiesJSON, &s.Status,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			s.OrganizationID = orgID.String
		}
		if state.Valid {
			s.State = state.String
		}
		if address.Valid {
			s.Address = address.String
		}
		if timezone.Valid {
			s.Timezone = timezone.String
		}
		if latitude.Valid {
			s.Latitude = latitude.Float64
		}
		if longitude.Valid {
			s.Longitude = longitude.Float64
		}
		json.Unmarshal(amenitiesJSON, &s.Amenities)
		stations = append(stations, &s)
	}

	return stations, total, nil
}

func (r *PostgresStationRepository) Update(ctx context.Context, station *domain.Station) error {
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

func (r *PostgresStationRepository) Delete(ctx context.Context, id, orgID string) error {
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

// PostgresRouteRepository handles route persistence
type PostgresRouteRepository struct {
	DB *sql.DB
}

func NewRouteRepository(db *sql.DB) *PostgresRouteRepository {
	return &PostgresRouteRepository{DB: db}
}

func (r *PostgresRouteRepository) Create(ctx context.Context, route *domain.Route) error {
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

func (r *PostgresRouteRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Route, error) {
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

func (r *PostgresRouteRepository) List(ctx context.Context, orgID, originID, destID string, limit, offset int) ([]*domain.Route, int, error) {
	var args []interface{}
	var whereClause string
	argIdx := 1

	// Build WHERE clause dynamically
	if orgID != "" {
		whereClause = fmt.Sprintf("WHERE organization_id = $%d", argIdx)
		args = append(args, orgID)
		argIdx++
	} else {
		whereClause = "WHERE 1=1"
	}

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
		var orgID sql.NullString
		if err := rows.Scan(
			&rt.ID, &orgID, &rt.Code, &rt.Name, &rt.OriginStationID,
			&rt.DestinationStationID, &stopsJSON, &rt.DistanceKm, &rt.EstimatedDurationMin,
			&rt.Status, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			rt.OrganizationID = orgID.String
		}
		json.Unmarshal(stopsJSON, &rt.IntermediateStops)
		routes = append(routes, &rt)
	}

	return routes, total, nil
}

// PostgresTripRepository handles trip persistence
type PostgresTripRepository struct {
	DB *sql.DB
}

func NewTripRepository(db *sql.DB) *PostgresTripRepository {
	return &PostgresTripRepository{DB: db}
}

func (r *PostgresTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	trip.ID = uuid.New().String()
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	trip.Status = domain.TripStatusScheduled
	trip.AvailableSeats = trip.TotalSeats

	pricingJSON, _ := json.Marshal(trip.Pricing)

	query := `INSERT INTO trips (id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status,
			  created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.DB.ExecContext(ctx, query,
		trip.ID, trip.OrganizationID, trip.RouteID, trip.VehicleID, trip.VehicleType, trip.VehicleClass,
		trip.DepartureTime, trip.ArrivalTime, trip.TotalSeats, trip.AvailableSeats, pricingJSON,
		trip.Status, trip.CreatedAt, trip.UpdatedAt,
	)
	return err
}

func (r *PostgresTripRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	query := `SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status,
			  created_at, updated_at
			  FROM trips WHERE id = $1 AND organization_id = $2`

	var trip domain.Trip
	var pricingJSON []byte

	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&trip.ID, &trip.OrganizationID, &trip.RouteID, &trip.VehicleID, &trip.VehicleType,
		&trip.VehicleClass, &trip.DepartureTime, &trip.ArrivalTime, &trip.TotalSeats,
		&trip.AvailableSeats, &pricingJSON, &trip.Status,
		&trip.CreatedAt, &trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, err
	}

	json.Unmarshal(pricingJSON, &trip.Pricing)
	// Segments not in DB yet
	return &trip, nil
}

func (r *PostgresTripRepository) List(ctx context.Context, orgID, routeID string, limit, offset int) ([]*domain.Trip, int, error) {
	var args []interface{}
	var whereClause string
	argIdx := 1

	// Build WHERE clause dynamically
	if orgID != "" {
		whereClause = fmt.Sprintf("WHERE organization_id = $%d", argIdx)
		args = append(args, orgID)
		argIdx++
	} else {
		whereClause = "WHERE 1=1"
	}

	if routeID != "" {
		whereClause += fmt.Sprintf(" AND route_id = $%d", argIdx)
		args = append(args, routeID)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM trips %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch data
	query := fmt.Sprintf(`SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class,
		  departure_time, arrival_time, total_seats, available_seats, pricing, status,
		  created_at, updated_at
		  FROM trips %s ORDER BY departure_time DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var trips []*domain.Trip
	for rows.Next() {
		var t domain.Trip
		var pricingJSON []byte
		var orgID sql.NullString
		if err := rows.Scan(
			&t.ID, &orgID, &t.RouteID, &t.VehicleID, &t.VehicleType, &t.VehicleClass,
			&t.DepartureTime, &t.ArrivalTime, &t.TotalSeats, &t.AvailableSeats, &pricingJSON,
			&t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			t.OrganizationID = orgID.String
		}
		json.Unmarshal(pricingJSON, &t.Pricing)
		// Segments not in DB
		trips = append(trips, &t)
	}

	return trips, total, nil
}

func (r *PostgresTripRepository) Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error) {
	// Complex search joining trips, routes, stations
	query := `
		SELECT t.id, t.organization_id, t.route_id, t.vehicle_id, t.vehicle_type, t.vehicle_class,
			   t.departure_time, t.arrival_time, t.total_seats, t.available_seats, t.pricing, 
			   t.status, t.created_at, t.updated_at
		FROM trips t
		JOIN routes r ON t.route_id = r.id
		JOIN stations origin ON r.origin_station_id = origin.id
		JOIN stations dest ON r.destination_station_id = dest.id
		WHERE (origin.city = $1 OR origin.id::text = $1) 
		  AND (dest.city = $2 OR dest.id::text = $2) 
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
		var pricingJSON []byte
		var orgID sql.NullString
		if err := rows.Scan(
			&t.ID, &orgID, &t.RouteID, &t.VehicleID, &t.VehicleType, &t.VehicleClass,
			&t.DepartureTime, &t.ArrivalTime, &t.TotalSeats, &t.AvailableSeats, &pricingJSON,
			&t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			t.OrganizationID = orgID.String
		}
		json.Unmarshal(pricingJSON, &t.Pricing)
		// Segments not yet implemented in DB
		trips = append(trips, &t)
	}

	// Count total (simplified)
	return trips, len(trips), nil
}

func (r *PostgresTripRepository) UpdateStatus(ctx context.Context, id, orgID, status string) error {
	query := `UPDATE trips SET status = $1, updated_at = $2 WHERE id = $3 AND organization_id = $4`
	_, err := r.DB.ExecContext(ctx, query, status, time.Now(), id, orgID)
	return err
}

func (r *PostgresTripRepository) DecrementSeats(ctx context.Context, id string, count int) error {
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
