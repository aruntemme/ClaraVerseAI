package models

// Config represents API configuration (legacy, kept for backward compatibility)
type Config struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}
