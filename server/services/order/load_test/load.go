package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	OrderURL       = "http://localhost:8085/v1/orders" // Updated port
	IdempotencyKey = "order-idempotency-test-123"
	TotalReqs      = 10
)

func main() {
	fmt.Println("🚀 Starting Order Service Idempotency Test...")
	fmt.Printf("Target: %s\n", OrderURL)
	fmt.Printf("Key: %s\n", IdempotencyKey)

	var (
		processedCount int32
		conflictCount  int32
		hitCount       int32
		errorCount     int32
	)

	var wg sync.WaitGroup
	start := time.Now()

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < TotalReqs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Mock Body
			body := []byte(`{"user_id":"test-user","trip_id":"trip-123","passengers":[]}`)
			// Note: This body likely fails validation, but middleware runs first!

			req, _ := http.NewRequest("POST", OrderURL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", IdempotencyKey)

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				fmt.Printf("[%d] Error: %v\n", id, err)
				return
			}
			defer resp.Body.Close()

			// Check Headers/Status
			if resp.Header.Get("X-Idempotency-Hit") == "true" {
				atomic.AddInt32(&hitCount, 1)
				fmt.Printf("[%d] 200 OK (Cache Hit)\n", id)
			} else if resp.StatusCode == http.StatusConflict {
				atomic.AddInt32(&conflictCount, 1)
				fmt.Printf("[%d] 409 Conflict (Processing)\n", id)
			} else if resp.StatusCode == 200 || resp.StatusCode == 201 {
				atomic.AddInt32(&processedCount, 1)
				fmt.Printf("[%d] 200 OK (PROCESSED)\n", id)
			} else {
				// We expect middleware to run regardless of backend validity,
				// but backend might return 400 for bad body, which middleware caches!
				bodyBytes, _ := io.ReadAll(resp.Body)
				fmt.Printf("[%d] Status %d: %s\n", id, resp.StatusCode, string(bodyBytes))
				if resp.StatusCode >= 400 && resp.StatusCode < 500 {
					// Middleware caches 4xx too? No, currently only 2xx.
					// If backend returns 400, middleware deletes lock.
					atomic.AddInt32(&errorCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Total Requests: %d\n", TotalReqs)
	fmt.Printf("Processed (First): %d (Should be 1)\n", processedCount)
	fmt.Printf("Cached Hits: %d\n", hitCount)
	fmt.Printf("Conflicts (Processing): %d\n", conflictCount)
	fmt.Printf("Errors: %d\n", errorCount)
}
