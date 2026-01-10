# Fraud Service gRPC Documentation

**Package:** `fraud.v1`  
**Internal DNS:** `fraud:50090`  
**Proto File:** `server/api/proto/fraud/v1/fraud.proto`

## Overview
The Fraud Service provides intelligent risk assessment using a hybrid of **Deterministic Rules**, **User Profiling**, and **Generative AI**. It also handles document verification (NID/Passport) using Vision models.

## Key Behaviors

### AI-Driven Risk Analysis
The `AnalyzeBooking` pipeline combines three inputs for the LLM:
1. **Transaction Data:** Current booking details (Velocity, IP, Amount).
2. **User Profiling:** Deviation analysis from historical user behavior (e.g., "Is this IP new for this user?", "Is the spending 5x their average?").
3. **RAG (Retrieval-Augmented Generation):** Retrieves semantically similar past fraud cases to provide context to the model.

### Fail-Safe Logic
- **Primary:** AI Model assessment (Score 0-100).
- **Fallback:** If the AI provider is unavailable, the service falls back to **Statistical Deviation Scoring** logic to ensure business continuity ("Fail-Open" or "Fail-Closed" denotes logic).

### Blocking
- Transactions are automatically flagged for blocking if `risk_score >= 70` (Default threshold).

---

## RPC Methods

### `AnalyzeBooking`
Evaluates a transaction for fraud risk.

- **Request:** `AnalyzeBookingRequest`
- **Response:** `FraudResult`
  - `risk_score`: 0-100.
  - `should_block`: boolean advice.
  - `risk_factors`: List of specific concerns (e.g., `velocity_abuse`).

### `VerifyDocument`
Analyzes an image document for authenticity.

- **Request:** `VerifyDocumentRequest` (Base64 image).
- **Response:** `VerifyDocumentResponse` (Tampering score, OCR extraction).

---

## Message Definitions

### Risk Levels
- `LOW` (0-20): Safe.
- `MEDIUM` (21-69): Review recommended.
- `HIGH` (70-89): Likely fraud.
- `CRITICAL` (90+): Definite Blocking.
