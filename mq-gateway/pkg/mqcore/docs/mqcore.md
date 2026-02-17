# mqcore.go description

Source: `mq-gateway/pkg/mqcore/mqcore.go`

## Purpose

Public, stable facade over `internal/mqcore` for consumers outside internal packages.

## Main flow

- `NewGateway` creates internal implementation.
- Wrapper methods delegate calls to the internal gateway.
- `InquireQueue` maps internal queue info into exported `QueueInfo` type.

## Notes

- Separates external API contract from internal implementation details.
