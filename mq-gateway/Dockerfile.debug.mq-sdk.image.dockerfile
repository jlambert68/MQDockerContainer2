# 1) Source stage: IBM MQ image that contains /opt/mqm (incl headers + libs)
FROM icr.io/ibm-messaging/mq:latest AS mqsrc

# 2) Build stage: Go build with CGO + MQ headers/libs
FROM golang:1.25 AS build
WORKDIR /src

# tools for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev pkg-config ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Copy your source
COPY go.mod ./
COPY . .
RUN go mod download

# Copy MQ installation (headers + libs) from mqsrc
COPY --from=mqsrc /opt/mqm /opt/mqm

# Sanity check
RUN test -f /opt/mqm/inc/cmqc.h

# Tell cgo where MQ headers/libs are
ENV CGO_CFLAGS="-I/opt/mqm/inc"
ENV CGO_LDFLAGS="-L/opt/mqm/lib64 -Wl,-rpath,/opt/mqm/lib64"

# Build with debug flags
RUN CGO_ENABLED=1 go build -gcflags="all=-N -l" -o /out/app .

# 3) Delve stage
FROM golang:1.25 AS dlv
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# 4) Runtime stage: minimal Debian + your app + dlv + MQ libs
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Need MQ runtime libs at runtime:
COPY --from=mqsrc /opt/mqm /opt/mqm

# App + debugger
COPY --from=build /out/app /app/app
COPY --from=dlv /go/bin/dlv /usr/local/bin/dlv

ENV LD_LIBRARY_PATH="/opt/mqm/lib64"
EXPOSE 8080 2345
CMD ["dlv","exec","/app/app","--headless","--listen=:2345","--api-version=2","--accept-multiclient","--continue","--log"]
