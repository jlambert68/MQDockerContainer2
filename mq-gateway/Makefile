Build_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml build --no-cache --progress=plain

Run_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml up

Run_docker-compose.debug.mq-sdk.local_with_tls_env:
	docker compose --env-file .env.mq-client-tls.example -f docker-compose.debug.mq-sdk.local.yml up -d

Exit_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml down

Recreate_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml up --force-recreate

Show_Logs_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml logs -f

Run_UnitTests_docker-compose.debug.mq.local:
	docker run --rm -v "$$PWD":/src/mq-gateway -w /src/mq-gateway \
      -e CGO_CFLAGS="-I/src/mq-gateway/mq-client/tools/c/include -I/src/mq-gateway/mq-client/tools/inc" \
      -e CGO_LDFLAGS="-L/src/mq-gateway/mq-client/tools/lib64 -Wl,-rpath,/src/mq-gateway/mq-client/tools/lib64" \
      golang:1.25 bash -c \
      "apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev pkg-config ca-certificates && go test ./..."

Run_UnitTests_Verbose_docker-compose.debug.mq.local:
	docker run --rm -v "$$PWD":/src/mq-gateway -w /src/mq-gateway \
      -e CGO_CFLAGS="-I/src/mq-gateway/mq-client/tools/c/include -I/src/mq-gateway/mq-client/tools/inc" \
      -e CGO_LDFLAGS="-L/src/mq-gateway/mq-client/tools/lib64 -Wl,-rpath,/src/mq-gateway/mq-client/tools/lib64" \
      golang:1.25 bash -c \
      "apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev pkg-config ca-certificates && go test -v ./..."

Run_UnitTests_Local:
	CGO_CFLAGS="-I$$PWD/mq-client/tools/c/include -I$$PWD/mq-client/tools/inc" \
	CGO_LDFLAGS="-L$$PWD/mq-client/tools/lib64 -Wl,-rpath,$$PWD/mq-client/tools/lib64" \
	go test ./...

# One-command setup for MQ client key repository (.kdb/.sth) from a CA cert + client P12.
# Required variables:
#   CA_CRT_PATH=/path/to/ca.crt
#   CLIENT_P12_PATH=/path/to/client.p12
#   CLIENT_P12_PASSWORD=...
#   KDB_PASSWORD=...
# Optional variables:
#   KDB_DIR=./pki-local/client
#   KDB_BASENAME=key
#   ROOT_CA_LABEL=rootca
Setup_MQ_Client_TLS:
	CA_CRT_PATH="$(CA_CRT_PATH)" \
	CLIENT_P12_PATH="$(CLIENT_P12_PATH)" \
	CLIENT_P12_PASSWORD="$(CLIENT_P12_PASSWORD)" \
	KDB_PASSWORD="$(KDB_PASSWORD)" \
	KDB_DIR="$(KDB_DIR)" \
	KDB_BASENAME="$(KDB_BASENAME)" \
	ROOT_CA_LABEL="$(ROOT_CA_LABEL)" \
	bash ./scripts/setup-mq-client-tls.sh
