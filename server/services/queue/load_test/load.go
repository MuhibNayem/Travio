package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MuhibNayem/Travio/server/services/queue/internal/repository"
)

func main() {
	// Connect to Redis
	repo, err := repository.NewQueueRepository("localhost:6380")
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	eventID := "load-test-event"

	// Cleanup previous run
	repo.Leave(ctx, eventID, "cleanup")

	totalUsers := 1000000
	concurrency := 10000

	fmt.Printf("Starting load test: %d users, %d concurrent workers\n", totalUsers, concurrency)

	start := time.Now()
	var wg sync.WaitGroup
	var errorCount int64

	// Channel for jobs
	jobs := make(chan int, totalUsers)

	// Workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				userID := fmt.Sprintf("user-%d", id)
				_, err := repo.Join(ctx, eventID, userID, "session-id")
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					log.Printf("Join error: %v", err)
				}
			}
		}()
	}

	// Enqueue jobs
	for i := 0; i < totalUsers; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("Join phase completed in %v. Errors: %d\n", duration, errorCount)

	// Verify Queue Size
	stats, _ := repo.GetStats(ctx, eventID)
	if stats.TotalWaiting != totalUsers {
		fmt.Printf("FAIL: Expected %d waiting, got %d\n", totalUsers, stats.TotalWaiting)
	} else {
		fmt.Printf("PASS: Queue integrity verified. %d users in queue.\n", stats.TotalWaiting)
	}

	// Test Admission
	fmt.Println("Testing Admission...")
	admissionCount := 5000
	admitted, err := repo.AdmitNext(ctx, eventID, admissionCount, 1*time.Minute)
	if err != nil {
		log.Fatalf("Admission failed: %v", err)
	}

	if len(admitted) != admissionCount {
		fmt.Printf("FAIL: Expected %d admitted, got %d\n", admissionCount, len(admitted))
	} else {
		fmt.Printf("PASS: Admitted %d users successfully.\n", admissionCount)
	}

	// Verify updated stats
	stats, _ = repo.GetStats(ctx, eventID)
	expectedWaiting := totalUsers - admissionCount
	if stats.TotalWaiting != expectedWaiting {
		fmt.Printf("FAIL: Expected %d waiting after admission, got %d\n", expectedWaiting, stats.TotalWaiting)
	} else {
		fmt.Printf("PASS: Queue state updated correctly.\n")
	}
}
