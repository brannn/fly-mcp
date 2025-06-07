#!/bin/bash

# Test script for MCP tools functionality

set -e

echo "🧪 Testing fly-mcp MCP tools..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Check if server is running
SERVER_URL="http://localhost:8080"

# Function to test MCP endpoint
test_mcp_request() {
    local method="$1"
    local params="$2"
    local description="$3"
    
    print_status "Testing: $description"
    
    local request_body=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "$method"
}
EOF
)
    
    if [ -n "$params" ]; then
        request_body=$(cat <<EOF
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "$method",
    "params": $params
}
EOF
)
    fi
    
    local response=$(curl -s -X POST "$SERVER_URL/mcp" \
        -H "Content-Type: application/json" \
        -d "$request_body" 2>/dev/null || echo "ERROR")
    
    if [ "$response" = "ERROR" ]; then
        print_error "$description - Server not responding"
        return 1
    fi
    
    # Check if response contains error
    if echo "$response" | grep -q '"error"'; then
        print_warning "$description - Returned error (expected with dummy credentials)"
        echo "Response: $response" | head -c 200
        echo "..."
        return 0
    fi
    
    # Check if response contains result
    if echo "$response" | grep -q '"result"'; then
        print_success "$description - Success"
        return 0
    fi
    
    print_warning "$description - Unexpected response"
    echo "Response: $response" | head -c 200
    echo "..."
    return 0
}

# Function to check if server is running
check_server() {
    print_status "Checking if server is running..."
    
    local health_response=$(curl -s "$SERVER_URL/health" 2>/dev/null || echo "ERROR")
    
    if [ "$health_response" = "ERROR" ]; then
        print_error "Server is not running at $SERVER_URL"
        echo ""
        echo "To start the server:"
        echo "1. Set environment variables:"
        echo "   export FLY_MCP_FLY_API_TOKEN=fo1_your_token_here"
        echo "   export FLY_MCP_FLY_ORGANIZATION=your-org-here"
        echo "2. Run: make dev"
        echo ""
        return 1
    fi
    
    print_success "Server is running"
    echo "Health response: $health_response"
    return 0
}

# Main test execution
main() {
    echo "🚀 fly-mcp MCP Tools Test Suite"
    echo "================================"
    echo ""
    
    # Check server
    if ! check_server; then
        exit 1
    fi
    
    echo ""
    print_status "Testing MCP protocol endpoints..."
    echo ""
    
    # Test initialize
    test_mcp_request "initialize" '{"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}' "MCP Initialize"
    
    echo ""
    
    # Test tools list
    test_mcp_request "tools/list" "" "List Available Tools"
    
    echo ""
    
    # Test ping tool
    test_mcp_request "tools/call" '{"name": "ping", "arguments": {"message": "Hello from test!"}}' "Ping Tool"
    
    echo ""
    
    # Test fly tools (these will fail with dummy credentials, but should show the tools exist)
    print_status "Testing Fly.io tools (expected to fail with authentication errors)..."
    
    test_mcp_request "tools/call" '{"name": "fly_list_apps", "arguments": {}}' "List Apps Tool"
    
    test_mcp_request "tools/call" '{"name": "fly_app_info", "arguments": {"app_name": "test-app"}}' "App Info Tool"
    
    test_mcp_request "tools/call" '{"name": "fly_status", "arguments": {"app_name": "test-app"}}' "App Status Tool"
    
    test_mcp_request "tools/call" '{"name": "fly_scale", "arguments": {"app_name": "test-app", "action": "status"}}' "App Scale Tool"
    
    test_mcp_request "tools/call" '{"name": "fly_restart", "arguments": {"app_name": "test-app", "confirm": false}}' "App Restart Tool (without confirmation)"
    
    echo ""
    echo "🎉 Test suite completed!"
    echo ""
    echo "📝 Notes:"
    echo "- Ping tool should work without credentials"
    echo "- Fly.io tools should fail with authentication errors (expected with dummy credentials)"
    echo "- All tools should be registered and callable"
    echo ""
    echo "🔧 To test with real credentials:"
    echo "1. Set your real Fly.io API token: export FLY_MCP_FLY_API_TOKEN=fo1_your_real_token"
    echo "2. Set your organization: export FLY_MCP_FLY_ORGANIZATION=your-real-org"
    echo "3. Restart the server: make dev"
    echo "4. Run this test again"
}

# Run main function
main "$@"
