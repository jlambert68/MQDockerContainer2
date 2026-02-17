# mqgrpc_test.go

Source: `mq-gateway/pkg/mqgrpc/mqgrpc_test.go`

## Purpose

Tests `pkg/mqgrpc` client wrapper against an in-memory gRPC server.

## Coverage

- Client creation with custom dial options and transport creds
- `Put` call via wrapper client returns expected response

## Run

```bash
go test ./pkg/mqgrpc
```
