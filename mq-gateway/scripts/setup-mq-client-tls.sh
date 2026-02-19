#!/usr/bin/env bash
set -euo pipefail

CA_CRT_PATH="${CA_CRT_PATH:-}"
CLIENT_P12_PATH="${CLIENT_P12_PATH:-}"
CLIENT_P12_PASSWORD="${CLIENT_P12_PASSWORD:-}"
KDB_PASSWORD="${KDB_PASSWORD:-}"
KDB_DIR="${KDB_DIR:-./pki-local/client}"
KDB_BASENAME="${KDB_BASENAME:-key}"
ROOT_CA_LABEL="${ROOT_CA_LABEL:-rootca}"

if [[ -z "$CA_CRT_PATH" || -z "$CLIENT_P12_PATH" || -z "$CLIENT_P12_PASSWORD" || -z "$KDB_PASSWORD" ]]; then
  cat >&2 <<'USAGE'
Usage (via env vars):
  CA_CRT_PATH=/path/to/ca.crt \
  CLIENT_P12_PATH=/path/to/client.p12 \
  CLIENT_P12_PASSWORD='...' \
  KDB_PASSWORD='...' \
  [KDB_DIR=./pki-local/client] \
  [KDB_BASENAME=key] \
  [ROOT_CA_LABEL=rootca] \
  bash ./scripts/setup-mq-client-tls.sh

Example:
  make Setup_MQ_Client_TLS \
    CA_CRT_PATH=./certs/mq/ca.crt \
    CLIENT_P12_PATH=./certs/mq/client.p12 \
    CLIENT_P12_PASSWORD='changeit' \
    KDB_PASSWORD='changeit'
USAGE
  exit 2
fi

if ! command -v runmqakm >/dev/null 2>&1; then
  echo "ERROR: runmqakm not found in PATH." >&2
  echo "Install IBM MQ client tools or run this from an MQ-enabled container." >&2
  exit 1
fi

mkdir -p "$KDB_DIR"
KDB_BASE="$KDB_DIR/$KDB_BASENAME"
KDB_FILE="${KDB_BASE}.kdb"

if [[ ! -f "$KDB_FILE" ]]; then
  runmqakm -keydb -create -db "$KDB_FILE" -pw "$KDB_PASSWORD" -type cms -stash
fi

# Keep CA label idempotent by replacing it when it already exists.
if runmqakm -cert -list -db "$KDB_FILE" -pw "$KDB_PASSWORD" 2>/dev/null | grep -Fxq "$ROOT_CA_LABEL"; then
  runmqakm -cert -delete -db "$KDB_FILE" -pw "$KDB_PASSWORD" -label "$ROOT_CA_LABEL"
fi

runmqakm -cert -add \
  -db "$KDB_FILE" -pw "$KDB_PASSWORD" \
  -label "$ROOT_CA_LABEL" -file "$CA_CRT_PATH" -format ascii

runmqakm -cert -import \
  -db "$KDB_FILE" -pw "$KDB_PASSWORD" \
  -file "$CLIENT_P12_PATH" -type pkcs12 -pw "$CLIENT_P12_PASSWORD"

echo
echo "TLS key repository is ready."
echo "Set MQ_KEY_REPOSITORY=$KDB_BASE"
