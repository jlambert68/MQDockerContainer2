# logging_test.go

Source: `mq-gateway/internal/logging/logging_test.go`

## Purpose

Smoke test for logging initialization.

## Coverage

- `Init` configures global default `slog` logger
- Logging call after initialization does not panic

## Run

```bash
go test ./internal/logging
```
