Build_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml build --no-cache --progress=plain

Run_docker-compose.debug.mq-sdk.local:
	docker compose -f docker-compose.debug.mq-sdk.local.yml up

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
