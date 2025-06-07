# fly-mcp

An open-source MCP (Model Context Protocol) server for Fly.io infrastructure management, enabling AI-driven DevOps workflows through natural language interactions.

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Fly.io account and API token
- Git

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/brannn/fly-mcp.git
   cd fly-mcp
   ```

2. **Set up environment variables**
   ```bash
   export FLY_MCP_FLY_API_TOKEN="your_fly_api_token_here"
   export FLY_MCP_FLY_ORGANIZATION="your_fly_org_here"
   ```

3. **Build and run**
   ```bash
   make build
   make run
   ```

   Or for development with hot reload:
   ```bash
   make dev
   ```

### Production Deployment on Fly.io

Deploy using Fly.io's MCP infrastructure:

```bash
fly mcp launch \
  "github.com/brannn/fly-mcp" \
  --claude --cursor --zed \
  --server fly-infrastructure \
  --secret FLY_API_TOKEN=fo1_your_token \
  --secret FLY_ORG=your-org-name
```

## 🏗️ Architecture

### Project Structure

```
fly-mcp/
├── cmd/fly-mcp/              # Main application entry point
├── pkg/
│   ├── mcp/                  # MCP protocol implementation
│   ├── fly/                  # Fly.io API client (coming soon)
│   ├── auth/                 # Authentication (coming soon)
│   ├── tools/                # MCP tool implementations (coming soon)
│   └── config/               # Configuration management
├── internal/
│   ├── server/               # HTTP server implementation
│   ├── security/             # Security utilities (coming soon)
│   └── logger/               # Structured logging
├── config.local.yaml         # Local development configuration
├── config.production.yaml    # Production configuration
└── Makefile                  # Build automation
```

### Configuration

The application supports flexible configuration through:

- **YAML files**: `config.local.yaml` for development, `config.production.yaml` for production
- **Environment variables**: All config values can be overridden with `FLY_MCP_` prefixed env vars
- **Command line flags**: `--config` and `--log-level` flags

#### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `FLY_MCP_FLY_API_TOKEN` | Fly.io API token | Yes |
| `FLY_MCP_FLY_ORGANIZATION` | Fly.io organization name | Yes |
| `FLY_MCP_ENVIRONMENT` | Environment (local/production) | No |
| `FLY_MCP_LOGGING_LEVEL` | Log level (debug/info/warn/error) | No |

## 🛠️ Development

### Available Make Targets

```bash
make build          # Build the binary
make build-all      # Build for all platforms
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linters
make fmt            # Format code
make clean          # Clean build artifacts
make dev            # Run in development mode
make validate-config # Validate configuration
make docker-build   # Build Docker image
make help           # Show all available targets
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run all checks
make check
```

## 🔧 Configuration Examples

### Local Development

Create a `.env` file or set environment variables:

```bash
export FLY_MCP_FLY_API_TOKEN="fo1_your_development_token"
export FLY_MCP_FLY_ORGANIZATION="your-dev-org"
export FLY_MCP_LOGGING_LEVEL="debug"
```

### Production on Fly.io

Set secrets in your Fly.io app:

```bash
fly secrets set FLY_API_TOKEN=fo1_your_production_token
fly secrets set FLY_ORG=your-production-org
```

## 🧪 Testing the MCP Server

Once running, you can test the MCP server:

1. **Health Check**
   ```bash
   curl http://localhost:8080/health
   ```

2. **MCP Initialize**
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "id": 1,
       "method": "initialize",
       "params": {
         "protocolVersion": "2024-11-05",
         "capabilities": {},
         "clientInfo": {"name": "test-client", "version": "1.0.0"}
       }
     }'
   ```

3. **List Tools**
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "id": 2,
       "method": "tools/list"
     }'
   ```

## 🎯 Current Status

This is the initial foundation setup. Currently implemented:

- ✅ Project structure and build system
- ✅ Configuration management (local/production)
- ✅ HTTP server with middleware
- ✅ Basic MCP protocol handler
- ✅ Structured logging
- ✅ Health checks and metrics endpoints
- ✅ Simple ping tool for testing

### Coming Next

- 🔄 Fly.io API client integration
- 🔄 Core Fly.io management tools (deploy, scale, logs, etc.)
- 🔄 Authentication and security
- 🔄 Comprehensive testing
- 🔄 Docker containerization
- 🔄 CI/CD pipeline

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📞 Support

- GitHub Issues: [Report bugs or request features](https://github.com/brannn/fly-mcp/issues)
- Documentation: [Full documentation](docs/)
- Community: [Discussions](https://github.com/brannn/fly-mcp/discussions)
