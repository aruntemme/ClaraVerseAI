package bridge

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// ToolCall represents a tool execution request from backend
type ToolCall struct {
	CallID    string                 `json:"call_id"`
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	Timeout   int                    `json:"timeout"`
}

// Bridge manages the WebSocket connection to the backend
type Bridge struct {
	backendURL     string
	authToken      string
	conn           *websocket.Conn
	writeChan      chan Message
	stopChan       chan struct{}
	reconnectDelay time.Duration
	maxReconnect   time.Duration
	connected      bool
	mutex          sync.RWMutex
	onToolCall     func(ToolCall)
	verbose        bool
}

// NewBridge creates a new WebSocket bridge
func NewBridge(backendURL, authToken string, verbose bool) *Bridge {
	return &Bridge{
		backendURL:     backendURL,
		authToken:      authToken,
		writeChan:      make(chan Message, 100),
		stopChan:       make(chan struct{}),
		reconnectDelay: 1 * time.Second,
		maxReconnect:   60 * time.Second,
		verbose:        verbose,
	}
}

// SetToolCallHandler sets the callback for tool call events
func (b *Bridge) SetToolCallHandler(handler func(ToolCall)) {
	b.onToolCall = handler
}

// Connect establishes the WebSocket connection
func (b *Bridge) Connect() error {
	url := fmt.Sprintf("%s?token=%s", b.backendURL, b.authToken)

	if b.verbose {
		log.Printf("[Bridge] Connecting to %s", b.backendURL)
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	b.mutex.Lock()
	b.conn = conn
	b.connected = true
	b.reconnectDelay = 1 * time.Second // Reset reconnect delay on successful connection
	b.mutex.Unlock()

	log.Println("✅ Connected to backend")

	// Start read and write loops
	go b.readLoop()
	go b.writeLoop()

	return nil
}

// ConnectWithRetry connects with automatic retry and exponential backoff
func (b *Bridge) ConnectWithRetry() {
	attempt := 0
	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		err := b.Connect()
		if err == nil {
			return
		}

		attempt++
		log.Printf("❌ Connection failed (attempt %d): %v", attempt, err)
		log.Printf("🔄 Retrying in %v...", b.reconnectDelay)

		time.Sleep(b.reconnectDelay)

		// Exponential backoff
		b.reconnectDelay = time.Duration(math.Min(
			float64(b.reconnectDelay*2),
			float64(b.maxReconnect),
		))
	}
}

// readLoop handles incoming messages
func (b *Bridge) readLoop() {
	defer func() {
		b.handleDisconnect()
	}()

	for {
		var msg Message
		err := b.conn.ReadJSON(&msg)
		if err != nil {
			if b.verbose {
				log.Printf("[Bridge] Read error: %v", err)
			}
			return
		}

		b.handleMessage(msg)
	}
}

// writeLoop handles outgoing messages
func (b *Bridge) writeLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-b.writeChan:
			err := b.conn.WriteJSON(msg)
			if err != nil {
				if b.verbose {
					log.Printf("[Bridge] Write error: %v", err)
				}
				return
			}

		case <-ticker.C:
			// Send heartbeat
			if err := b.SendHeartbeat(); err != nil {
				return
			}

		case <-b.stopChan:
			return
		}
	}
}

// handleMessage processes incoming messages
func (b *Bridge) handleMessage(msg Message) {
	if b.verbose {
		log.Printf("[Bridge] Received: %s", msg.Type)
	}

	switch msg.Type {
	case "ack":
		log.Printf("✅ Registration acknowledged")
		if status, ok := msg.Payload["status"].(string); ok {
			log.Printf("   Status: %s", status)
		}
		if toolsReg, ok := msg.Payload["tools_registered"].(float64); ok {
			log.Printf("   Tools registered: %.0f", toolsReg)
		}

	case "tool_call":
		// Parse tool call
		callID := msg.Payload["call_id"].(string)
		toolName := msg.Payload["tool_name"].(string)
		args, _ := msg.Payload["arguments"].(map[string]interface{})
		timeout, _ := msg.Payload["timeout"].(float64)

		toolCall := ToolCall{
			CallID:    callID,
			ToolName:  toolName,
			Arguments: args,
			Timeout:   int(timeout),
		}

		log.Printf("🔧 Tool call: %s (call_id: %s)", toolName, callID)

		// Call handler if set
		if b.onToolCall != nil {
			b.onToolCall(toolCall)
		}

	case "error":
		errMsg := msg.Payload["message"].(string)
		log.Printf("❌ Error from backend: %s", errMsg)

	default:
		if b.verbose {
			log.Printf("[Bridge] Unknown message type: %s", msg.Type)
		}
	}
}

// handleDisconnect handles disconnection and reconnection
func (b *Bridge) handleDisconnect() {
	b.mutex.Lock()
	b.connected = false
	if b.conn != nil {
		b.conn.Close()
	}
	b.mutex.Unlock()

	log.Println("🔌 Disconnected from backend")
	log.Println("🔄 Attempting to reconnect...")

	// Reconnect with exponential backoff
	b.ConnectWithRetry()
}

// RegisterTools sends tool registration message
func (b *Bridge) RegisterTools(clientID, clientVersion, platform string, tools []interface{}) error {
	msg := Message{
		Type: "register_tools",
		Payload: map[string]interface{}{
			"client_id":      clientID,
			"client_version": clientVersion,
			"platform":       platform,
			"tools":          tools,
		},
	}

	b.writeChan <- msg
	return nil
}

// SendToolResult sends tool execution result back to backend
func (b *Bridge) SendToolResult(callID string, success bool, result, errorMsg string) error {
	msg := Message{
		Type: "tool_result",
		Payload: map[string]interface{}{
			"call_id": callID,
			"success": success,
			"result":  result,
			"error":   errorMsg,
		},
	}

	b.writeChan <- msg
	return nil
}

// SendHeartbeat sends a heartbeat message
func (b *Bridge) SendHeartbeat() error {
	msg := Message{
		Type: "heartbeat",
		Payload: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	b.writeChan <- msg
	return nil
}

// Close gracefully closes the bridge
func (b *Bridge) Close() error {
	// Send disconnect message
	msg := Message{
		Type:    "disconnect",
		Payload: map[string]interface{}{},
	}
	b.writeChan <- msg

	// Wait a bit for message to send
	time.Sleep(100 * time.Millisecond)

	close(b.stopChan)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.conn != nil {
		return b.conn.Close()
	}

	return nil
}

// IsConnected returns whether the bridge is currently connected
func (b *Bridge) IsConnected() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.connected
}
