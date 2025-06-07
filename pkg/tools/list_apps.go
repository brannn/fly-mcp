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

// ListAppsTool implements the fly_list_apps MCP tool
type ListAppsTool struct {
	flyClient   *fly.Client
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewListAppsTool creates a new list apps tool
func NewListAppsTool(flyClient *fly.Client, authManager *auth.Manager, logger *logger.Logger) *ListAppsTool {
	return &ListAppsTool{
		flyClient:   flyClient,
		authManager: authManager,
		logger:      logger,
	}
}

// Name returns the tool name
func (t *ListAppsTool) Name() string {
	return "fly_list_apps"
}

// Description returns the tool description
func (t *ListAppsTool) Description() string {
	return "List all applications in your Fly.io organization with their current status, deployment state, and basic information"
}

// InputSchema returns the JSON schema for the tool's input
func (t *ListAppsTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status_filter": map[string]interface{}{
				"type":        "string",
				"description": "Filter apps by status (running, stopped, suspended, etc.)",
				"enum":        []string{"running", "stopped", "suspended", "deployed", "pending"},
			},
			"include_details": map[string]interface{}{
				"type":        "boolean",
				"description": "Include detailed information about each app",
				"default":     false,
			},
			"organization": map[string]interface{}{
				"type":        "string",
				"description": "Organization slug to list apps from (optional, uses configured org if not specified)",
			},
		},
		"additionalProperties": false,
	}
}

// Execute executes the list apps tool
func (t *ListAppsTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
	// Validate permissions
	if err := t.authManager.ValidateRequest(ctx, "read", "apps"); err != nil {
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Permission denied: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Extract arguments
	statusFilter := ""
	if filter, ok := args["status_filter"].(string); ok {
		statusFilter = filter
	}

	includeDetails := false
	if details, ok := args["include_details"].(bool); ok {
		includeDetails = details
	}

	organization := ""
	if org, ok := args["organization"].(string); ok {
		organization = org
	}

	// Log the operation
	userID, _ := t.authManager.ExtractUserFromContext(ctx)
	t.logger.Info().
		Str("user_id", userID).
		Str("tool", "fly_list_apps").
		Str("status_filter", statusFilter).
		Bool("include_details", includeDetails).
		Str("organization", organization).
		Msg("Executing list apps tool")

	// Get apps from Fly.io
	apps, err := t.flyClient.GetApps(ctx)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "list_apps", "apps", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Failed to retrieve apps from Fly.io: %v", err),
			}},
			IsError: true,
		}, nil
	}

	// Filter apps by status if specified
	if statusFilter != "" {
		filteredApps := make([]fly.App, 0)
		for _, app := range apps {
			if app.Status == statusFilter {
				filteredApps = append(filteredApps, app)
			}
		}
		apps = filteredApps
	}

	// Log successful operation
	t.authManager.AuditLog(ctx, userID, "list_apps", "apps", "success", map[string]interface{}{
		"app_count":       len(apps),
		"status_filter":   statusFilter,
		"include_details": includeDetails,
	})

	// Format response
	if len(apps) == 0 {
		message := "No applications found"
		if statusFilter != "" {
			message = fmt.Sprintf("No applications found with status '%s'", statusFilter)
		}
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: message,
			}},
		}, nil
	}

	// Create response content
	var responseText string
	var responseData interface{}

	if includeDetails {
		// Detailed response with JSON data
		responseData = map[string]interface{}{
			"apps":        apps,
			"total_count": len(apps),
			"filter":      statusFilter,
		}

		jsonData, err := json.MarshalIndent(responseData, "", "  ")
		if err != nil {
			return &interfaces.ToolResult{
				Content: []interfaces.ContentBlock{{
					Type: "text",
					Text: fmt.Sprintf("Error formatting response: %v", err),
				}},
				IsError: true,
			}, nil
		}

		responseText = fmt.Sprintf("Found %d applications:\n\n```json\n%s\n```", len(apps), string(jsonData))
	} else {
		// Simple text response
		responseText = fmt.Sprintf("Found %d applications:\n\n", len(apps))
		
		for i, app := range apps {
			status := "🔴 stopped"
			if app.Status == "running" {
				status = "🟢 running"
			} else if app.Status == "suspended" {
				status = "🟡 suspended"
			} else if app.Deployed {
				status = "🔵 deployed"
			}

			responseText += fmt.Sprintf("%d. **%s** (%s)\n", i+1, app.Name, status)
			responseText += fmt.Sprintf("   - URL: %s\n", app.AppURL)
			responseText += fmt.Sprintf("   - Hostname: %s\n", app.Hostname)
			if app.Organization != nil {
				responseText += fmt.Sprintf("   - Organization: %s\n", app.Organization.Name)
			}
			responseText += fmt.Sprintf("   - Updated: %s\n\n", app.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	t.logger.Debug().
		Str("user_id", userID).
		Int("app_count", len(apps)).
		Msg("Successfully listed apps")

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: responseText,
		}},
	}, nil
}
