---
sidebar_position: 2
---

# Authentication

The API uses JWT-based authentication with session tracking.

## Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │     │   API    │     │ Database │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     │ POST /login    │                │
     │ {email, pass}  │                │
     │───────────────▶│                │
     │                │ Verify creds   │
     │                │───────────────▶│
     │                │ Create session │
     │                │───────────────▶│
     │ JWT + Refresh  │                │
     │◀───────────────│                │
     │                │                │
     │ API Request    │                │
     │ Bearer <JWT>   │                │
     │───────────────▶│                │
     │                │ Verify JWT     │
     │                │ Load session   │
     │ Response       │                │
     │◀───────────────│                │
```

## JWT Structure

```json
{
  "sub": "employee-uuid",
  "org_id": "organization-uuid",
  "role": "admin",
  "exp": 1234567890,
  "iat": 1234567800
}
```

## Token Lifecycle

| Token | Expiration | Storage |
|-------|------------|---------|
| Access Token (JWT) | 1 year | Memory/localStorage |
| Refresh Token | 30 days | HttpOnly cookie |
| Session | Until logout | Database |

## Security

- JWTs signed with HS256
- Passwords hashed with bcrypt (cost 12)
- Sessions can be revoked server-side
- Rate limiting on login endpoints
