package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"claraverse/internal/database"
	"claraverse/internal/models"
	"claraverse/internal/tools"
	"github.com/google/uuid"
)

// MCPBridgeService manages MCP client connections and tool routing
type MCPBridgeService struct {
	db          *database.DB
	connections map[string]*models.MCPConnection // clientID -> connection
	userConns   map[string]string                // userID -> clientID
	registry    *tools.Registry
	mutex       sync.RWMutex
}

// NewMCPBridgeService creates a new MCP bridge service
func NewMCPBridgeService(db *database.DB, registry *tools.Registry) *MCPBridgeService {
	return &MCPBridgeService{
		db:          db,
		connections: make(map[string]*models.MCPConnection),
		userConns:   make(map[string]string),
		registry:    registry,
	}
}

// RegisterClient registers a new MCP client connection
func (s *MCPBridgeService) RegisterClient(userID string, registration *models.MCPToolRegistration) (*models.MCPConnection, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if user already has a connection
	if existingClientID, exists := s.userConns[userID]; exists {
		// Disconnect existing connection
		if existingConn, ok := s.connections[existingClientID]; ok {
			log.Printf("Disconnecting existing MCP client for user %s", userID)
			s.disconnectClientLocked(existingClientID, existingConn)
		}
	}

	// Create new connection
	conn := &models.MCPConnection{
		ID:             uuid.New().String(),
		UserID:         userID,
		ClientID:       registration.ClientID,
		ClientVersion:  registration.ClientVersion,
		Platform:       registration.Platform,
		ConnectedAt:    time.Now(),
		LastHeartbeat:  time.Now(),
		IsActive:       true,
		Tools:          registration.Tools,
		WriteChan:      make(chan models.MCPServerMessage, 100),
		StopChan:       make(chan bool, 1),
		PendingResults: make(map[string]chan models.MCPToolResult),
	}

	// Store in memory
	s.connections[registration.ClientID] = conn
	s.userConns[userID] = registration.ClientID

	// Store in database
	_, err := s.db.Exec(`
		INSERT INTO mcp_connections (user_id, client_id, client_version, platform, connected_at, last_heartbeat, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, registration.ClientID, registration.ClientVersion, registration.Platform, conn.ConnectedAt, conn.LastHeartbeat, true)

	if err != nil {
		delete(s.connections, registration.ClientID)
		delete(s.userConns, userID)
		return nil, fmt.Errorf("failed to store connection in database: %w", err)
	}

	// Get connection ID from database
	var dbConnID int64
	err = s.db.QueryRow("SELECT id FROM mcp_connections WHERE client_id = ?", registration.ClientID).Scan(&dbConnID)
	if err != nil {
		log.Printf("Warning: Failed to get connection ID from database: %v", err)
	}

	// Register tools in registry and database
	for _, tool := range registration.Tools {
		// Register in registry
		err := s.registry.RegisterUserTool(userID, &tools.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
			Source:      tools.ToolSourceMCPLocal,
			UserID:      userID,
			Execute:     nil, // MCP tools don't have direct execute functions
		})

		if err != nil {
			log.Printf("Warning: Failed to register tool %s: %v", tool.Name, err)
			continue
		}

		// Store tool in database
		toolDefJSON, _ := json.Marshal(tool)
		_, err = s.db.Exec(`
			INSERT OR REPLACE INTO mcp_tools (user_id, connection_id, tool_name, tool_definition)
			VALUES (?, ?, ?, ?)
		`, userID, dbConnID, tool.Name, string(toolDefJSON))

		if err != nil {
			log.Printf("Warning: Failed to store tool %s in database: %v", tool.Name, err)
		}
	}

	log.Printf("✅ MCP client registered: user=%s, client=%s, tools=%d", userID, registration.ClientID, len(registration.Tools))

	// Send acknowledgment
	go func() {
		conn.WriteChan <- models.MCPServerMessage{
			Type: "ack",
			Payload: map[string]interface{}{
				"status":          "connected",
				"tools_registered": len(registration.Tools),
			},
		}
	}()

	return conn, nil
}

// DisconnectClient handles client disconnection
func (s *MCPBridgeService) DisconnectClient(clientID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, exists := s.connections[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	s.disconnectClientLocked(clientID, conn)
	return nil
}

// disconnectClientLocked handles disconnection (must be called with lock held)
func (s *MCPBridgeService) disconnectClientLocked(clientID string, conn *models.MCPConnection) {
	// Mark as inactive in database
	_, err := s.db.Exec("UPDATE mcp_connections SET is_active = 0 WHERE client_id = ?", clientID)
	if err != nil {
		log.Printf("Warning: Failed to mark connection as inactive: %v", err)
	}

	// Unregister all tools
	s.registry.UnregisterAllUserTools(conn.UserID)

	// Clean up memory
	delete(s.connections, clientID)
	delete(s.userConns, conn.UserID)

	// Close channels
	close(conn.StopChan)
	close(conn.WriteChan)

	log.Printf("🔌 MCP client disconnected: user=%s, client=%s", conn.UserID, clientID)
}

// UpdateHeartbeat updates the last heartbeat time for a client
func (s *MCPBridgeService) UpdateHeartbeat(clientID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	conn, exists := s.connections[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	conn.LastHeartbeat = time.Now()

	// Update in database
	_, err := s.db.Exec("UPDATE mcp_connections SET last_heartbeat = ? WHERE client_id = ?", conn.LastHeartbeat, clientID)
	return err
}

// ExecuteToolOnClient sends a tool execution request to the MCP client
func (s *MCPBridgeService) ExecuteToolOnClient(userID string, toolName string, args map[string]interface{}, timeout time.Duration) (string, error) {
	s.mutex.RLock()
	clientID, exists := s.userConns[userID]
	if !exists {
		s.mutex.RUnlock()
		return "", fmt.Errorf("no MCP client connected for user %s", userID)
	}

	conn, connExists := s.connections[clientID]
	s.mutex.RUnlock()

	if !connExists {
		return "", fmt.Errorf("MCP client connection not found")
	}

	// Generate unique call ID
	callID := uuid.New().String()

	// Create result channel for this call
	resultChan := make(chan models.MCPToolResult, 1)
	conn.PendingResults[callID] = resultChan

	// Create tool call message
	toolCall := models.MCPToolCall{
		CallID:    callID,
		ToolName:  toolName,
		Arguments: args,
		Timeout:   int(timeout.Seconds()),
	}

	// Send to client
	select {
	case conn.WriteChan <- models.MCPServerMessage{
		Type: "tool_call",
		Payload: map[string]interface{}{
			"call_id":   toolCall.CallID,
			"tool_name": toolCall.ToolName,
			"arguments": toolCall.Arguments,
			"timeout":   toolCall.Timeout,
		},
	}:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		delete(conn.PendingResults, callID)
		return "", fmt.Errorf("timeout sending tool call to client")
	}

	// Wait for result with timeout
	select {
	case result := <-resultChan:
		delete(conn.PendingResults, callID)
		if result.Success {
			return result.Result, nil
		} else {
			return "", fmt.Errorf("%s", result.Error)
		}
	case <-time.After(timeout):
		delete(conn.PendingResults, callID)
		return "", fmt.Errorf("tool execution timeout after %v", timeout)
	}
}

// GetConnection retrieves a connection by client ID
func (s *MCPBridgeService) GetConnection(clientID string) (*models.MCPConnection, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	conn, exists := s.connections[clientID]
	return conn, exists
}

// GetUserConnection retrieves a connection by user ID
func (s *MCPBridgeService) GetUserConnection(userID string) (*models.MCPConnection, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clientID, exists := s.userConns[userID]
	if !exists {
		return nil, false
	}

	conn, connExists := s.connections[clientID]
	return conn, connExists
}

// IsUserConnected checks if a user has an active MCP client
func (s *MCPBridgeService) IsUserConnected(userID string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.userConns[userID]
	return exists
}

// GetConnectionCount returns the number of active connections
func (s *MCPBridgeService) GetConnectionCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.connections)
}

// LogToolExecution logs a tool execution for audit purposes
func (s *MCPBridgeService) LogToolExecution(userID, toolName, conversationID string, executionTimeMs int, success bool, errorMsg string) {
	_, err := s.db.Exec(`
		INSERT INTO mcp_audit_log (user_id, tool_name, conversation_id, execution_time_ms, success, error_message)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, toolName, conversationID, executionTimeMs, success, errorMsg)

	if err != nil {
		log.Printf("Warning: Failed to log tool execution: %v", err)
	}
}
