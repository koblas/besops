BINARY := rupert
PKG := github.com/koblas/rupert
DB_URL ?= sqlite://./data/besops.db

.PHONY: build run test lint fmt vet generate clean migrate-up migrate-down bobgen oas

build:
	go build -o bin/$(BINARY) ./cmd/uptime

run:
	go run ./cmd/uptime

test:
	go test ./... -race -count=1

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

generate: oas bobgen

oas:
	go generate ./internal/api/

migrate-up:
	go run ./cmd/migrate -dir up

migrate-down:
	go run ./cmd/migrate -dir down

bobgen: migrate-up
	go generate ./internal/database/bobgen/

clean:
	rm -rf bin/ data/
