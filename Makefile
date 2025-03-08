BIN_DIR = bin
APP_NAME = voconsteroid
BINARY = $(BIN_DIR)/$(APP_NAME)

.PHONY: run
run: ## Run the application with air for live reload
	cd back && air

.PHONY: build
build: ## Build the application
	@mkdir -p $(BIN_DIR)
	cd back && go build -o ../$(BINARY) ./api/main.go

.PHONY: lint
lint: ## Run golangci-lint
	cd back && golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	cd back && go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
