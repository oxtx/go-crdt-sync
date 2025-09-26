# Makefile for github.com/oxtx/go-crdt-sync

# Config
BIN_NAME := server
BIN_DIR  := bin
PKG_MAIN := ./cmd/server
ADDR     := :8080
GO       := go

# Read module path (for info)
MOD := $(shell awk '/^module /{print $$2}' go.mod 2>/dev/null)

.PHONY: all tidy deps build run run-bg stop test testv vet fmt fmt-check lint clean docker-image docker-run

all: build

# Resolve and tidy dependencies
tidy:
	$(GO) mod tidy

deps: tidy

# Build server binary
build: deps
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(BIN_NAME) $(PKG_MAIN)
	@echo "Built $(BIN_DIR)/$(BIN_NAME) for module $(MOD)"

# Run foreground
run: build
	ADDR=$(ADDR) $(BIN_DIR)/$(BIN_NAME)

# Run background with logs and PID
run-bg: build
	@mkdir -p $(BIN_DIR)
	ADDR=$(ADDR) nohup $(BIN_DIR)/$(BIN_NAME) > $(BIN_DIR)/server.out 2>&1 & echo $$! > .server.pid
	@echo "Server started on $(ADDR) (pid $$(cat .server.pid)). Logs: $(BIN_DIR)/server.out"

# Stop background server
stop:
	@if [ -f .server.pid ]; then \
		kill $$(cat .server.pid) || true; \
		rm -f .server.pid; \
		echo "Server stopped."; \
	else \
		echo "No .server.pid found."; \
	fi

# Tests
test:
	$(GO) test ./...

testv:
	$(GO) test -v ./...

# Static analysis
vet:
	$(GO) vet ./...

lint: vet

# Formatting
fmt:
	$(GO) fmt ./...

fmt-check:
	@diff -u <(echo -n) <($(GO) fmt ./... | tee /dev/stderr) || true

# Clean artifacts
clean:
	@rm -rf $(BIN_DIR) .server.pid
	@echo "Cleaned."

# Optional Docker targets (if you add a Dockerfile)
docker-image:
	@docker build -t oxtx/go-crdt-sync:latest .

docker-run:
	@docker run --rm -p 8080:8080 -e ADDR=$(ADDR) oxtx/go-crdt-sync:latest
