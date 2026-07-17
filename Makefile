.PHONY: build test run clean lint fmt vet

# Variables
BINARY_NAME=bakery-app
BUILD_DIR=bin
MAIN_PATH=./cmd/server

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Run all tests
test:
	go test ./... -v -count=1

# Run tests with race detection
test-race:
	go test ./... -v -race -count=1

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Run vet
vet:
	go vet ./...

# Run lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Tidy dependencies
tidy:
	go mod tidy

# Database migrations (requires goose: go install github.com/pressly/goose/v3/cmd/goose@latest)
DB_DSN ?= postgres://localhost:5432/bakery_app?sslmode=disable
MIGRATIONS_DIR=db/migrations

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" reset

# Run all checks (fmt, vet, test)
check: fmt vet test
