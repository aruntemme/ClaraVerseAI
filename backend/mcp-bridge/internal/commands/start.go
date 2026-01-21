package commands

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/claraverse/mcp-client/internal/bridge"
	"github.com/claraverse/mcp-client/internal/config"
	"github.com/claraverse/mcp-client/internal/registry"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP client and connect to backend",
	Long: `Starts the MCP client daemon, connects to the ClaraVerse backend,
and registers all enabled MCP servers. The client will run in the foreground
and handle tool execution requests from the backend.`,
	RunE: runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if authenticated
	if cfg.AuthToken == "" {
		return fmt.Errorf("not authenticated. Please run 'mcp-client login' first")
	}

	verbose, _ := cmd.Flags().GetBool("verbose")

	log.Println("🚀 Starting ClaraVerse MCP Client")
	log.Printf("📍 Config: %s", config.GetConfigPath())
	log.Printf("🌐 Backend: %s", cfg.BackendURL)

	// Create server registry
	reg := registry.NewRegistry(verbose)

	// Start all enabled MCP servers
	enabledServers := cfg.GetEnabledServers()
	if len(enabledServers) == 0 {
		log.Println("⚠️  No MCP servers configured. Add servers with 'mcp-client add'")
	}

	for _, server := range enabledServers {
		if err := reg.StartServer(server); err != nil {
			log.Printf("❌ Failed to start %s: %v", server.Name, err)
			continue
		}
	}

	if reg.GetServerCount() == 0 {
		return fmt.Errorf("no MCP servers started successfully")
	}

	log.Printf("✅ Started %d MCP servers with %d total tools", reg.GetServerCount(), reg.GetToolCount())

	// Create WebSocket bridge
	b := bridge.NewBridge(cfg.BackendURL, cfg.AuthToken, verbose)

	// Set tool call handler
	b.SetToolCallHandler(func(tc bridge.ToolCall) {
		handleToolCall(reg, b, tc)
	})

	// Connect to backend
	log.Println("🔌 Connecting to backend...")
	if err := b.Connect(); err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}

	// Register tools
	clientID := uuid.New().String()
	tools := reg.GetAllTools()

	log.Printf("📦 Registering %d tools...", len(tools))
	if err := b.RegisterTools(clientID, "1.0.0", runtime.GOOS, convertTools(tools)); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	log.Println("✅ MCP client running. Press Ctrl+C to exit.")
	log.Println("💡 Tools are now available in your web chat!")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Println("\n🛑 Shutting down...")
	b.Close()
	reg.StopAll()
	log.Println("✅ Goodbye!")

	return nil
}

func handleToolCall(reg *registry.Registry, b *bridge.Bridge, tc bridge.ToolCall) {
	log.Printf("🔧 Executing tool: %s (call_id: %s)", tc.ToolName, tc.CallID)

	// Execute the tool
	result, err := reg.ExecuteTool(tc.ToolName, tc.Arguments)

	if err != nil {
		log.Printf("❌ Tool execution failed: %v", err)
		b.SendToolResult(tc.CallID, false, "", err.Error())
		return
	}

	log.Printf("✅ Tool executed successfully: %s", tc.ToolName)
	b.SendToolResult(tc.CallID, true, result, "")
}

func convertTools(tools []map[string]interface{}) []interface{} {
	result := make([]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = tool
	}
	return result
}
