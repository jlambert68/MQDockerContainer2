# rest.go description

Source: `mq-gateway/api/rest/rest.go`

## Purpose

Provides HTTP JSON handlers for MQ operations.

## Endpoints

- `POST /put`
- `POST /get`
- `POST /browse/first`
- `POST /browse/next`
- `POST /inquire/queue`

## Main flow per handler

1. Decode JSON request body.
2. Validate required fields.
3. Call `mqcore.Gateway` method.
4. Return JSON response (`status`, result fields, optional `error`).

## Error handling

- Invalid input: HTTP `400`.
- MQ/backend failure: HTTP `502` and `status="error"`.
