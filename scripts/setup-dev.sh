#!/bin/bash

# Development environment setup script for fly-mcp

set -e

echo "🚀 Setting up fly-mcp development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    print_status "Checking Go installation..."
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_status "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION is installed"
}

# Check if required tools are installed
check_tools() {
    print_status "Checking required tools..."
    
    # Check make
    if ! command -v make &> /dev/null; then
        print_warning "make is not installed. Please install make."
    else
        print_success "make is available"
    fi
    
    # Check git
    if ! command -v git &> /dev/null; then
        print_error "git is not installed. Please install git."
        exit 1
    else
        print_success "git is available"
    fi
    
    # Check curl
    if ! command -v curl &> /dev/null; then
        print_warning "curl is not installed. Some testing features may not work."
    else
        print_success "curl is available"
    fi
}

# Download dependencies
setup_dependencies() {
    print_status "Downloading Go dependencies..."
    go mod download
    go mod tidy
    print_success "Dependencies downloaded"
}

# Create .env file template
create_env_template() {
    print_status "Creating environment template..."
    
    if [ ! -f ".env.example" ]; then
        cat > .env.example << EOF
# Fly.io Configuration
# Get your API token from: https://fly.io/user/personal_access_tokens
FLY_MCP_FLY_API_TOKEN=fo1_your_token_here
FLY_MCP_FLY_ORGANIZATION=your-org-name

# Development Configuration
FLY_MCP_ENVIRONMENT=local
FLY_MCP_LOGGING_LEVEL=debug

# Server Configuration
FLY_MCP_SERVER_HOST=127.0.0.1
FLY_MCP_SERVER_PORT=8080
EOF
        print_success "Created .env.example file"
    else
        print_status ".env.example already exists"
    fi
    
    if [ ! -f ".env" ]; then
        cp .env.example .env
        print_warning "Created .env file from template. Please update with your actual values."
        print_status "Edit .env file and add your Fly.io API token and organization"
    else
        print_status ".env file already exists"
    fi
}

# Build the application
build_app() {
    print_status "Building the application..."
    make build
    print_success "Application built successfully"
}

# Run basic tests
run_tests() {
    print_status "Running tests..."
    if make test; then
        print_success "All tests passed"
    else
        print_warning "Some tests failed. This is expected if Fly.io credentials are not configured."
    fi
}

# Validate configuration
validate_config() {
    print_status "Validating configuration..."
    
    # Source .env file if it exists
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
    fi
    
    if [ -z "$FLY_MCP_FLY_API_TOKEN" ] || [ "$FLY_MCP_FLY_API_TOKEN" = "fo1_your_token_here" ]; then
        print_warning "Fly.io API token not configured. Configuration validation will fail."
        print_status "To get your API token:"
        print_status "1. Visit https://fly.io/user/personal_access_tokens"
        print_status "2. Create a new token"
        print_status "3. Update the FLY_MCP_FLY_API_TOKEN in .env file"
        return 0
    fi
    
    if ./dist/fly-mcp validate --config config.local.yaml; then
        print_success "Configuration is valid"
    else
        print_error "Configuration validation failed"
        return 1
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    print_success "Development environment setup complete!"
    echo ""
    print_status "Next steps:"
    echo "1. Update .env file with your Fly.io credentials"
    echo "2. Run 'make dev' to start the development server"
    echo "3. Test the server with 'curl http://localhost:8080/health'"
    echo ""
    print_status "Available commands:"
    echo "  make dev            - Run in development mode"
    echo "  make build          - Build the application"
    echo "  make test           - Run tests"
    echo "  make validate-config - Validate configuration"
    echo "  make help           - Show all available commands"
    echo ""
    print_status "Documentation:"
    echo "  README.md           - Getting started guide"
    echo "  docs/               - Detailed documentation"
    echo ""
}

# Main execution
main() {
    echo "🔧 fly-mcp Development Environment Setup"
    echo "========================================"
    echo ""
    
    check_go
    check_tools
    setup_dependencies
    create_env_template
    build_app
    run_tests
    validate_config
    print_next_steps
}

# Run main function
main "$@"
