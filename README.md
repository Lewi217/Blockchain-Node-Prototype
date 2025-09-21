# Blockchain Node Prototype

A lightweight blockchain implementation written in Go, featuring a complete blockchain node and cryptocurrency wallet functionality.

## Features

### Blockchain Core
- ✅ **Block Structure**: Complete block headers with merkle roots, timestamps, and proof-of-work
- ✅ **Transaction System**: Support for regular transactions and coinbase (mining reward) transactions
- ✅ **Proof of Work**: Simple mining algorithm with adjustable difficulty
- ✅ **UTXO Model**: Unspent Transaction Output tracking for balance calculation
- ✅ **Chain Validation**: Full blockchain and transaction validation
- ✅ **Persistence**: Pluggable storage system (Memory and LevelDB support)

### Cryptography
- ✅ **Digital Signatures**: ECDSA signature creation and verification
- ✅ **Key Generation**: Public/private key pair generation
- ✅ **Address Creation**: Bitcoin-style address generation from public keys
- ✅ **Hash Functions**: SHA-256, RIPEMD-160, and Hash160 implementations

### Wallet Functionality
- ✅ **Wallet Management**: Create, save, and load wallets
- ✅ **Balance Tracking**: Real-time balance calculation from UTXO set
- ✅ **Transaction Creation**: Build and sign transactions
- ✅ **Key Storage**: Secure private key storage and management

### Network & API
- ✅ **REST API**: Complete HTTP API for blockchain interaction
- ✅ **Node Operations**: Mining, transaction broadcasting, block retrieval
- ✅ **Wallet Operations**: Balance queries, transaction sending
- ✅ **Real-time Updates**: Auto-mining and balance updates

## Project Structure

```
blockchain-node/
├── cmd/
│   ├── node/          # Blockchain node entry point
│   ├── wallet/        # Wallet CLI entry point
│   └── miner/         # Mining node entry point
├── pkg/
│   ├── blockchain/    # Core blockchain logic
│   ├── crypto/        # Cryptographic functions
│   ├── wallet/        # Wallet functionality
│   ├── storage/       # Storage implementations
│   └── utils/         # Utility functions
├── internal/
│   └── api/           # Internal API handlers and middleware
├── build/             # Compiled binaries
├── go.mod            # Go module definition
├── Makefile          # Build automation
└── README.md         # This file
```

## Quick Start

### Prerequisites
- Go 1.21 or later
- Make (optional, but recommended)

### Installation

1. **Clone the repository:**
```bash
git clone <repository-url>
cd blockchain-node
```

2. **Install dependencies:**
```bash
make deps
# or manually:
go mod tidy
```

3. **Build the project:**
```bash
make build
# This creates binaries in ./build/ directory
```

### Running the Blockchain Node

```bash
# Start the blockchain node
make run-node
# or directly:
./build/blockchain-node
```

The node will start on `http://localhost:8080` with the following endpoints:
- `GET /api/v1/info` - Node information
- `GET /api/v1/blocks` - All blocks
- `GET /api/v1/blocks/latest` - Latest block
- `POST /api/v1/blocks/mine` - Mine a new block
- `POST /api/v1/transactions` - Create transaction
- `GET /api/v1/wallet/balance/{address}` - Get balance

### Using the Wallet CLI

```bash
# Create a new wallet
make create-wallet
# or:
./build/blockchain-wallet create my-wallet.wallet

# Check wallet balance
./build/blockchain-wallet balance <address>

# Send transaction
./build/blockchain-wallet send <from_address> <to_address> <amount>

# List wallet files
./build/blockchain-wallet list

# Get help
./build/blockchain-wallet help
```

### Using the Mining Node

```bash
# Start dedicated miner
./build/blockchain-miner

# Custom mining options
./build/blockchain-miner -wallet miner.wallet -interval 15s

# Show mining statistics
./build/blockchain-miner -stats
```

## Usage Examples

### 1. Basic Blockchain Operations

```bash
# Terminal 1: Start the node
make run-node

# Terminal 2: Interact with the blockchain
# Get node info
curl http://localhost:8080/api/v1/info

# Mine a block
curl -X POST http://localhost:8080/api/v1/blocks/mine

# Get all blocks
curl http://localhost:8080/api/v1/blocks
```

### 2. Wallet Operations

```bash
# Create two wallets
./build/blockchain-wallet create alice.wallet
./build/blockchain-wallet create bob.wallet

# Get wallet info (shows address)
./build/blockchain-wallet info alice.wallet

# Check balance (will be 0 initially)
./build/blockchain-wallet balance <alice_address>

# Send coins (after mining some blocks to Alice's address)
./build/blockchain-wallet send <alice_address> <bob_address> 1000000
```

### 3. Mining and Transactions

```bash
# Mine a block (rewards go to node's wallet)
make mine

# Create a custom transaction via API
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "from": "source_address",
    "to": "destination_address", 
    "amount": 1000000
  }'
```

## API Reference

### Node Information
```bash
GET /api/v1/info
```
Returns current blockchain height, difficulty, and latest block hash.

### Blocks
```bash
GET /api/v1/blocks                    # Get all blocks
GET /api/v1/blocks/{height}           # Get block by height
GET /api/v1/blocks/latest             # Get latest block
POST /api/v1/blocks/mine              # Mine a new block
```

### Transactions
```bash
POST /api/v1/transactions             # Create new transaction
GET /api/v1/transactions/{txid}       # Get transaction by ID
```

### Wallet
```bash
GET /api/v1/wallet/balance/{address}  # Get address balance
POST /api/v1/wallet/new               # Create new wallet
```

### Mining
```bash
POST /api/v1/mining/mine              # Mine a single block
GET /api/v1/mining/stats              # Get mining statistics
GET /api/v1/mining/info               # Get mining information
```

## Configuration

### Node Configuration
The node accepts the following environment variables:
- `PORT`: Server port (default: 8080)
- `DIFFICULTY`: Mining difficulty (default: 4)
- `BLOCK_TIME`: Target block time in seconds (default: 30)

### Wallet Configuration
The wallet CLI accepts:
- `BLOCKCHAIN_NODE_URL`: Node URL (default: http://localhost:8080)

### Miner Configuration
The miner accepts command-line options:
- `-node <url>`: Node URL (default: http://localhost:8080)
- `-wallet <file>`: Wallet file (default: miner.wallet)
- `-interval <time>`: Mining interval (default: 10s)

## Development

### Building
```bash
make build          # Build all binaries
make build-node     # Build node only
make build-wallet   # Build wallet only
make build-miner    # Build miner only
```

### Code Quality
```bash
make fmt            # Format code
make lint           # Run linter (requires golangci-lint)
```

### Testing
```bash
go test ./...       # Run all tests
```

## Architecture

### Blockchain Components
1. **Block**: Contains header (metadata) and transactions
2. **Transaction**: Inputs (spending) and outputs (receiving)
3. **UTXO Set**: Tracks unspent transaction outputs
4. **Proof of Work**: Simple hash-based mining algorithm

### Storage Layer
- **Interface**: Pluggable storage system
- **Memory Storage**: Fast in-memory storage for development
- **LevelDB Storage**: Persistent storage for production (planned)

### Cryptographic Security
- **ECDSA**: Elliptic Curve Digital Signature Algorithm
- **SHA-256**: Primary hash function
- **RIPEMD-160**: Used in address generation
- **Hash160**: SHA-256 + RIPEMD-160 for Bitcoin-style addresses

## Security Considerations

⚠️ **This is a prototype implementation for educational purposes. Do not use in production without proper security auditing.**

Current limitations:
- Simplified key management
- No network security protocols
- Basic transaction validation
- Limited DoS protection

## Roadmap

### Phase 1 (Current)
- [x] Basic blockchain functionality
- [x] Wallet operations
- [x] REST API
- [x] Proof of Work mining

### Phase 2 (Planned)
- [ ] P2P networking
- [ ] Advanced transaction types
- [ ] Improved security
- [ ] Performance optimizations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Bitcoin whitepaper for blockchain concepts
- Go community for excellent libraries
- Educational resources on cryptocurrency development
