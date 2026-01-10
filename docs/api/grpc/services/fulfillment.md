# Fulfillment Service gRPC Documentation

**Package:** `fulfillment.v1`  
**Internal DNS:** `fulfillment:9086`  
**Proto File:** `server/api/proto/fulfillment/v1/fulfillment.proto`

## Overview
The Fulfillment Service is responsible for the digital delivery of travel assets (Tickets, Confirmations). It handles QR code generation, PDF rendering, and secure asset storage.

## Key Behaviors

### Ticket Generation Pipeline
1. **Creation:** Records ticket entities in the database.
2. **QR Generation:** Generates a cryptographic QR code containing a signed payload.
3. **PDF Rendering:** Compiles all tickets for a booking into a single PDF document.
4. **Storage:** Uploads the PDF to Object Storage (MinIO/S3) with a private ACL.
5. **Delivery:** Returns a **Presigned URL** (temporal access) to the client.

### Validation Logic
- **Validity Window:** Tickets are valid until **24 hours after** the scheduled departure time.
- **Boarding:** Scanning a ticket marks it as `USED` (IsBoarded=true). Re-scanning a used ticket returns an error to prevent reuse.

---

## RPC Methods

### `GenerateTickets`
Triggers the generation pipeline for a confirmed booking.

- **Request:** `GenerateTicketsRequest` (Passenger details, Route info).
- **Response:** `GenerateTicketsResponse` (Includes `pdf_url`).

### `ValidateTicket`
Used by field operators (conductors/gate agents) to verify a ticket.

- **Request:** `ValidateTicketRequest` (QR Payload).
- **Response:** `Ticket` (with status `VALID`, `EXPIRED`, `USED`).
- **Side Effect:** Marks ticket as boarded if valid.

### `GetTicketPDF`
Regenerates or retrieves the PDF for a specific ticket.

---

## Message Definitions

### Ticket
| Field | Type | Description |
|-------|------|-------------|
| `qr_code_data` | `bytes` | Raw PNG data of the QR code |
| `status` | `string` | `ACTIVE`, `USED`, `CANCELLED` |
| `valid_until` | `string` | Expiry timestamp |
