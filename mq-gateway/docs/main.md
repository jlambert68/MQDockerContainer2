# main.go description

Source: `mq-gateway/main.go`

## Purpose

Application entrypoint that wires MQ core, REST API, gRPC API, and graceful shutdown.

## Startup sequence

1. Initialize structured logging.
2. Sleep briefly to let MQ container initialize.
3. Create shared `mqcore.Gateway` connection.
4. Start REST server on `REST_PORT` (default `:8080`).
5. Start gRPC server on `GRPC_PORT` (default `:9090`).

## Shutdown sequence

1. Wait for `SIGINT`/`SIGTERM`.
2. Gracefully stop gRPC server.
3. Gracefully shutdown REST server.
4. Close MQ gateway.
