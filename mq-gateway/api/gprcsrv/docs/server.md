# server.go description

Source: `mq-gateway/api/gprcsrv/server.go`

## Purpose

Implements the gRPC service endpoints and maps each RPC to `mqcore.Gateway` operations.

## Main flow

1. Validate request fields (`queue` or `browse_id`).
2. Call MQ gateway (`Put`, `Get`, `BrowseFirst`, `BrowseNext`, `InquireQueue`).
3. Convert result to protobuf response.
4. Return `status="ok"` or `status="error"` with `error` text.

## Notes

- Errors are returned in response payloads rather than as transport errors in normal operation.
- Logs failures with structured `slog` entries.
