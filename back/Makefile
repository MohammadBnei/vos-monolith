BIN_DIR = bin
APP_NAME = voconsteroid
BINARY = $(BIN_DIR)/$(APP_NAME)
DATABASE_URL ?= postgres://user:password@localhost:5432/voconsteroid?sslmode=disable

.PHONY: run
run: ## Run the application with air for live reload
	air

.PHONY: build
build: ## Build the application
	@mkdir -p $(BIN_DIR)
	go build -o ../$(BINARY) ./api/main.go

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: migrate
migrate: ## Run database migrations
	migrate -path internal/infrastructure/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	migrate -path internal/infrastructure/migrations -database "$(DATABASE_URL)" down

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
