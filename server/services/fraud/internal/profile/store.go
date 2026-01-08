package profile

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store handles user profile persistence.
type Store struct {
	db *gorm.DB
}

// NewStore creates a new profile store.
func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// AutoMigrate creates the necessary tables.
func (s *Store) AutoMigrate() error {
	return s.db.AutoMigrate(&UserProfile{})
}

// GetProfile retrieves a user profile by ID.
func (s *Store) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	var profile UserProfile
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// CreateOrUpdateProfile creates or updates a user profile.
func (s *Store) CreateOrUpdateProfile(ctx context.Context, profile *UserProfile) error {
	profile.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(profile).Error
}

// UpdateFromEvent updates a profile based on a booking event.
func (s *Store) UpdateFromEvent(ctx context.Context, event *BookingEvent) error {
	profile, err := s.GetProfile(ctx, event.UserID)
	if err != nil {
		return err
	}

	now := time.Now()

	if profile == nil {
		// New user
		profile = &UserProfile{
			UserID:             event.UserID,
			TotalBookings:      1,
			AvgBookingValue:    float64(event.AmountPaisa),
			BookingVelocity24h: 1.0,
			BookingVelocity7d:  1.0,
			CommonRoutes:       []string{event.Route},
			CommonTimes:        []int{event.BookingTime.Hour()},
			DeviceFingerprints: []string{event.UserAgent},
			CommonIPs:          []string{event.IPAddress},
			RiskScores:         []float64{event.RiskScore},
			AvgRiskScore:       event.RiskScore,
			FirstSeen:          now,
			LastSeen:           now,
			CreatedAt:          now,
		}
		if event.WasBlocked {
			profile.BlockedCount = 1
		}
	} else {
		// Update existing profile
		profile.TotalBookings++
		profile.LastSeen = now

		// Rolling average for booking value
		profile.AvgBookingValue = (profile.AvgBookingValue*float64(profile.TotalBookings-1) + float64(event.AmountPaisa)) / float64(profile.TotalBookings)

		// Update velocity (simple exponential moving average)
		alpha := 0.1
		profile.BookingVelocity24h = profile.BookingVelocity24h*(1-alpha) + 1.0*alpha

		// Add route if not in common routes (keep top 10)
		profile.CommonRoutes = appendUnique(profile.CommonRoutes, event.Route, 10)

		// Add time if not in common times
		profile.CommonTimes = appendUniqueInt(profile.CommonTimes, event.BookingTime.Hour(), 24)

		// Add device fingerprint
		profile.DeviceFingerprints = appendUnique(profile.DeviceFingerprints, event.UserAgent, 5)

		// Add IP
		profile.CommonIPs = appendUnique(profile.CommonIPs, event.IPAddress, 10)

		// Update risk history (keep last 10)
		profile.RiskScores = append(profile.RiskScores, event.RiskScore)
		if len(profile.RiskScores) > 10 {
			profile.RiskScores = profile.RiskScores[len(profile.RiskScores)-10:]
		}

		// Recalculate average risk score
		sum := 0.0
		for _, s := range profile.RiskScores {
			sum += s
		}
		profile.AvgRiskScore = sum / float64(len(profile.RiskScores))

		if event.WasBlocked {
			profile.BlockedCount++
		}
	}

	logger.Debug("Updated user profile",
		"user_id", event.UserID,
		"total_bookings", profile.TotalBookings,
		"avg_risk", profile.AvgRiskScore,
	)

	return s.CreateOrUpdateProfile(ctx, profile)
}

// appendUnique appends a string to a slice if not already present, keeping max items.
func appendUnique(slice []string, item string, max int) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	slice = append(slice, item)
	if len(slice) > max {
		slice = slice[len(slice)-max:]
	}
	return slice
}

// appendUniqueInt appends an int to a slice if not already present.
func appendUniqueInt(slice []int, item int, max int) []int {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	slice = append(slice, item)
	if len(slice) > max {
		slice = slice[len(slice)-max:]
	}
	return slice
}
