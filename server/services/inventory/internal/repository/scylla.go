package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/inventory/internal/domain"
	"github.com/gocql/gocql"
)

// ScyllaRepository handles high-throughput inventory operations
// Designed for 100K+ reads/second during peak booking
type ScyllaRepository struct {
	session *gocql.Session
}

func NewScyllaRepository(session *gocql.Session) *ScyllaRepository {
	return &ScyllaRepository{session: session}
}

// InitializeTrip creates all segment-seat records for a new trip
func (r *ScyllaRepository) InitializeTrip(ctx context.Context, orgID, tripID string, segments []domain.Segment, seats []domain.SeatInventory) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)

	// Insert segment metadata
	for _, seg := range segments {
		batch.Query(`INSERT INTO segments (organization_id, trip_id, segment_index, from_station_id, to_station_id, departure_time, arrival_time) 
					 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			orgID, seg.TripID, seg.SegmentIndex, seg.FromStationID, seg.ToStationID, seg.DepartureTime, seg.ArrivalTime)
	}

	// Insert seat inventory for each segment
	for _, seg := range segments {
		for _, seat := range seats {
			batch.Query(`INSERT INTO seat_inventory (organization_id, trip_id, segment_index, seat_id, seat_number, seat_class, 
						 seat_type, status, price_paisa, updated_at) 
						 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				orgID, tripID, seg.SegmentIndex, seat.SeatID, seat.SeatNumber, seat.SeatClass,
				seat.SeatType, domain.SeatStatusAvailable, seat.PricePaisa, time.Now())
		}
	}

	return r.session.ExecuteBatch(batch)
}

// GetSeatAvailability returns availability for specific segments
func (r *ScyllaRepository) GetSeatAvailability(ctx context.Context, orgID, tripID string, segmentIndices []int) ([]domain.SeatInventory, error) {
	// Build query for multiple segments
	// Using IN clause for segment indices (efficient in Scylla with partition key)
	query := `SELECT trip_id, segment_index, seat_id, seat_number, seat_class, seat_type, 
			  status, hold_id, hold_user_id, hold_expiry, booking_id, price_paisa, updated_at 
			  FROM seat_inventory 
			  WHERE organization_id = ? AND trip_id = ? AND segment_index IN ?`

	iter := r.session.Query(query, orgID, tripID, segmentIndices).WithContext(ctx).Iter()

	var seats []domain.SeatInventory
	var seat domain.SeatInventory

	for iter.Scan(
		&seat.TripID, &seat.SegmentIndex, &seat.SeatID, &seat.SeatNumber, &seat.SeatClass,
		&seat.SeatType, &seat.Status, &seat.HoldID, &seat.HoldUserID, &seat.HoldExpiry,
		&seat.BookingID, &seat.PricePaisa, &seat.UpdatedAt,
	) {
		seat.OrganizationID = orgID
		seats = append(seats, seat)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return seats, nil
}

// CheckSeatsAvailableForSegments checks if specific seats are available across all required segments
func (r *ScyllaRepository) CheckSeatsAvailableForSegments(ctx context.Context, orgID, tripID string, seatIDs []string, segmentIndices []int) (bool, map[string]string, error) {
	// Check each seat-segment combination
	unavailable := make(map[string]string) // seat_id -> reason

	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			query := `SELECT status, hold_expiry FROM seat_inventory 
					  WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?`

			var status string
			var holdExpiry time.Time

			err := r.session.Query(query, orgID, tripID, segIdx, seatID).WithContext(ctx).Scan(&status, &holdExpiry)
			if err != nil {
				if err == gocql.ErrNotFound {
					unavailable[seatID] = "seat not found"
					continue
				}
				return false, nil, err
			}

			// Check status
			switch status {
			case domain.SeatStatusBooked:
				unavailable[seatID] = fmt.Sprintf("booked on segment %d", segIdx)
			case domain.SeatStatusBlocked:
				unavailable[seatID] = fmt.Sprintf("blocked on segment %d", segIdx)
			case domain.SeatStatusHeld:
				// Check if hold has expired
				if time.Now().Before(holdExpiry) {
					unavailable[seatID] = fmt.Sprintf("held on segment %d", segIdx)
				}
				// Else: hold expired, treat as available
			}
		}
	}

	return len(unavailable) == 0, unavailable, nil
}

// ReleaseExpiredHolds clears expired holds for specific seats and segments
func (r *ScyllaRepository) ReleaseExpiredHolds(ctx context.Context, orgID, tripID string, seatIDs []string, segmentIndices []int) error {
	if len(seatIDs) == 0 || len(segmentIndices) == 0 {
		return nil
	}

	batch := r.session.NewBatch(gocql.LoggedBatch)
	now := time.Now()

	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			batch.Query(`UPDATE seat_inventory
						 SET status = ?, hold_id = '', hold_user_id = '', hold_expiry = ?, updated_at = ?
						 WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?
						 IF status = ? AND hold_expiry < ?`,
				domain.SeatStatusAvailable, time.Time{}, now,
				orgID, tripID, segIdx, seatID,
				domain.SeatStatusHeld, now)
		}
	}

	return r.session.ExecuteBatch(batch)
}

// HoldSeats marks seats as held across all required segments atomically
func (r *ScyllaRepository) HoldSeats(ctx context.Context, orgID, holdID, tripID, userID string, seatIDs []string, segmentIndices []int, expiry time.Time) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)

	now := time.Now()
	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			// Use lightweight transaction (LWT) for atomicity
			batch.Query(`UPDATE seat_inventory 
						 SET status = ?, hold_id = ?, hold_user_id = ?, hold_expiry = ?, updated_at = ?
						 WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?
						 IF status = ?`,
				domain.SeatStatusHeld, holdID, userID, expiry, now,
				orgID, tripID, segIdx, seatID,
				domain.SeatStatusAvailable)
		}
	}

	return r.session.ExecuteBatch(batch)
}

// ReleaseHold marks seats as available again
func (r *ScyllaRepository) ReleaseHold(ctx context.Context, orgID, tripID, holdID string, segmentIndices []int, seatIDs []string) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)

	now := time.Now()
	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			batch.Query(`UPDATE seat_inventory 
						 SET status = ?, hold_id = '', hold_user_id = '', hold_expiry = ?, updated_at = ?
						 WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?
						 IF hold_id = ?`,
				domain.SeatStatusAvailable, time.Time{}, now,
				orgID, tripID, segIdx, seatID,
				holdID)
		}
	}

	return r.session.ExecuteBatch(batch)
}

// ConfirmBooking converts held seats to booked status
func (r *ScyllaRepository) ConfirmBooking(ctx context.Context, orgID, tripID, holdID, bookingID string, segmentIndices []int, seatIDs []string) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)

	now := time.Now()
	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			batch.Query(`UPDATE seat_inventory 
						 SET status = ?, booking_id = ?, hold_id = '', hold_user_id = '', updated_at = ?
						 WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?
						 IF hold_id = ?`,
				domain.SeatStatusBooked, bookingID, now,
				orgID, tripID, segIdx, seatID,
				holdID)
		}
	}

	return r.session.ExecuteBatch(batch)
}

// CancelBooking releases booked seats back to available
func (r *ScyllaRepository) CancelBooking(ctx context.Context, orgID, tripID, bookingID string, segmentIndices []int, seatIDs []string) error {
	batch := r.session.NewBatch(gocql.LoggedBatch)

	now := time.Now()
	for _, segIdx := range segmentIndices {
		for _, seatID := range seatIDs {
			batch.Query(`UPDATE seat_inventory 
						 SET status = ?, booking_id = '', updated_at = ?
						 WHERE organization_id = ? AND trip_id = ? AND segment_index = ? AND seat_id = ?
						 IF booking_id = ?`,
				domain.SeatStatusAvailable, now,
				orgID, tripID, segIdx, seatID,
				bookingID)
		}
	}

	return r.session.ExecuteBatch(batch)
}

// GetSegments returns segment metadata for a trip
func (r *ScyllaRepository) GetSegments(ctx context.Context, orgID, tripID string) ([]domain.Segment, error) {
	query := `SELECT trip_id, segment_index, from_station_id, to_station_id, departure_time, arrival_time 
			  FROM segments WHERE organization_id = ? AND trip_id = ? ORDER BY segment_index`

	iter := r.session.Query(query, orgID, tripID).WithContext(ctx).Iter()

	var segments []domain.Segment
	var seg domain.Segment

	for iter.Scan(&seg.TripID, &seg.SegmentIndex, &seg.FromStationID, &seg.ToStationID, &seg.DepartureTime, &seg.ArrivalTime) {
		seg.OrganizationID = orgID
		segments = append(segments, seg)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return segments, nil
}

// CountAvailableSeats counts seats available across all required segments
func (r *ScyllaRepository) CountAvailableSeats(ctx context.Context, orgID, tripID string, segmentIndices []int, seatClass string) (int, error) {
	// This is a simplification - in production, you'd use a materialized view or counter table
	seats, err := r.GetSeatAvailability(ctx, orgID, tripID, segmentIndices)
	if err != nil {
		return 0, err
	}

	// Group seats by seat_id and check if available across ALL segments
	seatSegmentCount := make(map[string]int)
	seatInfo := make(map[string]domain.SeatInventory)

	now := time.Now()
	for _, seat := range seats {
		if seatClass != "" && seat.SeatClass != seatClass {
			continue
		}

		isAvailable := seat.Status == domain.SeatStatusAvailable ||
			(seat.Status == domain.SeatStatusHeld && now.After(seat.HoldExpiry))

		if isAvailable {
			seatSegmentCount[seat.SeatID]++
			seatInfo[seat.SeatID] = seat
		}
	}

	// Count seats that are available on ALL required segments
	requiredSegments := len(segmentIndices)
	available := 0
	for _, count := range seatSegmentCount {
		if count == requiredSegments {
			available++
		}
	}

	return available, nil
}
