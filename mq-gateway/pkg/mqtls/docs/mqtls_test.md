# mqtls_test.go

Source: `mq-gateway/pkg/mqtls/mqtls_test.go`

## Purpose

Tests TLS helper functions for cert/key loading and key parsing.

## Coverage

- Load TLS config from PEM files
- Missing password env handling for encrypted keys
- Encrypted PEM private key parsing behavior

## Run

```bash
go test ./pkg/mqtls
```
