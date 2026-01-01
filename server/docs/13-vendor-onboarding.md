# Vendor Onboarding - Production Grade Design

This document defines KYC/AML provider integrations, vendor SLAs,
onboarding state machine, lifecycle events, and compliance requirements.

## 1) KYC/AML Provider Integration

### 1.1 Provider Requirements

- Support document verification and business verification.
- Provide webhook callbacks for status updates.
- Offer audit logs and proof bundles.
- Support regional compliance (OFAC, PEP, sanctions).

### 1.2 Integration Pattern

- Vendor Service creates KYC session with provider.
- Provider returns `provider_ref` and submission URL.
- Vendor submits documents directly to provider.
- Provider calls webhook to update status.

### 1.3 Webhooks

Endpoint:
- POST `/v1/vendors/kyc/webhook`

Payload:
- vendor_id
- provider_ref
- status (pending, approved, rejected)
- reason codes (if rejected)
- reviewed_at timestamp

Security:
- HMAC signature verification.
- IP allowlist or mTLS.

## 2) Payout Provider Integration

### 2.1 Provider Requirements

- Payout account tokenization.
- Bank account verification callbacks.
- Support multi-currency payouts.

### 2.2 Integration Pattern

- Vendor submits payout details.
- Provider returns payout account reference.
- Provider verifies and sends webhook update.

Webhook:
- POST `/v1/vendors/payouts/webhook`

## 3) Vendor SLAs

- Onboarding verification time: 24-72 hours.
- Payout setup verification: 1-3 business days.
- Support response: 24 hours (business), 4 hours (critical incidents).
- Data accuracy: 99.9% for payout calculations.

## 4) Lifecycle State Machine

States:
- PENDING: Created but not submitted.
- SUBMITTED: KYC and documents submitted.
- VERIFIED: KYC approved.
- ACTIVE: Contracts signed + payout verified.
- SUSPENDED: Compliance or operational hold.
- REJECTED: KYC failed or contract declined.

Transitions:
- PENDING -> SUBMITTED (KYC started)
- SUBMITTED -> VERIFIED (KYC approved)
- VERIFIED -> ACTIVE (contract + payout verified)
- ACTIVE -> SUSPENDED (risk/compliance)
- SUBMITTED -> REJECTED (KYC failed)
- SUSPENDED -> ACTIVE (reinstated)

## 5) Lifecycle Events (Kafka)

Topics:
- `vendor.created`
- `vendor.kyc.submitted`
- `vendor.kyc.verified`
- `vendor.kyc.rejected`
- `vendor.contract.accepted`
- `vendor.payout.verified`
- `vendor.activated`
- `vendor.suspended`

Event rules:
- Include vendor_id, status, timestamp, and trace_id.
- Versioned using schema registry policies.

## 6) Compliance Requirements

- Maintain proof bundles for KYC verification.
- Store audit records for all vendor status changes.
- Retain vendor contracts for 7 years.
- Enforce least privilege for vendor data access.

