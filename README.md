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

## 🛠️ Available MCP Tools

### Core Tools

| Tool | Description | Example Usage |
|------|-------------|---------------|
| `ping` | Test connectivity and server response | `{"name": "ping", "arguments": {"message": "Hello!"}}` |

### Fly.io Management Tools

| Tool | Description | Example Usage |
|------|-------------|---------------|
| `fly_list_apps` | List all applications with filtering | `{"name": "fly_list_apps", "arguments": {"status_filter": "running"}}` |
| `fly_app_info` | Get detailed application information | `{"name": "fly_app_info", "arguments": {"app_name": "my-app"}}` |
| `fly_status` | Real-time application and machine status | `{"name": "fly_status", "arguments": {"app_name": "my-app"}}` |
| `fly_restart` | Restart applications with confirmation | `{"name": "fly_restart", "arguments": {"app_name": "my-app", "confirm": true}}` |
| `fly_scale` | Scaling status and recommendations | `{"name": "fly_scale", "arguments": {"app_name": "my-app", "action": "status"}}` |

### Tool Features

- **🔒 Security**: All tools require proper authentication and permissions
- **📝 Audit Logging**: All operations are logged for compliance and debugging
- **⚡ Real-time**: Status and machine information is fetched in real-time
- **🛡️ Safety**: Destructive operations require explicit confirmation
- **📊 Rich Output**: Human-readable responses with actionable recommendations

## 🧪 Testing the MCP Server

### Automated Testing

Use the provided test script to verify all tools:

```bash
# Start the server first
make dev

# In another terminal, run tests
./scripts/test-mcp-tools.sh
```

### Manual Testing

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

3. **List Available Tools**
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "id": 2,
       "method": "tools/list"
     }'
   ```

4. **Test Ping Tool**
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "id": 3,
       "method": "tools/call",
       "params": {
         "name": "ping",
         "arguments": {"message": "Hello from fly-mcp!"}
       }
     }'
   ```

5. **Test Fly.io Tools** (requires valid credentials)
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{
       "jsonrpc": "2.0",
       "id": 4,
       "method": "tools/call",
       "params": {
         "name": "fly_list_apps",
         "arguments": {}
       }
     }'
   ```

## 🎯 Current Status

**Phase 2 Complete**: Fly.io API Integration & Core Tools

### ✅ Implemented Features

- ✅ **Project structure and build system**
- ✅ **Configuration management** (local/production environments)
- ✅ **HTTP server** with middleware (CORS, rate limiting, logging)
- ✅ **MCP protocol handler** with full request/response handling
- ✅ **Structured logging** with audit trails and security events
- ✅ **Fly.io API integration** (hybrid approach: fly-go + Machines API)
- ✅ **Authentication & authorization** with permissions and audit logging
- ✅ **Core MCP tools**:
  - `ping` - Test tool for connectivity
  - `fly_list_apps` - List all applications with filtering
  - `fly_app_info` - Get detailed application information
  - `fly_status` - Real-time application and machine status
  - `fly_restart` - Restart applications with confirmation
  - `fly_scale` - Scaling status and recommendations
- ✅ **Health checks and metrics** endpoints
- ✅ **Comprehensive error handling** and validation
- ✅ **Security features** (rate limiting, CORS, audit logging)

### 🔄 Coming Next (Phase 3)

- 🔄 **Additional tools**: logs, secrets, volumes, certificates
- 🔄 **Deploy tool** for application deployment
- 🔄 **Advanced scaling** with auto-scaling recommendations
- 🔄 **Monitoring integration** with alerts and dashboards
- 🔄 **CI/CD pipeline** and automated testing
- 🔄 **Documentation** and usage examples

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📞 Support

- GitHub Issues: [Report bugs or request features](https://github.com/brannn/fly-mcp/issues)
- Documentation: [Full documentation](docs/)
- Community: [Discussions](https://github.com/brannn/fly-mcp/discussions)
