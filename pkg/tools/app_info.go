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

// AppInfoTool implements the fly_app_info MCP tool
type AppInfoTool struct {
	flyClient   *fly.Client
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewAppInfoTool creates a new app info tool
func NewAppInfoTool(flyClient *fly.Client, authManager *auth.Manager, logger *logger.Logger) *AppInfoTool {
	return &AppInfoTool{
		flyClient:   flyClient,
		authManager: authManager,
		logger:      logger,
	}
}

// Name returns the tool name
func (t *AppInfoTool) Name() string {
	return "fly_app_info"
}

// Description returns the tool description
func (t *AppInfoTool) Description() string {
	return "Get detailed information about a specific Fly.io application including configuration, status, and deployment details"
}

// InputSchema returns the JSON schema for the tool's input
func (t *AppInfoTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the application to get information about",
			},
			"include_status": map[string]interface{}{
				"type":        "boolean",
				"description": "Include current status and machine information",
				"default":     true,
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Response format (text or json)",
				"enum":        []string{"text", "json"},
				"default":     "text",
			},
		},
		"required":             []string{"app_name"},
		"additionalProperties": false,
	}
}

// Execute executes the app info tool
func (t *AppInfoTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
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

	includeStatus := true
	if status, ok := args["include_status"].(bool); ok {
		includeStatus = status
	}

	format := "text"
	if fmt, ok := args["format"].(string); ok {
		format = fmt
	}

	// Log the operation
	userID, _ := t.authManager.ExtractUserFromContext(ctx)
	t.logger.Info().
		Str("user_id", userID).
		Str("tool", "fly_app_info").
		Str("app_name", appName).
		Bool("include_status", includeStatus).
		Str("format", format).
		Msg("Executing app info tool")

	// Get app information from Fly.io
	app, err := t.flyClient.GetApp(ctx, appName)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "get_app_info", appName, "failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Failed to retrieve app information for '%s': %v", appName, err),
			}},
			IsError: true,
		}, nil
	}

	// Get status information if requested
	var appStatus *fly.AppStatus
	if includeStatus {
		status, err := t.flyClient.GetAppStatus(ctx, appName)
		if err != nil {
			t.logger.Warn().
				Str("app_name", appName).
				Err(err).
				Msg("Failed to get app status, continuing without status info")
		} else {
			appStatus = status
		}
	}

	// Log successful operation
	t.authManager.AuditLog(ctx, userID, "get_app_info", appName, "success", map[string]interface{}{
		"include_status": includeStatus,
		"format":         format,
	})

	// Format response based on requested format
	if format == "json" {
		return t.formatJSONResponse(app, appStatus)
	}
	
	return t.formatTextResponse(app, appStatus)
}

// formatJSONResponse formats the response as JSON
func (t *AppInfoTool) formatJSONResponse(app *fly.App, status *fly.AppStatus) (*interfaces.ToolResult, error) {
	response := map[string]interface{}{
		"app": app,
	}
	
	if status != nil {
		response["status"] = status
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
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
			Text: fmt.Sprintf("Application information for '%s':\n\n```json\n%s\n```", app.Name, string(jsonData)),
		}},
	}, nil
}

// formatTextResponse formats the response as human-readable text
func (t *AppInfoTool) formatTextResponse(app *fly.App, status *fly.AppStatus) (*interfaces.ToolResult, error) {
	var response string
	
	// App header
	response += fmt.Sprintf("# Application: %s\n\n", app.Name)
	
	// Basic information
	response += "## Basic Information\n"
	response += fmt.Sprintf("- **ID**: %s\n", app.ID)
	response += fmt.Sprintf("- **Name**: %s\n", app.Name)
	response += fmt.Sprintf("- **Status**: %s\n", app.Status)
	response += fmt.Sprintf("- **Deployed**: %t\n", app.Deployed)
	response += fmt.Sprintf("- **Hostname**: %s\n", app.Hostname)
	response += fmt.Sprintf("- **App URL**: %s\n", app.AppURL)
	
	if app.Organization != nil {
		response += fmt.Sprintf("- **Organization**: %s\n", app.Organization.Name)
	}
	
	response += fmt.Sprintf("- **Created**: %s\n", app.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	response += fmt.Sprintf("- **Updated**: %s\n", app.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
	
	// Status information
	if status != nil {
		response += "\n## Current Status\n"
		
		statusIcon := "🔴"
		if status.Status == "running" {
			statusIcon = "🟢"
		} else if status.Status == "suspended" {
			statusIcon = "🟡"
		} else if status.Deployed {
			statusIcon = "🔵"
		}
		
		response += fmt.Sprintf("- **Status**: %s %s\n", statusIcon, status.Status)
		response += fmt.Sprintf("- **Deployed**: %t\n", status.Deployed)
		response += fmt.Sprintf("- **Machine Count**: %d\n", status.MachineCount)
		
		if len(status.MachineStates) > 0 {
			response += "- **Machine States**:\n"
			for state, count := range status.MachineStates {
				stateIcon := "⚪"
				switch state {
				case "started":
					stateIcon = "🟢"
				case "stopped":
					stateIcon = "🔴"
				case "starting":
					stateIcon = "🟡"
				case "stopping":
					stateIcon = "🟠"
				}
				response += fmt.Sprintf("  - %s %s: %d\n", stateIcon, state, count)
			}
		}
		
		response += fmt.Sprintf("- **Last Updated**: %s\n", status.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
	}
	
	// URLs and access
	response += "\n## Access Information\n"
	response += fmt.Sprintf("- **Primary URL**: https://%s\n", app.Hostname)
	if app.AppURL != "" && app.AppURL != app.Hostname {
		response += fmt.Sprintf("- **App URL**: %s\n", app.AppURL)
	}
	
	// Quick actions
	response += "\n## Quick Actions\n"
	response += "You can perform the following actions on this app:\n"
	response += "- Use `fly_status` to get real-time status\n"
	response += "- Use `fly_restart` to restart the application\n"
	response += "- Use `fly_scale` to scale the application\n"
	response += "- Use `fly_logs` to view application logs\n"

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: response,
		}},
	}, nil
}
