# Makefile
build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

test:
	go test -v ./...

test-integration:
	go test -v -tags=integration integration_test.go

lint:
	golangci-lint run ./...

docker-build:
	docker build -t pr-reviewer-service .

docker-run:
	docker-compose up --build

.PHONY: build run test test-integration lint docker-build docker-run