.PHONY: help up down logs run test lint build migrate-up migrate-down docker-build clean

POSTGRES_DSN ?= postgres://linkforge:linkforge@localhost:5432/linkforge?sslmode=disable
REDIS_ADDR   ?= localhost:6379

help:
	@echo "Targets:"
	@echo "  up            start postgres + redis (docker compose)"
	@echo "  down          stop and remove containers"
	@echo "  logs          tail compose logs"
	@echo "  migrate-up    apply migrations (needs golang-migrate CLI)"
	@echo "  migrate-down  rollback last migration"
	@echo "  run           run api locally"
	@echo "  test          run tests with race detector"
	@echo "  lint          go vet"
	@echo "  build         build binary into bin/"
	@echo "  docker-build  build docker image linkforge:dev"
	@echo "  clean         remove bin/"

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

run:
	POSTGRES_DSN="$(POSTGRES_DSN)" REDIS_ADDR="$(REDIS_ADDR)" go run ./cmd/api

test:
	go test ./... -race -count=1

lint:
	go vet ./...

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

migrate-up:
	migrate -database "$(POSTGRES_DSN)" -path ./migrations up

migrate-down:
	migrate -database "$(POSTGRES_DSN)" -path ./migrations down 1

docker-build:
	docker build -t linkforge:dev .

clean:
	rm -rf bin/
