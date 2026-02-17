# mqrest_test.go

Source: `mq-gateway/pkg/mqrest/mqrest_test.go`

## Purpose

Tests REST client wrapper behavior with `httptest` server.

## Coverage

- Transport selection based on TLS config presence
- Successful `Put` request path and method verification
- Error return when HTTP status is `>= 400`

## Run

```bash
go test ./pkg/mqrest
```
