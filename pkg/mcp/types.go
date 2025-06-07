package mcp

import (
	"github.com/brannn/fly-mcp/pkg/interfaces"
)

// MCPRequest represents an MCP protocol request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP protocol response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Type aliases for interfaces package types
type Tool = interfaces.Tool
type ToolResult = interfaces.ToolResult
type ContentBlock = interfaces.ContentBlock

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptArgument represents an argument for a prompt template
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// ServerCapabilities represents the capabilities of the MCP server
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

// ToolsCapability represents tools-related capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources-related capabilities
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts-related capabilities
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeParams represents parameters for the initialize request
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities represents the capabilities of the MCP client
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability represents roots-related capabilities
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents sampling-related capabilities
type SamplingCapability struct{}

// ClientInfo represents information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ToolsListParams represents parameters for the tools/list request
type ToolsListParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ToolsCallParams represents parameters for the tools/call request
type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ResourcesListParams represents parameters for the resources/list request
type ResourcesListParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ResourcesReadParams represents parameters for the resources/read request
type ResourcesReadParams struct {
	URI string `json:"uri"`
}

// PromptsListParams represents parameters for the prompts/list request
type PromptsListParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// PromptsGetParams represents parameters for the prompts/get request
type PromptsGetParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}
