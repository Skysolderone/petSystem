CONFIG ?= ./config/config.yaml
GOCACHE ?= /tmp/go-build
SMOKE_BASE_URL ?= http://127.0.0.1:8080/api/v1
SMOKE_WAIT ?= 30s
SMOKE_SKIP_WS ?= false

.PHONY: build run test smoke migrate-up migrate-down tidy

tidy:
	GOCACHE=$(GOCACHE) go mod tidy

build:
	GOCACHE=$(GOCACHE) go build ./...

run: migrate-up
	GOCACHE=$(GOCACHE) go run ./cmd/api -config $(CONFIG)

test:
	GOCACHE=$(GOCACHE) go test ./...

smoke:
	GOCACHE=$(GOCACHE) go run ./cmd/smoke -base-url $(SMOKE_BASE_URL) -wait $(SMOKE_WAIT) -skip-ws=$(SMOKE_SKIP_WS)

migrate-up:
	GOCACHE=$(GOCACHE) go run ./cmd/migrate -config $(CONFIG) up

migrate-down:
	GOCACHE=$(GOCACHE) go run ./cmd/migrate -config $(CONFIG) down
