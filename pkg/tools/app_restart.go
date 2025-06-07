package tools

import (
	"context"
	"fmt"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/auth"
	"github.com/brannn/fly-mcp/pkg/fly"
	"github.com/brannn/fly-mcp/pkg/interfaces"
)

// AppRestartTool implements the fly_restart MCP tool
type AppRestartTool struct {
	flyClient   *fly.Client
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewAppRestartTool creates a new app restart tool
func NewAppRestartTool(flyClient *fly.Client, authManager *auth.Manager, logger *logger.Logger) *AppRestartTool {
	return &AppRestartTool{
		flyClient:   flyClient,
		authManager: authManager,
		logger:      logger,
	}
}

// Name returns the tool name
func (t *AppRestartTool) Name() string {
	return "fly_restart"
}

// Description returns the tool description
func (t *AppRestartTool) Description() string {
	return "Restart a Fly.io application by restarting all of its machines. This is useful for applying configuration changes or recovering from issues."
}

// InputSchema returns the JSON schema for the tool's input
func (t *AppRestartTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the application to restart",
			},
			"confirm": map[string]interface{}{
				"type":        "boolean",
				"description": "Confirmation that you want to restart the application (required for safety)",
				"default":     false,
			},
			"reason": map[string]interface{}{
				"type":        "string",
				"description": "Optional reason for the restart (for audit logging)",
			},
		},
		"required":             []string{"app_name", "confirm"},
		"additionalProperties": false,
	}
}

// Execute executes the app restart tool
func (t *AppRestartTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
	// Validate permissions
	if err := t.authManager.ValidateRequest(ctx, "restart", "app"); err != nil {
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

	confirm, ok := args["confirm"].(bool)
	if !ok || !confirm {
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: "⚠️ **Restart Confirmation Required**\n\nRestarting an application will cause temporary downtime. To proceed, you must set `confirm: true` in your request.\n\nExample:\n```json\n{\n  \"app_name\": \"" + appName + "\",\n  \"confirm\": true,\n  \"reason\": \"Applying configuration changes\"\n}\n```",
			}},
			IsError: true,
		}, nil
	}

	reason := ""
	if r, ok := args["reason"].(string); ok {
		reason = r
	}

	// Log the operation
	userID, _ := t.authManager.ExtractUserFromContext(ctx)
	t.logger.Info().
		Str("user_id", userID).
		Str("tool", "fly_restart").
		Str("app_name", appName).
		Str("reason", reason).
		Msg("Executing app restart tool")

	// Get current app status before restart
	statusBefore, err := t.flyClient.GetAppStatus(ctx, appName)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "restart_app", appName, "failed_status_check", map[string]interface{}{
			"error":  err.Error(),
			"reason": reason,
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Failed to check app status before restart for '%s': %v\n\nThe restart was not performed for safety reasons.", appName, err),
			}},
			IsError: true,
		}, nil
	}

	// Perform the restart
	err = t.flyClient.RestartApp(ctx, appName)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "restart_app", appName, "failed", map[string]interface{}{
			"error":          err.Error(),
			"reason":         reason,
			"machines_before": statusBefore.MachineCount,
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("❌ **Restart Failed**\n\nFailed to restart app '%s': %v\n\nThe application may still be in its previous state. You can check the status using `fly_status`.", appName, err),
			}},
			IsError: true,
		}, nil
	}

	// Log successful operation
	t.authManager.AuditLog(ctx, userID, "restart_app", appName, "success", map[string]interface{}{
		"reason":          reason,
		"machines_before": statusBefore.MachineCount,
		"status_before":   statusBefore.Status,
	})

	// Format success response
	var response string
	
	response += fmt.Sprintf("✅ **Application '%s' Restart Initiated**\n\n", appName)
	
	response += "## Restart Summary\n"
	response += fmt.Sprintf("- **Application**: %s\n", appName)
	response += fmt.Sprintf("- **Status Before**: %s\n", statusBefore.Status)
	response += fmt.Sprintf("- **Machines Restarted**: %d\n", statusBefore.MachineCount)
	if reason != "" {
		response += fmt.Sprintf("- **Reason**: %s\n", reason)
	}
	response += fmt.Sprintf("- **Initiated By**: %s\n", userID)
	
	response += "\n## What Happens Next\n"
	response += "1. 🔄 All machines are being restarted\n"
	response += "2. ⏱️ There may be brief downtime during the restart\n"
	response += "3. 🟢 Machines will come back online automatically\n"
	response += "4. 🌐 Traffic will resume once machines are healthy\n"
	
	response += "\n## Monitoring the Restart\n"
	response += "- Use `fly_status` to check the current status\n"
	response += "- Use `fly_logs` to monitor the restart process\n"
	response += "- The restart typically completes within 1-2 minutes\n"
	
	if statusBefore.Hostname != "" {
		response += fmt.Sprintf("\n## Access\n")
		response += fmt.Sprintf("- **URL**: https://%s\n", statusBefore.Hostname)
		response += "- The application should be accessible at this URL once the restart completes\n"
	}

	t.logger.Info().
		Str("user_id", userID).
		Str("app_name", appName).
		Int("machine_count", statusBefore.MachineCount).
		Msg("Successfully initiated app restart")

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: response,
		}},
	}, nil
}
