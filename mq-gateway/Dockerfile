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

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libstdc++6 \
    libgcc-s1 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /opt/mqm /opt/mqm

# Make /opt/mqm/tools/lib64 part of the dynamic linker search path
RUN echo "/opt/mqm/tools/lib64" > /etc/ld.so.conf.d/mqm.conf && ldconfig

# Optional (kept for clarity; ldconfig is the important part)
ENV LD_LIBRARY_PATH=/opt/mqm/tools/lib64

COPY --from=builder /out/mq-gateway /usr/local/bin/mq-gateway
EXPOSE 8080 9090
CMD ["mq-gateway"]



FROM golang:1.25-bookworm AS builder
RUN apt-get update && apt-get install -y gcc ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o /out/mq-gateway ./cmd/mq-gateway

FROM icr.io/ibm-messaging/mq:9.3.2.0-r2
COPY --from=builder /out/mq-gateway /usr/local/bin/mq-gateway
EXPOSE 8080 9090
CMD ["mq-gateway"]

