# Nationwide Ticketing System - Protobuf Specs

This file defines concrete protobuf messages and gRPC services for core flows.
Kafka events should reuse these message definitions where possible.

## 1) Common Types

```proto
syntax = "proto3";

package ticketing.common.v1;

message Money {
  int64 amount_micros = 1;
  string currency = 2;
}

message Timestamp {
  int64 unix_ms = 1;
}
```

## 2) Inventory Service

```proto
syntax = "proto3";

package ticketing.inventory.v1;

service InventoryService {
  rpc ReserveTickets(ReserveTicketsRequest) returns (ReserveTicketsResponse);
  rpc ReleaseHold(ReleaseHoldRequest) returns (ReleaseHoldResponse);
  rpc ConfirmAllocation(ConfirmAllocationRequest) returns (ConfirmAllocationResponse);
}

message ReserveTicketsRequest {
  string event_id = 1;
  repeated string seat_ids = 2;
  int64 hold_ttl_seconds = 3;
  string user_id = 4;
}

message ReserveTicketsResponse {
  string hold_id = 1;
  int64 expires_at_unix = 2;
}

message ReleaseHoldRequest {
  string hold_id = 1;
}

message ReleaseHoldResponse {
  bool released = 1;
}

message ConfirmAllocationRequest {
  string order_id = 1;
  string hold_id = 2;
}

message ConfirmAllocationResponse {
  bool confirmed = 1;
}
```

## 3) Order Service

```proto
syntax = "proto3";

package ticketing.order.v1;

service OrderService {
  rpc CreateDraft(CreateDraftRequest) returns (Order);
  rpc ConfirmOrder(ConfirmOrderRequest) returns (Order);
}

message CreateDraftRequest {
  string user_id = 1;
  string event_id = 2;
  repeated string seat_ids = 3;
  string promo_code = 4;
  string idempotency_key = 5;
}

message ConfirmOrderRequest {
  string order_id = 1;
  string payment_id = 2;
}

message Order {
  string id = 1;
  string status = 2;
  int64 created_at_unix = 3;
  int64 updated_at_unix = 4;
  ticketing.common.v1.Money total = 5;
}
```

## 4) Payment Service

```proto
syntax = "proto3";

package ticketing.payment.v1;

service PaymentService {
  rpc Authorize(AuthorizeRequest) returns (AuthorizeResponse);
  rpc Capture(CaptureRequest) returns (CaptureResponse);
  rpc Refund(RefundRequest) returns (RefundResponse);
}

message AuthorizeRequest {
  string order_id = 1;
  ticketing.common.v1.Money amount = 2;
  string payment_token = 3;
}

message AuthorizeResponse {
  string payment_id = 1;
  string status = 2;
  int64 authorized_at_unix = 3;
}

message CaptureRequest {
  string payment_id = 1;
}

message CaptureResponse {
  string payment_id = 1;
  string status = 2;
}

message RefundRequest {
  string payment_id = 1;
  ticketing.common.v1.Money amount = 2;
}

message RefundResponse {
  string refund_id = 1;
  string status = 2;
}
```

## 5) Fraud Service

```proto
syntax = "proto3";

package ticketing.fraud.v1;

service FraudService {
  rpc ScoreTransaction(ScoreTransactionRequest) returns (ScoreTransactionResponse);
}

message ScoreTransactionRequest {
  string order_id = 1;
  string user_id = 2;
  string ip = 3;
  string device_id = 4;
}

message ScoreTransactionResponse {
  double risk_score = 1;
  string decision = 2;
}
```

## 6) Kafka Event Samples

```proto
syntax = "proto3";

package ticketing.events.v1;

message OrderConfirmed {
  string order_id = 1;
  string user_id = 2;
  ticketing.common.v1.Money total = 3;
  int64 confirmed_at_unix = 4;
  string trace_id = 5;
}

message InventoryHoldExpired {
  string hold_id = 1;
  string event_id = 2;
  int64 expired_at_unix = 3;
  string trace_id = 4;
}
```

