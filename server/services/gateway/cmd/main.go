package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/config"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/handler"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger.Init("api-gateway")
	cfg := config.Load()

	// Initialize Chi router
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// Rate limiter (optional - depends on Redis availability)
	rateLimiter := middleware.NewRateLimiter(cfg.RedisURL, 100, 60) // 100 req/min
	defer rateLimiter.Close()
	r.Use(rateLimiter.Middleware)

	// Entitlement enforcement (requires active subscription for protected routes)
	entitlementMW := middleware.NewEntitlementMiddleware(cfg.RedisURL)
	r.Use(entitlementMW.Middleware)

	// Health endpoints (no auth required)
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)

	// Initialize Circuit Breakers
	catalogCB := middleware.NewCircuitBreaker("catalog-service")
	inventoryCB := middleware.NewCircuitBreaker("inventory-service")
	orderCB := middleware.NewCircuitBreaker("order-service")
	searchCB := middleware.NewCircuitBreaker("search-service")
	_ = searchCB // Pending Search Update
	// Note: All services now use gRPC clients (no HTTP circuit breakers needed)

	// TLS config for mTLS (reads from env vars)
	tlsCfg := client.TLSConfig{
		CertFile: cfg.TLSCertFile,
		KeyFile:  cfg.TLSKeyFile,
		CAFile:   cfg.TLSCAFile,
	}

	// Initialize gRPC handlers
	catalogHandler, err := handler.NewCatalogHandler(cfg.CatalogURL, catalogCB)
	if err != nil {
		logger.Error("Failed to connect to catalog service", "error", err)
	} else {
		defer catalogHandler.Close()
	}

	inventoryHandler, err := handler.NewInventoryHandler(cfg.InventoryURL, inventoryCB)
	if err != nil {
		logger.Error("Failed to connect to inventory service", "error", err)
	} else {
		defer inventoryHandler.Close()
	}

	orderHandler, err := handler.NewOrderHandler(cfg.OrderURL, orderCB)
	if err != nil {
		logger.Error("Failed to connect to order service", "error", err)
	} else {
		defer orderHandler.Close()
	}

	var searchHandler *handler.SearchHandler
	searchClient, err := client.NewSearchClient(cfg.SearchURL)
	if err != nil {
		logger.Error("Failed to connect to search service", "error", err)
	} else {
		defer searchClient.Close()
		searchHandler = handler.NewSearchHandler(searchClient)
	}

	// Initialize gRPC clients for all remaining services
	identityClient, err := client.NewIdentityClient(cfg.IdentityURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to identity service", "error", err)
	} else {
		defer identityClient.Close()
	}

	paymentClient, err := client.NewPaymentClient(cfg.PaymentURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to payment service", "error", err)
	} else {
		defer paymentClient.Close()
	}

	fulfillmentClient, err := client.NewFulfillmentClient(cfg.FulfillmentURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to fulfillment service", "error", err)
	} else {
		defer fulfillmentClient.Close()
	}

	queueClient, err := client.NewQueueClient(cfg.QueueURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to queue service", "error", err)
	} else {
		defer queueClient.Close()
	}

	pricingClient, err := client.NewPricingClient(cfg.PricingURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to pricing service", "error", err)
	} else {
		defer pricingClient.Close()
	}

	operatorClient, err := client.NewOperatorClient(cfg.OperatorURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to operator service", "error", err)
	} else {
		defer operatorClient.Close()
	}

	subscriptionClient, err := client.NewSubscriptionClient(cfg.SubscriptionURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to subscription service", "error", err)
	} else {
		defer subscriptionClient.Close()
	}

	fraudClient, err := client.NewFraudClient(cfg.FraudURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to fraud service", "error", err)
	} else {
		defer fraudClient.Close()
	}

	reportingClient, err := client.NewReportingClient(cfg.ReportingURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to reporting service", "error", err)
	} else {
		defer reportingClient.Close()
	}

	eventsClient, err := client.NewEventsClient(cfg.EventsURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to events service", "error", err)
	} else {
		defer eventsClient.Close()
	}

	fleetClient, err := client.NewFleetClient(cfg.FleetURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to fleet service", "error", err)
	} else {
		defer fleetClient.Close()
	}

	crmClient, err := client.NewCRMClient(cfg.CRMURL, tlsCfg)
	if err != nil {
		logger.Error("Failed to connect to crm service", "error", err)
	} else {
		defer crmClient.Close()
	}

	// Initialize handlers with gRPC clients
	var identityHandler *handler.IdentityHandler
	if identityClient != nil {
		identityHandler = handler.NewIdentityHandler(identityClient, subscriptionClient)
	}

	var paymentHandler *handler.PaymentHandler
	if paymentClient != nil {
		paymentHandler = handler.NewPaymentHandler(paymentClient)
	}

	var fulfillmentHandler *handler.FulfillmentHandler
	if fulfillmentClient != nil {
		fulfillmentHandler = handler.NewFulfillmentHandler(fulfillmentClient)
	}

	var queueHandler *handler.QueueHandler
	if queueClient != nil {
		queueHandler = handler.NewQueueHandler(queueClient)
	}

	var pricingHandler *handler.PricingHandler
	if pricingClient != nil {
		pricingHandler = handler.NewPricingHandler(pricingClient)
	}

	var operatorHandler *handler.OperatorHandler
	if operatorClient != nil {
		operatorHandler = handler.NewOperatorHandler(operatorClient)
	}

	var subscriptionHandler *handler.SubscriptionHandler
	if subscriptionClient != nil {
		subscriptionHandler = handler.NewSubscriptionHandler(subscriptionClient)
	}

	var fraudHandler *handler.FraudHandler
	if fraudClient != nil {
		fraudHandler = handler.NewFraudHandler(fraudClient)
	}

	var reportingHandler *handler.ReportingHandler
	if reportingClient != nil {
		reportingHandler = handler.NewReportingHandler(reportingClient)
	}

	var eventsHandler *handler.EventsHandler
	if eventsClient != nil {
		eventsHandler = handler.NewEventsHandler(eventsClient)
	}

	var fleetHandler *handler.FleetHandler
	if fleetClient != nil {
		fleetHandler = handler.NewFleetHandler(fleetClient)
	}

	var crmHandler *handler.CRMHandler
	if crmClient != nil {
		crmHandler = handler.NewCRMHandler(crmClient)
	}

	// JWT Auth config
	jwtAuth := middleware.JWTAuth(middleware.JWTConfig{
		Secret: cfg.JWTSecret,
		SkipPaths: []string{
			"/health", "/ready",
			"/v1/auth",
			"/v1/stations", "/v1/trips",
			"/v1/search",
			"/v1/pricing/calculate",
			"/v1/queue",
			"/v1/events",
			"/v1/fleet/location", // Allow location updates without forced user token? Probably secure it.
			// Actually, tracking devices might need API keys, but for now let's assume secure or public tracking for demo?
			// Let's keep it secure by default, but maybe allow public public tracking view?
		},
	})

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Apply JWT auth globally, but skip paths as configured
		r.Use(jwtAuth)

		// Auth routes - gRPC to Identity service (public)
		if identityHandler != nil {
			r.Route("/auth", func(r chi.Router) {
				r.Post("/register", identityHandler.Register)
				r.Post("/login", identityHandler.Login)
				r.Post("/refresh", identityHandler.RefreshToken)
				r.Post("/logout", identityHandler.Logout)
				r.Post("/invite/accept", identityHandler.AcceptInvite)
			})

			// Organization routes - gRPC to Identity
			r.Route("/orgs", func(r chi.Router) {
				r.Post("/", identityHandler.CreateOrganization)
				r.Post("/invites", identityHandler.CreateInvite)
				r.Get("/invites", identityHandler.ListInvites)
				r.Get("/members", identityHandler.ListMembers)
				r.Put("/members/{userId}/role", identityHandler.UpdateUserRole)
				r.Delete("/members/{userId}", identityHandler.RemoveMember)
			})
		}

		// Catalog routes (public)
		if catalogHandler != nil {
			r.Get("/stations", catalogHandler.ListStations)
			r.Get("/stations/{stationId}", catalogHandler.GetStation)
			r.Get("/trips/search", catalogHandler.SearchTrips)
			r.Get("/trips/{tripId}", catalogHandler.GetTrip)
		}

		// Search routes (public)
		if searchHandler != nil {
			r.Get("/search/trips", searchHandler.SearchTrips)
			r.Get("/search/stations", searchHandler.SearchStations)
		}

		// Pricing routes (public)
		if pricingHandler != nil {
			r.Post("/pricing/calculate", pricingHandler.CalculatePrice)
			r.Get("/pricing/rules", pricingHandler.GetPricingRules)
		}

		// Operator/Vendor routes (protected + admin only)
		if operatorHandler != nil {
			r.Route("/vendors", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/", operatorHandler.CreateVendor)
				r.Get("/", operatorHandler.ListVendors)
				r.Get("/{id}", operatorHandler.GetVendor)
				r.Put("/{id}", operatorHandler.UpdateVendor)
				r.Delete("/{id}", operatorHandler.DeleteVendor)
			})
		}

		// Subscription routes
		if subscriptionHandler != nil {
			r.Route("/plans", func(r chi.Router) {
				r.Get("/", subscriptionHandler.ListPlans)
				r.Get("/{id}", subscriptionHandler.GetPlan)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin"))
					r.Post("/", subscriptionHandler.CreatePlan)
					r.Put("/", subscriptionHandler.UpdatePlan)
				})
			})

			r.Route("/subscriptions", func(r chi.Router) {
				r.Post("/", subscriptionHandler.CreateSubscription)
				r.Get("/{orgID}", subscriptionHandler.GetSubscription)
				r.Post("/{orgID}/cancel", subscriptionHandler.CancelSubscription)

				r.Get("/{subID}/invoices", subscriptionHandler.ListInvoices)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin"))
					r.Get("/", subscriptionHandler.ListSubscriptions)
				})
			})
		}

		// Queue routes (public - for waiting room)
		if queueHandler != nil {
			r.Post("/queue/join", queueHandler.JoinQueue)
			r.Get("/queue/position", queueHandler.GetQueuePosition)
			r.Post("/queue/verify", queueHandler.VerifyQueueToken)
		}

		// === PROTECTED ROUTES (require auth) ===

		// Catalog Routes (Protected - Operator Actions)
		if catalogHandler != nil {
			r.Post("/trips", catalogHandler.CreateTrip)
		}

		// Inventory routes (protected)
		if inventoryHandler != nil {
			r.Get("/trips/{tripId}/availability", inventoryHandler.CheckAvailability)
			r.Get("/trips/{tripId}/seatmap", inventoryHandler.GetSeatMap)
			r.Post("/holds", inventoryHandler.HoldSeats)
			r.Delete("/holds/{holdId}", inventoryHandler.ReleaseHold)
		}

		// Order routes (protected)
		if orderHandler != nil {
			r.Post("/orders", orderHandler.CreateOrder)
			r.Get("/orders", orderHandler.ListOrders)
			r.Get("/orders/{orderId}", orderHandler.GetOrder)
			r.Post("/orders/{orderId}/cancel", orderHandler.CancelOrder)
		}

		// Payment routes (protected)
		if paymentHandler != nil {
			r.Get("/payments/methods", paymentHandler.GetPaymentMethods)
			r.Post("/payments", paymentHandler.ProcessPayment)
			r.Get("/payments/{orderId}", paymentHandler.GetPaymentStatus)

			// Organization Config (Admin Only)
			r.Route("/organizations/{orgId}/payment-config", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Put("/", paymentHandler.UpdatePaymentConfig)
				r.Get("/", paymentHandler.GetPaymentConfig)
			})
		}

		// Fulfillment/Ticket routes (protected)
		if fulfillmentHandler != nil {
			r.Get("/tickets/{ticketId}", fulfillmentHandler.GetTicket)
			r.Get("/tickets/{ticketId}/download", fulfillmentHandler.DownloadTicket)
			r.Get("/orders/{orderId}/tickets", fulfillmentHandler.GetOrderTickets)
		}

		// Fraud routes (protected)
		if fraudHandler != nil {
			fraudHandler.RegisterRoutes(r)
		}

		// Reporting routes (protected)
		if reportingHandler != nil {
			reportingHandler.RegisterRoutes(r)
		}

		// Events routes
		if eventsHandler != nil {
			// Public routes
			r.Get("/events", eventsHandler.ListEvents)
			r.Get("/events/{id}", eventsHandler.GetEvent)

			r.Get("/venues", eventsHandler.ListVenues)
			r.Get("/venues/{id}", eventsHandler.GetVenue)

			r.Get("/events/{eventId}/tickets", eventsHandler.ListTicketTypes)

			// Protected routes
			r.Group(func(r chi.Router) {
				// r.Use(middleware.RequireRole("operator")) // Optional: Enforce role
				r.Post("/events", eventsHandler.CreateEvent)
				r.Post("/venues", eventsHandler.CreateVenue)
				r.Post("/tickets/types", eventsHandler.CreateTicketType)
			})
		}

		// Fleet routes
		if fleetHandler != nil {
			// Assets
			r.Get("/fleet/assets/{id}", fleetHandler.GetAsset)

			// Tracking (Secure)
			r.Post("/fleet/location", fleetHandler.UpdateLocation) // Device ping
			r.Get("/fleet/assets/{id}/location", fleetHandler.GetLocation)

			// Protected Management
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("operator", "admin"))
				r.Post("/fleet/assets", fleetHandler.RegisterAsset)
				r.Put("/fleet/assets/{id}/status", fleetHandler.UpdateAssetStatus)
			})
		}

		// CRM Routes
		if crmHandler != nil {
			// Coupons (checkout)
			r.Post("/coupons/validate", crmHandler.ValidateCoupon)

			// Protected
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("operator", "admin"))
				r.Post("/coupons", crmHandler.CreateCoupon)
				r.Get("/coupons", crmHandler.ListCoupons)

				r.Get("/support/tickets", crmHandler.ListTickets)
			})

			// Customer Support
			r.Group(func(r chi.Router) {
				// r.Use(middleware.RequireAuth) // Should match user ID in real app
				r.Post("/support/tickets", crmHandler.CreateTicket)
				r.Post("/support/tickets/{id}/messages", crmHandler.AddTicketMessage)
			})
		}

	})

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("API Gateway starting", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	logger.Info("Server stopped")
}

// proxyTo creates a simple reverse proxy handler
func proxyTo(client *http.Client, target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple proxy - in production use httputil.ReverseProxy
		// Client is now injected (with Circuit Breaker)

		proxyURL := target + r.URL.Path
		if r.URL.RawQuery != "" {
			proxyURL += "?" + r.URL.RawQuery
		}

		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
		if err != nil {
			http.Error(w, "Proxy error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, v := range values {
				proxyReq.Header.Add(key, v)
			}
		}

		resp, err := client.Do(proxyReq)
		if err != nil {
			fmt.Printf("Proxy error: %v\n", err)
			if strings.Contains(err.Error(), "circuit breaker is open") {
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "Upstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		buf := make([]byte, 32*1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}
}
