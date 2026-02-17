# server_test.go

Source: `mq-gateway/api/gprcsrv/server_test.go`

## Purpose

Tests gRPC handler behavior using a fake `mqcore.Gateway`.

## Coverage

- Validation errors (missing required fields)
- Gateway failure mapping to `status="error"`
- Successful `InquireQueue` mapping to response fields
- BrowseNext validation behavior

## Run

```bash
go test ./api/gprcsrv
```
