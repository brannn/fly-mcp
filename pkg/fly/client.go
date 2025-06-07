package fly

import (
	"context"
	"fmt"
	"time"

	"github.com/superfly/fly-go"
	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/config"
)

// Client wraps the Fly.io API client with additional functionality
type Client struct {
	flyClient *fly.Client
	logger    *logger.Logger
	config    *config.FlyConfig
}

// NewClient creates a new Fly.io API client
func NewClient(cfg *config.FlyConfig, log *logger.Logger) (*Client, error) {
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("Fly.io API token is required")
	}

	// Create Fly.io client
	flyClient := fly.NewClientFromOptions(fly.ClientOptions{
		AccessToken: cfg.APIToken,
		BaseURL:     cfg.BaseURL,
		Name:        "fly-mcp",
		Version:     "0.1.0",
	})

	client := &Client{
		flyClient: flyClient,
		logger:    log,
		config:    cfg,
	}

	// Validate the client by checking authentication
	if err := client.validateAuth(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to validate Fly.io authentication: %w", err)
	}

	log.Info().
		Str("base_url", cfg.BaseURL).
		Str("organization", cfg.Organization).
		Msg("Fly.io client initialized successfully")

	return client, nil
}

// validateAuth validates the API token by making a simple API call
func (c *Client) validateAuth(ctx context.Context) error {
	start := time.Now()
	
	// Try to get the current user to validate the token
	_, err := c.flyClient.GetCurrentUser(ctx)
	duration := time.Since(start)
	
	c.logger.LogFlyAPICall("/user", "GET", getStatusCode(err), duration)
	
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	
	return nil
}

// GetApps retrieves all applications for the organization
func (c *Client) GetApps(ctx context.Context) ([]App, error) {
	start := time.Now()

	var apps []fly.App
	var err error

	if c.config.Organization != "" {
		apps, err = c.flyClient.GetAppsForOrganization(ctx, c.config.Organization)
	} else {
		apps, err = c.flyClient.GetApps(ctx, nil)
	}

	duration := time.Since(start)
	c.logger.LogFlyAPICall("/apps", "GET", getStatusCode(err), duration)

	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}

	// Convert to our App type
	result := make([]App, len(apps))
	for i, app := range apps {
		result[i] = App{
			ID:       app.ID,
			Name:     app.Name,
			Status:   app.Status,
			Deployed: app.Deployed,
			Hostname: app.Hostname,
			AppURL:   app.AppURL,
			// Note: Organization and timestamps may not be available in all API responses
		}
	}

	c.logger.Debug().
		Int("count", len(result)).
		Str("organization", c.config.Organization).
		Msg("Retrieved apps from Fly.io")

	return result, nil
}

// GetApp retrieves detailed information about a specific application
func (c *Client) GetApp(ctx context.Context, appName string) (*App, error) {
	start := time.Now()

	app, err := c.flyClient.GetAppCompact(ctx, appName)
	duration := time.Since(start)

	c.logger.LogFlyAPICall(fmt.Sprintf("/apps/%s", appName), "GET", getStatusCode(err), duration)

	if err != nil {
		return nil, fmt.Errorf("failed to get app %s: %w", appName, err)
	}

	result := &App{
		ID:       app.ID,
		Name:     app.Name,
		Status:   app.Status,
		Deployed: app.Deployed,
		Hostname: app.Hostname,
		AppURL:   app.AppURL,
		// Note: Organization and timestamps may not be available in AppCompact
	}

	c.logger.Debug().
		Str("app_name", appName).
		Str("status", app.Status).
		Msg("Retrieved app details from Fly.io")

	return result, nil
}

// GetAppStatus retrieves the current status of an application
func (c *Client) GetAppStatus(ctx context.Context, appName string) (*AppStatus, error) {
	start := time.Now()

	app, err := c.flyClient.GetAppCompact(ctx, appName)
	duration := time.Since(start)

	c.logger.LogFlyAPICall(fmt.Sprintf("/apps/%s", appName), "GET", getStatusCode(err), duration)

	if err != nil {
		return nil, fmt.Errorf("failed to get app status for %s: %w", appName, err)
	}

	// For now, we'll create a basic status without machine details
	// TODO: Implement machine API integration for detailed machine states
	status := &AppStatus{
		AppName:       appName,
		Status:        app.Status,
		Deployed:      app.Deployed,
		MachineCount:  0, // Will be updated when machine API is integrated
		MachineStates: make(map[string]int),
		Hostname:      app.Hostname,
		UpdatedAt:     time.Now(), // Use current time since UpdatedAt may not be available
	}

	c.logger.Debug().
		Str("app_name", appName).
		Str("status", app.Status).
		Msg("Retrieved app status from Fly.io")

	return status, nil
}

// RestartApp restarts an application
func (c *Client) RestartApp(ctx context.Context, appName string) error {
	start := time.Now()

	// For now, we'll return a placeholder implementation
	// TODO: Implement actual restart functionality using the machines API
	// This would typically involve:
	// 1. Getting all machines for the app
	// 2. Restarting each machine individually
	// 3. Waiting for machines to come back online

	duration := time.Since(start)
	c.logger.LogFlyAPICall(fmt.Sprintf("/apps/%s/restart", appName), "POST", 200, duration)

	c.logger.Info().
		Str("app_name", appName).
		Msg("Restart initiated for app (placeholder implementation)")

	// Return an error indicating this is not yet implemented
	return fmt.Errorf("restart functionality not yet implemented - requires machines API integration")
}

// getStatusCode extracts HTTP status code from error or returns 200 for success
func getStatusCode(err error) int {
	if err == nil {
		return 200
	}
	
	// Try to extract status code from error
	// This is a simplified approach - in a real implementation,
	// you might want to parse the error more carefully
	return 500
}
