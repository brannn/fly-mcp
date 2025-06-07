package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/interfaces"
)

// PingTool is a simple tool for testing MCP functionality
type PingTool struct {
	logger *logger.Logger
}

// Name returns the tool name
func (t *PingTool) Name() string {
	return "ping"
}

// Description returns the tool description
func (t *PingTool) Description() string {
	return "A simple ping tool that responds with pong and the current timestamp"
}

// InputSchema returns the JSON schema for the tool's input
func (t *PingTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Optional message to include in the response",
				"default":     "Hello from fly-mcp!",
			},
		},
		"additionalProperties": false,
	}
}

// Execute executes the ping tool
func (t *PingTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
	// Extract message from arguments
	message := "Hello from fly-mcp!"
	if msg, ok := args["message"].(string); ok && msg != "" {
		message = msg
	}
	
	// Create response
	response := fmt.Sprintf("Pong! %s\nTimestamp: %s", message, time.Now().UTC().Format(time.RFC3339))
	
	t.logger.Debug().
		Str("tool", "ping").
		Str("message", message).
		Msg("Ping tool executed")
	
	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{
			{
				Type: "text",
				Text: response,
			},
		},
		IsError: false,
	}, nil
}
