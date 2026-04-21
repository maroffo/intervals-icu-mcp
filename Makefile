# ABOUTME: Quality gates + build targets for intervals-icu-mcp
# ABOUTME: make check (vet + test), make build, make install, make integration, make lint

BINARY := intervals-icu-mcp
GOBIN  ?= $(shell go env GOPATH)/bin

.PHONY: help check vet test test-e2e integration lint build install clean

.DEFAULT_GOAL := help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check: vet test ## Run vet + unit tests (default CI gate)

vet: ## Run `go vet` across all packages
	go vet ./...

test: ## Run unit tests with fresh cache
	go test -count=1 ./...

integration: ## Run integration tests (requires INTERVALS_API_KEY)
	go test -count=1 -tags=integration ./internal/icu/ -run Integration -v

test-e2e: ## End-to-end tests (integration build tag; skips gracefully without INTERVALS_API_KEY)
	go test -count=1 -tags=integration ./internal/icu/ -run Integration

lint: ## Run staticcheck (fails if not installed)
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed; run: go install honnef.co/go/tools/cmd/staticcheck@latest" >&2; \
		exit 1; \
	fi

build: ## Build the server binary to bin/$(BINARY)
	go build -o bin/$(BINARY) .

install: ## Install the server to $(GOBIN)
	go install .
	@echo "Installed to $(GOBIN)/$(BINARY)"

clean: ## Remove build artifacts and test cache
	rm -rf bin/
	go clean -testcache
