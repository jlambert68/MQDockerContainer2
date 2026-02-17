# mqtls.go description

Source: `mq-gateway/pkg/mqtls/mqtls.go`

## Purpose

Utility helpers for building TLS client configuration for REST/gRPC/MQ client use.

## Supported inputs

- PEM cert + PEM key (+ optional encrypted key password)
- PKCS#12 (`.p12`/`.pfx`) bundle
- Password from environment variable helpers

## Main flow

1. Load certificate/key material.
2. Parse/decrypt key where required.
3. Optionally load CA bundle into root cert pool.
4. Return `tls.Config` with `MinVersion = TLS1.2`.

## Notes

- Encrypted PKCS#8 PEM is explicitly unsupported.
- Includes helper functions for safe key/cert parsing.
