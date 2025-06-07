package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/auth"
	"github.com/brannn/fly-mcp/pkg/fly"
	"github.com/brannn/fly-mcp/pkg/interfaces"
)

// AppStatusTool implements the fly_status MCP tool
type AppStatusTool struct {
	flyClient   *fly.Client
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewAppStatusTool creates a new app status tool
func NewAppStatusTool(flyClient *fly.Client, authManager *auth.Manager, logger *logger.Logger) *AppStatusTool {
	return &AppStatusTool{
		flyClient:   flyClient,
		authManager: authManager,
		logger:      logger,
	}
}

// Name returns the tool name
func (t *AppStatusTool) Name() string {
	return "fly_status"
}

// Description returns the tool description
func (t *AppStatusTool) Description() string {
	return "Get real-time status information for a Fly.io application including machine states, health checks, and deployment status"
}

// InputSchema returns the JSON schema for the tool's input
func (t *AppStatusTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the application to check status for",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Response format (text or json)",
				"enum":        []string{"text", "json"},
				"default":     "text",
			},
			"detailed": map[string]interface{}{
				"type":        "boolean",
				"description": "Include detailed machine information",
				"default":     false,
			},
		},
		"required":             []string{"app_name"},
		"additionalProperties": false,
	}
}

// Execute executes the app status tool
func (t *AppStatusTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
	// Validate permissions
	if err := t.authManager.ValidateRequest(ctx, "read", "app"); err != nil {
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Permission denied: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Extract and validate arguments
	appName, ok := args["app_name"].(string)
	if !ok || appName == "" {
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: "Error: app_name is required and must be a non-empty string",
			}},
			IsError: true,
		}, nil
	}

	format := "text"
	if fmt, ok := args["format"].(string); ok {
		format = fmt
	}

	detailed := false
	if det, ok := args["detailed"].(bool); ok {
		detailed = det
	}

	// Log the operation
	userID, _ := t.authManager.ExtractUserFromContext(ctx)
	t.logger.Info().
		Str("user_id", userID).
		Str("tool", "fly_status").
		Str("app_name", appName).
		Str("format", format).
		Bool("detailed", detailed).
		Msg("Executing app status tool")

	// Get app status from Fly.io
	status, err := t.flyClient.GetAppStatus(ctx, appName)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "get_app_status", appName, "failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Failed to retrieve status for app '%s': %v", appName, err),
			}},
			IsError: true,
		}, nil
	}

	// Log successful operation
	t.authManager.AuditLog(ctx, userID, "get_app_status", appName, "success", map[string]interface{}{
		"format":        format,
		"detailed":      detailed,
		"machine_count": status.MachineCount,
		"status":        status.Status,
	})

	// Format response based on requested format
	if format == "json" {
		return t.formatJSONResponse(status)
	}
	
	return t.formatTextResponse(status, detailed)
}

// formatJSONResponse formats the response as JSON
func (t *AppStatusTool) formatJSONResponse(status *fly.AppStatus) (*interfaces.ToolResult, error) {
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Error formatting JSON response: %v", err),
			}},
			IsError: true,
		}, nil
	}

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: fmt.Sprintf("Status for application '%s':\n\n```json\n%s\n```", status.AppName, string(jsonData)),
		}},
	}, nil
}

// formatTextResponse formats the response as human-readable text
func (t *AppStatusTool) formatTextResponse(status *fly.AppStatus, detailed bool) (*interfaces.ToolResult, error) {
	var response string
	
	// Status header with emoji
	statusIcon := "🔴"
	statusColor := "stopped"
	switch status.Status {
	case "running":
		statusIcon = "🟢"
		statusColor = "running"
	case "suspended":
		statusIcon = "🟡"
		statusColor = "suspended"
	case "deployed":
		statusIcon = "🔵"
		statusColor = "deployed"
	}
	
	response += fmt.Sprintf("# %s Status: %s %s\n\n", status.AppName, statusIcon, statusColor)
	
	// Overview section
	response += "## Overview\n"
	response += fmt.Sprintf("- **Application**: %s\n", status.AppName)
	response += fmt.Sprintf("- **Status**: %s %s\n", statusIcon, status.Status)
	response += fmt.Sprintf("- **Deployed**: %t\n", status.Deployed)
	response += fmt.Sprintf("- **Total Machines**: %d\n", status.MachineCount)
	response += fmt.Sprintf("- **Hostname**: %s\n", status.Hostname)
	response += fmt.Sprintf("- **Last Updated**: %s\n", status.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
	
	// Machine states section
	if len(status.MachineStates) > 0 {
		response += "\n## Machine States\n"
		
		totalHealthy := 0
		totalUnhealthy := 0
		
		for state, count := range status.MachineStates {
			stateIcon := "⚪"
			stateDescription := state
			
			switch state {
			case "started":
				stateIcon = "🟢"
				stateDescription = "Started (Healthy)"
				totalHealthy += count
			case "stopped":
				stateIcon = "🔴"
				stateDescription = "Stopped"
				totalUnhealthy += count
			case "starting":
				stateIcon = "🟡"
				stateDescription = "Starting"
			case "stopping":
				stateIcon = "🟠"
				stateDescription = "Stopping"
			case "replacing":
				stateIcon = "🔄"
				stateDescription = "Replacing"
			case "destroyed":
				stateIcon = "💀"
				stateDescription = "Destroyed"
				totalUnhealthy += count
			}
			
			response += fmt.Sprintf("- %s **%s**: %d machine(s)\n", stateIcon, stateDescription, count)
		}
		
		// Health summary
		response += "\n### Health Summary\n"
		if totalHealthy > 0 && totalUnhealthy == 0 {
			response += "🟢 **All machines are healthy**\n"
		} else if totalHealthy > 0 && totalUnhealthy > 0 {
			response += fmt.Sprintf("🟡 **Partially healthy**: %d healthy, %d unhealthy\n", totalHealthy, totalUnhealthy)
		} else if totalHealthy == 0 && totalUnhealthy > 0 {
			response += "🔴 **No healthy machines**\n"
		} else {
			response += "⚪ **Status unknown**\n"
		}
	}
	
	// Access information
	response += "\n## Access\n"
	response += fmt.Sprintf("- **Primary URL**: https://%s\n", status.Hostname)
	
	// Quick status interpretation
	response += "\n## Status Interpretation\n"
	if status.Status == "running" && status.Deployed {
		response += "✅ **Application is running and deployed successfully**\n"
		response += "- Your app should be accessible at the URLs above\n"
		response += "- All systems appear to be operational\n"
	} else if status.Status == "suspended" {
		response += "⏸️ **Application is suspended**\n"
		response += "- The app is not currently serving traffic\n"
		response += "- Use `fly_restart` to resume the application\n"
	} else if !status.Deployed {
		response += "⚠️ **Application is not deployed**\n"
		response += "- The app may need to be deployed\n"
		response += "- Use `fly_deploy` to deploy the application\n"
	} else {
		response += "ℹ️ **Application status requires attention**\n"
		response += "- Check the machine states above for more details\n"
		response += "- Consider restarting if machines are in an unhealthy state\n"
	}
	
	// Suggested actions
	response += "\n## Suggested Actions\n"
	response += "- Use `fly_restart` to restart the application\n"
	response += "- Use `fly_logs` to view recent application logs\n"
	response += "- Use `fly_scale` to adjust the number of machines\n"
	response += "- Use `fly_app_info` for detailed application information\n"

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: response,
		}},
	}, nil
}
