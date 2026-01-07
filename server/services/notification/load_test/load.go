package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MuhibNayem/Travio/server/services/notification/internal/provider"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/service"
)

const (
	TotalNotifications = 100
	Workers            = 50
	EmailRatePerSecond = 10
)

func main() {
	fmt.Println("🚀 Starting Notification Load Test")
	fmt.Printf("Sending %d emails with rate limit of %d/s\n", TotalNotifications, EmailRatePerSecond)

	// Create rate-limited provider
	baseProvider := provider.NewConsoleEmailProvider()
	rateLimitedProvider := provider.NewRateLimitedEmailProvider(baseProvider, EmailRatePerSecond)

	notificationSvc := service.NewNotificationService(rateLimitedProvider, nil)

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < TotalNotifications; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			notificationSvc.SendEmail(ctx, &service.EmailRequest{
				To:       fmt.Sprintf("user%d@example.com", idx),
				Subject:  "Load Test Email",
				Template: "order_confirmed",
				Data: map[string]interface{}{
					"order_id":   fmt.Sprintf("ORD-%d", idx),
					"booking_id": fmt.Sprintf("BKG-%d", idx),
					"total":      100.00,
				},
			})
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	expectedMinDuration := time.Duration(TotalNotifications/EmailRatePerSecond) * time.Second
	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Expected minimum (with rate limiting): %v\n", expectedMinDuration)

	if duration >= expectedMinDuration {
		fmt.Println("✅ PASS: Rate limiting is working correctly")
	} else {
		fmt.Println("❌ FAIL: Rate limiting may not be active")
	}
}
