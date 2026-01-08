package profile

import (
	"math"
	"time"
)

// Analyzer calculates behavioral deviation from user profiles.
type Analyzer struct {
	deviationThreshold float64
}

// NewAnalyzer creates a new profile analyzer.
func NewAnalyzer(deviationThreshold float64) *Analyzer {
	if deviationThreshold <= 0 {
		deviationThreshold = 2.0 // Default: 2 standard deviations
	}
	return &Analyzer{
		deviationThreshold: deviationThreshold,
	}
}

// AnalyzeDeviation calculates how much the current booking deviates from the user's profile.
func (a *Analyzer) AnalyzeDeviation(profile *UserProfile, event *BookingEvent) *DeviationResult {
	result := &DeviationResult{}

	// New user - limited analysis
	if profile == nil || profile.TotalBookings < 3 {
		result.IsNewUser = true
		result.Score = 10 // Small baseline risk for new users
		return result
	}

	// 1. Value deviation
	if profile.AvgBookingValue > 0 {
		stdDev := profile.AvgBookingValue * 0.3 // Assume 30% std dev
		result.ValueDeviation = math.Abs(float64(event.AmountPaisa)-profile.AvgBookingValue) / stdDev
		if result.ValueDeviation > a.deviationThreshold {
			result.Score += 15
		}
	}

	// 2. Velocity deviation
	if profile.BookingVelocity24h > 0 {
		// Check if booking 24h ago, estimate current velocity
		hoursSinceLastBooking := time.Since(profile.LastSeen).Hours()
		if hoursSinceLastBooking < 24 {
			currentVelocity := 24.0 / hoursSinceLastBooking // Projected daily rate
			result.VelocityDeviation = currentVelocity / profile.BookingVelocity24h
			if result.VelocityDeviation > 3.0 { // 3x normal velocity
				result.Score += 25
			} else if result.VelocityDeviation > 2.0 {
				result.Score += 15
			}
		}
	}

	// 3. Time deviation
	bookingHour := event.BookingTime.Hour()
	isCommonTime := false
	for _, h := range profile.CommonTimes {
		if h == bookingHour || abs(h-bookingHour) <= 2 { // Within 2 hours
			isCommonTime = true
			break
		}
	}
	if !isCommonTime && len(profile.CommonTimes) > 0 {
		result.TimeDeviation = 1.0
		result.Score += 10
	}

	// 4. Route deviation
	isKnownRoute := false
	for _, r := range profile.CommonRoutes {
		if r == event.Route {
			isKnownRoute = true
			break
		}
	}
	if !isKnownRoute && len(profile.CommonRoutes) > 0 {
		result.RouteDeviation = 1.0
		result.Score += 5 // New route is mildly suspicious
	}

	// 5. IP deviation
	isKnownIP := false
	for _, ip := range profile.CommonIPs {
		if ip == event.IPAddress {
			isKnownIP = true
			break
		}
	}
	if !isKnownIP && len(profile.CommonIPs) > 0 {
		result.IPDeviation = 1.0
		result.Score += 10
	}

	// 6. Device deviation
	isKnownDevice := false
	for _, d := range profile.DeviceFingerprints {
		if d == event.UserAgent {
			isKnownDevice = true
			break
		}
	}
	if !isKnownDevice && len(profile.DeviceFingerprints) > 0 {
		result.DeviceDeviation = 1.0
		result.Score += 10
	}

	// 7. High risk history
	if profile.AvgRiskScore > 50 {
		result.HighRiskHistory = true
		result.Score += 15
	}
	if profile.BlockedCount > 0 {
		result.Score += float64(profile.BlockedCount) * 10
	}

	// Cap at 100
	if result.Score > 100 {
		result.Score = 100
	}

	result.IsAnomalous = result.Score >= 30

	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
