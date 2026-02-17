# MQ Gateway User Guide

This guide shows how to use `mq-gateway` end to end:

1. Install prerequisites
2. Start IBM MQ + `mq-gateway`
3. Call the REST and gRPC APIs
4. Implement small Go clients

## 1. What `mq-gateway` does

`mq-gateway` provides a simple HTTP (REST) and gRPC interface on top of IBM MQ.

Operations exposed:

- `Put`
- `Get`
- `BrowseFirst`
- `BrowseNext`
- `InquireQueue`

Default local ports:

- REST: `8080`
- gRPC: `9090`

## 2. Prerequisites

- Docker + Docker Compose
- Go `1.25.x` (only needed for local client examples)

## 3. Start the stack

Run from `mq-gateway/`.

### Option A (recommended quick start): non-TLS MQ channel

This uses `docker-compose.debug.mq-sdk.image.yml` and is the easiest path.

```bash
cd mq-gateway
docker compose -f docker-compose.debug.mq-sdk.image.yml up --build
```

This starts:

- IBM MQ container (`mq-local_host`)
- `mq-gateway` container

### Option B: TLS between `mq-gateway` and MQ

This uses `docker-compose.debug.mq-sdk.local.yml` (TLS-enabled channel and key repository).

```bash
cd mq-gateway
docker compose -f docker-compose.debug.mq-sdk.local.yml up --build
```

### Optional: use Makefile shortcuts

From `mq-gateway/`:

```bash
make Build_docker-compose.debug.mq-sdk.local
make Run_docker-compose.debug.mq-sdk.local
make Show_Logs_docker-compose.debug.mq-sdk.local
make Exit_docker-compose.debug.mq-sdk.local
```

## 4. Verify the gateway is running

In a second terminal:

```bash
curl -sS -X POST http://localhost:8080/inquire/queue \
  -H 'Content-Type: application/json' \
  -d '{"queue":"DEV.QUEUE.1"}' | jq .
```

Expected: JSON with `"status":"ok"`.

## 5. REST API quick examples

### 5.1 Put

```bash
curl -sS -X POST http://localhost:8080/put \
  -H 'Content-Type: application/json' \
  -d '{"queue":"DEV.QUEUE.1","message":"Hello from REST"}' | jq .
```

### 5.2 BrowseFirst and BrowseNext (non-destructive)

```bash
BROWSE_ID=$(curl -sS -X POST http://localhost:8080/browse/first \
  -H 'Content-Type: application/json' \
  -d '{"queue":"DEV.QUEUE.1","wait_ms":1000,"max_msg_bytes":65536}' | jq -r .browse_id)

curl -sS -X POST http://localhost:8080/browse/next \
  -H 'Content-Type: application/json' \
  -d "{\"browse_id\":\"${BROWSE_ID}\",\"wait_ms\":1000,\"max_msg_bytes\":65536}" | jq .
```

Notes:

- `BrowseFirst` returns a `browse_id`.
- Browse sessions expire after idle timeout (currently 5 minutes in implementation).

### 5.3 Get (destructive)

```bash
curl -sS -X POST http://localhost:8080/get \
  -H 'Content-Type: application/json' \
  -d '{"queue":"DEV.QUEUE.1","wait_ms":5000,"max_msg_bytes":65536}' | jq .
```

`empty=true` means no message was available.

### 5.4 Inquire queue attributes

```bash
curl -sS -X POST http://localhost:8080/inquire/queue \
  -H 'Content-Type: application/json' \
  -d '{"queue":"DEV.QUEUE.1"}' | jq .
```

## 6. gRPC quick example (Go)

A complete sample exists at:

- `business-clients/business-client-gRPC/main_gRPC.go`

Run it:

```bash
cd business-clients/business-client-gRPC
go mod tidy
go run .
```

It exercises:

- `Put` twice
- `BrowseFirst`/`BrowseNext`
- `InquireQueue`
- `Get` twice

## 7. REST client quick example (Go)

A complete sample exists at:

- `business-clients/business-client-rest/main_rest.go`

Run it:

```bash
cd business-clients/business-client-rest
go mod tidy
go run .
```

## 8. Minimal implementation snippets

### 8.1 Minimal REST client (Go)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqrest"
)

func main() {
    c := mqrest.NewClient(mqrest.Config{
        BaseURL: "http://localhost:8080",
        Timeout: 10 * time.Second,
    })

    putResp, _ := c.Put(context.Background(), mqrest.PutRequest{
        Queue:   "DEV.QUEUE.1",
        Message: "hello",
    })
    fmt.Println("put:", putResp.Status)

    getResp, _ := c.Get(context.Background(), mqrest.GetRequest{
        Queue:       "DEV.QUEUE.1",
        WaitMs:      1000,
        MaxMsgBytes: 65536,
    })
    fmt.Printf("get: empty=%v message=%q\n", getResp.Empty, getResp.Message)
}
```

### 8.2 Minimal gRPC client (Go)

```go
package main

import (
    "context"
    "fmt"

    "github.com/jlambert68/MQDockerContainer2/mq-gateway/api/proto/mq_grpc_api"
    "github.com/jlambert68/MQDockerContainer2/mq-gateway/pkg/mqgrpc"
)

func main() {
    c, err := mqgrpc.NewClient(mqgrpc.Config{Address: "localhost:9090"})
    if err != nil {
        panic(err)
    }
    defer c.Close()

    putResp, err := c.Put(context.Background(), &mq_grpc_api.PutRequest{
        Queue: "DEV.QUEUE.1", Message: "hello via grpc",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("put:", putResp.Status)
}
```

## 9. Configuration reference (`mq-gateway`)

Main environment variables:

- `MQ_QMGR` (default: `QM1`)
- `MQ_HOST` (default: `mq-local_host`)
- `MQ_PORT` (default: `1414`)
- `MQ_CHANNEL` (default: `DEV.TLS.SVRCONN`)
- `MQ_USER` (default: `app`)
- `MQ_PASSWORD` (default: `passw0rd`)
- `MQ_TLS_ENABLED` (`true`/`false`, default: `false`)
- `MQ_SSLCIPH` (required when TLS enabled)
- `MQ_KEY_REPOSITORY` (required when TLS enabled)
- `REST_PORT` (default: `:8080`)
- `GRPC_PORT` (default: `:9090`)

If `MQ_TLS_ENABLED=true`, `MQ_SSLCIPH` and `MQ_KEY_REPOSITORY` must be set or startup fails.

## 10. Error behavior

REST:

- `400` for invalid input (for example, missing `queue`)
- `502` when MQ operation fails

gRPC:

- RPC typically returns a response with `status="error"` and `error` field populated

## 11. Troubleshooting

### Gateway cannot connect to MQ

Check container logs:

```bash
cd mq-gateway
docker compose -f docker-compose.debug.mq-sdk.image.yml logs -f mq-gateway
```

For TLS mode, verify:

- Correct channel (for example `DEV.TLS.SVRCONN`)
- `MQ_SSLCIPH` matches queue manager channel cipher
- `MQ_KEY_REPOSITORY` path exists in container

### `browse_id not found or expired`

Start a new browse session using `BrowseFirst` and use the new `browse_id` quickly.

### `empty=true` on `Get`

No messages were available in the queue during `wait_ms`.

## 12. Next steps

- Use `pkg/mqrest` for HTTP-based integrations.
- Use `pkg/mqgrpc` for strongly-typed gRPC integrations.
- Keep queue names/config in environment variables for deployment portability.
