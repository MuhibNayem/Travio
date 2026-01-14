package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	orderpb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/service"
)

// OrderEventConsumer consumes order events and triggers ticket generation
type OrderEventConsumer struct {
	consumer           *kafka.Consumer
	fulfillmentService *service.FulfillmentService
	catalogClient      CatalogClient
	orderClient        OrderClient
}

// NewOrderEventConsumer creates a new consumer for order events
func NewOrderEventConsumer(brokers []string, fulfillmentSvc *service.FulfillmentService, catalogClient CatalogClient, orderClient OrderClient) (*OrderEventConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "fulfillment-service", []string{kafka.TopicOrders})
	if err != nil {
		return nil, err
	}

	c := &OrderEventConsumer{
		consumer:           consumer,
		fulfillmentService: fulfillmentSvc,
		catalogClient:      catalogClient,
		orderClient:        orderClient,
	}

	// Register handlers
	consumer.RegisterHandler(kafka.EventOrderConfirmed, c.handleOrderConfirmed)

	return c, nil
}

// OrderConfirmedPayload matches the event structure from order service
type OrderConfirmedPayload struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	TripID         string `json:"trip_id"`
	BookingID      string `json:"booking_id"`
	PaymentID      string `json:"payment_id"`
	TotalPaisa     int64  `json:"total_paisa"`
	ContactEmail   string `json:"contact_email"`
	ContactPhone   string `json:"contact_phone"`
}

// handleOrderConfirmed processes OrderConfirmed events
func (c *OrderEventConsumer) handleOrderConfirmed(ctx context.Context, event *kafka.Event) error {
	logger.Info("received OrderConfirmed event",
		"event_id", event.ID,
		"order_id", event.AggregateID,
	)

	// Parse payload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		logger.Error("failed to marshal payload", "error", err)
		return err
	}

	var payload OrderConfirmedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		logger.Error("failed to unmarshal OrderConfirmed payload", "error", err)
		return err
	}

    order, err := c.orderClient.GetOrder(ctx, payload.OrderID, payload.UserID)
	if err != nil {
		logger.Error("failed to fetch order", "error", err)
		return err
	}

	trip, err := c.catalogClient.GetTrip(ctx, payload.OrganizationID, payload.TripID)
	if err != nil {
		logger.Error("failed to fetch trip", "error", err)
		return err
	}

	route, err := c.catalogClient.GetRoute(ctx, payload.OrganizationID, trip.RouteId)
	if err != nil {
		logger.Error("failed to fetch route", "error", err)
		return err
	}

	origin, err := c.catalogClient.GetStation(ctx, payload.OrganizationID, order.FromStationId)
	if err != nil {
		logger.Error("failed to fetch origin station", "error", err)
		return err
	}

	destination, err := c.catalogClient.GetStation(ctx, payload.OrganizationID, order.ToStationId)
	if err != nil {
		logger.Error("failed to fetch destination station", "error", err)
		return err
	}

	passengers := buildPassengerSeats(order, payload.TotalPaisa)
	if len(passengers) == 0 {
		logger.Error("no passengers found for order", "order_id", payload.OrderID)
		return fmt.Errorf("no passengers found for order %s", payload.OrderID)
	}

	req := &service.GenerateTicketsReq{
		BookingID:      payload.BookingID,
		OrderID:        payload.OrderID,
		OrganizationID: payload.OrganizationID,
		TripID:         payload.TripID,
		RouteName:      route.Name,
		FromStation:    origin.Name,
		ToStation:      destination.Name,
		DepartureTime:  time.Unix(trip.DepartureTime, 0),
		ArrivalTime:    time.Unix(trip.ArrivalTime, 0),
		Passengers:     passengers,
		ContactEmail:   order.ContactEmail,
		ContactPhone:   order.ContactPhone,
	}

	// Generate tickets
	result, err := c.fulfillmentService.GenerateTickets(ctx, req)
	if err != nil {
		logger.Error("failed to generate tickets",
			"order_id", payload.OrderID,
			"error", err,
		)
		return err
	}

	logger.Info("tickets generated successfully",
		"order_id", payload.OrderID,
		"ticket_count", len(result.Tickets),
	)

	return nil
}

func buildPassengerSeats(order *orderpb.Order, totalPaisa int64) []service.PassengerSeat {
	if order == nil {
		return nil
	}

	seatPrices := make(map[string]int64)
	seatClass := make(map[string]string)
	seatNumber := make(map[string]string)
	for _, seat := range order.Seats {
		seatPrices[seat.SeatId] = seat.PricePaisa
		seatClass[seat.SeatId] = seat.SeatClass
		seatNumber[seat.SeatId] = seat.SeatNumber
	}

	var passengers []service.PassengerSeat
	for _, p := range order.Passengers {
		price := seatPrices[p.SeatId]
		if price == 0 && totalPaisa > 0 {
			price = totalPaisa / int64(max(1, len(order.Passengers)))
		}
		seatNum := p.SeatNumber
		if seatNum == "" {
			seatNum = seatNumber[p.SeatId]
		}
		seatCls := p.SeatClass
		if seatCls == "" {
			seatCls = seatClass[p.SeatId]
		}
		passengers = append(passengers, service.PassengerSeat{
			NID:        p.Nid,
			Name:       p.Name,
			SeatID:     p.SeatId,
			SeatNumber: seatNum,
			SeatClass:  seatCls,
			PricePaisa: price,
		})
	}

	if len(passengers) == 0 && len(order.Seats) > 0 {
		for _, seat := range order.Seats {
			price := seat.PricePaisa
			if price == 0 && totalPaisa > 0 {
				price = totalPaisa / int64(max(1, len(order.Seats)))
			}
			passengers = append(passengers, service.PassengerSeat{
				SeatID:     seat.SeatId,
				SeatNumber: seat.SeatNumber,
				SeatClass:  seat.SeatClass,
				PricePaisa: price,
			})
		}
	}

	return passengers
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Start begins consuming events
func (c *OrderEventConsumer) Start() error {
	return c.consumer.Start()
}

// Stop stops the consumer
func (c *OrderEventConsumer) Stop() error {
	return c.consumer.Stop()
}
