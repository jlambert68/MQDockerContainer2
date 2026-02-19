#!/usr/bin/env bash
# Exit on first error, treat unset vars as errors, and fail on pipe errors.
set -euo pipefail

# Input variables (required + optional defaults).
CA_CRT_PATH="${CA_CRT_PATH:-}"
CLIENT_P12_PATH="${CLIENT_P12_PATH:-}"
CLIENT_P12_PASSWORD="${CLIENT_P12_PASSWORD:-}"
KDB_PASSWORD="${KDB_PASSWORD:-}"
KDB_DIR="${KDB_DIR:-./pki-local/client}"
KDB_BASENAME="${KDB_BASENAME:-key}"
ROOT_CA_LABEL="${ROOT_CA_LABEL:-rootca}"

# Validate required inputs and print usage when missing.
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

# Ensure IBM MQ key management tool is available.
if ! command -v runmqakm >/dev/null 2>&1; then
  echo "ERROR: runmqakm not found in PATH." >&2
  echo "Install IBM MQ client tools or run this from an MQ-enabled container." >&2
  exit 1
fi

# Prepare key repository paths.
mkdir -p "$KDB_DIR"
KDB_BASE="$KDB_DIR/$KDB_BASENAME"
KDB_FILE="${KDB_BASE}.kdb"

# Create key database + stash file once; reuse if already present.
if [[ ! -f "$KDB_FILE" ]]; then
  runmqakm -keydb -create -db "$KDB_FILE" -pw "$KDB_PASSWORD" -type cms -stash
fi

# Keep CA signer import idempotent by replacing the label if it exists.
if runmqakm -cert -list -db "$KDB_FILE" -pw "$KDB_PASSWORD" 2>/dev/null | grep -Fxq "$ROOT_CA_LABEL"; then
  runmqakm -cert -delete -db "$KDB_FILE" -pw "$KDB_PASSWORD" -label "$ROOT_CA_LABEL"
fi

# Add CA signer certificate used to trust the queue manager cert chain.
runmqakm -cert -add \
  -db "$KDB_FILE" -pw "$KDB_PASSWORD" \
  -label "$ROOT_CA_LABEL" -file "$CA_CRT_PATH" -format ascii

# Import client personal certificate + private key from PKCS#12 bundle.
runmqakm -cert -import \
  -db "$KDB_FILE" -pw "$KDB_PASSWORD" \
  -file "$CLIENT_P12_PATH" -type pkcs12 -pw "$CLIENT_P12_PASSWORD"

# Print the repository base path expected by MQ_KEY_REPOSITORY (no .kdb suffix).
echo
echo "TLS key repository is ready."
echo "Set MQ_KEY_REPOSITORY=$KDB_BASE"
