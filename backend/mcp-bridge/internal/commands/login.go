package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/claraverse/mcp-client/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with ClaraVerse",
	Long: `Authenticate with your ClaraVerse account using email and password.
Your credentials will be used to obtain a JWT token from Supabase.`,
	RunE: runLogin,
}

type SupabaseAuthResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

func runLogin(cmd *cobra.Command, args []string) error {
	fmt.Println("🔐 ClaraVerse Authentication")
	fmt.Println()

	// Get email
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Get password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	fmt.Println() // New line after password input

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Authenticate with Supabase
	fmt.Println("🔄 Authenticating...")

	supabaseURL := "https://ocqoqjafmjuiywsppwkw.supabase.co"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im9jcW9xamFmbWp1aXl3c3Bwd2t3Iiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjI5Njk1NTQsImV4cCI6MjA3ODU0NTU1NH0.LwM-n70KvdPpU6-lnMMgphGUPQIk62otNreXpsplYeA"

	// Create auth request
	authData := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send auth request to Supabase
	req, err := http.NewRequest("POST", supabaseURL+"/auth/v1/token?grant_type=password", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", supabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var authResp SupabaseAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if authResp.AccessToken == "" {
		return fmt.Errorf("no access token received")
	}

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Save token and user info
	cfg.AuthToken = authResp.AccessToken
	cfg.UserID = authResp.User.ID
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Authentication successful!")
	fmt.Printf("📧 Logged in as: %s\n", authResp.User.Email)
	fmt.Printf("👤 User ID: %s\n", authResp.User.ID)
	fmt.Printf("📁 Config saved to: %s\n", config.GetConfigPath())
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Add MCP servers: mcp-client add <name> --path <server-path>")
	fmt.Println("2. Start client: mcp-client start")

	return nil
}
