# Nationwide Ticketing System - Software Requirements Specification

## 1) Introduction

This SRS defines functional and non-functional requirements for a
nationwide ticketing system with microservices and high scalability.

## 2) Users and Roles

- Buyer: purchases tickets for events.
- Organizer: creates events and manages inventory.
- Vendor/Partner: venue or promoter integrations.
- Admin: system operations and compliance.

## 3) Functional Requirements

### 3.1 Identity and Access

- FR-IA-01: Support email/phone registration with MFA.
- FR-IA-02: Support OAuth for major providers.
- FR-IA-03: Role-based authorization for all endpoints.

### 3.2 Catalog and Search

- FR-CS-01: Create and manage events, venues, performers.
- FR-CS-02: Index events with metadata for search.
- FR-CS-03: Support geo and time-based filtering.

### 3.3 Inventory and Holds

- FR-IH-01: Reserve inventory with TTL holds.
- FR-IH-02: Prevent oversell across regions.
- FR-IH-03: Support seat-based and GA inventory.

### 3.4 Pricing and Promotions

- FR-PP-01: Support base price and price tiers.
- FR-PP-02: Apply promo codes and bundles.
- FR-PP-03: Enable dynamic pricing by demand.

### 3.5 Orders and Checkout

- FR-OC-01: Maintain cart state with idempotency keys.
- FR-OC-02: Support cancel and refund flows.
- FR-OC-03: Maintain audit trails for all changes.

### 3.6 Payments

- FR-PAY-01: Support multiple gateways and payment methods.
- FR-PAY-02: Tokenize payment data; no raw PAN storage.
- FR-PAY-03: Support 3DS/SCA flows.

### 3.7 Fulfillment

- FR-FUL-01: Generate secure tickets (QR/barcode).
- FR-FUL-02: Allow reissue with audit trail.
- FR-FUL-03: Support ticket transfer and resale workflows.

### 3.8 Fraud and Risk

- FR-FRD-01: Velocity limits by user/device/IP.
- FR-FRD-02: Risk scoring with rules and ML hooks.
- FR-FRD-03: Manual review and blocklist support.

### 3.9 Notifications and Webhooks

- FR-NOT-01: Send confirmations and reminders.
- FR-NOT-02: Partner webhooks for order events.

### 3.10 Reporting and Analytics

- FR-REP-01: Sales dashboards by event and vendor.
- FR-REP-02: Export financial summaries.
- FR-REP-03: Audit reports for compliance.

### 3.11 Vendor Registration and Onboarding

- FR-VEN-01: Vendors can register with organization profile and contacts.
- FR-VEN-02: Collect KYC/AML artifacts and verification status.
- FR-VEN-03: Configure payout accounts with verification checks.
- FR-VEN-04: Accept contracts and revenue share terms.
- FR-VEN-05: Vendor status lifecycle (pending, verified, active, suspended).

## 4) Non-Functional Requirements

- NFR-PERF-01: p95 latency < 200ms for read paths.
- NFR-PERF-02: p95 latency < 400ms for checkout.
- NFR-AVAIL-01: 99.95%+ availability for purchase flow.
- NFR-SEC-01: PCI-DSS, GDPR compliance.
- NFR-OBS-01: End-to-end tracing for purchase flows.
- NFR-SCALE-01: 10x burst scaling for peak events.

## 5) Data and Consistency

- DCR-01: Inventory holds must be globally consistent per event.
- DCR-02: Orders must be idempotent and recoverable.
- DCR-03: Payment authorization is required before fulfillment.

## 6) Constraints

- Backend services must be implemented in Go.
- Polyglot persistence required for scalability.
- Kafka must be used for event-driven coordination.
