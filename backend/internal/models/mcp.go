package models

import "time"

// MCPConnection represents an active MCP client connection
type MCPConnection struct {
	ID             string                       `json:"id"`
	UserID         string                       `json:"user_id"`
	ClientID       string                       `json:"client_id"`
	ClientVersion  string                       `json:"client_version"`
	Platform       string                       `json:"platform"`
	ConnectedAt    time.Time                    `json:"connected_at"`
	LastHeartbeat  time.Time                    `json:"last_heartbeat"`
	IsActive       bool                         `json:"is_active"`
	Tools          []MCPTool                    `json:"tools"`
	WriteChan      chan MCPServerMessage        `json:"-"`
	StopChan       chan bool                    `json:"-"`
	PendingResults map[string]chan MCPToolResult `json:"-"` // call_id -> result channel
}

// MCPTool represents a tool registered by an MCP client
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
	Source      string                 `json:"source"`     // "mcp_local"
	UserID      string                 `json:"user_id"`
}

// MCPClientMessage represents messages from MCP client to backend
type MCPClientMessage struct {
	Type    string                 `json:"type"` // "register_tools", "tool_result", "heartbeat", "disconnect"
	Payload map[string]interface{} `json:"payload"`
}

// MCPServerMessage represents messages from backend to MCP client
type MCPServerMessage struct {
	Type    string                 `json:"type"` // "tool_call", "ack", "error"
	Payload map[string]interface{} `json:"payload"`
}

// MCPToolRegistration represents the registration payload from client
type MCPToolRegistration struct {
	ClientID      string    `json:"client_id"`
	ClientVersion string    `json:"client_version"`
	Platform      string    `json:"platform"`
	Tools         []MCPTool `json:"tools"`
}

// MCPToolCall represents a tool execution request to client
type MCPToolCall struct {
	CallID    string                 `json:"call_id"`
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	Timeout   int                    `json:"timeout"` // seconds
}

// MCPToolResult represents a tool execution result from client
type MCPToolResult struct {
	CallID  string `json:"call_id"`
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error,omitempty"`
}

// MCPHeartbeat represents a heartbeat message
type MCPHeartbeat struct {
	Timestamp time.Time `json:"timestamp"`
}
