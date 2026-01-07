package searcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/opensearch-project/opensearch-go/v2"
)

type Searcher struct {
	client *opensearch.Client
}

func New(client *opensearch.Client) *Searcher {
	return &Searcher{client: client}
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

	return trips, total, nil
}

type StationDocument struct {
	StationID string `json:"station_id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Division  string `json:"division"`
}

func (s *Searcher) SearchStations(ctx context.Context, query string, limit int) ([]StationDocument, error) {
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

	return stations, nil
}
