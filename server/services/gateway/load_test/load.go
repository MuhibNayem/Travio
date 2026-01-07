package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Config
const (
	TargetURL = "http://localhost:8888/v1/auth/sessions" // Example endpoint
	TotalReqs = 100
	Workers   = 10
)

func main() {
	fmt.Println("🚀 Starting Gateway Resiliency Load Test...")
	fmt.Printf("Target: %s\n", TargetURL)
	fmt.Println("Instructions: Stop the Identity Service during this test to see Circuit Breaker open (503).")
	time.Sleep(2 * time.Second)

	var wg sync.WaitGroup
	results := make(chan int, TotalReqs)

	start := time.Now()

	for i := 0; i < Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < TotalReqs/Workers; j++ {
				resp, err := http.Get(TargetURL)
				if err != nil {
					fmt.Printf("Request failed: %v\n", err)
					results <- 0
					continue
				}
				results <- resp.StatusCode
				resp.Body.Close()
				time.Sleep(100 * time.Millisecond) // Slight delay
			}
		}()
	}

	wg.Wait()
	close(results)
	duration := time.Since(start)

	statusCodes := make(map[int]int)
	for code := range results {
		statusCodes[code]++
	}

	fmt.Printf("\n✅ Load Test Completed in %v\n", duration)
	fmt.Println("Status Code Distribution:")
	for code, count := range statusCodes {
		fmt.Printf("[%d]: %d requests\n", code, count)
		if code == 503 {
			fmt.Println("   -> Circuit Breaker ACTIVATED (Service Unavailable)")
		}
	}
}
