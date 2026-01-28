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
