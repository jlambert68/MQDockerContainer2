# mqgrpc.go description

Source: `mq-gateway/pkg/mqgrpc/mqgrpc.go`

## Purpose

Provides a thin, ergonomic gRPC client wrapper for MQ gateway RPCs.

## Main flow

1. Configure address and transport (custom creds, TLS, or insecure).
2. Dial server and build generated protobuf client.
3. Delegate all RPC calls (`Put`, `Get`, `Browse*`, `InquireQueue`).
4. Close connection when done.

## Notes

- `TransportCredentials` takes precedence over `TLSConfig`.
- Defaults to `localhost:9090` when no address is provided.
