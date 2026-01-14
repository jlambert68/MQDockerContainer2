FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod download

# verktyg för cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev pkg-config ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# 1) Installera IBM MQ Client/SDK här så att cmqc.h finns.
#    Efter installation ska du ha: /opt/mqm/inc/cmqc.h
#    (exakt installsteg beror på hur du hämtar MQ-klienten)

# 2) tala om var headers & libs finns (vanligt för MQ)
#ENV MQ_INSTALLATION_PATH=/opt/mqm
#ENV CGO_CFLAGS="-I/opt/mqm/inc"
#ENV CGO_LDFLAGS="-L/opt/mqm/lib64 -Wl,-rpath,/opt/mqm/lib64"
ENV MQ_INSTALLATION_PATH=/src/mq-client/tools/
ENV CGO_CFLAGS="-I${MQ_INSTALLATION_PATH}/c/include"
ENV CGO_LDFLAGS="-L${MQ_INSTALLATION_PATH}/lib64 -Wl,-rpath,${MQ_INSTALLATION_PATH}/lib64"

RUN test -f "${MQ_INSTALLATION_PATH}/c/include/cmqc.h" || (echo "cmqc.h missing" && exit 1)


#RUN ls -la /opt/mqm/inc/cmqc.h || (echo "cmqc.h missing" && exit 1)
#RUN ls -la src/mq-client/tools/c/include/cmqc.h || (echo "cmqc.h missing" && exit 1)

RUN echo "$CGO_CFLAGS" && echo "$CGO_LDFLAGS" && ls -la /src/mq-client/tools/c/include/cmqc.h

RUN ls -la ${MQ_INSTALLATION_PATH}/lib64 | head -n 200
RUN ls -la ${MQ_INSTALLATION_PATH}/lib64/libmqm_r* || true
RUN ls -la ${MQ_INSTALLATION_PATH}/lib64/libmqm* || true


RUN CGO_ENABLED=1 go build -gcflags="all=-N -l" -o /out/app .

# dlv stage
FROM golang:1.25 AS dlv
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# runtime
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/app /app/app

# copy MQ shared libs
COPY --from=build /src/mq-client/tools/lib64 /opt/mqm/lib64
ENV LD_LIBRARY_PATH=/opt/mqm/lib64

# copy full MQ client tree (not just lib64)
COPY --from=build /src/mq-client/tools /opt/mqm

ENV MQ_INSTALLATION_PATH=/opt/mqm
ENV LD_LIBRARY_PATH=/opt/mqm/lib64




COPY --from=dlv /go/bin/dlv /usr/local/bin/dlv


EXPOSE 8080 2345
CMD ["dlv","exec","/app/app","--headless","--listen=:2345","--api-version=2","--accept-multiclient","--continue","--log"]
