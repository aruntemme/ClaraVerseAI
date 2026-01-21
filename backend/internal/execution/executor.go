package execution

import (
	"claraverse/internal/models"
	"claraverse/internal/services"
	"claraverse/internal/tools"
	"context"
	"fmt"
)

// BlockExecutor interface for all block types
type BlockExecutor interface {
	Execute(ctx context.Context, block models.Block, inputs map[string]any) (map[string]any, error)
}

// ExecutorRegistry maps block types to executors
type ExecutorRegistry struct {
	executors map[string]BlockExecutor
}

// NewExecutorRegistry creates a new executor registry with all block type executors
// Hybrid Architecture: Supports variable, llm_inference, and code_block types.
// - variable: Input/output data handling
// - llm_inference: AI reasoning with tool access
// - code_block: Direct tool execution (no LLM, faster & deterministic)
func NewExecutorRegistry(
	chatService *services.ChatService,
	providerService *services.ProviderService,
	toolRegistry *tools.Registry,
	credentialService *services.CredentialService,
) *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: map[string]BlockExecutor{
			// Variable blocks handle input/output data
			"variable": NewVariableExecutor(),
			// LLM blocks handle all intelligent actions via tools
			// Tools available: search_web, scrape_web, send_webhook, send_discord_message, send_slack_message, etc.
			"llm_inference": NewAgentBlockExecutor(chatService, providerService, toolRegistry, credentialService),
			// Code blocks execute tools directly without LLM (faster, deterministic)
			// Use for mechanical tasks that don't need AI reasoning
			"code_block": NewToolExecutor(toolRegistry, credentialService),
		},
	}
}

// Get retrieves an executor for a block type
func (r *ExecutorRegistry) Get(blockType string) (BlockExecutor, error) {
	exec, ok := r.executors[blockType]
	if !ok {
		return nil, fmt.Errorf("no executor registered for block type: %s", blockType)
	}
	return exec, nil
}

// Register adds a new executor for a block type
func (r *ExecutorRegistry) Register(blockType string, executor BlockExecutor) {
	r.executors[blockType] = executor
}
