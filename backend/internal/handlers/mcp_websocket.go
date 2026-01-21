package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"claraverse/internal/models"
	"claraverse/internal/services"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// MCPWebSocketHandler handles MCP client WebSocket connections
type MCPWebSocketHandler struct {
	mcpService *services.MCPBridgeService
}

// NewMCPWebSocketHandler creates a new MCP WebSocket handler
func NewMCPWebSocketHandler(mcpService *services.MCPBridgeService) *MCPWebSocketHandler {
	return &MCPWebSocketHandler{
		mcpService: mcpService,
	}
}

// HandleConnection handles incoming MCP client WebSocket connections
func (h *MCPWebSocketHandler) HandleConnection(c *websocket.Conn) {
	// Get user from fiber context (set by auth middleware)
	userID := c.Locals("user_id").(string)

	if userID == "" || userID == "anonymous" {
		log.Printf("❌ MCP connection rejected: no authenticated user")
		c.WriteJSON(fiber.Map{
			"type": "error",
			"payload": map[string]interface{}{
				"message": "Authentication required",
			},
		})
		c.Close()
		return
	}

	log.Printf("🔌 MCP client connecting: user=%s", userID)

	var mcpConn *models.MCPConnection
	var clientID string

	// Read loop
	for {
		var msg models.MCPClientMessage
		err := c.ReadJSON(&msg)
		if err != nil {
			if mcpConn != nil {
				log.Printf("MCP client disconnected: %v", err)
				h.mcpService.DisconnectClient(clientID)
			}
			break
		}

		switch msg.Type {
		case "register_tools":
			// Parse registration payload
			regData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Failed to marshal registration payload: %v", err)
				continue
			}

			var registration models.MCPToolRegistration
			err = json.Unmarshal(regData, &registration)
			if err != nil {
				log.Printf("Failed to unmarshal registration: %v", err)
				c.WriteJSON(models.MCPServerMessage{
					Type: "error",
					Payload: map[string]interface{}{
						"message": "Invalid registration format",
					},
				})
				continue
			}

			// Register client
			conn, err := h.mcpService.RegisterClient(userID, &registration)
			if err != nil {
				log.Printf("Failed to register MCP client: %v", err)
				c.WriteJSON(models.MCPServerMessage{
					Type: "error",
					Payload: map[string]interface{}{
						"message": fmt.Sprintf("Registration failed: %v", err),
					},
				})
				continue
			}

			mcpConn = conn
			clientID = registration.ClientID

			// Start write loop
			go h.writeLoop(c, conn)

			log.Printf("✅ MCP client registered successfully: user=%s, client=%s", userID, clientID)

		case "tool_result":
			// Handle tool execution result
			resultData, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("Failed to marshal tool result: %v", err)
				continue
			}

			var result models.MCPToolResult
			err = json.Unmarshal(resultData, &result)
			if err != nil {
				log.Printf("Failed to unmarshal tool result: %v", err)
				continue
			}

			// Log execution for audit
			execTime := 0 // We don't track this yet, but could add it
			h.mcpService.LogToolExecution(userID, "", "", execTime, result.Success, result.Error)

			log.Printf("Tool result received: call_id=%s, success=%v", result.CallID, result.Success)

			// Forward result to pending result channel
			if conn, exists := h.mcpService.GetConnection(clientID); exists {
				if resultChan, pending := conn.PendingResults[result.CallID]; pending {
					// Non-blocking send to result channel
					select {
					case resultChan <- result:
						log.Printf("✅ Tool result forwarded to waiting channel: %s", result.CallID)
					default:
						log.Printf("⚠️  Result channel full or closed for call_id: %s", result.CallID)
					}
				} else {
					log.Printf("⚠️  No pending result channel for call_id: %s", result.CallID)
				}
			}

		case "heartbeat":
			// Update heartbeat
			if clientID != "" {
				err := h.mcpService.UpdateHeartbeat(clientID)
				if err != nil {
					log.Printf("Failed to update heartbeat: %v", err)
				}
			}

		case "disconnect":
			// Client is gracefully disconnecting
			if clientID != "" {
				h.mcpService.DisconnectClient(clientID)
			}
			c.Close()
			return

		default:
			log.Printf("Unknown message type from MCP client: %s", msg.Type)
			c.WriteJSON(models.MCPServerMessage{
				Type: "error",
				Payload: map[string]interface{}{
					"message": "Unknown message type",
				},
			})
		}
	}
}

// writeLoop handles outgoing messages to the MCP client
func (h *MCPWebSocketHandler) writeLoop(c *websocket.Conn, conn *models.MCPConnection) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-conn.WriteChan:
			if !ok {
				// Channel closed
				return
			}

			err := c.WriteJSON(msg)
			if err != nil {
				log.Printf("Failed to write message to MCP client: %v", err)
				return
			}

		case <-conn.StopChan:
			// Stop signal received
			return

		case <-ticker.C:
			// Send ping to keep connection alive
			err := c.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("Failed to send ping to MCP client: %v", err)
				return
			}
		}
	}
}
