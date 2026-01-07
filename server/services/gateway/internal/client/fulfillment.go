package client

import (
	"context"
	"time"

	fulfillmentv1 "github.com/MuhibNayem/Travio/server/api/proto/fulfillment/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"google.golang.org/grpc"
)

// FulfillmentClient wraps the gRPC fulfillment client
type FulfillmentClient struct {
	conn   *grpc.ClientConn
	client fulfillmentv1.FulfillmentServiceClient
}

// NewFulfillmentClient creates a new gRPC client for the fulfillment service
// Uses mTLS if TLS config is provided
func NewFulfillmentClient(address string, tlsCfg TLSConfig) (*FulfillmentClient, error) {
	opts := GetDialOptions(tlsCfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to fulfillment service", "address", address, "tls", tlsCfg.CertFile != "")
	return &FulfillmentClient{
		conn:   conn,
		client: fulfillmentv1.NewFulfillmentServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *FulfillmentClient) Close() error {
	return c.conn.Close()
}

// GetTicket retrieves a ticket by ID
func (c *FulfillmentClient) GetTicket(ctx context.Context, ticketID string) (*fulfillmentv1.Ticket, error) {
	return c.client.GetTicket(ctx, &fulfillmentv1.GetTicketRequest{TicketId: ticketID})
}

// ListTickets lists tickets for an order
func (c *FulfillmentClient) ListTickets(ctx context.Context, orderID string) ([]*fulfillmentv1.Ticket, error) {
	resp, err := c.client.ListTickets(ctx, &fulfillmentv1.ListTicketsRequest{OrderId: orderID})
	if err != nil {
		return nil, err
	}
	return resp.Tickets, nil
}

// GetTicketPDF retrieves ticket PDF
func (c *FulfillmentClient) GetTicketPDF(ctx context.Context, ticketID string) (*fulfillmentv1.TicketPDFResponse, error) {
	return c.client.GetTicketPDF(ctx, &fulfillmentv1.GetTicketPDFRequest{TicketId: ticketID})
}
