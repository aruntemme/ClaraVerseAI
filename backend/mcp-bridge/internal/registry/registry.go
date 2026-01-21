package registry

import (
	"fmt"
	"log"
	"sync"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/mcp"
)

// ServerInstance represents a running MCP server
type ServerInstance struct {
	Config   config.MCPServer
	Executor *mcp.Executor
	Tools    []mcp.Tool
}

// Registry manages all MCP server instances
type Registry struct {
	servers map[string]*ServerInstance
	mutex   sync.RWMutex
	verbose bool
}

// NewRegistry creates a new server registry
func NewRegistry(verbose bool) *Registry {
	return &Registry{
		servers: make(map[string]*ServerInstance),
		verbose: verbose,
	}
}

// StartServer starts an MCP server
func (r *Registry) StartServer(cfg config.MCPServer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if already running
	if _, exists := r.servers[cfg.Name]; exists {
		return fmt.Errorf("server %s is already running", cfg.Name)
	}

	// Only support stdio for now
	if cfg.Type != "stdio" {
		return fmt.Errorf("only stdio servers are supported (server %s uses %s)", cfg.Name, cfg.Type)
	}

	log.Printf("🚀 Starting MCP server: %s", cfg.Name)

	// Create executor - check if command-based or path-based
	var executor *mcp.Executor
	var err error

	if cfg.Command != "" {
		// Command-based server (e.g., npx @browsermcp/mcp@latest)
		executor, err = mcp.NewExecutorWithCommand(cfg.Name, cfg.Command, cfg.Args, r.verbose)
	} else if cfg.Path != "" {
		// Path-based server (e.g., /path/to/server.exe)
		executor, err = mcp.NewExecutor(cfg.Path, r.verbose)
	} else {
		return fmt.Errorf("server %s must have either 'path' or 'command' configured", cfg.Name)
	}

	if err != nil {
		return fmt.Errorf("failed to start server %s: %w", cfg.Name, err)
	}

	// List tools
	tools, err := executor.ListTools()
	if err != nil {
		executor.Close()
		return fmt.Errorf("failed to list tools from %s: %w", cfg.Name, err)
	}

	instance := &ServerInstance{
		Config:   cfg,
		Executor: executor,
		Tools:    tools,
	}

	r.servers[cfg.Name] = instance

	log.Printf("✅ Server %s started with %d tools", cfg.Name, len(tools))
	for _, tool := range tools {
		log.Printf("   - %s: %s", tool.Name, tool.Description)
	}

	return nil
}

// StopServer stops an MCP server
func (r *Registry) StopServer(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.servers[name]
	if !exists {
		return fmt.Errorf("server %s is not running", name)
	}

	log.Printf("🛑 Stopping MCP server: %s", name)

	if err := instance.Executor.Close(); err != nil {
		log.Printf("Warning: error closing executor for %s: %v", name, err)
	}

	delete(r.servers, name)

	log.Printf("✅ Server %s stopped", name)
	return nil
}

// StopAll stops all running servers
func (r *Registry) StopAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for name, instance := range r.servers {
		log.Printf("🛑 Stopping server: %s", name)
		instance.Executor.Close()
	}

	r.servers = make(map[string]*ServerInstance)
}

// GetAllTools returns all tools from all running servers
func (r *Registry) GetAllTools() []map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var allTools []map[string]interface{}

	for _, instance := range r.servers {
		for _, tool := range instance.Tools {
			// Convert MCP tool to OpenAI format
			toolDef := map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.InputSchema,
			}
			allTools = append(allTools, toolDef)
		}
	}

	return allTools
}

// ExecuteTool executes a tool by finding which server provides it
func (r *Registry) ExecuteTool(toolName string, arguments map[string]interface{}) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Find which server has this tool
	for serverName, instance := range r.servers {
		for _, tool := range instance.Tools {
			if tool.Name == toolName {
				log.Printf("🔧 Executing %s on server %s", toolName, serverName)
				return instance.Executor.CallTool(toolName, arguments)
			}
		}
	}

	return "", fmt.Errorf("tool %s not found in any running server", toolName)
}

// GetServerCount returns the number of running servers
func (r *Registry) GetServerCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.servers)
}

// GetToolCount returns the total number of tools across all servers
func (r *Registry) GetToolCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, instance := range r.servers {
		count += len(instance.Tools)
	}
	return count
}

// GetServerNames returns names of all running servers
func (r *Registry) GetServerNames() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.servers))
	for name := range r.servers {
		names = append(names, name)
	}
	return names
}

// GetServer returns a server instance by name
func (r *Registry) GetServer(name string) (*ServerInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.servers[name]
	if !exists {
		return nil, fmt.Errorf("server %s not found", name)
	}
	return instance, nil
}
