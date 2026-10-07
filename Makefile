BIN_DIR=./bin

DOCKER_IMAGE ?= alisa-gpt:latest
DOCKER_CONTAINER ?= alisa-gpt
DOCKER_ENV_FILE ?= .env
DOCKER_STOP_TIMEOUT ?= 15

GHCR_IMAGE ?= ghcr.io/OWNER/alisa-gpt

GREEN := \033[32m
RED := \033[31m
RESET := \033[0m

# .env опционален: нет файла — ключ приходит из окружения напрямую.
ifneq ($(wildcard $(DOCKER_ENV_FILE)),)
DOCKER_ENV_FLAGS := --env-file $(DOCKER_ENV_FILE)
else
DOCKER_ENV_FLAGS :=
endif

.PHONY: run
run: ## Run the application
	@echo "$(GREEN)▶$(RESET) Running application..."
	@go run ./cmd/app/main.go

.PHONY: build
build: ## Build the binary
	go build -o $(BIN_DIR)/alisa-gpt ./cmd/app/main.go

.PHONY: lint
lint: ## Run linters on all files
	@echo "$(GREEN)▶$(RESET) Running linters..."
	@go tool golangci-lint run \
		--fix \
		--config=.golangci.yaml \
		--max-issues-per-linter=1000 \
		--max-same-issues=1000 \
		./...

.PHONY: lint-full
lint-full: ## Run linters on all files (alias of lint)
	@$(MAKE) lint

.PHONY: clean-mocks
clean-mocks: ## Clean old mocks
	@echo "$(RED)▶$(RESET) Cleaning old mocks..."
	@rm -rf $$(find . -type d -name mocks)

.PHONY: gen-mocks
gen-mocks: clean-mocks ## Generate mocks
	@echo "$(GREEN)▶$(RESET) Generating mocks..."
	@go tool mockery --log-level="error" || (echo "$(RED)✖\033[0m Mockery failed" && exit 1)

.PHONY: gen-swagger
gen-swagger: ## Generate swagger docs
	@echo "$(GREEN)▶$(RESET) Generating swagger..."
	@go tool swag fmt
	@go tool swag init --quiet --parseDependency --parseInternal -g cmd/app/main.go

.PHONY: format
format: ## Run code formatting
	@echo "$(GREEN)▶$(RESET) Running code formatting..."
	@go tool gofumpt -l -w -extra .

.PHONY: fix
fix: ## Run go fix
	@echo "$(GREEN)▶$(RESET) Running code fixing..."
	@go fix ./...

.PHONY: generate
generate: ## Code generation
	@echo "$(GREEN)▶$(RESET) Code generation..."
	@$(MAKE) gen-mocks gen-swagger fix format
	@echo "$(GREEN)▶$(RESET) Code generation completed successfully"

.PHONY: test
test: ## Run tests with coverage
	@echo "$(GREEN)▶$(RESET) Running tests with coverage..."
	@go test -race -count=1 ./... -coverprofile coverage.out.tmp
	@grep -vE "mock.go" coverage.out.tmp > coverage.out
	@rm -f coverage.out.tmp
	@go tool cover -func coverage.out | grep total: | awk '{print "Test coverage percent: " $$3}'
	@rm -f coverage.out

.PHONY: docker-build
docker-build: ## Build docker image
	@echo "$(GREEN)▶$(RESET) Building docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) .

.PHONY: docker-run
docker-run: ## Run the webhook container in background (recreates existing)
	@echo "$(GREEN)▶$(RESET) Running container $(DOCKER_CONTAINER)..."
	@docker rm -f $(DOCKER_CONTAINER) 2>/dev/null || true
	docker run -d \
		--name $(DOCKER_CONTAINER) \
		--restart unless-stopped \
		--stop-timeout $(DOCKER_STOP_TIMEOUT) \
		$(DOCKER_ENV_FLAGS) \
		-p 127.0.0.1:8080:8080 \
		$(DOCKER_IMAGE)

.PHONY: docker-stop
docker-stop: ## Stop the webhook container (container kept for logs)
	@echo "$(RED)▶$(RESET) Stopping container $(DOCKER_CONTAINER)..."
	@docker stop -t $(DOCKER_STOP_TIMEOUT) $(DOCKER_CONTAINER)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-14s\033[0m %s\n", $$1, $$2}'
