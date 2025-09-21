# Blockchain Node Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
NODE_BINARY=blockchain-node
WALLET_BINARY=blockchain-wallet

# Build directories
BUILD_DIR=build

# Default target
.PHONY: all
all: build

# Build all binaries
.PHONY: build
build: build-node build-wallet

# Build node binary
.PHONY: build-node
build-node:
	@echo "Building blockchain node..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(NODE_BINARY) ./cmd/node

# Build wallet binary
.PHONY: build-wallet
build-wallet:
	@echo "Building blockchain wallet..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(WALLET_BINARY) ./cmd/wallet

# Run node
.PHONY: run-node
run-node: build-node
	@echo "Starting blockchain node..."
	./$(BUILD_DIR)/$(NODE_BINARY)

# Run wallet CLI
.PHONY: run-wallet
run-wallet: build-wallet
	@echo "Running wallet CLI..."
	./$(BUILD_DIR)/$(WALLET_BINARY) $(ARGS)

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u github.com/gorilla/mux
	$(GOGET) -u github.com/syndtr/goleveldb/leveldb
	$(GOGET) -u github.com/gorilla/websocket
	$(GOGET) -u golang.org/x/crypto

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f *.wallet

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Creating sample configuration files..."
	@mkdir -p configs
	@echo "Development setup complete!"

# Quick demo
.PHONY: demo
demo: build-node build-wallet
	@echo "Starting demo..."
	@echo "Starting node in background..."
	@./$(BUILD_DIR)/$(NODE_BINARY) &
	@NODE_PID=$!; \
	sleep 3; \
	echo "Creating demo wallet..."; \
	./$(BUILD_DIR)/$(WALLET_BINARY) create demo.wallet; \
	echo "Demo complete. Node running in background (PID: $NODE_PID)"; \
	echo "Use 'kill $NODE_PID' to stop the node"

# Stop any running processes
.PHONY: stop
stop:
	@echo "Stopping blockchain processes..."
	@pkill -f $(NODE_BINARY) || true

# Create example wallet
.PHONY: create-wallet
create-wallet: build-wallet
	@echo "Creating new wallet..."
	./$(BUILD_DIR)/$(WALLET_BINARY) create example.wallet

# Show wallet balance (requires ADDRESS variable)
.PHONY: balance
balance: build-wallet
	@if [ -z "$(ADDRESS)" ]; then \
		echo "Error: ADDRESS variable required"; \
		echo "Usage: make balance ADDRESS=<wallet_address>"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(WALLET_BINARY) balance $(ADDRESS)

# Send transaction (requires FROM, TO, AMOUNT variables)
.PHONY: send
send: build-wallet
	@if [ -z "$(FROM)" ] || [ -z "$(TO)" ] || [ -z "$(AMOUNT)" ]; then \
		echo "Error: FROM, TO, and AMOUNT variables required"; \
		echo "Usage: make send FROM=<from_address> TO=<to_address> AMOUNT=<amount>"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(WALLET_BINARY) send $(FROM) $(TO) $(AMOUNT)

# Mine a block (requires running node)
.PHONY: mine
mine:
	@echo "Mining a new block..."
	@curl -X POST http://localhost:8080/api/v1/blocks/mine || echo "Make sure the node is running"

# Get node info
.PHONY: node-info
node-info:
	@echo "Getting node information..."
	@curl -s http://localhost:8080/api/v1/info | python3 -m json.tool || echo "Make sure the node is running"

# Get all blocks
.PHONY: blocks
blocks:
	@echo "Getting all blocks..."
	@curl -s http://localhost:8080/api/v1/blocks | python3 -m json.tool || echo "Make sure the node is running"

# Get latest block
.PHONY: latest-block
latest-block:
	@echo "Getting latest block..."
	@curl -s http://localhost:8080/api/v1/blocks/latest | python3 -m json.tool || echo "Make sure the node is running"

# Help target
.PHONY: help
help:
	@echo "Blockchain Node - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build all binaries"
	@echo "  build-node     - Build node binary only"
	@echo "  build-wallet   - Build wallet binary only"
	@echo ""
	@echo "Run targets:"
	@echo "  run-node       - Build and run the blockchain node"
	@echo "  run-wallet     - Build and run wallet CLI (use ARGS for arguments)"
	@echo "  demo           - Quick demo setup"
	@echo ""
	@echo "Development targets:"
	@echo "  deps           - Install dependencies"
	@echo "  dev-setup      - Setup development environment"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Wallet operations:"
	@echo "  create-wallet  - Create a new wallet"
	@echo "  balance        - Get wallet balance (ADDRESS=<address>)"
	@echo "  send           - Send transaction (FROM=<addr> TO=<addr> AMOUNT=<amount>)"
	@echo ""
	@echo "Node operations:"
	@echo "  mine           - Mine a new block"
	@echo "  node-info      - Get node information"
	@echo "  blocks         - Get all blocks"
	@echo "  latest-block   - Get latest block"
	@echo "  stop           - Stop running processes"
	@echo ""
	@echo "Examples:"
	@echo "  make run-node"
	@echo "  make create-wallet"
	@echo "  make balance ADDRESS=1A2B3C..."
	@echo "  make send FROM=1A2B... TO=1X2Y... AMOUNT=1000000"
	@echo "  make mine"