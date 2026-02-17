# mqrest.go description

Source: `mq-gateway/pkg/mqrest/mqrest.go`

## Purpose

Exposes a small HTTP/JSON client API for `mq-gateway`.

## Main flow

1. Build client with base URL, timeout, optional TLS transport.
2. High-level methods call shared `post` helper with endpoint path.
3. `post` marshals JSON, sends request, decodes JSON response.
4. Returns typed response and error on HTTP failure status.

## Notes

- Default base URL: `http://localhost:8080`.
- Default timeout: `10s`.
