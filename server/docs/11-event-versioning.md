# Nationwide Ticketing System - Event Versioning and Schema Evolution

## 1) Versioning Principles

- Events are immutable; corrections are new events.
- Version fields are explicit: `schema_version` and `event_type`.
- Backward compatibility is required for all consumers.
- Deprecation uses scheduled end-of-life dates.

## 2) Event Envelope

All Kafka events must follow a shared envelope:

```
{
  "event_id": "uuid",
  "event_type": "order.confirmed",
  "schema_version": 1,
  "source": "order-service",
  "occurred_at": "2025-01-01T12:00:00Z",
  "trace_id": "trace-id",
  "payload": { ... }
}
```

## 3) Compatibility Rules

- Additive changes only (add optional fields).
- No field renames or removals in place.
- Deprecate fields by leaving them unused for N releases.
- Consumers must ignore unknown fields.

## 4) Schema Registry Policy

- Schema registry enforces compatibility checks per topic.
- Default compatibility: BACKWARD for core topics.
- Breaking changes require a new topic with version suffix.

Example:
- `order.confirmed.v1`
- `order.confirmed.v2`

## 7) Vendor Lifecycle Event Types

- `vendor.created`
- `vendor.kyc.submitted`
- `vendor.kyc.verified`
- `vendor.kyc.rejected`
- `vendor.contract.accepted`
- `vendor.payout.verified`
- `vendor.activated`
- `vendor.suspended`

Versioning:
- Increment `schema_version` on any non-additive change.
- Prefer additive changes and keep events backward compatible.

## 5) Deprecation Workflow

1) Announce deprecation with end-of-life date.
2) Run parallel publish for old and new schema.
3) Track consumer adoption.
4) Decommission old schema after adoption and EOL.

## 6) Testing Requirements

- Contract tests per producer/consumer pair.
- Replay tests with historical data.
- Schema validation in CI for every change.
