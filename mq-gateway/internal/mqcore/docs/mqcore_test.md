# mqcore_test.go description

Source: `mq-gateway/internal/mqcore/mqcore_test.go`

## Purpose

Validates helper behavior in `internal/mqcore`.

## Covered tests

- `TestNewBrowseID`
- `TestIntAttr`
- `TestStringAttr`
- `TestGetenv`
- `TestGetbool`

## Notes

- Focuses on deterministic utility behavior.
- Does not require live MQ connectivity.
