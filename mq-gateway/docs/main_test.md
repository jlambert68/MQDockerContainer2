# main_test.go

Source: `mq-gateway/main_test.go`

## Purpose

Tests helper behavior in `main.go` without starting servers.

## Coverage

- `getenv` returns environment value when set
- `getenv` returns fallback default when missing

## Run

```bash
go test .
```
