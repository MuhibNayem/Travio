package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/events"
	"github.com/google/uuid"
)

var (
	ErrStationNotFound  = errors.New("station not found")
	ErrRouteNotFound    = errors.New("route not found")
	ErrTripNotFound     = errors.New("trip not found")
	ErrScheduleNotFound = errors.New("schedule not found")
	ErrDuplicateCode    = errors.New("code already exists")
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

func (r *PostgresStationRepository) List(ctx context.Context, orgID, city, searchQuery string, limit, offset int) ([]*domain.Station, int, error) {
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

	if searchQuery != "" {
		// Fuzzy search on Name, Code, City, or State
		searchTerm := "%" + searchQuery + "%"
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d OR city ILIKE $%d OR state ILIKE $%d)", argIdx, argIdx, argIdx, argIdx)
		// We use the same arg 4 times, but pgx/database/sql usually requires positional args.
		// Actually, standard SQL with $1 requires the value to be passed for EACH usage if using simple drivers, but usually one arg per index.
		// Wait, if I use $3 for all, I only pass it once.
		// Let's verify standard postgres param usage. Yes, I can reuse $N.
		args = append(args, searchTerm)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stations %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch data
	query := fmt.Sprintf(`SELECT id, organization_id, code, name, city, state, country, 
		  latitude, longitude, timezone, address, amenities, status, created_at, updated_at 
		  FROM stations %s ORDER BY name ASC, id ASC LIMIT $%d OFFSET $%d`,
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
	DB        *sql.DB
	publisher *events.Publisher
}

func NewTripRepository(db *sql.DB, publisher *events.Publisher) *PostgresTripRepository {
	return &PostgresTripRepository{DB: db, publisher: publisher}
}

func (r *PostgresTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	trip.ID = uuid.New().String()
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	trip.Status = domain.TripStatusScheduled
	trip.AvailableSeats = trip.TotalSeats

	pricingJSON, _ := json.Marshal(trip.Pricing)

	var scheduleID interface{} = nil
	if trip.ScheduleID != "" {
		scheduleID = trip.ScheduleID
	}

	var serviceDate interface{} = nil
	if trip.ServiceDate != "" {
		parsedDate, err := time.Parse("2006-01-02", trip.ServiceDate)
		if err != nil {
			return err
		}
		serviceDate = parsedDate
	}

	query := `INSERT INTO trips (id, organization_id, schedule_id, service_date, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status,
			  created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.DB.ExecContext(ctx, query,
		trip.ID, trip.OrganizationID, scheduleID, serviceDate, trip.RouteID, trip.VehicleID, trip.VehicleType,
		trip.VehicleClass, trip.DepartureTime, trip.ArrivalTime, trip.TotalSeats, trip.AvailableSeats,
		pricingJSON, trip.Status, trip.CreatedAt, trip.UpdatedAt,
	)
	return err
}

func (r *PostgresTripRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	query := `SELECT id, organization_id, schedule_id, service_date, route_id, vehicle_id, vehicle_type, vehicle_class,
			  departure_time, arrival_time, total_seats, available_seats, pricing, status,
			  created_at, updated_at
			  FROM trips WHERE id = $1 AND organization_id = $2`

	var trip domain.Trip
	var pricingJSON []byte
	var scheduleID sql.NullString
	var serviceDate sql.NullTime

	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&trip.ID, &trip.OrganizationID, &scheduleID, &serviceDate, &trip.RouteID, &trip.VehicleID, &trip.VehicleType,
		&trip.VehicleClass, &trip.DepartureTime, &trip.ArrivalTime, &trip.TotalSeats, &trip.AvailableSeats,
		&pricingJSON, &trip.Status,
		&trip.CreatedAt, &trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	json.Unmarshal(pricingJSON, &trip.Pricing)
	if scheduleID.Valid {
		trip.ScheduleID = scheduleID.String
	}
	if serviceDate.Valid {
		trip.ServiceDate = serviceDate.Time.Format("2006-01-02")
	}

	segments, err := r.GetSegments(ctx, trip.ID)
	if err == nil {
		trip.Segments = segments
	}
	return &trip, nil
}

func (r *PostgresTripRepository) List(ctx context.Context, orgID, routeID, scheduleID, serviceDateFrom, serviceDateTo string, limit, offset int) ([]*domain.Trip, int, error) {
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
	if scheduleID != "" {
		whereClause += fmt.Sprintf(" AND schedule_id = $%d", argIdx)
		args = append(args, scheduleID)
		argIdx++
	}
	if serviceDateFrom != "" {
		parsedDate, err := time.Parse("2006-01-02", serviceDateFrom)
		if err != nil {
			return nil, 0, err
		}
		whereClause += fmt.Sprintf(" AND service_date >= $%d", argIdx)
		args = append(args, parsedDate)
		argIdx++
	}
	if serviceDateTo != "" {
		parsedDate, err := time.Parse("2006-01-02", serviceDateTo)
		if err != nil {
			return nil, 0, err
		}
		whereClause += fmt.Sprintf(" AND service_date <= $%d", argIdx)
		args = append(args, parsedDate)
		argIdx++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM trips %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	// Fetch data
	query := fmt.Sprintf(`SELECT id, organization_id, schedule_id, service_date, route_id, vehicle_id, vehicle_type, vehicle_class,
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
		var scheduleID sql.NullString
		var serviceDate sql.NullTime
		if err := rows.Scan(
			&t.ID, &orgID, &scheduleID, &serviceDate, &t.RouteID, &t.VehicleID, &t.VehicleType,
			&t.VehicleClass, &t.DepartureTime, &t.ArrivalTime, &t.TotalSeats, &t.AvailableSeats,
			&pricingJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			t.OrganizationID = orgID.String
		}
		if scheduleID.Valid {
			t.ScheduleID = scheduleID.String
		}
		if serviceDate.Valid {
			t.ServiceDate = serviceDate.Time.Format("2006-01-02")
		}
		json.Unmarshal(pricingJSON, &t.Pricing)
		trips = append(trips, &t)
	}

	return trips, total, nil
}

func (r *PostgresTripRepository) Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error) {
	// Complex search joining trips, routes, stations
	query := `
		SELECT t.id, t.organization_id, t.schedule_id, t.service_date, t.route_id, t.vehicle_id, t.vehicle_type, t.vehicle_class,
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
		var scheduleID sql.NullString
		var serviceDate sql.NullTime
		if err := rows.Scan(
			&t.ID, &orgID, &scheduleID, &serviceDate, &t.RouteID, &t.VehicleID, &t.VehicleType,
			&t.VehicleClass, &t.DepartureTime, &t.ArrivalTime, &t.TotalSeats, &t.AvailableSeats,
			&pricingJSON, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if orgID.Valid {
			t.OrganizationID = orgID.String
		}
		if scheduleID.Valid {
			t.ScheduleID = scheduleID.String
		}
		if serviceDate.Valid {
			t.ServiceDate = serviceDate.Time.Format("2006-01-02")
		}
		json.Unmarshal(pricingJSON, &t.Pricing)
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

func (r *PostgresTripRepository) CreateSegments(ctx context.Context, tripID string, segments []domain.TripSegment) error {
	if len(segments) == 0 {
		return nil
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO trip_segments (
		trip_id, segment_index, from_station_id, to_station_id, departure_time, arrival_time, available_seats
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, seg := range segments {
		if _, err := stmt.ExecContext(ctx, tripID, seg.SegmentIndex, seg.FromStationID, seg.ToStationID, seg.DepartureTime, seg.ArrivalTime, seg.AvailableSeats); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresTripRepository) GetSegments(ctx context.Context, tripID string) ([]domain.TripSegment, error) {
	query := `SELECT segment_index, from_station_id, to_station_id, departure_time, arrival_time, available_seats
			  FROM trip_segments WHERE trip_id = $1 ORDER BY segment_index ASC`

	rows, err := r.DB.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var segments []domain.TripSegment
	for rows.Next() {
		var seg domain.TripSegment
		if err := rows.Scan(&seg.SegmentIndex, &seg.FromStationID, &seg.ToStationID, &seg.DepartureTime, &seg.ArrivalTime, &seg.AvailableSeats); err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}

	return segments, nil
}

func (r *PostgresTripRepository) BatchCreate(ctx context.Context, trips []*domain.Trip) error {
	if len(trips) == 0 {
		return nil
	}

	// Assign IDs if missing
	for _, t := range trips {
		if t.ID == "" {
			t.ID = uuid.New().String()
		}
		if t.Status == "" {
			t.Status = domain.TripStatusScheduled
		}
		if t.CreatedAt.IsZero() {
			t.CreatedAt = time.Now()
		}
		if t.UpdatedAt.IsZero() {
			t.UpdatedAt = time.Now()
		}
		if t.AvailableSeats == 0 {
			t.AvailableSeats = t.TotalSeats
		}
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Trips
	chunkSize := 50
	for i := 0; i < len(trips); i += chunkSize {
		end := i + chunkSize
		if end > len(trips) {
			end = len(trips)
		}
		batchTrips := trips[i:end]

		if err := r.batchInsertTrips(ctx, tx, batchTrips); err != nil {
			return err
		}

		// Publish events for these trips
		if r.publisher != nil {
			for _, t := range batchTrips {
				if err := r.publisher.PublishTripCreated(ctx, tx, t); err != nil {
					return fmt.Errorf("failed to publish trip created event: %w", err)
				}
			}
		}
	}

	// 2. Insert Segments
	var allSegments []struct {
		TripID string
		Seg    domain.TripSegment
	}
	for _, t := range trips {
		for _, s := range t.Segments {
			allSegments = append(allSegments, struct {
				TripID string
				Seg    domain.TripSegment
			}{t.ID, s})
		}
	}

	chunkSizeSeg := 100
	for i := 0; i < len(allSegments); i += chunkSizeSeg {
		end := i + chunkSizeSeg
		if end > len(allSegments) {
			end = len(allSegments)
		}
		batchSegs := allSegments[i:end]
		if err := r.batchInsertSegments(ctx, tx, batchSegs); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresTripRepository) batchInsertTrips(ctx context.Context, tx *sql.Tx, trips []*domain.Trip) error {
	placeholders := 16
	valueStrings := make([]string, 0, len(trips))
	valueArgs := make([]interface{}, 0, len(trips)*placeholders)

	for i, t := range trips {
		n := i * placeholders
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n+1, n+2, n+3, n+4, n+5, n+6, n+7, n+8, n+9, n+10, n+11, n+12, n+13, n+14, n+15, n+16))

		pricingJSON, _ := json.Marshal(t.Pricing)
		var scheduleID interface{} = nil
		if t.ScheduleID != "" {
			scheduleID = t.ScheduleID
		}
		var serviceDate interface{} = nil
		if t.ServiceDate != "" {
			if parsed, err := time.Parse("2006-01-02", t.ServiceDate); err == nil {
				serviceDate = parsed
			}
		}

		valueArgs = append(valueArgs,
			t.ID, t.OrganizationID, scheduleID, serviceDate, t.RouteID, t.VehicleID, t.VehicleType,
			t.VehicleClass, t.DepartureTime, t.ArrivalTime, t.TotalSeats, t.AvailableSeats,
			pricingJSON, t.Status, t.CreatedAt, t.UpdatedAt)
	}

	query := fmt.Sprintf("INSERT INTO trips (id, organization_id, schedule_id, service_date, route_id, vehicle_id, vehicle_type, vehicle_class, departure_time, arrival_time, total_seats, available_seats, pricing, status, created_at, updated_at) VALUES %s", strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, valueArgs...)
	return err
}

func (r *PostgresTripRepository) batchInsertSegments(ctx context.Context, tx *sql.Tx, segments []struct {
	TripID string
	Seg    domain.TripSegment
}) error {
	if len(segments) == 0 {
		return nil
	}
	placeholders := 7
	valueStrings := make([]string, 0, len(segments))
	valueArgs := make([]interface{}, 0, len(segments)*placeholders)

	for i, entry := range segments {
		n := i * placeholders
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			n+1, n+2, n+3, n+4, n+5, n+6, n+7))

		valueArgs = append(valueArgs,
			entry.TripID, entry.Seg.SegmentIndex, entry.Seg.FromStationID, entry.Seg.ToStationID,
			entry.Seg.DepartureTime, entry.Seg.ArrivalTime, entry.Seg.AvailableSeats)
	}

	query := fmt.Sprintf("INSERT INTO trip_segments (trip_id, segment_index, from_station_id, to_station_id, departure_time, arrival_time, available_seats) VALUES %s", strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, valueArgs...)
	return err
}

func (r *PostgresTripRepository) CheckVehicleAvailability(ctx context.Context, vehicleID string, startTime, endTime time.Time) (bool, error) {
	// Check for any non-cancelled trip that overlaps with the requested time window
	// Overlap logic: (StartA < EndB) AND (EndA > StartB)
	query := `SELECT COUNT(*) FROM trips 
			  WHERE vehicle_id = $1 
			  AND status != $2 
			  AND (departure_time < $3 AND arrival_time > $4)`

	var count int
	err := r.DB.QueryRowContext(ctx, query, vehicleID, domain.TripStatusCancelled, endTime, startTime).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// PostgresScheduleRepository handles schedule persistence
type PostgresScheduleRepository struct {
	DB *sql.DB
}

func NewScheduleRepository(db *sql.DB) *PostgresScheduleRepository {
	return &PostgresScheduleRepository{DB: db}
}

func (r *PostgresScheduleRepository) Create(ctx context.Context, schedule *domain.ScheduleTemplate) error {
	schedule.ID = uuid.New().String()
	schedule.CreatedAt = time.Now()
	schedule.UpdatedAt = time.Now()
	schedule.Status = domain.ScheduleStatusActive
	if schedule.Version <= 0 {
		schedule.Version = 1
	}

	fmt.Printf("DEBUG Schedule.Create: ID=%s, OrgID=%s, RouteID=%s, VehicleID=%s\n",
		schedule.ID, schedule.OrganizationID, schedule.RouteID, schedule.VehicleID)

	pricingJSON, _ := json.Marshal(schedule.Pricing)

	query := `INSERT INTO schedule_templates (
		id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class, total_seats, pricing,
		departure_time, arrival_offset_minutes, timezone, start_date, end_date,
		days_of_week, status, version, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.DB.ExecContext(ctx, query,
		schedule.ID, schedule.OrganizationID, schedule.RouteID, schedule.VehicleID,
		schedule.VehicleType, schedule.VehicleClass, schedule.TotalSeats, pricingJSON, minutesToTime(schedule.DepartureMinutes),
		schedule.ArrivalOffsetMinutes, schedule.Timezone, schedule.StartDate, schedule.EndDate,
		schedule.DaysOfWeek, schedule.Status, schedule.Version, schedule.CreatedAt, schedule.UpdatedAt,
	)
	if err != nil {
		fmt.Printf("DEBUG Schedule.Create ERROR: %v\n", err)
	}
	return err
}

func (r *PostgresScheduleRepository) GetByID(ctx context.Context, id, orgID string) (*domain.ScheduleTemplate, error) {
	query := `SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class, total_seats, pricing,
		departure_time, arrival_offset_minutes, timezone, start_date, end_date, days_of_week,
		status, version, created_at, updated_at
		FROM schedule_templates WHERE id = $1 AND organization_id = $2`

	var schedule domain.ScheduleTemplate
	var departureTime time.Time
	var pricingJSON []byte
	err := r.DB.QueryRowContext(ctx, query, id, orgID).Scan(
		&schedule.ID, &schedule.OrganizationID, &schedule.RouteID, &schedule.VehicleID,
		&schedule.VehicleType, &schedule.VehicleClass, &schedule.TotalSeats, &pricingJSON, &departureTime, &schedule.ArrivalOffsetMinutes,
		&schedule.Timezone, &schedule.StartDate, &schedule.EndDate, &schedule.DaysOfWeek,
		&schedule.Status, &schedule.Version, &schedule.CreatedAt, &schedule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	schedule.DepartureMinutes = timeToMinutes(departureTime)
	json.Unmarshal(pricingJSON, &schedule.Pricing)
	return &schedule, nil
}

func (r *PostgresScheduleRepository) List(ctx context.Context, orgID, routeID, status string, limit, offset int) ([]*domain.ScheduleTemplate, int, error) {
	var args []interface{}
	var whereClause string
	argIdx := 1

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

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schedule_templates %s", whereClause)
	var total int
	r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := fmt.Sprintf(`SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class, total_seats, pricing,
		departure_time, arrival_offset_minutes, timezone, start_date, end_date, days_of_week,
		status, version, created_at, updated_at
		FROM schedule_templates %s ORDER BY start_date ASC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var schedules []*domain.ScheduleTemplate
	for rows.Next() {
		var schedule domain.ScheduleTemplate
		var departureTimeStr string // Scan as string since DB returns TIME as string
		var pricingJSON []byte
		if err := rows.Scan(
			&schedule.ID, &schedule.OrganizationID, &schedule.RouteID, &schedule.VehicleID,
			&schedule.VehicleType, &schedule.VehicleClass, &schedule.TotalSeats, &pricingJSON, &departureTimeStr, &schedule.ArrivalOffsetMinutes,
			&schedule.Timezone, &schedule.StartDate, &schedule.EndDate, &schedule.DaysOfWeek,
			&schedule.Status, &schedule.Version, &schedule.CreatedAt, &schedule.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		// Parse TIME string (HH:MM:SS) to minutes since midnight
		if departureTimeStr != "" {
			t, err := time.Parse("15:04:05", departureTimeStr)
			if err == nil {
				schedule.DepartureMinutes = timeToMinutes(t)
			}
		}

		json.Unmarshal(pricingJSON, &schedule.Pricing)
		schedules = append(schedules, &schedule)
	}

	return schedules, total, nil
}

func (r *PostgresScheduleRepository) HasVehicleConflict(ctx context.Context, schedule *domain.ScheduleTemplate, excludeID string) (bool, error) {
	if schedule == nil {
		return false, nil
	}
	newStart := schedule.DepartureMinutes
	newEnd := schedule.DepartureMinutes + schedule.ArrivalOffsetMinutes
	if newEnd <= newStart {
		return false, nil
	}

	// Build query dynamically based on whether we need to exclude a specific ID
	queryBase := `
		SELECT COUNT(*) FROM schedule_templates
		WHERE organization_id = $1
		  AND vehicle_id = $2`

	var query string
	var args []interface{}

	if excludeID != "" {
		query = queryBase + `
		  AND id != $3
		  AND status = 'active'
		  AND start_date <= $4
		  AND end_date >= $5
		  AND (days_of_week & $6) != 0
		  AND (
				(EXTRACT(EPOCH FROM departure_time)/60) < $7
				AND (EXTRACT(EPOCH FROM departure_time)/60 + COALESCE(arrival_offset_minutes, 0)) > $8
		  )`
		args = []interface{}{
			schedule.OrganizationID,
			schedule.VehicleID,
			excludeID,
			schedule.EndDate,
			schedule.StartDate,
			schedule.DaysOfWeek,
			float64(newEnd) / 60.0,
			float64(newStart) / 60.0,
		}
	} else {
		query = queryBase + `
		  AND status = 'active'
		  AND start_date <= $3
		  AND end_date >= $4
		  AND (days_of_week & $5) != 0
		  AND (
				(EXTRACT(EPOCH FROM departure_time)/60) < $6
				AND (EXTRACT(EPOCH FROM departure_time)/60 + COALESCE(arrival_offset_minutes, 0)) > $7
		  )`
		args = []interface{}{
			schedule.OrganizationID,
			schedule.VehicleID,
			schedule.EndDate,
			schedule.StartDate,
			schedule.DaysOfWeek,
			float64(newEnd) / 60.0,
			float64(newStart) / 60.0,
		}
	}

	var count int
	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&count)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresScheduleRepository) Update(ctx context.Context, schedule *domain.ScheduleTemplate) error {
	// 1. Fetch current version within a transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current state for versioning
	var currentVersion int

	// Simply selecting all fields to store as snapshot
	queryGet := `SELECT id, organization_id, route_id, vehicle_id, vehicle_type, vehicle_class, total_seats, pricing,
		departure_time, arrival_offset_minutes, timezone, start_date, end_date, days_of_week, status, version 
		FROM schedule_templates WHERE id = $1 AND organization_id = $2 FOR UPDATE`

	var current domain.ScheduleTemplate
	var pricingJSON []byte
	var departureTime time.Time

	err = tx.QueryRowContext(ctx, queryGet, schedule.ID, schedule.OrganizationID).Scan(
		&current.ID, &current.OrganizationID, &current.RouteID, &current.VehicleID,
		&current.VehicleType, &current.VehicleClass, &current.TotalSeats, &pricingJSON,
		&departureTime, &current.ArrivalOffsetMinutes, &current.Timezone, &current.StartDate,
		&current.EndDate, &current.DaysOfWeek, &current.Status, &currentVersion,
	)
	if err != nil {
		return err
	}
	current.DepartureMinutes = timeToMinutes(departureTime)
	json.Unmarshal(pricingJSON, &current.Pricing)

	// 2. Insert into schedule_versions
	snapshot, _ := json.Marshal(current)
	queryVersion := `INSERT INTO schedule_versions (schedule_id, version, snapshot) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, queryVersion, current.ID, currentVersion, snapshot)
	if err != nil {
		return err
	}

	// 3. Update the schedule_templates table and increment version
	newVersion := currentVersion + 1
	schedule.Version = newVersion
	schedule.UpdatedAt = time.Now()
	newPricingJSON, _ := json.Marshal(schedule.Pricing)

	queryUpdate := `UPDATE schedule_templates SET 
		route_id = $1, vehicle_id = $2, vehicle_type = $3, vehicle_class = $4, total_seats = $5, pricing = $6,
		departure_time = $7, arrival_offset_minutes = $8, timezone = $9, start_date = $10, end_date = $11,
		days_of_week = $12, status = $13, version = $14, updated_at = $15
		WHERE id = $16 AND organization_id = $17`

	res, err := tx.ExecContext(ctx, queryUpdate,
		schedule.RouteID, schedule.VehicleID, schedule.VehicleType, schedule.VehicleClass, schedule.TotalSeats, newPricingJSON,
		minutesToTime(schedule.DepartureMinutes), schedule.ArrivalOffsetMinutes, schedule.Timezone, schedule.StartDate, schedule.EndDate,
		schedule.DaysOfWeek, schedule.Status, newVersion, schedule.UpdatedAt,
		schedule.ID, schedule.OrganizationID,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrScheduleNotFound
	}

	return tx.Commit()
}

func (r *PostgresScheduleRepository) Delete(ctx context.Context, id, orgID string) error {
	query := `DELETE FROM schedule_templates WHERE id = $1 AND organization_id = $2`
	res, err := r.DB.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrScheduleNotFound
	}
	return nil
}

func (r *PostgresScheduleRepository) AddException(ctx context.Context, exception *domain.ScheduleException) error {
	exception.ID = uuid.New().String()
	exception.CreatedAt = time.Now()
	query := `INSERT INTO schedule_exceptions (id, schedule_id, service_date, is_added, reason, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.DB.ExecContext(ctx, query,
		exception.ID, exception.ScheduleID, exception.ServiceDate, exception.IsAdded, exception.Reason, exception.CreatedAt)
	return err
}

func (r *PostgresScheduleRepository) ListExceptions(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleException, error) {
	// Verify ownership first
	var exists bool
	r.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schedule_templates WHERE id=$1 AND organization_id=$2)", scheduleID, orgID).Scan(&exists)
	if !exists {
		return nil, ErrScheduleNotFound
	}

	query := `SELECT id, schedule_id, service_date, is_added, reason, created_at FROM schedule_exceptions WHERE schedule_id = $1`
	rows, err := r.DB.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exceptions []*domain.ScheduleException
	for rows.Next() {
		var ex domain.ScheduleException
		var serviceDate time.Time // To parse DB date format
		if err := rows.Scan(&ex.ID, &ex.ScheduleID, &serviceDate, &ex.IsAdded, &ex.Reason, &ex.CreatedAt); err != nil {
			return nil, err
		}
		ex.ServiceDate = serviceDate.Format("2006-01-02")
		exceptions = append(exceptions, &ex)
	}
	return exceptions, nil
}

func (r *PostgresScheduleRepository) GetHistory(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleVersion, error) {
	// Verify ownership
	var exists bool
	err := r.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schedule_templates WHERE id=$1 AND organization_id=$2)", scheduleID, orgID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrScheduleNotFound
	}

	query := `SELECT id, schedule_id, version, snapshot, created_at FROM schedule_versions WHERE schedule_id = $1 ORDER BY version DESC`
	rows, err := r.DB.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*domain.ScheduleVersion
	for rows.Next() {
		var v domain.ScheduleVersion
		var snapshotJSON []byte
		if err := rows.Scan(&v.ID, &v.ScheduleID, &v.Version, &snapshotJSON, &v.CreatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(snapshotJSON, &v.Snapshot)
		versions = append(versions, &v)
	}
	return versions, nil
}

func minutesToTime(minutes int) time.Time {
	if minutes < 0 {
		minutes = 0
	}
	hours := minutes / 60
	mins := minutes % 60
	return time.Date(2000, 1, 1, hours, mins, 0, 0, time.UTC)
}

func timeToMinutes(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}
