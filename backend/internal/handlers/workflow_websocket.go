package handlers

import (
	"claraverse/internal/execution"
	"claraverse/internal/middleware"
	"claraverse/internal/models"
	"claraverse/internal/services"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WorkflowWebSocketHandler handles WebSocket connections for workflow execution
type WorkflowWebSocketHandler struct {
	agentService      *services.AgentService
	executionService  *services.ExecutionService
	workflowEngine    *execution.WorkflowEngine
	executionLimiter  *middleware.ExecutionLimiter
}

// NewWorkflowWebSocketHandler creates a new workflow WebSocket handler
func NewWorkflowWebSocketHandler(
	agentService *services.AgentService,
	workflowEngine *execution.WorkflowEngine,
	executionLimiter *middleware.ExecutionLimiter,
) *WorkflowWebSocketHandler {
	return &WorkflowWebSocketHandler{
		agentService:     agentService,
		workflowEngine:   workflowEngine,
		executionLimiter: executionLimiter,
	}
}

// SetExecutionService sets the execution service (optional, for MongoDB execution tracking)
func (h *WorkflowWebSocketHandler) SetExecutionService(svc *services.ExecutionService) {
	h.executionService = svc
}

// WorkflowClientMessage represents a message from the client
type WorkflowClientMessage struct {
	Type    string         `json:"type"` // execute_workflow, cancel_execution
	AgentID string         `json:"agent_id,omitempty"`
	Input   map[string]any `json:"input,omitempty"`

	// EnableBlockChecker enables block completion validation (optional)
	// When true, each block is checked to ensure it accomplished its job
	EnableBlockChecker bool `json:"enable_block_checker,omitempty"`

	// CheckerModelID is the model to use for block checking (optional)
	// Defaults to gpt-4o-mini for fast, cheap validation
	CheckerModelID string `json:"checker_model_id,omitempty"`
}

// WorkflowServerMessage represents a message to send to the client
type WorkflowServerMessage struct {
	Type        string         `json:"type"` // connected, execution_started, execution_update, execution_complete, error
	ExecutionID string         `json:"execution_id,omitempty"`
	BlockID     string         `json:"block_id,omitempty"`
	Status      string         `json:"status,omitempty"`
	Inputs      map[string]any `json:"inputs,omitempty"`
	Output      map[string]any `json:"output,omitempty"`
	FinalOutput map[string]any `json:"final_output,omitempty"`
	Duration    int64          `json:"duration_ms,omitempty"`
	Error       string         `json:"error,omitempty"`

	// APIResponse is the standardized, clean response for API consumers
	// This provides a well-structured output with result, artifacts, files, etc.
	APIResponse *models.ExecutionAPIResponse `json:"api_response,omitempty"`
}

// Handle handles a new WebSocket connection for workflow execution
func (h *WorkflowWebSocketHandler) Handle(c *websocket.Conn) {
	userID := c.Locals("user_id").(string)
	connID := uuid.New().String()

	log.Printf("🔌 [WORKFLOW-WS] New connection: connID=%s, userID=%s", connID, userID)

	// Send connected message
	if err := c.WriteJSON(WorkflowServerMessage{
		Type: "connected",
	}); err != nil {
		log.Printf("❌ [WORKFLOW-WS] Failed to send connected message: %v", err)
		return
	}

	// Context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Read loop
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("❌ [WORKFLOW-WS] Read error for %s: %v", connID, err)
			break
		}

		var clientMsg WorkflowClientMessage
		if err := json.Unmarshal(msg, &clientMsg); err != nil {
			log.Printf("⚠️ [WORKFLOW-WS] Invalid message format from %s: %v", connID, err)
			c.WriteJSON(WorkflowServerMessage{
				Type:  "error",
				Error: "Invalid message format",
			})
			continue
		}

		switch clientMsg.Type {
		case "execute_workflow":
			h.handleExecuteWorkflow(ctx, c, userID, clientMsg)
		case "cancel_execution":
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
		default:
			log.Printf("⚠️ [WORKFLOW-WS] Unknown message type: %s", clientMsg.Type)
		}
	}
}

// handleExecuteWorkflow handles a workflow execution request
func (h *WorkflowWebSocketHandler) handleExecuteWorkflow(
	ctx context.Context,
	c *websocket.Conn,
	userID string,
	msg WorkflowClientMessage,
) {
	startTime := time.Now()

	log.Printf("🔍 [WORKFLOW-WS] Received execute request: AgentID=%s, Input=%+v", msg.AgentID, msg.Input)

	// Check daily execution limit
	if h.executionLimiter != nil {
		remaining, err := h.executionLimiter.GetRemainingExecutions(userID)
		if err != nil {
			log.Printf("⚠️  [WORKFLOW-WS] Failed to check execution limit: %v", err)
			// Continue on error, don't block execution
		} else if remaining == 0 {
			log.Printf("⚠️  [WORKFLOW-WS] User %s exceeded daily execution limit", userID)
			c.WriteJSON(WorkflowServerMessage{
				Type:  "error",
				Error: "Daily execution limit exceeded. Please upgrade your plan or wait until tomorrow.",
			})
			return
		} else if remaining > 0 {
			log.Printf("✅ [WORKFLOW-WS] User %s has %d executions remaining today", userID, remaining)
		}
	}

	// Get agent and workflow
	agent, err := h.agentService.GetAgent(msg.AgentID, userID)
	if err != nil {
		log.Printf("❌ [WORKFLOW-WS] Agent not found: %s", msg.AgentID)
		c.WriteJSON(WorkflowServerMessage{
			Type:  "error",
			Error: "Agent not found: " + err.Error(),
		})
		return
	}

	if agent.Workflow == nil {
		log.Printf("❌ [WORKFLOW-WS] No workflow for agent: %s", msg.AgentID)
		c.WriteJSON(WorkflowServerMessage{
			Type:  "error",
			Error: "Agent has no workflow defined",
		})
		return
	}

	// Create execution record using ExecutionService (MongoDB) if available
	var execID string
	var execObjectID primitive.ObjectID

	if h.executionService != nil {
		execRecord, err := h.executionService.Create(ctx, &services.CreateExecutionRequest{
			AgentID:         msg.AgentID,
			UserID:          userID,
			WorkflowVersion: agent.Workflow.Version,
			TriggerType:     "manual",
			Input:           msg.Input,
		})
		if err != nil {
			log.Printf("❌ [WORKFLOW-WS] Failed to create execution: %v", err)
			c.WriteJSON(WorkflowServerMessage{
				Type:  "error",
				Error: "Failed to create execution: " + err.Error(),
			})
			return
		}
		execID = execRecord.ID.Hex()
		execObjectID = execRecord.ID
	} else {
		// Fallback: generate a local ID if ExecutionService is not available
		execID = uuid.New().String()
		log.Printf("⚠️ [WORKFLOW-WS] ExecutionService not available, using local ID: %s", execID)
	}

	log.Printf("🚀 [WORKFLOW-WS] Starting execution %s for agent %s", execID, msg.AgentID)

	// Send execution started message
	c.WriteJSON(WorkflowServerMessage{
		Type:        "execution_started",
		ExecutionID: execID,
	})

	// Increment execution counter for today
	if h.executionLimiter != nil {
		if err := h.executionLimiter.IncrementCount(userID); err != nil {
			log.Printf("⚠️  [WORKFLOW-WS] Failed to increment execution count: %v", err)
			// Don't fail the execution if counter increment fails
		}
	}

	// Create status channel
	statusChan := make(chan models.ExecutionUpdate, 100)

	// Start goroutine to forward status updates to WebSocket
	go func() {
		for update := range statusChan {
			update.ExecutionID = execID
			c.WriteJSON(WorkflowServerMessage{
				Type:        "execution_update",
				ExecutionID: execID,
				BlockID:     update.BlockID,
				Status:      update.Status,
				Inputs:      update.Inputs,
				Output:      update.Output,
				Error:       update.Error,
			})
		}
	}()

	// Inject user context into input for credential resolution and tool execution
	if msg.Input == nil {
		msg.Input = make(map[string]interface{})
	}
	msg.Input["__user_id__"] = userID

	// Build execution options - block checker is controlled by client request
	// When enabled, it validates that each block actually accomplished its job
	execOptions := &execution.ExecutionOptions{
		WorkflowGoal:       agent.Description,      // Use agent description as workflow goal
		EnableBlockChecker: msg.EnableBlockChecker, // Controlled by frontend toggle
		CheckerModelID:     msg.CheckerModelID,
	}
	if msg.EnableBlockChecker {
		log.Printf("🔍 [WORKFLOW-WS] Block checker ENABLED (model: %s)", execOptions.CheckerModelID)
	} else {
		log.Printf("🔍 [WORKFLOW-WS] Block checker DISABLED")
	}

	// Execute workflow
	log.Printf("🔍 [WORKFLOW-WS] Executing with input: %+v", msg.Input)
	result, err := h.workflowEngine.ExecuteWithOptions(ctx, agent.Workflow, msg.Input, statusChan, execOptions)
	close(statusChan)

	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		log.Printf("❌ [WORKFLOW-WS] Execution failed: %v", err)

		// Update execution status using ExecutionService if available
		if h.executionService != nil {
			h.executionService.Complete(ctx, execObjectID, &services.ExecutionCompleteRequest{
				Status: "failed",
				Error:  err.Error(),
			})
		}

		c.WriteJSON(WorkflowServerMessage{
			Type:        "execution_complete",
			ExecutionID: execID,
			Status:      "failed",
			Duration:    duration,
			Error:       err.Error(),
		})
		return
	}

	// Build the standardized API response
	apiResponse := h.workflowEngine.BuildAPIResponse(result, agent.Workflow, execID, duration)
	apiResponse.Metadata.AgentID = msg.AgentID

	// Update execution status in database using ExecutionService if available
	if h.executionService != nil {
		h.executionService.Complete(ctx, execObjectID, &services.ExecutionCompleteRequest{
			Status:      result.Status,
			Output:      result.Output,
			BlockStates: result.BlockStates,
			Error:       result.Error,
			// Store clean API response fields
			Result:    apiResponse.Result,
			Artifacts: apiResponse.Artifacts,
			Files:     apiResponse.Files,
		})
	}

	log.Printf("✅ [WORKFLOW-WS] Execution %s completed: status=%s, duration=%dms, result=%d chars",
		execID, result.Status, duration, len(apiResponse.Result))

	// Send completion message with both legacy and new API response format
	c.WriteJSON(WorkflowServerMessage{
		Type:        "execution_complete",
		ExecutionID: execID,
		Status:      result.Status,
		FinalOutput: result.Output,      // Legacy format (backward compat)
		Duration:    duration,
		Error:       result.Error,
		APIResponse: apiResponse,         // New standardized format
	})
}
