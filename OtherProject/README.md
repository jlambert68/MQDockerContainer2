# OtherProject (Simulation)

This folder simulates a separate project that only needs to start the two
containers (`mq` and `mq-gateway`) using Docker Compose.

## Start

```bash
cd OtherProject
docker compose up -d --build
```

## Stop

```bash
docker compose down
```

## Notes

- gRPC listens on `localhost:9090`
- REST listens on `localhost:8080`
- MQ listener on `localhost:1414`

## TLS client examples

TLS examples live in:
- `OtherProject/grpc-client-tls`
- `OtherProject/rest-client-tls`

Sample files were generated in `OtherProject/tls`:

- `OtherProject/tls/appclient.crt`
- `OtherProject/tls/appclient.key` (unencrypted)
- `OtherProject/tls/appclient.enc.key` (encrypted, password: `passw0rd`)
- `OtherProject/tls/appclient.p12` (password: `passw0rd`)
- `OtherProject/tls/qm1server.crt` (use as CA)

Run them by setting environment variables:

```bash
export MQ_TLS_CERT=../tls/appclient.crt
export MQ_TLS_KEY=../tls/appclient.enc.key
export MQ_TLS_CA=../tls/qm1server.crt
export MQ_TLS_KEY_PASSWORD_ENV=MQ_KEY_PASSWORD
export MQ_KEY_PASSWORD=passw0rd

# If you use the unencrypted key:
# export MQ_TLS_KEY=../tls/appclient.key
# unset MQ_TLS_KEY_PASSWORD_ENV

cd OtherProject/grpc-client-tls && go run .
cd OtherProject/rest-client-tls && go run .
```

To use a PKCS#12 bundle instead:

```bash
export MQ_TLS_P12=../tls/appclient.p12
export MQ_TLS_P12_PASSWORD_ENV=MQ_P12_PASSWORD
export MQ_P12_PASSWORD=passw0rd
```
