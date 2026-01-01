package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLimitExceeded = errors.New("ticket limit exceeded")
)

// TicketLimitChecker enforces per-user/IP/NID ticket limits using Redis
type TicketLimitChecker struct {
	client *redis.Client
}

func NewTicketLimitChecker(client *redis.Client) *TicketLimitChecker {
	return &TicketLimitChecker{client: client}
}

// CheckAndIncrement verifies limits and increments counters atomically
// Returns error if any limit is exceeded
func (c *TicketLimitChecker) CheckAndIncrement(ctx context.Context, tripID, userID, ipAddress, nid string, quantity int, limits TicketLimit) error {
	// Use Redis transaction (MULTI/EXEC) for atomic check-and-increment

	// Keys
	userKey := fmt.Sprintf("ticketlimit:%s:user:%s", tripID, userID)
	ipKey := fmt.Sprintf("ticketlimit:%s:ip:%s", tripID, ipAddress)
	nidKey := fmt.Sprintf("ticketlimit:%s:nid:%s", tripID, nid)
	hourlyKey := fmt.Sprintf("ticketlimit:hourly:%s", userID)

	// Lua script for atomic check-and-increment
	script := redis.NewScript(`
		local userKey = KEYS[1]
		local ipKey = KEYS[2]
		local nidKey = KEYS[3]
		local hourlyKey = KEYS[4]
		
		local quantity = tonumber(ARGV[1])
		local maxUser = tonumber(ARGV[2])
		local maxIP = tonumber(ARGV[3])
		local maxNID = tonumber(ARGV[4])
		local maxHourly = tonumber(ARGV[5])
		local ttl = tonumber(ARGV[6])
		
		-- Check current values
		local userCount = tonumber(redis.call('GET', userKey) or 0)
		local ipCount = tonumber(redis.call('GET', ipKey) or 0)
		local nidCount = tonumber(redis.call('GET', nidKey) or 0)
		local hourlyCount = tonumber(redis.call('GET', hourlyKey) or 0)
		
		-- Validate limits
		if userCount + quantity > maxUser then
			return {err = "user_limit", current = userCount, max = maxUser}
		end
		if ipCount + quantity > maxIP then
			return {err = "ip_limit", current = ipCount, max = maxIP}
		end
		if nidCount + quantity > maxNID then
			return {err = "nid_limit", current = nidCount, max = maxNID}
		end
		if hourlyCount + quantity > maxHourly then
			return {err = "hourly_limit", current = hourlyCount, max = maxHourly}
		end
		
		-- All checks passed, increment counters
		redis.call('INCRBY', userKey, quantity)
		redis.call('EXPIRE', userKey, ttl)
		
		redis.call('INCRBY', ipKey, quantity)
		redis.call('EXPIRE', ipKey, ttl)
		
		redis.call('INCRBY', nidKey, quantity)
		redis.call('EXPIRE', nidKey, ttl)
		
		redis.call('INCRBY', hourlyKey, quantity)
		redis.call('EXPIRE', hourlyKey, 3600)  -- 1 hour for hourly limit
		
		return {ok = true}
	`)

	// 24 hour TTL for trip-specific limits
	ttl := 24 * 3600

	result, err := script.Run(ctx, c.client,
		[]string{userKey, ipKey, nidKey, hourlyKey},
		quantity,
		limits.MaxTicketsPerUser,
		limits.MaxTicketsPerIP,
		limits.MaxTicketsPerNID,
		limits.MaxTicketsPerHour,
		ttl,
	).Result()

	if err != nil {
		return fmt.Errorf("limit check failed: %w", err)
	}

	// Parse result
	if resultMap, ok := result.(map[interface{}]interface{}); ok {
		if errType, hasErr := resultMap["err"]; hasErr {
			return fmt.Errorf("%w: %s", ErrLimitExceeded, errType)
		}
	}

	return nil
}

// GetCurrentCounts returns current ticket counts for display
type TicketCounts struct {
	UserCount   int64 `json:"user_count"`
	IPCount     int64 `json:"ip_count"`
	NIDCount    int64 `json:"nid_count"`
	HourlyCount int64 `json:"hourly_count"`
}

func (c *TicketLimitChecker) GetCurrentCounts(ctx context.Context, tripID, userID, ipAddress, nid string) (*TicketCounts, error) {
	userKey := fmt.Sprintf("ticketlimit:%s:user:%s", tripID, userID)
	ipKey := fmt.Sprintf("ticketlimit:%s:ip:%s", tripID, ipAddress)
	nidKey := fmt.Sprintf("ticketlimit:%s:nid:%s", tripID, nid)
	hourlyKey := fmt.Sprintf("ticketlimit:hourly:%s", userID)

	pipe := c.client.Pipeline()
	userCmd := pipe.Get(ctx, userKey)
	ipCmd := pipe.Get(ctx, ipKey)
	nidCmd := pipe.Get(ctx, nidKey)
	hourlyCmd := pipe.Get(ctx, hourlyKey)

	pipe.Exec(ctx)

	counts := &TicketCounts{}
	counts.UserCount, _ = userCmd.Int64()
	counts.IPCount, _ = ipCmd.Int64()
	counts.NIDCount, _ = nidCmd.Int64()
	counts.HourlyCount, _ = hourlyCmd.Int64()

	return counts, nil
}

// ReleaseTickets decrements counters (for order cancellation)
func (c *TicketLimitChecker) ReleaseTickets(ctx context.Context, tripID, userID, ipAddress, nid string, quantity int) error {
	userKey := fmt.Sprintf("ticketlimit:%s:user:%s", tripID, userID)
	ipKey := fmt.Sprintf("ticketlimit:%s:ip:%s", tripID, ipAddress)
	nidKey := fmt.Sprintf("ticketlimit:%s:nid:%s", tripID, nid)

	pipe := c.client.Pipeline()
	pipe.DecrBy(ctx, userKey, int64(quantity))
	pipe.DecrBy(ctx, ipKey, int64(quantity))
	pipe.DecrBy(ctx, nidKey, int64(quantity))
	_, err := pipe.Exec(ctx)

	return err
}

// --- NID Deduplication ---

// CheckNIDUsed checks if an NID is already used for a trip
func (c *TicketLimitChecker) CheckNIDUsed(ctx context.Context, tripID string, nids []string) ([]string, error) {
	usedNIDs := []string{}

	for _, nid := range nids {
		key := fmt.Sprintf("ticketlimit:%s:nid:%s", tripID, nid)
		count, err := c.client.Get(ctx, key).Int64()
		if err == nil && count > 0 {
			usedNIDs = append(usedNIDs, nid)
		}
	}

	return usedNIDs, nil
}

// --- Hold Management (with TTL) ---

type HoldManager struct {
	client *redis.Client
}

func NewHoldManager(client *redis.Client) *HoldManager {
	return &HoldManager{client: client}
}

// CreateHold creates a seat hold with automatic TTL expiration
func (h *HoldManager) CreateHold(ctx context.Context, holdID, tripID, seatID, userID, sessionID, ip string, ttl time.Duration) error {
	// Key for the specific seat
	seatKey := fmt.Sprintf("hold:seat:%s:%s", tripID, seatID)
	userHoldsKey := fmt.Sprintf("hold:user:%s", userID)

	// Check if seat already held
	exists, err := h.client.Exists(ctx, seatKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return errors.New("seat already held")
	}

	// Check user's hold count
	userHoldCount, _ := h.client.SCard(ctx, userHoldsKey).Result()
	if userHoldCount >= 2 { // MaxHoldsPerUser
		return errors.New("maximum concurrent holds exceeded")
	}

	// Create hold with TTL
	holdData := map[string]interface{}{
		"hold_id":    holdID,
		"trip_id":    tripID,
		"seat_id":    seatID,
		"user_id":    userID,
		"session_id": sessionID,
		"ip":         ip,
		"created_at": time.Now().Unix(),
	}

	pipe := h.client.Pipeline()
	pipe.HSet(ctx, seatKey, holdData)
	pipe.Expire(ctx, seatKey, ttl)
	pipe.SAdd(ctx, userHoldsKey, holdID)
	pipe.Expire(ctx, userHoldsKey, ttl)
	_, err = pipe.Exec(ctx)

	return err
}

// ReleaseHold manually releases a hold (for checkout or cancel)
func (h *HoldManager) ReleaseHold(ctx context.Context, tripID, seatID, userID, holdID string) error {
	seatKey := fmt.Sprintf("hold:seat:%s:%s", tripID, seatID)
	userHoldsKey := fmt.Sprintf("hold:user:%s", userID)

	pipe := h.client.Pipeline()
	pipe.Del(ctx, seatKey)
	pipe.SRem(ctx, userHoldsKey, holdID)
	_, err := pipe.Exec(ctx)

	return err
}

// GetHold retrieves hold information
func (h *HoldManager) GetHold(ctx context.Context, tripID, seatID string) (map[string]string, error) {
	seatKey := fmt.Sprintf("hold:seat:%s:%s", tripID, seatID)
	return h.client.HGetAll(ctx, seatKey).Result()
}

// GetUserHolds returns all active holds for a user
func (h *HoldManager) GetUserHolds(ctx context.Context, userID string) ([]string, error) {
	userHoldsKey := fmt.Sprintf("hold:user:%s", userID)
	return h.client.SMembers(ctx, userHoldsKey).Result()
}

type TicketLimit struct {
	MaxTicketsPerUser int
	MaxTicketsPerIP   int
	MaxTicketsPerNID  int
	MaxTicketsPerHour int
	MaxHoldsPerUser   int
}
