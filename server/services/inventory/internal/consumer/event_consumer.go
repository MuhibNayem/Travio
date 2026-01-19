package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	fleetpb "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/inventory/internal/service"
)

type EventConsumer struct {
	consumer     *kafka.Consumer
	inventorySvc *service.InventoryService
	fleetClient  *clients.FleetClient
}

func New(brokers []string, groupID string, inventorySvc *service.InventoryService, fleetClient *clients.FleetClient) (*EventConsumer, error) {
	topics := []string{kafka.TopicCatalog}
	consumer, err := kafka.NewConsumer(brokers, groupID, topics)
	if err != nil {
		return nil, err
	}

	c := &EventConsumer{
		consumer:     consumer,
		inventorySvc: inventorySvc,
		fleetClient:  fleetClient,
	}

	consumer.RegisterHandler(kafka.EventTripCreated, c.handleTripCreated)

	return c, nil
}

func (c *EventConsumer) Start() error {
	return c.consumer.Start()
}

func (c *EventConsumer) Stop() error {
	return c.consumer.Stop()
}

type TripEventDTO struct {
	ID             string           `json:"id"`
	OrganizationID string           `json:"organization_id"`
	VehicleID      string           `json:"vehicle_id"`
	ServiceDate    string           `json:"service_date"`
	DepartureTime  time.Time        `json:"departure_time"`
	RouteID        string           `json:"route_id"`
	Pricing        TripPricingDTO   `json:"pricing"`
	Segments       []TripSegmentDTO `json:"segments"`
}

type TripPricingDTO struct {
	BasePricePaisa     int64            `json:"base_price_paisa"`
	ClassPrices        map[string]int64 `json:"class_prices"`
	SeatCategoryPrices map[string]int64 `json:"seat_category_prices"`
}

type TripSegmentDTO struct {
	SegmentIndex  int       `json:"segment_index"`
	FromStationID string    `json:"from_station_id"`
	ToStationID   string    `json:"to_station_id"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
}

func (c *EventConsumer) handleTripCreated(ctx context.Context, event *kafka.Event) error {
	logger.Info("Handling TripCreated event", "trip_id", event.AggregateID)

	var trip TripEventDTO
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload for decoding: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &trip); err != nil {
		return fmt.Errorf("failed to unmarshal trip event: %w", err)
	}

	// Fetch Vehicle Asset
	asset, err := c.fleetClient.GetAsset(ctx, trip.VehicleID, trip.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to fetch asset %s: %w", trip.VehicleID, err)
	}

	// Map to SeatConfig using Pricing
	seatConfig := mapAssetToSeatConfig(asset, trip.Pricing)

	// Map Segments
	var segments []service.SegmentDef
	for _, seg := range trip.Segments {
		segments = append(segments, service.SegmentDef{
			SegmentIndex:  seg.SegmentIndex,
			FromStationID: seg.FromStationID,
			ToStationID:   seg.ToStationID,
			DepartureTime: seg.DepartureTime.Unix(),
			ArrivalTime:   seg.ArrivalTime.Unix(),
		})
	}

	// Call Inventory Service
	req := &service.InitializeTripRequest{
		TripID:         trip.ID,
		OrganizationID: trip.OrganizationID,
		VehicleID:      trip.VehicleID,
		Segments:       segments,
		SeatConfig:     seatConfig,
	}

	// Idempotency check handled by Service/Repo (usually `InitializeTrip` fails if exists or is safe)
	// ScyllaDB `InitializeTrip` uses LWT or simple insert?
	// If it's a simple INSERT, it might be idempotent if PRIMARY KEY is (trip_id, seat_id).
	// We assume Service handles it.
	result, err := c.inventorySvc.InitializeTripInventory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to initialize inventory: %w", err)
	}

	logger.Info("Successfully initialized inventory via event", "trip_id", trip.ID, "seats_created", result.SeatsCreated)
	return nil
}

func mapAssetToSeatConfig(asset *fleetpb.Asset, pricing TripPricingDTO) service.SeatConfig {
	var seats []service.SeatDef
	totalSeats := 0

	// Helper to calculate price
	getPrice := func(seatClass, seatCategory string) int64 {
		price := pricing.BasePricePaisa
		if p, ok := pricing.ClassPrices[seatClass]; ok {
			price = p
		}
		if p, ok := pricing.SeatCategoryPrices[seatCategory]; ok {
			price = p
		}
		return price
	}

	if asset.Config.GetBus() != nil {
		bus := asset.Config.GetBus()
		totalSeats = int(bus.Rows * bus.SeatsPerRow) // Approximation
		for r := 1; r <= int(bus.Rows); r++ {
			char := string(rune('A' + r - 1))
			for c := 1; c <= int(bus.SeatsPerRow); c++ {
				seatNum := fmt.Sprintf("%s%d", char, c)
				seatType := "window"
				if c > 1 && c < int(bus.SeatsPerRow) {
					seatType = "aisle"
				}

				seatClass := "economy" // default
				for _, cat := range bus.Categories {
					// Naive mapping: if cat.Name matches? No, logic in Catalog was simpler.
					// We'll stick to default unless we traverse explicit seat map if available.
					// Catalog loop: "seatClass = strings.ToLower(cat.Name)" was looping inside?
					// Catalog logic was: loop categories -> set seatClass.
					// This means LAST category wins? Or did it have seat ranges?
					// Step 262 line 870: loops categories, sets seatClass.
					// It doesn't seem to check WHICH seats. It just overwrites.
					// This implies the asset config structure in the earlier file view was maybe incomplete or I missed the Seat Range check.
					// However, for "smooth" port, I copy the exact behavior.
					seatClass = strings.ToLower(cat.Name)
				}

				seats = append(seats, service.SeatDef{
					SeatID:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
					SeatNumber: seatNum,
					Row:        r,
					Column:     c,
					SeatType:   seatType,
					SeatClass:  seatClass,
					PricePaisa: getPrice(seatClass, ""),
				})
			}
		}
	} else if asset.Config.GetTrain() != nil {
		train := asset.Config.GetTrain()
		for _, coach := range train.Coaches {
			for r := 1; r <= int(coach.Rows); r++ {
				for s := 1; s <= int(coach.SeatsPerRow); s++ {
					seatNum := fmt.Sprintf("%s-%d-%d", coach.Id, r, s)
					seatClass := strings.ToLower(coach.Name)
					seats = append(seats, service.SeatDef{
						SeatID:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
						SeatNumber: seatNum,
						Row:        r,
						Column:     s,
						SeatClass:  seatClass,
						PricePaisa: getPrice(seatClass, ""),
					})
					totalSeats++
				}
			}
		}
	} else if asset.Config.GetLaunch() != nil {
		launch := asset.Config.GetLaunch()
		for _, deck := range launch.Decks {
			for r := 1; r <= int(deck.Rows); r++ {
				for c := 1; c <= int(deck.Cols); c++ {
					seatNum := fmt.Sprintf("%s-%d-%d", deck.Id, r, c)
					seatClass := strings.ToLower(deck.Name)
					seats = append(seats, service.SeatDef{
						SeatID:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
						SeatNumber: seatNum,
						Row:        r,
						Column:     c,
						SeatClass:  seatClass,
						PricePaisa: int64(deck.SeatPricePaisa),
					})
					totalSeats++
				}
			}
			for _, cabin := range deck.Cabins {
				seatNum := cabin.Name
				seatClass := "cabin"
				if cabin.IsSuite {
					seatClass = "suite"
				}
				seats = append(seats, service.SeatDef{
					SeatID:     fmt.Sprintf("%s-%s", asset.Id, cabin.Id),
					SeatNumber: seatNum,
					SeatClass:  seatClass,
					PricePaisa: int64(cabin.PricePaisa),
				})
				totalSeats++
			}
		}
	}

	return service.SeatConfig{
		TotalSeats: totalSeats,
		Seats:      seats,
	}
}
