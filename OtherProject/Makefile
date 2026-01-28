.PHONY: build up down logs test-grpc

build:
	docker compose build --no-cache

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

test-grpc:
	cd grpc-client && go run .
