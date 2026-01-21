package main

import (
	"fmt"
	"os"

	"github.com/claraverse/mcp-client/internal/commands"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "mcp-client",
	Short: "ClaraVerse MCP Client - Connect local tools to cloud chat",
	Long: `ClaraVerse MCP Client allows you to connect local MCP (Model Context Protocol)
servers to your ClaraVerse cloud chat, giving the AI access to your local tools,
filesystems, databases, and custom integrations.`,
	Version: version,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	// Add all commands
	rootCmd.AddCommand(commands.LoginCmd)
	rootCmd.AddCommand(commands.StartCmd)
	rootCmd.AddCommand(commands.AddCmd)
	rootCmd.AddCommand(commands.ListCmd)
	rootCmd.AddCommand(commands.RemoveCmd)
	rootCmd.AddCommand(commands.StatusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
