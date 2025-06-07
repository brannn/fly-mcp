package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/internal/server"
	"github.com/brannn/fly-mcp/pkg/config"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fly-mcp",
	Short: "MCP server for Fly.io infrastructure management",
	Long: `fly-mcp is an MCP (Model Context Protocol) server that enables AI assistants
to manage Fly.io infrastructure through natural language interactions.

It provides tools for deploying applications, scaling resources, managing secrets,
and monitoring your Fly.io infrastructure.`,
	RunE: runServer,
}

var (
	configFile string
	logLevel   string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level (debug, info, warn, error)")
	
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(validateCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("fly-mcp %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		fmt.Println("Configuration is valid!")
		fmt.Printf("Environment: %s\n", cfg.Environment)
		fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
		fmt.Printf("Fly.io Organization: %s\n", cfg.Fly.Organization)
		fmt.Printf("Log Level: %s\n", cfg.Logging.Level)
		
		return nil
	},
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// Override log level if specified
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	
	// Initialize logger
	log, err := logger.New(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	log.Info().
		Str("version", version).
		Str("commit", commit).
		Str("environment", cfg.Environment).
		Msg("Starting fly-mcp server")
	
	// Create server
	srv, err := server.New(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(ctx); err != nil {
			serverErr <- err
		}
	}()
	
	log.Info().
		Str("host", cfg.Server.Host).
		Int("port", cfg.Server.Port).
		Msg("Server started successfully")
	
	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
		
		// Give server time to shutdown gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Error during server shutdown")
			return err
		}
		
		log.Info().Msg("Server shutdown complete")
		
	case err := <-serverErr:
		log.Error().Err(err).Msg("Server error")
		return err
	}
	
	return nil
}

func loadConfig() (*config.Config, error) {
	if configFile != "" {
		// Load specific config file
		return config.LoadFromFile(configFile)
	}
	
	// Load config using standard discovery
	return config.Load()
}
