.DEFAULT_GOAL := build

# Build variables
BINARY_NAME  := lango
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME   := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS      := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go parameters (CGO required for sqlite3/sqlite-vec)
GOCMD    := go
GOBUILD  := CGO_ENABLED=1 $(GOCMD) build
GOCLEAN  := $(GOCMD) clean
GOTEST   := CGO_ENABLED=1 $(GOCMD) test
GOMOD    := $(GOCMD) mod

# Docker
REGISTRY     ?=
COVERAGE_DIR := .coverage

# ─── Build & Install ─────────────────────────────────────────────────────────

## build: Build binary for current platform
build:
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/lango

## build-linux: Cross-compile for Linux amd64
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOCMD) build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/lango

## build-darwin: Cross-compile for macOS arm64
build-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GOCMD) build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/lango

## build-all: Build for all platforms
build-all: build-linux build-darwin

## install: Install binary to $GOPATH/bin
install:
	$(GOBUILD) $(LDFLAGS) -o $(shell $(GOCMD) env GOPATH)/bin/$(BINARY_NAME) ./cmd/lango

# ─── Development ─────────────────────────────────────────────────────────────

## dev: Build and run server locally
dev:
	$(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/lango && ./bin/$(BINARY_NAME) serve

## run: Run server from existing binary (skip build)
run:
	./bin/$(BINARY_NAME) serve

# ─── Testing ─────────────────────────────────────────────────────────────────

## test: Run tests with race detector and coverage
test:
	$(GOTEST) -v -race -cover ./...

## test-short: Run short tests only
test-short:
	$(GOTEST) -v -short ./...

## bench: Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

## coverage: Generate HTML coverage report
coverage:
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

# ─── Code Quality ────────────────────────────────────────────────────────────

## fmt: Format code
fmt:
	gofmt -w -s .

## fmt-check: Check code formatting (CI, no modification)
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## lint: Run golangci-lint (auto-installs if missing)
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## generate: Run go generate
generate:
	$(GOCMD) generate ./...

## ci: Run full local CI pipeline (fmt-check → vet → lint → test)
ci: fmt-check vet lint test

# ─── Dependencies ────────────────────────────────────────────────────────────

## deps: Download and tidy dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# ─── Docker Build ────────────────────────────────────────────────────────────

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

## docker-push: Push image to registry (requires REGISTRY variable)
docker-push:
	@test -n "$(REGISTRY)" || (echo "Error: REGISTRY is not set. Usage: make docker-push REGISTRY=your.registry.io"; exit 1)
	docker tag $(BINARY_NAME):$(VERSION) $(REGISTRY)/$(BINARY_NAME):$(VERSION)
	docker tag $(BINARY_NAME):latest $(REGISTRY)/$(BINARY_NAME):latest
	docker push $(REGISTRY)/$(BINARY_NAME):$(VERSION)
	docker push $(REGISTRY)/$(BINARY_NAME):latest

# ─── Docker Compose ──────────────────────────────────────────────────────────

## docker-up: Start containers
docker-up:
	docker compose up -d

## docker-down: Stop containers
docker-down:
	docker compose down

## docker-logs: Tail container logs
docker-logs:
	docker compose logs -f

# ─── Utility ─────────────────────────────────────────────────────────────────

## health: Check running server health
health:
	@curl -sf http://localhost:18789/health && echo "" || echo "Server is not responding"

## clean: Remove build artifacts and coverage reports
clean:
	$(GOCLEAN)
	rm -rf bin/ $(COVERAGE_DIR)/

## help: Show available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | column -t -s ':'

.PHONY: build build-linux build-darwin build-all install \
        dev run \
        test test-short bench coverage \
        fmt fmt-check vet lint generate ci \
        deps \
        docker-build docker-push \
        docker-up docker-down docker-logs \
        health clean help
