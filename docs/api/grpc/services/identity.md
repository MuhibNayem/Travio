# Identity Service gRPC Documentation

**Package:** `identity.v1`  
**Internal DNS:** `identity:9081`  
**Proto File:** `server/api/proto/identity/v1/identity.proto`

## Overview
The Identity Service is the IAM (Identity and Access Management) core of the Travio ecosystem. It handles:
- **Authentication:** JWT-based login with Refresh Token Rotation.
- **Authorization:** Role-based access control (RBAC) at the organization level.
- **Tenancy:** Organization creation and member management.

## Key Behaviors

### Token Rotation & Security
- **Refresh Token Rotation:** Every time a refresh token is used, it is invalidated and replaced by a new one within the same `FamilyID`.
- **Reuse Detection:** If an invalidated refresh token is used again (indicating theft), the entire token family is revoked, forcing the user to re-login.
- **Blacklisting:** Logout operations add the JTI to a Redis blacklist for immediate revocation before the token naturally expires.

### Data Model
- **Tenancy:** Users are scoped to a primary `organization_id` but can be invited to others.
- **Roles:** Current support for `admin`, `agent`, and `user` roles.

---

## RPC Methods

### `Register`
Registers a new user and links them to an organization.
> **Note:** Currently accepts passwords of any length (client-side validation expected).

- **Request:** `RegisterRequest`
- **Response:** `RegisterResponse`

### `Login`
Authenticates user credentials and issues a `TokenPair`.

- **Request:** `LoginRequest`
- **Response:** `LoginResponse`
- **Errors:** `INVALID_CREDENTIALS` (401)

### `RefreshToken`
Rotates the session tokens.

- **Request:** `RefreshTokenRequest`
- **Response:** `RefreshTokenResponse`
- **Errors:** `REFRESH_TOKEN_REUSED` (401) - Critical security event.

### `Logout`
Revokes the refresh token and terminates the session.

- **Request:** `LogoutRequest`
- **Response:** `LogoutResponse`

### `CreateOrganization`
Creates a new tenant and an associated subscription record.

- **Request:** `CreateOrgRequest`
- **Response:** `CreateOrgResponse`

---

## Message Definitions

### RegisterRequest
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `email` | `string` | - | Valid email format required |
| `password` | `string` | - | Plaintext password |
| `organization_id` | `string` | - | Optional. If provided, adds user to org. |

### RegisterResponse
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `user_id` | `string` | - | UUIDv4 of the new user |

### LoginRequest
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `email` | `string` | - | User email |
| `password` | `string` | - | User password |

### LoginResponse
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `access_token` | `string` | - | JWT (Stateless, 15m expiry) |
| `refresh_token` | `string` | - | Opaque Token (Stateful, 7d expiry) |
| `expires_in` | `int64` | - | Seconds until access token expiry |

### RefreshTokenRequest
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `refresh_token` | `string` | - | The current valid refresh token |

### CreateOrgRequest
| Field | Type | Label | Description |
|-------|------|-------|-------------|
| `name` | `string` | - | Organization display name |
| `plan_id` | `string` | - | Subscription plan (e.g., 'pro', 'enterprise') |
