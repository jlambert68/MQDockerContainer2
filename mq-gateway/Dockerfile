FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y gcc ca-certificates && rm -rf /var/lib/apt/lists/*

COPY mq-client/ /opt/mqm/

ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV CC=gcc
ENV CGO_CFLAGS="-I/opt/mqm/tools/c/include -D_REENTRANT"
ENV CGO_LDFLAGS="-L/opt/mqm/tools/lib64 -lmqm"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /out
RUN go build -v -o /out/mq-gateway ./cmd/mq-gateway

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/mqm /opt/mqm
ENV LD_LIBRARY_PATH=/opt/mqm/tools/lib64

COPY --from=builder /out/mq-gateway /usr/local/bin/mq-gateway
EXPOSE 8080 9090
CMD ["mq-gateway"]
