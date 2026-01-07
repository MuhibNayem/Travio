package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	PricingURL = "http://localhost:50058/api/v1/pricing/calculate"
	Workers    = 20
	Requests   = 200
)

type CalculatePriceRequest struct {
	TripID         string  `json:"trip_id"`
	SeatClass      string  `json:"seat_class"`
	Date           string  `json:"date"`
	Quantity       int32   `json:"quantity"`
	BasePricePaisa int64   `json:"base_price_paisa"`
	OccupancyRate  float64 `json:"occupancy_rate"`
}

type CalculatePriceResponse struct {
	FinalPricePaisa int64 `json:"final_price_paisa"`
	BasePricePaisa  int64 `json:"base_price_paisa"`
	AppliedRules    []struct {
		RuleID     string  `json:"rule_id"`
		RuleName   string  `json:"rule_name"`
		Multiplier float64 `json:"multiplier"`
	} `json:"applied_rules"`
}

func main() {
	fmt.Println("🚀 Starting Pricing Service Load Test")
	fmt.Printf("Sending %d requests with %d workers\n", Requests, Workers)

	testCases := []CalculatePriceRequest{
		{TripID: "TRIP-001", SeatClass: "economy", Date: "2026-01-11", Quantity: 2, BasePricePaisa: 100000, OccupancyRate: 0.5},  // Weekend
		{TripID: "TRIP-002", SeatClass: "business", Date: "2026-02-15", Quantity: 1, BasePricePaisa: 100000, OccupancyRate: 0.3}, // Early + Business
		{TripID: "TRIP-003", SeatClass: "economy", Date: "2026-01-09", Quantity: 1, BasePricePaisa: 100000, OccupancyRate: 0.95}, // Last minute + High demand
	}

	var wg sync.WaitGroup
	start := time.Now()
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < Requests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := testCases[idx%len(testCases)]

			body, _ := json.Marshal(req)
			resp, err := http.Post(PricingURL, "application/json", bytes.NewReader(body))
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()

				// Print first few responses
				if idx < 3 {
					respBody, _ := io.ReadAll(resp.Body)
					var r CalculatePriceResponse
					json.Unmarshal(respBody, &r)
					fmt.Printf("\n📊 Test Case %d:\n", idx+1)
					fmt.Printf("   Base: %d → Final: %d (%.0f%%)\n", r.BasePricePaisa, r.FinalPricePaisa, float64(r.FinalPricePaisa)/float64(r.BasePricePaisa)*100)
					for _, rule := range r.AppliedRules {
						fmt.Printf("   ✓ %s (×%.2f)\n", rule.RuleName, rule.Multiplier)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Test Completed in %v\n", duration)
	fmt.Printf("Success Rate: %d/%d (%.1f%%)\n", successCount, Requests, float64(successCount)/float64(Requests)*100)
	fmt.Printf("Throughput: %.1f req/s\n", float64(Requests)/duration.Seconds())
}
