.DEFAULT_GOAL := help
SHELL := /bin/bash

# Docker Compose file used for local development
COMPOSE_FILE := deployments/docker-compose/docker-compose.yml

# Helper: list Go workspace modules with POSIX-style paths
GO_MODULE_DIRS := go list -m -f '{{.Dir}}' | sed 's@\\\\@/@g'

.PHONY: help
help: ## Show this help message
	@echo "PlantX Kit Platform - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

## -------------------- Build --------------------

.PHONY: build
build: build-go build-web ## Build all Go modules and frontend packages

.PHONY: build-go
build-go: ## Build all Go workspace modules
	@set -e; $(GO_MODULE_DIRS) | while read dir; do \
		echo "==> building $$dir"; \
		(cd "$$dir" && GOTOOLCHAIN=local go build ./...); \
	done

.PHONY: build-web
build-web: ## Build all pnpm workspace frontend packages
	pnpm -r run build

## -------------------- Test / Lint --------------------

.PHONY: test
test: test-go ## Run all tests

.PHONY: test-go
test-go: ## Run Go tests for all workspace modules
	@set -e; $(GO_MODULE_DIRS) | while read dir; do \
		echo "==> testing $$dir"; \
		cd "$$dir"; \
		if [ -n "$$(GOTOOLCHAIN=local go list ./... 2>/dev/null)" ]; then \
			GOTOOLCHAIN=local go test ./...; \
		else \
			echo "    (no packages to test)"; \
		fi; \
	done

.PHONY: lint
lint: lint-go lint-web ## Lint Go and frontend code

.PHONY: lint-go
lint-go: ## Run go vet for all workspace modules
	@set -e; $(GO_MODULE_DIRS) | while read dir; do \
		echo "==> vetting $$dir"; \
		(cd "$$dir" && GOTOOLCHAIN=local go vet ./...); \
	done

.PHONY: lint-web
lint-web: ## Run pnpm lint across frontend packages
	pnpm -r run lint

.PHONY: typecheck-web
typecheck-web: ## Type-check all frontend packages
	pnpm -r run typecheck

.PHONY: fmt-go
fmt-go: ## Format all Go workspace modules
	@set -e; $(GO_MODULE_DIRS) | while read dir; do \
		echo "==> formatting $$dir"; \
		(cd "$$dir" && GOTOOLCHAIN=local gofmt -w .); \
	done

## -------------------- Code Generation --------------------

.PHONY: generate
generate: generate-kit generate-order generate-test ## Generate code for kit, order and test service

.PHONY: generate-kit
generate-kit: ## Generate kit proto code and sqlc models (requires sqlc)
	cd kit/kit-cli && go run . generate

.PHONY: generate-order
generate-order: ## Generate Go + grpc-gateway code from services/order/api/order.proto
	cd services/order && \
	protoc \
		--proto_path=../../proto \
		--proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		api/order.proto

.PHONY: generate-test
generate-test: ## Generate Go + grpc-gateway code from services/test-service/api/test.proto
	cd services/test-service && \
	protoc \
		--proto_path=../../proto \
		--proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		api/test.proto

## -------------------- Docker / Local Stack --------------------

.PHONY: up
up: ## Start the full Docker Compose stack
	docker-compose -f $(COMPOSE_FILE) up -d --build

.PHONY: down
down: ## Stop the Docker Compose stack
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans

.PHONY: logs
logs: ## Tail Docker Compose logs
	docker-compose -f $(COMPOSE_FILE) logs -f

.PHONY: ps
ps: ## Show running Docker Compose services
	docker-compose -f $(COMPOSE_FILE) ps

## -------------------- E2E / Verification --------------------

.PHONY: e2e
e2e: ## Run the end-to-end smoke test against the local stack
	bash scripts/e2e-smoke-test.sh

.PHONY: health
health: ## Check gateway health and readiness
	curl -fsS http://localhost/health && echo
	curl -fsS http://localhost/ready && echo

## -------------------- Maintenance --------------------

.PHONY: tidy
tidy: ## Run go mod tidy for all workspace modules
	@set -e; $(GO_MODULE_DIRS) | while read dir; do \
		echo "==> tidying $$dir"; \
		(cd "$$dir" && GOTOOLCHAIN=local go mod tidy); \
	done

.PHONY: clean
clean: ## Stop stack and remove frontend build artifacts
	docker-compose -f $(COMPOSE_FILE) down --remove-orphans || true
	rm -rf apps/portal/dist services/order/web/order-ui/dist services/test-service/web/test-ui/dist
