# rest_test.go

Source: `mq-gateway/api/rest/rest_test.go`

## Purpose

Tests REST handlers with `httptest` and a fake `mqcore.Gateway`.

## Coverage

- Success path for `PUT`
- Invalid JSON / invalid input handling (`400`)
- Backend failure mapping (`502` + error status)
- Route registration for all REST endpoints

## Run

```bash
go test ./api/rest
```
