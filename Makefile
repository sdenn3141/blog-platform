all: build test

build:
	@echo "Building..."
	@go build -o main.exe cmd/api/main.go

run:
	@go run cmd/api/main.go

docker-run:
	@docker compose up --build

docker-down:
	@docker compose down

test:
	@echo "Testing..."
	@go test ./...

itest:
	@echo "Running integration tests..."
	@go test -tags=integration ./...

clean:
	@echo "Cleaning..."
	@rm -f main

.PHONY: all build run test clean docker-run docker-down itest
