package service

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/inventory/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/repository"
	"github.com/google/uuid"
)

const (
	DefaultHoldDuration = 10 * time.Minute
	MaxHoldsPerUser     = 2
)

// InventoryService handles seat availability and booking operations
type InventoryService struct {
	scyllaRepo *repository.ScyllaRepository
	holdRepo   *repository.HoldRepository
	redisRepo  *repository.RedisRepository
}

func NewInventoryService(scyllaRepo *repository.ScyllaRepository, holdRepo *repository.HoldRepository, redisRepo *repository.RedisRepository) *InventoryService {
	return &InventoryService{
		scyllaRepo: scyllaRepo,
		holdRepo:   holdRepo,
		redisRepo:  redisRepo,
	}
}

// CheckAvailability returns seat availability for a journey
func (s *InventoryService) CheckAvailability(ctx context.Context, tripID, fromStation, toStation string, passengers int, seatClass string) (*AvailabilityResult, error) {
	// Get segments for this trip
	segments, err := s.scyllaRepo.GetSegments(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Calculate segment range for the journey
	stationOrder := extractStationOrder(segments)
	segmentRange, err := domain.CalculateSegmentRange(stationOrder, fromStation, toStation)
	if err != nil {
		return nil, err
	}

	// Get seat availability for these segments
	seats, err := s.scyllaRepo.GetSeatAvailability(ctx, tripID, segmentRange)
	if err != nil {
		return nil, err
	}

	// Filter and aggregate availability
	availableSeats := filterAvailableSeats(seats, segmentRange, seatClass)

	// Calculate pricing
	var totalPrice int64
	if len(availableSeats) > 0 && passengers > 0 {
		for i := 0; i < passengers && i < len(availableSeats); i++ {
			totalPrice += availableSeats[i].PricePaisa
		}
	}

	return &AvailabilityResult{
		IsAvailable:     len(availableSeats) >= passengers,
		AvailableCount:  len(availableSeats),
		Seats:           availableSeats,
		TotalPricePaisa: totalPrice,
		SegmentRange:    segmentRange,
		CheckedAt:       time.Now(),
	}, nil
}

// HoldSeats creates a temporary hold on seats
func (s *InventoryService) HoldSeats(ctx context.Context, req *HoldRequest) (*HoldResult, error) {
	// Check user's current hold count (anti-scalping)
	holdCount, err := s.holdRepo.CountUserActiveHolds(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if holdCount >= MaxHoldsPerUser {
		return &HoldResult{
			Success:       false,
			FailureReason: "maximum concurrent holds exceeded",
		}, nil
	}

	// Get segments
	segments, err := s.scyllaRepo.GetSegments(ctx, req.TripID)
	if err != nil {
		return nil, err
	}

	stationOrder := extractStationOrder(segments)
	segmentRange, err := domain.CalculateSegmentRange(stationOrder, req.FromStation, req.ToStation)
	if err != nil {
		return nil, err
	}

	// Optimistic Pre-Lock with Redis
	// Purpose: Fail fast if another user is processing the same seat, protecting DB from heavy LWTs
	lockedSeats := make([]string, 0, len(req.SeatIDs))
	// We use a clean-up function to release locks
	defer func() {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for _, seatID := range lockedSeats {
				for _, segIdx := range segmentRange {
					s.redisRepo.ReleaseSeatLock(bgCtx, req.TripID, seatID, segIdx, req.UserID)
				}
			}
		}()
	}()

	for _, seatID := range req.SeatIDs {
		// Acquire lock for ALL segments involved
		for _, segIdx := range segmentRange {
			acquired, err := s.redisRepo.AcquireSeatLock(ctx, req.TripID, seatID, segIdx, req.UserID, 10*time.Second)
			if err != nil {
				return nil, err
			}
			if !acquired {
				return &HoldResult{
					Success:       false,
					FailureReason: "seat query contention - please retry",
				}, nil
			}
		}
		// Track successfully locked seats for deferred release
		lockedSeats = append(lockedSeats, seatID)
	}

	// Check availability (Scylla)
	available, unavailableReasons, err := s.scyllaRepo.CheckSeatsAvailableForSegments(ctx, req.TripID, req.SeatIDs, segmentRange)
	if err != nil {
		return nil, err
	}

	if !available {
		var failedSeats []string
		for seatID := range unavailableReasons {
			failedSeats = append(failedSeats, seatID)
		}
		return &HoldResult{
			Success:       false,
			FailedSeatIDs: failedSeats,
			FailureReason: "some seats not available",
		}, nil
	}

	// Create hold
	holdID := uuid.New().String()
	holdDuration := DefaultHoldDuration
	if req.HoldDuration > 0 {
		holdDuration = req.HoldDuration
	}
	expiresAt := time.Now().Add(holdDuration)

	// Update ScyllaDB (mark as held)
	if err := s.scyllaRepo.HoldSeats(ctx, holdID, req.TripID, req.UserID, req.SeatIDs, segmentRange, expiresAt); err != nil {
		return nil, err
	}

	// Store hold metadata in Redis
	hold := &domain.SeatHold{
		HoldID:        holdID,
		TripID:        req.TripID,
		UserID:        req.UserID,
		SessionID:     req.SessionID,
		FromStationID: req.FromStation,
		ToStationID:   req.ToStation,
		SeatIDs:       req.SeatIDs,
		SegmentRange:  segmentRange,
		Status:        domain.HoldStatusActive,
		ExpiresAt:     expiresAt,
		CreatedAt:     time.Now(),
		IPAddress:     req.IPAddress,
	}

	if err := s.holdRepo.CreateHold(ctx, hold); err != nil {
		// Rollback ScyllaDB hold
		_ = s.scyllaRepo.ReleaseHold(ctx, req.TripID, holdID, segmentRange, req.SeatIDs)
		return nil, err
	}

	// Invalidate Cache after successful hold
	// We delete the whole trip cache to force refresh on next read
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.redisRepo.InvalidateSeatMap(bgCtx, req.TripID)
	}()

	return &HoldResult{
		Success:     true,
		HoldID:      holdID,
		HeldSeatIDs: req.SeatIDs,
		ExpiresAt:   expiresAt,
	}, nil
}

// ReleaseSeats releases a hold
func (s *InventoryService) ReleaseSeats(ctx context.Context, holdID, userID string) error {
	hold, err := s.holdRepo.GetHold(ctx, holdID)
	if err != nil {
		return err
	}

	// Verify ownership
	if hold.UserID != userID {
		return domain.ErrHoldNotFound
	}

	// Release in ScyllaDB
	if err := s.scyllaRepo.ReleaseHold(ctx, hold.TripID, holdID, hold.SegmentRange, hold.SeatIDs); err != nil {
		return err
	}

	// Mark hold as released
	if err := s.holdRepo.UpdateHoldStatus(ctx, holdID, domain.HoldStatusReleased); err != nil {
		return err
	}

	return nil
}

// ConfirmBooking converts a hold to a confirmed booking
func (s *InventoryService) ConfirmBooking(ctx context.Context, holdID, orderID, userID string, passengers []PassengerInfo) (*BookingResult, error) {
	hold, err := s.holdRepo.GetHold(ctx, holdID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if hold.UserID != userID {
		return nil, domain.ErrHoldNotFound
	}

	// Check hold is still active
	if hold.Status != domain.HoldStatusActive {
		return nil, domain.ErrHoldExpired
	}

	if time.Now().After(hold.ExpiresAt) {
		return nil, domain.ErrHoldExpired
	}

	// Validate passenger count matches seats
	if len(passengers) != len(hold.SeatIDs) {
		return &BookingResult{
			Success:       false,
			FailureReason: "passenger count does not match seat count",
		}, nil
	}

	bookingID := uuid.New().String()

	// Confirm in ScyllaDB
	if err := s.scyllaRepo.ConfirmBooking(ctx, hold.TripID, holdID, bookingID, hold.SegmentRange, hold.SeatIDs); err != nil {
		return nil, err
	}

	// Update hold status
	if err := s.holdRepo.UpdateHoldStatus(ctx, holdID, domain.HoldStatusConverted); err != nil {
		// Non-fatal, booking is already confirmed
	}

	// Build confirmed seats response
	var confirmedSeats []ConfirmedSeatInfo
	for i, seatID := range hold.SeatIDs {
		ticketID := uuid.New().String()
		confirmedSeats = append(confirmedSeats, ConfirmedSeatInfo{
			SeatID:   seatID,
			TicketID: ticketID,
		})
		if i < len(passengers) {
			confirmedSeats[i].PassengerName = passengers[i].Name
		}
	}

	return &BookingResult{
		Success:        true,
		BookingID:      bookingID,
		ConfirmedSeats: confirmedSeats,
	}, nil
}

// GetSeatMap returns the seat layout with availability status
func (s *InventoryService) GetSeatMap(ctx context.Context, tripID, fromStation, toStation string) (*SeatMapResult, error) {
	segments, err := s.scyllaRepo.GetSegments(ctx, tripID)
	if err != nil {
		return nil, err
	}

	stationOrder := extractStationOrder(segments)
	segmentRange, err := domain.CalculateSegmentRange(stationOrder, fromStation, toStation)
	if err != nil {
		return nil, err
	}

	// Try Cache First (Read-Through)
	// We cache the ENTIRE trip inventory to allow in-memory filtering for any segment range
	seats, err := s.redisRepo.GetCachedSeatMap(ctx, tripID)
	if err != nil || len(seats) == 0 {
		// Cache Miss: Fetch ALL segments to warm cache for everyone
		allSegmentIndices := make([]int, len(segments))
		for i := range segments {
			allSegmentIndices[i] = segments[i].SegmentIndex
		}

		seats, err = s.scyllaRepo.GetSeatAvailability(ctx, tripID, allSegmentIndices)
		if err != nil {
			return nil, err
		}

		// Update Cache asynchronously
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.redisRepo.CacheSeatMap(bgCtx, tripID, seats, 5*time.Second) // Short TTL for near-realtime
		}()
	}

	// Aggregate availability across segments
	seatMap := aggregateSeatMap(seats, segmentRange)

	return seatMap, nil
}

// --- Helper Types ---

type AvailabilityResult struct {
	IsAvailable     bool
	AvailableCount  int
	Seats           []SeatInfo
	TotalPricePaisa int64
	SegmentRange    []int
	CheckedAt       time.Time
}

type SeatInfo struct {
	SeatID     string
	SeatNumber string
	SeatClass  string
	SeatType   string
	Status     string
	PricePaisa int64
	HoldExpiry time.Time
}

type HoldRequest struct {
	TripID       string
	FromStation  string
	ToStation    string
	SeatIDs      []string
	UserID       string
	SessionID    string
	IPAddress    string
	HoldDuration time.Duration
}

type HoldResult struct {
	Success       bool
	HoldID        string
	HeldSeatIDs   []string
	FailedSeatIDs []string
	ExpiresAt     time.Time
	FailureReason string
}

type PassengerInfo struct {
	NID  string
	Name string
}

type BookingResult struct {
	Success        bool
	BookingID      string
	ConfirmedSeats []ConfirmedSeatInfo
	FailureReason  string
}

type ConfirmedSeatInfo struct {
	SeatID        string
	SeatNumber    string
	TicketID      string
	PassengerName string
}

type SeatMapResult struct {
	TripID string
	Rows   []SeatRow
	Legend map[string]string
}

type SeatRow struct {
	RowNumber int
	Seats     []SeatCell
}

type SeatCell struct {
	SeatID     string
	SeatNumber string
	Column     int
	SeatType   string
	SeatClass  string
	Status     string
	PricePaisa int64
}

// --- Helper Functions ---

func extractStationOrder(segments []domain.Segment) []string {
	if len(segments) == 0 {
		return nil
	}

	stations := []string{segments[0].FromStationID}
	for _, seg := range segments {
		stations = append(stations, seg.ToStationID)
	}
	return stations
}

func filterAvailableSeats(seats []domain.SeatInventory, segmentRange []int, seatClass string) []SeatInfo {
	// Group by seat_id, check availability across all segments
	seatSegmentAvail := make(map[string]int)
	seatDetails := make(map[string]domain.SeatInventory)

	now := time.Now()
	for _, seat := range seats {
		if seatClass != "" && seat.SeatClass != seatClass {
			continue
		}

		isAvail := seat.Status == domain.SeatStatusAvailable ||
			(seat.Status == domain.SeatStatusHeld && now.After(seat.HoldExpiry))

		if isAvail {
			seatSegmentAvail[seat.SeatID]++
			seatDetails[seat.SeatID] = seat
		}
	}

	requiredSegments := len(segmentRange)
	var available []SeatInfo

	for seatID, count := range seatSegmentAvail {
		if count == requiredSegments {
			detail := seatDetails[seatID]
			available = append(available, SeatInfo{
				SeatID:     seatID,
				SeatNumber: detail.SeatNumber,
				SeatClass:  detail.SeatClass,
				SeatType:   detail.SeatType,
				Status:     domain.SeatStatusAvailable,
				PricePaisa: detail.PricePaisa,
			})
		}
	}

	return available
}

func aggregateSeatMap(seats []domain.SeatInventory, segmentRange []int) *SeatMapResult {
	// Similar logic to filterAvailableSeats but builds a seat map structure
	seatSegmentAvail := make(map[string]int)
	seatDetails := make(map[string]domain.SeatInventory)

	now := time.Now()
	for _, seat := range seats {
		isAvail := seat.Status == domain.SeatStatusAvailable ||
			(seat.Status == domain.SeatStatusHeld && now.After(seat.HoldExpiry))

		if isAvail {
			seatSegmentAvail[seat.SeatID]++
		}
		seatDetails[seat.SeatID] = seat
	}

	requiredSegments := len(segmentRange)

	var cells []SeatCell
	for seatID, detail := range seatDetails {
		status := detail.Status
		if seatSegmentAvail[seatID] == requiredSegments {
			status = domain.SeatStatusAvailable
		}

		cells = append(cells, SeatCell{
			SeatID:     seatID,
			SeatNumber: detail.SeatNumber,
			SeatType:   detail.SeatType,
			SeatClass:  detail.SeatClass,
			Status:     status,
			PricePaisa: detail.PricePaisa,
		})
	}

	return &SeatMapResult{
		Rows: []SeatRow{{Seats: cells}}, // Simplified - in production, group by row
		Legend: map[string]string{
			"available": "#00FF00",
			"held":      "#FFFF00",
			"booked":    "#FF0000",
			"blocked":   "#808080",
		},
	}
}
