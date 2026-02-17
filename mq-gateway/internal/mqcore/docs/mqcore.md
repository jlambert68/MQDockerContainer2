# mqcore.go description

Source: `mq-gateway/internal/mqcore/mqcore.go`

## Purpose

Implements direct IBM MQ operations and browse session lifecycle.

## Key responsibilities

- Build MQ client connection from environment variables.
- Optional TLS configuration and channel status logging.
- Core operations:
  - `Put`
  - `Get`
  - `BrowseFirst`
  - `BrowseNext`
  - `InquireQueue`
- Browse cursor/session management with idle timeout cleanup.
- Graceful MQ disconnect and session cleanup.

## Important behaviors

- `MQ_TLS_ENABLED=true` requires `MQ_SSLCIPH` and `MQ_KEY_REPOSITORY`.
- `Get` and browse calls can return `empty=true` when no message is available.
- Browse sessions expire after configured TTL (`5m` currently).
