package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/config"
)

// Manager handles authentication and authorization
type Manager struct {
	config *config.Config
	logger *logger.Logger
}

// NewManager creates a new authentication manager
func NewManager(cfg *config.Config, log *logger.Logger) *Manager {
	return &Manager{
		config: cfg,
		logger: log,
	}
}

// ValidateAPIToken validates a Fly.io API token format
func (m *Manager) ValidateAPIToken(token string) error {
	if token == "" {
		return fmt.Errorf("API token cannot be empty")
	}
	
	// Fly.io tokens typically start with "fo1_" for personal access tokens
	if !strings.HasPrefix(token, "fo1_") && !strings.HasPrefix(token, "fly_") {
		m.logger.Warn().
			Str("token_prefix", getTokenPrefix(token)).
			Msg("API token does not match expected Fly.io format")
	}
	
	// Basic length validation - Fly.io tokens are typically longer than 20 characters
	if len(token) < 20 {
		return fmt.Errorf("API token appears to be too short")
	}
	
	m.logger.Debug().
		Str("token_prefix", getTokenPrefix(token)).
		Int("token_length", len(token)).
		Msg("API token format validation passed")
	
	return nil
}

// ValidatePermissions checks if a user has permission to perform an action
func (m *Manager) ValidatePermissions(ctx context.Context, userID, action, resource string) error {
	// Get user permissions from config
	permissions, exists := m.config.Security.Permissions[userID]
	if !exists {
		// Check default permissions
		permissions, exists = m.config.Security.Permissions["default"]
		if !exists {
			return fmt.Errorf("no permissions configured for user %s", userID)
		}
	}
	
	// Check if user has the required permission
	requiredPermission := fmt.Sprintf("%s:%s", action, resource)
	for _, permission := range permissions {
		if permission == requiredPermission || permission == fmt.Sprintf("%s:*", action) || permission == "*" {
			m.logger.Debug().
				Str("user_id", userID).
				Str("action", action).
				Str("resource", resource).
				Str("permission", permission).
				Msg("Permission granted")
			return nil
		}
	}
	
	m.logger.Warn().
		Str("user_id", userID).
		Str("action", action).
		Str("resource", resource).
		Strs("user_permissions", permissions).
		Msg("Permission denied")
	
	return fmt.Errorf("insufficient permissions: user %s cannot %s on %s", userID, action, resource)
}

// LogSecurityEvent logs a security-related event
func (m *Manager) LogSecurityEvent(ctx context.Context, eventType, userID, resource string, allowed bool, details map[string]interface{}) {
	event := m.logger.Warn()
	if !allowed {
		event = m.logger.Error()
	}
	
	logEvent := event.
		Str("event_type", eventType).
		Str("user_id", userID).
		Str("resource", resource).
		Bool("allowed", allowed).
		Str("action", "security_event")
	
	if details != nil {
		logEvent = logEvent.Interface("details", details)
	}
	
	logEvent.Msg("Security event")
}

// AuditLog logs an audit trail event
func (m *Manager) AuditLog(ctx context.Context, userID, action, resource, result string, metadata map[string]interface{}) {
	logEvent := m.logger.Info().
		Str("user_id", userID).
		Str("action", action).
		Str("resource", resource).
		Str("result", result).
		Str("event_type", "audit").
		Time("timestamp", time.Now())
	
	if metadata != nil {
		logEvent = logEvent.Interface("metadata", metadata)
	}
	
	logEvent.Msg("Audit event")
}

// ExtractUserFromContext extracts user information from request context
func (m *Manager) ExtractUserFromContext(ctx context.Context) (string, error) {
	// In a real implementation, this would extract user info from JWT token,
	// session, or other authentication mechanism
	
	// For now, we'll use a simple approach with a context value
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		return userID, nil
	}
	
	// Default to anonymous user for development
	return "anonymous", nil
}

// ValidateRequest performs comprehensive request validation
func (m *Manager) ValidateRequest(ctx context.Context, action, resource string) error {
	// Extract user from context
	userID, err := m.ExtractUserFromContext(ctx)
	if err != nil {
		m.LogSecurityEvent(ctx, "auth_extraction_failed", "unknown", resource, false, map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to extract user from context: %w", err)
	}
	
	// Validate permissions
	if err := m.ValidatePermissions(ctx, userID, action, resource); err != nil {
		m.LogSecurityEvent(ctx, "permission_denied", userID, resource, false, map[string]interface{}{
			"action": action,
			"error":  err.Error(),
		})
		return err
	}
	
	m.LogSecurityEvent(ctx, "request_authorized", userID, resource, true, map[string]interface{}{
		"action": action,
	})
	
	return nil
}

// CreateAuditContext creates a context with audit information
func (m *Manager) CreateAuditContext(ctx context.Context, userID, requestID string) context.Context {
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "request_id", requestID)
	ctx = context.WithValue(ctx, "audit_timestamp", time.Now())
	return ctx
}

// getTokenPrefix safely extracts the first few characters of a token for logging
func getTokenPrefix(token string) string {
	if len(token) < 8 {
		return "***"
	}
	return token[:8] + "***"
}

// TokenInfo represents information about an API token
type TokenInfo struct {
	Prefix    string    `json:"prefix"`
	Length    int       `json:"length"`
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

// GetTokenInfo returns safe information about a token
func (m *Manager) GetTokenInfo(token string) *TokenInfo {
	info := &TokenInfo{
		Prefix: getTokenPrefix(token),
		Length: len(token),
		Valid:  m.ValidateAPIToken(token) == nil,
	}
	
	return info
}

// Permission represents a permission string
type Permission string

const (
	// Fly.io permissions
	PermissionFlyRead    Permission = "fly:read"
	PermissionFlyDeploy  Permission = "fly:deploy"
	PermissionFlyScale   Permission = "fly:scale"
	PermissionFlyRestart Permission = "fly:restart"
	PermissionFlyLogs    Permission = "fly:logs"
	PermissionFlySecrets Permission = "fly:secrets"
	PermissionFlyVolumes Permission = "fly:volumes"
	PermissionFlyAll     Permission = "fly:*"
	
	// Admin permissions
	PermissionAdmin Permission = "*"
)

// HasPermission checks if a user has a specific permission
func (m *Manager) HasPermission(userID string, permission Permission) bool {
	permissions, exists := m.config.Security.Permissions[userID]
	if !exists {
		permissions, exists = m.config.Security.Permissions["default"]
		if !exists {
			return false
		}
	}
	
	permStr := string(permission)
	for _, p := range permissions {
		if p == permStr || p == "*" {
			return true
		}
		
		// Check wildcard permissions
		if strings.HasSuffix(p, ":*") {
			prefix := strings.TrimSuffix(p, ":*")
			if strings.HasPrefix(permStr, prefix+":") {
				return true
			}
		}
	}
	
	return false
}
