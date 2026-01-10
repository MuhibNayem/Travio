# Operator Service gRPC Documentation

**Package:** `operator.v1`  
**Internal DNS:** `operator:50059`  
**Proto File:** `server/api/proto/operator/v1/operator.proto`

## Overview
The Operator Service (also known as Vendor Service) manages the profiles and settings for transport providers (e.g., Bus Companies) on the platform.

## Key Behaviors

### Commission Model
- **Rate Management:** Each vendor has a configured `commission_rate` (percentage) which is used by the Billing system to calculate platform fees.

### Lifecycle
- **Status:** Vendors can be `active` or `suspended`. Suspended vendors cannot schedule new trips.

---

## RPC Methods

### `CreateVendor`
Onboards a new transport operator.

- **Request:** `CreateVendorRequest` (Name, Contact, Address).
- **Response:** `Vendor` (with generated ID).

### `UpdateVendor`
Modifies vendor profile or commission rates.

- **Request:** `UpdateVendorRequest`.

### `ListVendors`
Paginated list of all onboarded operators.

---

## Message Definitions

### Vendor
| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Display name (e.g. "Hanif Enterprise") |
| `commission_rate` | `double` | e.g. `0.05` for 5% |
| `status` | `string` | `active`, `suspended` |
