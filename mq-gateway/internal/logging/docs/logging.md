# logging.go description

Source: `mq-gateway/internal/logging/logging.go`

## Purpose

Centralizes logger initialization for the gateway.

## Main flow

1. Build a JSON `slog` handler bound to `os.Stdout`.
2. Attach common service metadata.
3. Set logger as global default.

## Notes

- Keeps logging consistent across `main`, REST, gRPC, and MQ core packages.
