package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brannn/fly-mcp/internal/logger"
	"github.com/brannn/fly-mcp/pkg/auth"
	"github.com/brannn/fly-mcp/pkg/config"
	"github.com/brannn/fly-mcp/pkg/fly"
	"github.com/brannn/fly-mcp/pkg/interfaces"
	"github.com/brannn/fly-mcp/pkg/tools"
)

// Handler handles MCP protocol requests
type Handler struct {
	config      *config.Config
	logger      *logger.Logger
	tools       map[string]interfaces.Tool
	flyClient   *fly.Client
	authManager *auth.Manager
}

// NewHandler creates a new MCP handler
func NewHandler(cfg *config.Config, log *logger.Logger) (*Handler, error) {
	// Create Fly.io client
	flyClient, err := fly.NewClient(&cfg.Fly, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create Fly.io client: %w", err)
	}

	// Create authentication manager
	authManager := auth.NewManager(cfg, log)

	handler := &Handler{
		config:      cfg,
		logger:      log,
		tools:       make(map[string]interfaces.Tool),
		flyClient:   flyClient,
		authManager: authManager,
	}

	// Register tools
	if err := handler.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return handler, nil
}

// HandleRequest handles an incoming MCP request
func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	start := time.Now()

	// Parse the MCP request
	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to decode MCP request")
		return h.sendError(w, -32700, "Parse error", nil)
	}

	h.logger.LogMCPRequest(req.Method, req.Params)

	// Handle the request based on method
	var response *MCPResponse
	var err error

	switch req.Method {
	case "initialize":
		response, err = h.handleInitialize(&req)
	case "tools/list":
		response, err = h.handleToolsList(&req)
	case "tools/call":
		response, err = h.handleToolsCall(r, &req)
	case "resources/list":
		response, err = h.handleResourcesList(&req)
	case "resources/read":
		response, err = h.handleResourcesRead(&req)
	default:
		err = fmt.Errorf("unsupported method: %s", req.Method)
	}
	
	duration := time.Since(start)
	
	if err != nil {
		h.logger.LogMCPResponse(req.Method, false, duration)
		return h.sendError(w, -32601, "Method not found", map[string]interface{}{
			"method": req.Method,
			"error":  err.Error(),
		})
	}
	
	h.logger.LogMCPResponse(req.Method, true, duration)
	return h.sendResponse(w, response)
}

// handleInitialize handles the initialize request
func (h *Handler) handleInitialize(req *MCPRequest) (*MCPResponse, error) {
	result := map[string]interface{}{
		"protocolVersion": h.config.MCP.Version,
		"capabilities":    h.config.MCP.Capabilities,
		"serverInfo":      h.config.MCP.ServerInfo,
	}
	
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// handleToolsList handles the tools/list request
func (h *Handler) handleToolsList(req *MCPRequest) (*MCPResponse, error) {
	tools := make([]map[string]interface{}, 0, len(h.tools))
	
	for _, tool := range h.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		})
	}
	
	result := map[string]interface{}{
		"tools": tools,
	}
	
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// handleToolsCall handles the tools/call request
func (h *Handler) handleToolsCall(r *http.Request, req *MCPRequest) (*MCPResponse, error) {
	// Parse parameters
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid parameters for tools/call")
	}
	
	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("tool name is required")
	}
	
	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}
	
	// Find and execute the tool
	tool, exists := h.tools[toolName]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}
	
	start := time.Now()
	result, err := tool.Execute(r.Context(), arguments)
	duration := time.Since(start)
	
	// Log tool execution
	h.logger.LogToolExecution("unknown", toolName, duration, err)
	
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// handleResourcesList handles the resources/list request
func (h *Handler) handleResourcesList(req *MCPRequest) (*MCPResponse, error) {
	// TODO: Implement resources listing
	result := map[string]interface{}{
		"resources": []interface{}{},
	}
	
	return &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// handleResourcesRead handles the resources/read request
func (h *Handler) handleResourcesRead(req *MCPRequest) (*MCPResponse, error) {
	// TODO: Implement resource reading
	return nil, fmt.Errorf("resources/read not implemented")
}

// registerTools registers all available tools
func (h *Handler) registerTools() error {
	h.logger.Info().Msg("Registering MCP tools")

	// Register ping tool for testing
	h.tools["ping"] = &PingTool{logger: h.logger}

	// Register Fly.io management tools
	h.tools["fly_list_apps"] = tools.NewListAppsTool(h.flyClient, h.authManager, h.logger)
	h.tools["fly_app_info"] = tools.NewAppInfoTool(h.flyClient, h.authManager, h.logger)
	h.tools["fly_status"] = tools.NewAppStatusTool(h.flyClient, h.authManager, h.logger)
	h.tools["fly_restart"] = tools.NewAppRestartTool(h.flyClient, h.authManager, h.logger)

	h.logger.Info().
		Int("total_tools", len(h.tools)).
		Strs("tool_names", h.getToolNames()).
		Msg("Tools registered successfully")

	return nil
}

// getToolNames returns a slice of registered tool names for logging
func (h *Handler) getToolNames() []string {
	names := make([]string, 0, len(h.tools))
	for name := range h.tools {
		names = append(names, name)
	}
	return names
}

// sendResponse sends a successful MCP response
func (h *Handler) sendResponse(w http.ResponseWriter, response *MCPResponse) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	return json.NewEncoder(w).Encode(response)
}

// sendError sends an MCP error response
func (h *Handler) sendError(w http.ResponseWriter, code int, message string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // MCP errors are still HTTP 200
	
	response := MCPResponse{
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	
	return json.NewEncoder(w).Encode(response)
}
