package fly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/config"
)

// MachinesClient handles direct HTTP calls to the Fly.io Machines API
type MachinesClient struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
	logger     *logger.Logger
}

// NewMachinesClient creates a new Machines API client
func NewMachinesClient(cfg *config.FlyConfig, log *logger.Logger) *MachinesClient {
	return &MachinesClient{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		baseURL:  "https://api.machines.dev",
		apiToken: cfg.APIToken,
		logger:   log,
	}
}

// Machine represents a Fly.io machine from the Machines API
type Machine struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	State      string                 `json:"state"`
	Region     string                 `json:"region"`
	InstanceID string                 `json:"instance_id"`
	PrivateIP  string                 `json:"private_ip"`
	Config     map[string]interface{} `json:"config"`
	ImageRef   ImageRef               `json:"image_ref"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Events     []MachineEvent         `json:"events"`
}

// ImageRef represents a container image reference
type ImageRef struct {
	Registry   string            `json:"registry"`
	Repository string            `json:"repository"`
	Tag        string            `json:"tag,omitempty"`
	Digest     string            `json:"digest"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// MachineEvent represents a machine event
type MachineEvent struct {
	Type      string `json:"type"`
	Status    string `json:"status"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
}

// ListMachines retrieves all machines for an app
func (c *MachinesClient) ListMachines(ctx context.Context, appName string) ([]Machine, error) {
	start := time.Now()
	
	url := fmt.Sprintf("%s/v1/apps/%s/machines", c.baseURL, appName)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	c.logger.LogFlyAPICall(fmt.Sprintf("/v1/apps/%s/machines", appName), "GET", getStatusCodeFromResp(resp, err), duration)
	
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var machines []Machine
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Debug().
		Str("app_name", appName).
		Int("machine_count", len(machines)).
		Msg("Retrieved machines from Fly.io Machines API")
	
	return machines, nil
}

// GetMachine retrieves a specific machine
func (c *MachinesClient) GetMachine(ctx context.Context, appName, machineID string) (*Machine, error) {
	start := time.Now()
	
	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s", c.baseURL, appName, machineID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	c.logger.LogFlyAPICall(fmt.Sprintf("/v1/apps/%s/machines/%s", appName, machineID), "GET", getStatusCodeFromResp(resp, err), duration)
	
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	c.logger.Debug().
		Str("app_name", appName).
		Str("machine_id", machineID).
		Str("state", machine.State).
		Msg("Retrieved machine from Fly.io Machines API")
	
	return &machine, nil
}

// StartMachine starts a machine
func (c *MachinesClient) StartMachine(ctx context.Context, appName, machineID string) error {
	start := time.Now()
	
	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s/start", c.baseURL, appName, machineID)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	c.logger.LogFlyAPICall(fmt.Sprintf("/v1/apps/%s/machines/%s/start", appName, machineID), "POST", getStatusCodeFromResp(resp, err), duration)
	
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to start machine: status %d: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info().
		Str("app_name", appName).
		Str("machine_id", machineID).
		Msg("Successfully started machine")
	
	return nil
}

// StopMachine stops a machine
func (c *MachinesClient) StopMachine(ctx context.Context, appName, machineID string) error {
	start := time.Now()
	
	url := fmt.Sprintf("%s/v1/apps/%s/machines/%s/stop", c.baseURL, appName, machineID)
	
	// Create request body with default stop configuration
	stopConfig := map[string]interface{}{
		"timeout": "30s",
	}
	
	body, err := json.Marshal(stopConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal stop config: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)
	
	c.logger.LogFlyAPICall(fmt.Sprintf("/v1/apps/%s/machines/%s/stop", appName, machineID), "POST", getStatusCodeFromResp(resp, err), duration)
	
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stop machine: status %d: %s", resp.StatusCode, string(body))
	}
	
	c.logger.Info().
		Str("app_name", appName).
		Str("machine_id", machineID).
		Msg("Successfully stopped machine")
	
	return nil
}

// RestartMachine restarts a machine by stopping and starting it
func (c *MachinesClient) RestartMachine(ctx context.Context, appName, machineID string) error {
	c.logger.Info().
		Str("app_name", appName).
		Str("machine_id", machineID).
		Msg("Restarting machine")
	
	// Stop the machine first
	if err := c.StopMachine(ctx, appName, machineID); err != nil {
		return fmt.Errorf("failed to stop machine during restart: %w", err)
	}
	
	// Wait a moment for the machine to fully stop
	time.Sleep(2 * time.Second)
	
	// Start the machine
	if err := c.StartMachine(ctx, appName, machineID); err != nil {
		return fmt.Errorf("failed to start machine during restart: %w", err)
	}
	
	c.logger.Info().
		Str("app_name", appName).
		Str("machine_id", machineID).
		Msg("Successfully restarted machine")
	
	return nil
}

// getStatusCodeFromResp extracts status code from HTTP response or returns 500 for errors
func getStatusCodeFromResp(resp *http.Response, err error) int {
	if err != nil {
		return 500
	}
	if resp == nil {
		return 500
	}
	return resp.StatusCode
}
