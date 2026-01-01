# Nationwide Ticketing System - OpenAPI Specs (Expanded)

This file aggregates canonical OpenAPI specs for public-facing endpoints.
Each service owns its spec; this document keeps them synchronized for review.

## 1) Identity Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Identity Service
  version: 1.0.0
paths:
  /v1/auth/register:
    post:
      summary: Register a new user
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/RegisterRequest"}
      responses:
        "201":
          description: Registered
          content:
            application/json:
              schema: {$ref: "#/components/schemas/AuthResponse"}
  /v1/auth/login:
    post:
      summary: Authenticate user
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/LoginRequest"}
      responses:
        "200":
          description: Authenticated
          content:
            application/json:
              schema: {$ref: "#/components/schemas/AuthResponse"}
  /v1/auth/refresh:
    post:
      summary: Refresh access token
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/RefreshRequest"}
      responses:
        "200":
          description: Refreshed
          content:
            application/json:
              schema: {$ref: "#/components/schemas/AuthResponse"}
  /v1/auth/logout:
    post:
      summary: Logout and revoke tokens
      responses:
        "204":
          description: Logged out
  /v1/auth/password/reset:
    post:
      summary: Start password reset
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PasswordResetRequest"}
      responses:
        "202":
          description: Reset initiated
  /v1/auth/mfa/enable:
    post:
      summary: Enable MFA
      responses:
        "200":
          description: MFA enabled
  /v1/users/{id}:
    get:
      summary: Get user profile
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/User"}
    patch:
      summary: Update user profile
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/UserUpdate"}
      responses:
        "200":
          description: Updated
          content:
            application/json:
              schema: {$ref: "#/components/schemas/User"}
components:
  schemas:
    RegisterRequest:
      type: object
      required: [email, password]
      properties:
        email: {type: string, format: email}
        password: {type: string, minLength: 8}
        phone: {type: string}
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email: {type: string, format: email}
        password: {type: string}
    RefreshRequest:
      type: object
      required: [refresh_token]
      properties:
        refresh_token: {type: string}
    PasswordResetRequest:
      type: object
      required: [email]
      properties:
        email: {type: string, format: email}
    UserUpdate:
      type: object
      properties:
        name: {type: string}
        phone: {type: string}
    User:
      type: object
      properties:
        id: {type: string}
        email: {type: string}
        name: {type: string}
        phone: {type: string}
    AuthResponse:
      type: object
      properties:
        access_token: {type: string}
        refresh_token: {type: string}
        expires_in: {type: integer}
```

## 2) Catalog Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Catalog Service
  version: 1.1.0  # Bumped for Transport
paths:
  /v1/routes:
    post:
      summary: Create route (Transport)
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/RouteCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Route"}
  /v1/trips:
    post:
      summary: Create trip instance
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/TripCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Trip"}
    get:
      summary: Search trips
      parameters:
        - in: query
          name: from_station
          schema: {type: string}
        - in: query
          name: to_station
          schema: {type: string}
        - in: query
          name: date
          schema: {type: string, format: date}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {$ref: "#/components/schemas/Trip"}
  /v1/events:
    post:
      summary: Create an event
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/EventCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Event"}
    get:
      summary: List events
      parameters:
        - in: query
          name: date
          schema: {type: string, format: date}
        - in: query
          name: geo
          schema: {type: string}
        - in: query
          name: performer_id
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {$ref: "#/components/schemas/Event"}
  /v1/events/{id}:
    get:
      summary: Get event by id
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Event"}
    patch:
      summary: Update event
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/EventUpdate"}
      responses:
        "200":
          description: Updated
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Event"}
  /v1/venues:
    post:
      summary: Create venue or station
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/VenueCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Venue"}
    get:
      summary: List venues
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {$ref: "#/components/schemas/Venue"}
  /v1/venues/{id}:
    get:
      summary: Get venue
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Venue"}
  /v1/performers:
    post:
      summary: Create performer/operator
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PerformerCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Performer"}
    get:
      summary: List performers
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {$ref: "#/components/schemas/Performer"}
components:
  schemas:
    RouteCreate:
      type: object
      required: [name, origin_station_id, destination_station_id, stops]
      properties:
        name: {type: string}
        origin_station_id: {type: string}
        destination_station_id: {type: string}
        stops:
          type: array
          items:
            type: object
            properties:
              station_id: {type: string}
              sequence_order: {type: integer}
              km_from_origin: {type: number}
    Route:
      allOf:
        - $ref: "#/components/schemas/RouteCreate"
        - type: object
          properties:
            id: {type: string}
    TripCreate:
      type: object
      required: [route_id, departure_time, vehicle_id]
      properties:
        route_id: {type: string}
        vehicle_id: {type: string}
        departure_time: {type: string, format: date-time}
        schedule:
          type: array
          items:
            type: object
            properties:
              station_id: {type: string}
              arrival_time: {type: string, format: date-time}
              departure_time: {type: string, format: date-time}
    Trip:
      allOf:
        - $ref: "#/components/schemas/TripCreate"
        - type: object
          properties:
            id: {type: string}
    EventCreate:
      type: object
      required: [name, venue_id, start_time]
      properties:
        name: {type: string}
        venue_id: {type: string}
        start_time: {type: string, format: date-time}
        metadata: {type: object, additionalProperties: true}
    EventUpdate:
      type: object
      properties:
        name: {type: string}
        start_time: {type: string, format: date-time}
        metadata: {type: object, additionalProperties: true}
    Event:
      allOf:
        - $ref: "#/components/schemas/EventCreate"
        - type: object
          properties:
            id: {type: string}
    VenueCreate:
      type: object
      required: [name, address]
      properties:
        name: {type: string}
        address: {type: string}
        geo: {type: string}
        type: {type: string, enum: [stadium, theater, station, terminal]}
    Venue:
      allOf:
        - $ref: "#/components/schemas/VenueCreate"
        - type: object
          properties:
            id: {type: string}
    PerformerCreate:
      type: object
      required: [name]
      properties:
        name: {type: string}
    Performer:
      allOf:
        - $ref: "#/components/schemas/PerformerCreate"
        - type: object
          properties:
            id: {type: string}
```

## 3) Inventory Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Inventory Service
  version: 1.0.0
paths:
  /v1/inventory/{event_id}:
    get:
      summary: Get inventory for event
      parameters:
        - in: path
          name: event_id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/InventorySnapshot"}
  /v1/holds:
    post:
      summary: Create a hold
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/HoldCreateRequest"}
      responses:
        "201":
          description: Hold created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Hold"}
  /v1/holds/{hold_id}:
    delete:
      summary: Release a hold
      parameters:
        - in: path
          name: hold_id
          required: true
          schema: {type: string}
      responses:
        "204":
          description: Released
components:
  schemas:
    InventorySnapshot:
      type: object
      properties:
        event_id: {type: string}
        available: {type: integer}
        held: {type: integer}
        sold: {type: integer}
    HoldCreateRequest:
      type: object
      required: [event_id, seat_ids, ttl_seconds]
      properties:
        event_id: {type: string}
        seat_ids:
          type: array
          items: {type: string}
        ttl_seconds: {type: integer}
    Hold:
      type: object
      properties:
        hold_id: {type: string}
        expires_at: {type: string, format: date-time}
```

## 4) Order Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Order Service
  version: 1.0.0
paths:
  /v1/orders/draft:
    post:
      summary: Create order draft
      parameters:
        - in: header
          name: Idempotency-Key
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/OrderDraftRequest"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Order"}
  /v1/orders/confirm:
    post:
      summary: Confirm order after payment
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/OrderConfirmRequest"}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Order"}
  /v1/orders/{id}:
    get:
      summary: Get order
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Order"}
  /v1/orders/{id}/cancel:
    post:
      summary: Cancel order
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: Canceled
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Order"}
components:
  schemas:
    OrderDraftRequest:
      type: object
      required: [event_id, seat_ids]
      properties:
        event_id: {type: string}
        seat_ids:
          type: array
          items: {type: string}
        promo_code: {type: string}
    OrderConfirmRequest:
      type: object
      required: [order_id, payment_id]
      properties:
        order_id: {type: string}
        payment_id: {type: string}
    Order:
      type: object
      properties:
        id: {type: string}
        status: {type: string}
        total_amount: {type: number}
        currency: {type: string}
```

## 5) Payment Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Payment Service
  version: 1.0.0
paths:
  /v1/payments/authorize:
    post:
      summary: Authorize payment
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PaymentAuthRequest"}
      responses:
        "200":
          description: Authorized
          content:
            application/json:
              schema: {$ref: "#/components/schemas/PaymentAuthResponse"}
  /v1/payments/capture:
    post:
      summary: Capture payment
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PaymentCaptureRequest"}
      responses:
        "200":
          description: Captured
          content:
            application/json:
              schema: {$ref: "#/components/schemas/PaymentCaptureResponse"}
  /v1/payments/refund:
    post:
      summary: Refund payment
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PaymentRefundRequest"}
      responses:
        "200":
          description: Refunded
          content:
            application/json:
              schema: {$ref: "#/components/schemas/PaymentRefundResponse"}
components:
  schemas:
    PaymentAuthRequest:
      type: object
      required: [order_id, amount, currency, payment_token]
      properties:
        order_id: {type: string}
        amount: {type: number}
        currency: {type: string}
        payment_token: {type: string}
    PaymentAuthResponse:
      type: object
      properties:
        payment_id: {type: string}
        status: {type: string}
        authorized_at: {type: string, format: date-time}
    PaymentCaptureRequest:
      type: object
      required: [payment_id]
      properties:
        payment_id: {type: string}
    PaymentCaptureResponse:
      type: object
      properties:
        payment_id: {type: string}
        status: {type: string}
    PaymentRefundRequest:
      type: object
      required: [payment_id, amount]
      properties:
        payment_id: {type: string}
        amount: {type: number}
    PaymentRefundResponse:
      type: object
      properties:
        refund_id: {type: string}
        status: {type: string}
```

## 6) Notification Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Notification Service
  version: 1.0.0
paths:
  /v1/notifications/email:
    post:
      summary: Send email
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/EmailRequest"}
      responses:
        "202":
          description: Accepted
  /v1/notifications/sms:
    post:
      summary: Send SMS
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/SmsRequest"}
      responses:
        "202":
          description: Accepted
components:
  schemas:
    EmailRequest:
      type: object
      required: [to, template_id, variables]
      properties:
        to: {type: string}
        template_id: {type: string}
        variables: {type: object, additionalProperties: true}
    SmsRequest:
      type: object
      required: [to, message]
      properties:
        to: {type: string}
        message: {type: string}
```

## 7) Fulfillment Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Fulfillment Service
  version: 1.0.0
paths:
  /v1/tickets/issue:
    post:
      summary: Issue tickets after confirmation
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/TicketIssueRequest"}
      responses:
        "201":
          description: Issued
          content:
            application/json:
              schema: {$ref: "#/components/schemas/TicketIssueResponse"}
  /v1/tickets/{id}:
    get:
      summary: Get ticket
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Ticket"}
  /v1/tickets/transfer:
    post:
      summary: Transfer ticket to another user
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/TicketTransferRequest"}
      responses:
        "200":
          description: Transferred
components:
  schemas:
    TicketIssueRequest:
      type: object
      required: [order_id]
      properties:
        order_id: {type: string}
    TicketIssueResponse:
      type: object
      properties:
        tickets:
          type: array
          items: {$ref: "#/components/schemas/Ticket"}
    Ticket:
      type: object
      properties:
        id: {type: string}
        barcode: {type: string}
        status: {type: string}
    TicketTransferRequest:
      type: object
      required: [ticket_id, to_user_id]
      properties:
        ticket_id: {type: string}
        to_user_id: {type: string}
```

## 8) Pricing and Promotions Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Pricing Service
  version: 1.0.0
paths:
  /v1/pricing/quote:
    post:
      summary: Quote final price
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PriceQuoteRequest"}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/PriceQuoteResponse"}
  /v1/promotions:
    post:
      summary: Create promotion
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PromotionCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Promotion"}
  /v1/promotions/{code}:
    get:
      summary: Get promotion by code
      parameters:
        - in: path
          name: code
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Promotion"}
components:
  schemas:
    PriceQuoteRequest:
      type: object
      required: [event_id, seat_ids]
      properties:
        event_id: {type: string}
        seat_ids:
          type: array
          items: {type: string}
        promo_code: {type: string}
    PriceQuoteResponse:
      type: object
      properties:
        total_amount: {type: number}
        currency: {type: string}
        breakdown: {type: object, additionalProperties: true}
    PromotionCreate:
      type: object
      required: [code, discount_type, value]
      properties:
        code: {type: string}
        discount_type: {type: string}
        value: {type: number}
    Promotion:
      allOf:
        - $ref: "#/components/schemas/PromotionCreate"
        - type: object
          properties:
            id: {type: string}
```

## 9) Search Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Search Service
  version: 1.0.0
paths:
  /v1/search/events:
    get:
      summary: Search events
      parameters:
        - in: query
          name: q
          schema: {type: string}
        - in: query
          name: filters
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/SearchResults"}
components:
  schemas:
    SearchResults:
      type: object
      properties:
        total: {type: integer}
        results:
          type: array
          items: {type: object}
```

## 10) Reporting Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Reporting Service
  version: 1.0.0
paths:
  /v1/reports/sales:
    get:
      summary: Sales report
      parameters:
        - in: query
          name: event_id
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/SalesReport"}
  /v1/reports/finance:
    get:
      summary: Finance report
      parameters:
        - in: query
          name: range
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/FinanceReport"}
components:
  schemas:
    SalesReport:
      type: object
      properties:
        event_id: {type: string}
        total_sold: {type: integer}
        revenue: {type: number}
    FinanceReport:
      type: object
      properties:
        range: {type: string}
        gross: {type: number}
        net: {type: number}
```

## 11) Fraud Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Fraud Service
  version: 1.0.0
paths:
  /v1/risk/score:
    post:
      summary: Score transaction risk
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/RiskScoreRequest"}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/RiskScoreResponse"}
components:
  schemas:
    RiskScoreRequest:
      type: object
      required: [order_id, user_id]
      properties:
        order_id: {type: string}
        user_id: {type: string}
        ip: {type: string}
        device_id: {type: string}
    RiskScoreResponse:
      type: object
      properties:
        risk_score: {type: number}
        decision: {type: string}
```

## 12) Queue Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Queue Service
  version: 1.0.0
paths:
  /v1/queue/enter:
    post:
      summary: Enter virtual queue
      responses:
        "200":
          description: Token issued
          content:
            application/json:
              schema: {$ref: "#/components/schemas/QueueToken"}
  /v1/queue/status/{token}:
    get:
      summary: Get queue status
      parameters:
        - in: path
          name: token
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/QueueStatus"}
components:
  schemas:
    QueueToken:
      type: object
      properties:
        token: {type: string}
        position: {type: integer}
        estimated_wait_seconds: {type: integer}
    QueueStatus:
      type: object
      properties:
        position: {type: integer}
        allowed: {type: boolean}
```

## 13) Seat Map Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Seat Map Service
  version: 1.0.0
paths:
  /v1/seatmaps/{venue_id}:
    get:
      summary: Get seat map for venue
      parameters:
        - in: path
          name: venue_id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/SeatMap"}
components:
  schemas:
    SeatMap:
      type: object
      properties:
        venue_id: {type: string}
        sections:
          type: array
          items: {type: object}
```

## 14) Vendor/Partner Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Vendor Service
  version: 1.0.0
paths:
  /v1/vendors:
    post:
      summary: Create vendor
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/VendorCreate"}
      responses:
        "201":
          description: Created
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Vendor"}
    get:
      summary: List vendors
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {$ref: "#/components/schemas/Vendor"}
  /v1/vendors/register:
    post:
      summary: Register vendor with onboarding details
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/VendorRegistration"}
      responses:
        "201":
          description: Registered
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Vendor"}
  /v1/vendors/{id}:
    get:
      summary: Get vendor
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Vendor"}
    patch:
      summary: Update vendor profile
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/VendorUpdate"}
      responses:
        "200":
          description: Updated
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Vendor"}
  /v1/vendors/{id}/kyc:
    post:
      summary: Submit KYC data
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/KycSubmit"}
      responses:
        "202":
          description: Accepted
  /v1/vendors/{id}/contracts:
    post:
      summary: Accept contract terms
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/ContractAccept"}
      responses:
        "200":
          description: Accepted
  /v1/vendors/{id}/payouts:
    post:
      summary: Configure payout account
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PayoutSetup"}
      responses:
        "202":
          description: Accepted
  /v1/vendors/{id}/status:
    get:
      summary: Get vendor onboarding status
      parameters:
        - in: path
          name: id
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/VendorStatus"}
components:
  schemas:
    VendorCreate:
      type: object
      required: [name]
      properties:
        name: {type: string}
        contact_email: {type: string}
        contact_phone: {type: string}
        legal_entity: {type: string}
    VendorRegistration:
      type: object
      required: [name, contact_email, legal_entity]
      properties:
        name: {type: string}
        contact_email: {type: string}
        contact_phone: {type: string}
        legal_entity: {type: string}
        address: {type: string}
        tax_id: {type: string}
    VendorUpdate:
      type: object
      properties:
        contact_email: {type: string}
        contact_phone: {type: string}
        address: {type: string}
    Vendor:
      allOf:
        - $ref: "#/components/schemas/VendorCreate"
        - type: object
          properties:
            id: {type: string}
            status: {type: string}
    KycSubmit:
      type: object
      required: [provider, document_refs]
      properties:
        provider: {type: string}
        document_refs:
          type: array
          items: {type: string}
    ContractAccept:
      type: object
      required: [version, accepted_at]
      properties:
        version: {type: string}
        accepted_at: {type: string, format: date-time}
        revenue_share: {type: number}
    PayoutSetup:
      type: object
      required: [provider, payout_account_ref]
      properties:
        provider: {type: string}
        payout_account_ref: {type: string}
    VendorStatus:
      type: object
      properties:
        vendor_id: {type: string}
        status: {type: string}
        kyc_status: {type: string}
        payout_status: {type: string}
        contract_status: {type: string}
```

## 15) Audit Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Audit Service
  version: 1.0.0
paths:
  /v1/audit/events:
    get:
      summary: Query audit events
      parameters:
        - in: query
          name: entity_id
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items: {type: object}
```

## 16) Feature Flag Service (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Feature Flag Service
  version: 1.0.0
paths:
  /v1/flags/{key}:
    get:
      summary: Get flag by key
      parameters:
        - in: path
          name: key
          required: true
          schema: {type: string}
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema: {$ref: "#/components/schemas/Flag"}
components:
  schemas:
    Flag:
      type: object
      properties:
        key: {type: string}
        enabled: {type: boolean}
```

## 17) Vendor Webhooks (OpenAPI 3.0)

```yaml
openapi: 3.0.3
info:
  title: Vendor Webhooks
  version: 1.0.0
paths:
  /v1/vendors/kyc/webhook:
    post:
      summary: KYC provider status update
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/KycWebhook"}
      responses:
        "200":
          description: OK
  /v1/vendors/payouts/webhook:
    post:
      summary: Payout provider status update
      requestBody:
        required: true
        content:
          application/json:
            schema: {$ref: "#/components/schemas/PayoutWebhook"}
      responses:
        "200":
          description: OK
components:
  schemas:
    KycWebhook:
      type: object
      required: [vendor_id, provider_ref, status]
      properties:
        vendor_id: {type: string}
        provider_ref: {type: string}
        status: {type: string}
        reason: {type: string}
        reviewed_at: {type: string, format: date-time}
    PayoutWebhook:
      type: object
      required: [vendor_id, payout_account_ref, status]
      properties:
        vendor_id: {type: string}
        payout_account_ref: {type: string}
        status: {type: string}
        verified_at: {type: string, format: date-time}
```
