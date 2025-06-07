package interfaces

import (
	"context"
)

// Tool represents an MCP tool that can be executed
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a piece of content in a tool result
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"`
}
