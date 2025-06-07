package tools

import (
	"context"
	"fmt"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/auth"
	"github.com/brannn/fly-mcp/pkg/fly"
	"github.com/brannn/fly-mcp/pkg/interfaces"
)

// AppScaleTool implements the fly_scale MCP tool
type AppScaleTool struct {
	flyClient   *fly.Client
	authManager *auth.Manager
	logger      *logger.Logger
}

// NewAppScaleTool creates a new app scale tool
func NewAppScaleTool(flyClient *fly.Client, authManager *auth.Manager, logger *logger.Logger) *AppScaleTool {
	return &AppScaleTool{
		flyClient:   flyClient,
		authManager: authManager,
		logger:      logger,
	}
}

// Name returns the tool name
func (t *AppScaleTool) Name() string {
	return "fly_scale"
}

// Description returns the tool description
func (t *AppScaleTool) Description() string {
	return "Scale a Fly.io application by showing current machine count and providing scaling recommendations. Note: Actual scaling requires manual intervention or deployment."
}

// InputSchema returns the JSON schema for the tool's input
func (t *AppScaleTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the application to check scaling for",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action to perform: 'status' to show current scale, 'recommend' for scaling recommendations",
				"enum":        []string{"status", "recommend"},
				"default":     "status",
			},
			"target_count": map[string]interface{}{
				"type":        "integer",
				"description": "Target number of machines (for recommendations)",
				"minimum":     0,
				"maximum":     100,
			},
		},
		"required":             []string{"app_name"},
		"additionalProperties": false,
	}
}

// Execute executes the app scale tool
func (t *AppScaleTool) Execute(ctx context.Context, args map[string]interface{}) (*interfaces.ToolResult, error) {
	// Validate permissions
	if err := t.authManager.ValidateRequest(ctx, "scale", "app"); err != nil {
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

	action := "status"
	if a, ok := args["action"].(string); ok {
		action = a
	}

	var targetCount *int
	if tc, ok := args["target_count"].(float64); ok {
		count := int(tc)
		targetCount = &count
	}

	// Log the operation
	userID, _ := t.authManager.ExtractUserFromContext(ctx)
	t.logger.Info().
		Str("user_id", userID).
		Str("tool", "fly_scale").
		Str("app_name", appName).
		Str("action", action).
		Msg("Executing app scale tool")

	// Get current app status with machine information
	status, err := t.flyClient.GetAppStatus(ctx, appName)
	if err != nil {
		t.authManager.AuditLog(ctx, userID, "scale_check", appName, "failed", map[string]interface{}{
			"error": err.Error(),
		})
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Failed to retrieve app status for '%s': %v", appName, err),
			}},
			IsError: true,
		}, nil
	}

	// Log successful operation
	t.authManager.AuditLog(ctx, userID, "scale_check", appName, "success", map[string]interface{}{
		"action":        action,
		"machine_count": status.MachineCount,
		"target_count":  targetCount,
	})

	// Handle different actions
	switch action {
	case "status":
		return t.formatStatusResponse(status)
	case "recommend":
		return t.formatRecommendationResponse(status, targetCount)
	default:
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: fmt.Sprintf("Unknown action: %s. Use 'status' or 'recommend'", action),
			}},
			IsError: true,
		}, nil
	}
}

// formatStatusResponse formats the current scaling status
func (t *AppScaleTool) formatStatusResponse(status *fly.AppStatus) (*interfaces.ToolResult, error) {
	var response string
	
	// Header
	response += fmt.Sprintf("# Scaling Status: %s\n\n", status.AppName)
	
	// Current scale
	response += "## Current Scale\n"
	response += fmt.Sprintf("- **Total Machines**: %d\n", status.MachineCount)
	response += fmt.Sprintf("- **App Status**: %s\n", status.Status)
	response += fmt.Sprintf("- **Deployed**: %t\n", status.Deployed)
	
	// Machine distribution
	if len(status.MachineStates) > 0 {
		response += "\n## Machine States\n"
		
		runningCount := 0
		stoppedCount := 0
		
		for state, count := range status.MachineStates {
			stateIcon := "⚪"
			switch state {
			case "started":
				stateIcon = "🟢"
				runningCount += count
			case "stopped":
				stateIcon = "🔴"
				stoppedCount += count
			case "starting":
				stateIcon = "🟡"
			case "stopping":
				stateIcon = "🟠"
			}
			response += fmt.Sprintf("- %s **%s**: %d machine(s)\n", stateIcon, state, count)
		}
		
		// Health summary
		response += "\n### Scale Health\n"
		if runningCount > 0 && stoppedCount == 0 {
			response += "✅ **All machines are running**\n"
		} else if runningCount > 0 && stoppedCount > 0 {
			response += fmt.Sprintf("⚠️ **Mixed state**: %d running, %d stopped\n", runningCount, stoppedCount)
		} else if runningCount == 0 {
			response += "🔴 **No machines are running**\n"
		}
	}
	
	// Scaling recommendations
	response += "\n## Scaling Recommendations\n"
	
	if status.MachineCount == 0 {
		response += "- ⚠️ **No machines found** - App may need to be deployed first\n"
		response += "- Use `fly_deploy` to deploy the application\n"
	} else if status.MachineCount == 1 {
		response += "- 📈 **Single machine setup** - Consider adding more machines for high availability\n"
		response += "- Recommended: 2-3 machines for production workloads\n"
	} else if status.MachineCount >= 2 && status.MachineCount <= 3 {
		response += "- ✅ **Good scaling setup** for most applications\n"
		response += "- Monitor performance and scale up if needed\n"
	} else if status.MachineCount > 10 {
		response += "- 📊 **High scale deployment** - Monitor costs and utilization\n"
		response += "- Consider if all machines are necessary\n"
	}
	
	// Scaling actions
	response += "\n## Scaling Actions\n"
	response += "To scale your application:\n"
	response += "1. **Manual scaling**: Use `flyctl scale count <number>` in your terminal\n"
	response += "2. **Check recommendations**: Use this tool with `action: recommend` and `target_count`\n"
	response += "3. **Auto-scaling**: Configure auto-scaling in your fly.toml file\n"
	
	response += "\n## Next Steps\n"
	response += "- Use `fly_status` to monitor machine health\n"
	response += "- Use `fly_restart` if machines are in unhealthy states\n"
	response += "- Monitor application performance and adjust scale as needed\n"

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: response,
		}},
	}, nil
}

// formatRecommendationResponse formats scaling recommendations
func (t *AppScaleTool) formatRecommendationResponse(status *fly.AppStatus, targetCount *int) (*interfaces.ToolResult, error) {
	var response string
	
	currentCount := status.MachineCount
	
	if targetCount == nil {
		response += fmt.Sprintf("# Scaling Recommendations: %s\n\n", status.AppName)
		response += "## General Recommendations\n"
		response += fmt.Sprintf("- **Current machines**: %d\n", currentCount)
		response += "\n**Recommended scaling based on use case:**\n"
		response += "- **Development**: 1 machine\n"
		response += "- **Staging**: 1-2 machines\n"
		response += "- **Production (small)**: 2-3 machines\n"
		response += "- **Production (medium)**: 3-5 machines\n"
		response += "- **Production (large)**: 5+ machines\n"
		response += "\nProvide `target_count` for specific scaling recommendations.\n"
		
		return &interfaces.ToolResult{
			Content: []interfaces.ContentBlock{{
				Type: "text",
				Text: response,
			}},
		}, nil
	}
	
	target := *targetCount
	response += fmt.Sprintf("# Scaling Recommendation: %s\n\n", status.AppName)
	response += fmt.Sprintf("**Current**: %d machines → **Target**: %d machines\n\n", currentCount, target)
	
	if target == currentCount {
		response += "✅ **No scaling needed** - You're already at the target count\n"
	} else if target > currentCount {
		diff := target - currentCount
		response += fmt.Sprintf("📈 **Scale Up Recommendation** (+%d machines)\n\n", diff)
		response += "**Benefits:**\n"
		response += "- Increased capacity and performance\n"
		response += "- Better fault tolerance and availability\n"
		response += "- Improved load distribution\n\n"
		response += "**Considerations:**\n"
		response += fmt.Sprintf("- Additional cost: ~$%d/month (estimated)\n", diff*15) // Rough estimate
		response += "- Ensure your application can handle distributed load\n"
		response += "- Monitor resource utilization after scaling\n"
	} else {
		diff := currentCount - target
		response += fmt.Sprintf("📉 **Scale Down Recommendation** (-%d machines)\n\n", diff)
		response += "**Benefits:**\n"
		response += "- Reduced operational costs\n"
		response += "- Simplified management\n\n"
		response += "**Considerations:**\n"
		response += "- Ensure remaining capacity can handle peak load\n"
		response += "- Consider keeping at least 2 machines for availability\n"
		response += "- Monitor performance after scaling down\n"
		
		if target == 0 {
			response += "\n⚠️ **Warning**: Scaling to 0 machines will make your app unavailable\n"
		} else if target == 1 {
			response += "\n⚠️ **Warning**: Single machine setup has no redundancy\n"
		}
	}
	
	response += "\n## How to Scale\n"
	response += fmt.Sprintf("Run this command in your terminal:\n```bash\nflyctl scale count %d\n```\n", target)
	response += "\nOr update your fly.toml file and redeploy:\n"
	response += "```toml\n[http_service]\n  min_machines_running = " + fmt.Sprintf("%d", target) + "\n```\n"
	
	response += "\n## Post-Scaling Checklist\n"
	response += "- [ ] Monitor application performance\n"
	response += "- [ ] Check machine health with `fly_status`\n"
	response += "- [ ] Verify load distribution\n"
	response += "- [ ] Update monitoring and alerting thresholds\n"

	return &interfaces.ToolResult{
		Content: []interfaces.ContentBlock{{
			Type: "text",
			Text: response,
		}},
	}, nil
}
