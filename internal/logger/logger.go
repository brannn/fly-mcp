package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/brannn/fly-mcp/pkg/config"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// New creates a new logger instance based on configuration
func New(cfg config.LoggingConfig) (*Logger, error) {
	// Set global log level
	level, err := parseLogLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)
	
	// Configure output
	var output io.Writer
	switch cfg.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", cfg.Output, err)
		}
		output = file
	}
	
	// Configure format
	var logger zerolog.Logger
	if cfg.Format == "text" || !cfg.Structured {
		// Human-readable format for local development
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    cfg.Output != "stdout" && cfg.Output != "stderr",
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// JSON format for production
		logger = zerolog.New(output).With().Timestamp().Logger()
	}
	
	return &Logger{Logger: &logger}, nil
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx map[string]interface{}) *Logger {
	event := l.Logger.With()
	for key, value := range ctx {
		event = event.Interface(key, value)
	}
	logger := event.Logger()
	return &Logger{Logger: &logger}
}

// LogToolExecution logs tool execution with timing and result
func (l *Logger) LogToolExecution(userID, toolName string, duration time.Duration, err error) {
	event := l.Info().
		Str("user_id", userID).
		Str("tool_name", toolName).
		Dur("duration", duration).
		Str("action", "tool_execution")
	
	if err != nil {
		event = l.Error().
			Str("user_id", userID).
			Str("tool_name", toolName).
			Dur("duration", duration).
			Str("action", "tool_execution").
			Err(err)
	}
	
	event.Msg("Tool execution completed")
}

// LogMCPRequest logs incoming MCP requests
func (l *Logger) LogMCPRequest(method string, params interface{}) {
	l.Debug().
		Str("method", method).
		Interface("params", params).
		Str("action", "mcp_request").
		Msg("Received MCP request")
}

// LogMCPResponse logs outgoing MCP responses
func (l *Logger) LogMCPResponse(method string, success bool, duration time.Duration) {
	event := l.Info().
		Str("method", method).
		Bool("success", success).
		Dur("duration", duration).
		Str("action", "mcp_response")
	
	if !success {
		event = l.Error().
			Str("method", method).
			Bool("success", success).
			Dur("duration", duration).
			Str("action", "mcp_response")
	}
	
	event.Msg("MCP response sent")
}

// LogFlyAPICall logs Fly.io API calls
func (l *Logger) LogFlyAPICall(endpoint, method string, statusCode int, duration time.Duration) {
	event := l.Info().
		Str("endpoint", endpoint).
		Str("method", method).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Str("action", "fly_api_call")
	
	if statusCode >= 400 {
		event = l.Error().
			Str("endpoint", endpoint).
			Str("method", method).
			Int("status_code", statusCode).
			Dur("duration", duration).
			Str("action", "fly_api_call")
	}
	
	event.Msg("Fly.io API call completed")
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(eventType, userID, resource string, allowed bool) {
	event := l.Warn().
		Str("event_type", eventType).
		Str("user_id", userID).
		Str("resource", resource).
		Bool("allowed", allowed).
		Str("action", "security_event")
	
	if !allowed {
		event = l.Error().
			Str("event_type", eventType).
			Str("user_id", userID).
			Str("resource", resource).
			Bool("allowed", allowed).
			Str("action", "security_event")
	}
	
	event.Msg("Security event")
}

// LogAuditEvent logs audit trail events
func (l *Logger) LogAuditEvent(userID, action, resource, result string) {
	l.Info().
		Str("user_id", userID).
		Str("action", action).
		Str("resource", resource).
		Str("result", result).
		Str("event_type", "audit").
		Msg("Audit event")
}

// Default returns the default logger instance
func Default() *Logger {
	return &Logger{Logger: &log.Logger}
}
