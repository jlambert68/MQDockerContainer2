# mq_proto_test.go

Source: `mq-gateway/api/proto/mq_grpc_api/mq_proto_test.go`

## Purpose

Tests generated protobuf/gRPC code paths and message helpers.

## Coverage

- Getter and `Reset` behavior for generated message types
- Unimplemented server default error behavior
- End-to-end generated client call over in-memory `bufconn`
- Service descriptor method count sanity check

## Run

```bash
go test ./api/proto/mq_grpc_api
```
