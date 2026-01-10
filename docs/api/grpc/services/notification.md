# Notification Service gRPC Documentation

**Package:** `notification.v1`  
**Internal DNS:** `notification:9090`  
**Proto File:** `server/api/proto/notification/v1/notification.proto`

## Overview
The Notification Service is a centralized dispatcher for transactional communications (Email, SMS). It decouples business services from delivery providers (SMTP, Twilio, etc.) and handles templating.

## Key Behaviors

### Templating
- **Engine:** Uses Go's `html/template` engine.
- **Storage:** Templates are embedded in the binary for fast access and include standard layouts for invoices, tickets, and OTPs.

### Providers
- **Email:** SMTP-based delivery (configurable).
- **SMS:** Provider abstraction (mock/Twilio).

---

## RPC Methods

### `SendEmail`
Dispatches a template-based email.

- **Request:** `SendEmailRequest` (To, Subject, TemplateName, DataMap).
- **Response:** `SendEmailResponse`.

### `SendSMS`
Dispatches a text message.

- **Request:** `SendSMSRequest`.

### `RegisterDevice` (Future)
For Push Notification token management.

---

## Message Definitions

### SendEmailRequest
| Field | Type | Description |
|-------|------|-------------|
| `to` | `string` | Recipient email |
| `template_name` | `string` | e.g., `ticket_confirmation` |
| `data` | `map<string,string>` | Dynamic values for the template |
