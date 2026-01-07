package searcher

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/redis/go-redis/v9"
)

const (
	TripCacheTTL    = 5 * time.Minute
	StationCacheTTL = 1 * time.Hour
)

type Searcher struct {
	client *opensearch.Client
	rdb    *redis.Client
}

func New(client *opensearch.Client, rdb *redis.Client) *Searcher {
	return &Searcher{client: client, rdb: rdb}
}

// generateCacheKey creates a deterministic hash for the query parameters
func generateCacheKey(prefix string, params ...interface{}) string {
	h := sha256.New()
	for _, p := range params {
		h.Write([]byte(fmt.Sprintf("%v", p)))
	}
	return prefix + ":" + hex.EncodeToString(h.Sum(nil))[:16]
}

type TripDocument struct {
	TripID          string `json:"trip_id"`
	RouteName       string `json:"route_name"`
	DepartureTime   string `json:"departure_time"`
	ArrivalTime     string `json:"arrival_time"`
	PricePaisa      int64  `json:"price_paisa"`
	OperatorName    string `json:"operator_name"`
	VehicleType     string `json:"vehicle_type"`
	AvailableSeats  int    `json:"available_seats"`
	FromStationName string `json:"from_station_name"`
	ToStationName   string `json:"to_station_name"`
	FromStationID   string `json:"from_station_id"`
	ToStationID     string `json:"to_station_id"`
	Date            string `json:"date"` // YYYY-MM-DD
}

func (s *Searcher) SearchTrips(ctx context.Context, query, fromID, toID, date string, limit, offset int) ([]TripDocument, int64, error) {
	// Generate cache key
	cacheKey := generateCacheKey("trips", query, fromID, toID, date, limit, offset)

	// Check Redis cache
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var result struct {
				Trips []TripDocument `json:"trips"`
				Total int64          `json:"total"`
			}
			if json.Unmarshal([]byte(cached), &result) == nil {
				return result.Trips, result.Total, nil
			}
		}
	}

	// Build query
	mustClauses := []map[string]interface{}{}

	if fromID != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{"from_station_id": fromID},
		})
	}
	if toID != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{"to_station_id": toID},
		})
	}
	if date != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{"date": date},
		})
	}
	if query != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"route_name", "operator_name", "vehicle_type"},
			},
		})
	}

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustClauses,
			},
		},
		"from": offset,
		"size": limit,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, 0, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex("trips"),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("search error: %s", res.Status())
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, 0, err
	}

	hits := r["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))

	var trips []TripDocument
	for _, hit := range hits["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, _ := json.Marshal(source)
		var trip TripDocument
		json.Unmarshal(sourceBytes, &trip)
		trips = append(trips, trip)
	}

	// Cache the result
	if s.rdb != nil {
		cacheData, _ := json.Marshal(struct {
			Trips []TripDocument `json:"trips"`
			Total int64          `json:"total"`
		}{Trips: trips, Total: total})
		s.rdb.Set(ctx, cacheKey, cacheData, TripCacheTTL)
	}

	return trips, total, nil
}

type StationDocument struct {
	StationID string `json:"station_id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Division  string `json:"division"`
}

func (s *Searcher) SearchStations(ctx context.Context, query string, limit int) ([]StationDocument, error) {
	// Generate cache key
	cacheKey := generateCacheKey("stations", query, limit)

	// Check Redis cache
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			var stations []StationDocument
			if json.Unmarshal([]byte(cached), &stations) == nil {
				return stations, nil
			}
		}
	}

	searchBody := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"name", "location", "division"},
				"fuzziness": "AUTO",
			},
		},
		"size": limit,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchBody)

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex("stations"),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	var r map[string]interface{}
	json.NewDecoder(res.Body).Decode(&r)

	hits := r["hits"].(map[string]interface{})

	var stations []StationDocument
	for _, hit := range hits["hits"].([]interface{}) {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, _ := json.Marshal(source)
		var station StationDocument
		json.Unmarshal(sourceBytes, &station)
		stations = append(stations, station)
	}

	// Cache the result
	if s.rdb != nil {
		cacheData, _ := json.Marshal(stations)
		s.rdb.Set(ctx, cacheKey, cacheData, StationCacheTTL)
	}

	return stations, nil
}
