/**
 * Agent Builder Type Definitions
 *
 * This file contains all type definitions for the Agent Builder feature.
 * Agents are workflow automations created through natural language conversation.
 */

// =============================================================================
// Agent Types
// =============================================================================

export type AgentStatus = 'draft' | 'deployed' | 'paused';

// Sync status for backend-first architecture
export type SyncStatus = 'local-only' | 'syncing' | 'synced' | 'error';

export interface Agent {
  id: string;
  userId: string;
  name: string;
  description: string;
  workflow: Workflow;
  status: AgentStatus;
  apiKey?: string; // For webhook authentication
  createdAt: Date;
  updatedAt: Date;
  // Backend-first persistence
  syncStatus: SyncStatus;
  lastSyncError?: string;
}

// =============================================================================
// Workflow Types
// =============================================================================

export interface Workflow {
  id: string;
  blocks: Block[];
  connections: Connection[];
  variables: WorkflowVariable[];
  version: number;
}

// Variable types supported in workflows
export type VariableType = 'string' | 'number' | 'boolean' | 'array' | 'object' | 'file';

export interface WorkflowVariable {
  name: string;
  type: VariableType;
  defaultValue?: unknown;
}

// =============================================================================
// File Reference Types
// =============================================================================

/**
 * Represents a file that can be passed between workflow blocks.
 * Files are uploaded to the backend and referenced by fileId.
 */
export interface FileReference {
  fileId: string;
  filename: string;
  mimeType: string;
  size: number;
  type: FileType;
}

/**
 * File type categories for workflow files.
 * - image: JPEG, PNG, GIF, WebP, SVG
 * - document: PDF, DOCX, PPTX
 * - audio: MP3, WAV, M4A, OGG, FLAC, WebM
 * - data: CSV, JSON, Excel, plain text
 */
export type FileType = 'image' | 'document' | 'audio' | 'data';

/**
 * File attachment for LLM blocks with vision support.
 * Currently only images are supported for OpenAI vision models.
 */
export interface FileAttachment {
  fileId: string;
  type: FileType;
}

// =============================================================================
// Block Types
// =============================================================================

// Hybrid Architecture: Supports variable, llm_inference, and code_block types.
// - variable: Input/output data handling
// - llm_inference: AI reasoning with tool access (for tasks needing decisions/interpretation)
// - code_block: Direct tool execution without LLM (faster, for mechanical tasks)
export type BlockType = 'llm_inference' | 'variable' | 'code_block';

export interface Block {
  id: string;
  normalizedId: string; // Normalized name for variable interpolation (e.g., "search-latest-news")
  type: BlockType;
  name: string;
  description: string;
  config: BlockConfig;
  position: { x: number; y: number };
  timeout: number; // Default 30, max 60 seconds
}

// Union type for block configs (Hybrid: LLM, Variable, and Code blocks)
export type BlockConfig = LLMInferenceConfig | VariableConfig | CodeBlockConfig;

// Deprecated config types (kept for backwards compatibility with existing workflows)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
type DeprecatedBlockConfig = ToolExecutionConfig | WebhookConfig | PythonToolConfig;

// Structured output metadata for models
export interface StructuredOutputMetadata {
  support: 'excellent' | 'good' | 'fair' | 'poor' | 'unknown';
  compliance?: number; // 0-100
  speed_ms?: number;
  badge?: 'FASTEST' | 'RECOMMENDED' | 'BETA';
  warning?: string;
}

export interface LLMInferenceConfig {
  type: 'llm_inference';
  modelId: string;
  systemPrompt: string;
  userPromptTemplate: string; // Supports {{input.fieldName}} interpolation
  temperature: number;
  maxTokens: number;
  enabledTools: string[]; // Tool IDs to enable (e.g., ["search_web", "send_discord_message"])
  credentials?: string[]; // Credential IDs selected by user for tool authentication
  outputFormat: 'text' | 'json';
  outputSchema?: Record<string, unknown>; // JSON Schema for structured output
  // Vision support: file attachments for OpenAI vision models
  attachments?: FileAttachment[];
}

export interface ToolExecutionConfig {
  type: 'tool_execution';
  toolId: string;
  argumentMapping: Record<string, string>; // Maps input ports to tool args
}

/**
 * Configuration for code_block - direct tool execution without LLM
 * Use when the task is purely mechanical and doesn't need AI reasoning.
 * Example: sending a pre-formatted message, getting current time, making API calls with known params
 */
export interface CodeBlockConfig {
  type: 'code_block';
  toolName: string; // The tool to execute (e.g., "send_discord_message", "get_current_time")
  argumentMapping: Record<string, string>; // Maps {{variables}} to tool arguments
}

export interface WebhookConfig {
  type: 'webhook';
  url: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  headers: Record<string, string>;
  bodyTemplate: string; // Supports {{variable}} and {{input.fieldName}}
  authType?: 'none' | 'bearer' | 'basic' | 'api_key';
  authConfig?: Record<string, string>;
}

export interface VariableConfig {
  type: 'variable';
  operation: 'set' | 'read';
  variableName: string;
  valueExpression?: string;
  defaultValue?: string;
  // Input type support for Start blocks
  inputType?: 'text' | 'file' | 'json'; // Default is 'text'
  fileValue?: FileReference | null; // File reference when inputType is 'file'
  acceptedFileTypes?: FileType[]; // Restrict file types (e.g., ['image'] for vision)
  // JSON input support
  jsonValue?: Record<string, unknown> | null; // Parsed JSON when inputType is 'json'
  jsonSchema?: Record<string, unknown> | null; // Optional JSON schema for validation/UI hints
}

export interface PythonToolConfig {
  type: 'python_tool';
  toolId: string; // Reference to custom Python tool
  argumentMapping: Record<string, string>;
}

// =============================================================================
// Connection Types
// =============================================================================

export interface Connection {
  id: string;
  sourceBlockId: string;
  sourceOutput: string; // Output port name
  targetBlockId: string;
  targetInput: string; // Named input: "fromWeatherAPI", "fromNewsAPI"
}

// =============================================================================
// Execution Types
// =============================================================================

export type ExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'partial_failure';

export type BlockExecutionStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';

export interface ExecutionContext {
  executionId: string;
  agentId: string;
  status: ExecutionStatus;
  blockStates: Record<string, BlockExecutionState>;
  variables: Record<string, unknown>;
  startedAt: Date;
  completedAt?: Date;
}

export interface BlockExecutionState {
  blockId: string;
  status: BlockExecutionStatus;
  inputs: Record<string, unknown>;
  outputs: Record<string, unknown>;
  error?: string;
  startedAt?: Date;
  completedAt?: Date;
}

// =============================================================================
// Standardized API Response Types
// Clean, well-structured output for API consumers
// =============================================================================

/**
 * ExecutionAPIResponse is the standardized response for workflow execution.
 * This provides a clean, predictable structure for API consumers.
 */
export interface ExecutionAPIResponse {
  /** Status of the execution: completed, failed, partial */
  status: string;

  /** The primary output from the workflow - the "answer" */
  result: string;

  /** All generated charts, images, visualizations */
  artifacts?: APIArtifact[];

  /** All generated files with download URLs */
  files?: APIFile[];

  /** Detailed output from each block (for debugging/advanced use) */
  blocks?: Record<string, APIBlockOutput>;

  /** Execution statistics */
  metadata: ExecutionMetadata;

  /** Error message if status is failed */
  error?: string;
}

/** Generated artifact (chart, image, etc.) */
export interface APIArtifact {
  type: string; // "chart", "image", "plot"
  format: string; // "png", "jpeg", "svg"
  data: string; // Base64 encoded data
  title?: string; // Description/title
  source_block?: string; // Which block generated this
}

/** Generated file with download URL */
export interface APIFile {
  file_id?: string;
  filename: string;
  download_url: string;
  mime_type?: string;
  size?: number;
  source_block?: string;
}

/** Clean representation of a block's output */
export interface APIBlockOutput {
  name: string;
  type: string;
  status: string;
  response?: string; // Primary text output
  data?: Record<string, unknown>; // Structured data
  error?: string;
  duration_ms?: number;
}

/** Execution statistics */
export interface ExecutionMetadata {
  execution_id: string;
  agent_id?: string;
  workflow_version?: number;
  duration_ms: number;
  total_tokens?: number;
  blocks_executed: number;
  blocks_failed: number;
}

// =============================================================================
// Builder Chat Types
// =============================================================================

export interface BuilderMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  workflowUpdate?: WorkflowUpdate;
  executionResult?: ExecutionResultMessage;
}

export interface ExecutionResultMessage {
  status: 'completed' | 'failed' | 'partial_failure';
  result?: string;
  error?: string;
  blocksExecuted?: number;
  blocksFailed?: number;
}

export interface WorkflowUpdate {
  action: 'create' | 'modify';
  workflow: Workflow;
  explanation: string;
  validationErrors?: ValidationError[];
}

export interface ValidationError {
  type: 'schema' | 'cycle' | 'type_mismatch' | 'missing_input';
  message: string;
  blockId?: string;
  connectionId?: string;
}

// =============================================================================
// Custom Python Tool Types
// =============================================================================

export interface PythonTool {
  id: string;
  userId: string;
  name: string;
  displayName: string;
  description: string;
  icon: string; // Lucide icon name
  parameters: ToolParameter[];
  pythonCode: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface ToolParameter {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'array' | 'object';
  description: string;
  required: boolean;
  default?: unknown;
}

// =============================================================================
// Reconnection Types
// =============================================================================

export interface ReconnectPayload {
  currentState: ExecutionContext;
  recentEvents: ExecutionEvent[];
  missedEventCount: number;
}

export interface ExecutionEvent {
  type: 'block_started' | 'block_completed' | 'block_failed' | 'variable_set';
  timestamp: Date;
  data: unknown;
}

// =============================================================================
// WebSocket Message Types (Agent-specific)
// =============================================================================

// Client → Server
export interface WorkflowGenerateMessage {
  type: 'workflow_generate';
  conversationId: string;
  userMessage: string;
  currentWorkflow?: Workflow;
}

export interface AgentExecuteMessage {
  type: 'agent_execute';
  agentId: string;
  inputs?: Record<string, unknown>;
}

export interface BlockExecuteMessage {
  type: 'block_execute';
  agentId: string;
  blockId: string;
  mockInputs?: Record<string, unknown>;
}

// Server → Client
export interface WorkflowUpdateMessage {
  type: 'workflow_update';
  workflow: Workflow;
  explanation: string;
  validationErrors?: ValidationError[];
}

export interface BlockExecutionMessage {
  type: 'block_execution';
  executionId: string;
  blockId: string;
  status: BlockExecutionStatus;
  inputs?: Record<string, unknown>;
  outputs?: Record<string, unknown>;
  error?: string;
}

export interface ExecutionCompleteMessage {
  type: 'execution_complete';
  executionId: string;
  status: ExecutionStatus;
  blockStates: Record<string, BlockExecutionState>;
}

export interface ReconnectStateMessage {
  type: 'reconnect_state';
  payload: ReconnectPayload;
}

// =============================================================================
// UI State Types
// =============================================================================

export interface AgentBuilderUIState {
  selectedAgentId: string | null;
  selectedBlockId: string | null;
  settingsModalBlockId: string | null;
  isGenerating: boolean;
  isSidebarOpen: boolean;
  canvasZoom: number;
  canvasPosition: { x: number; y: number };
}

// =============================================================================
// Block Display Helpers
// =============================================================================

// Hybrid Architecture: Three block types
export const BLOCK_TYPE_INFO: Record<BlockType, { label: string; icon: string; color: string }> = {
  llm_inference: {
    label: 'AI Agent',
    icon: 'Brain',
    color: 'bg-purple-500',
  },
  variable: {
    label: 'Input',
    icon: 'Variable',
    color: 'bg-orange-500',
  },
  code_block: {
    label: 'Tool',
    icon: 'Wrench',
    color: 'bg-blue-500',
  },
};

export const BLOCK_STATUS_INFO: Record<BlockExecutionStatus, { label: string; color: string }> = {
  pending: { label: 'Pending', color: 'text-gray-400' },
  running: { label: 'Running', color: 'text-blue-500' },
  completed: { label: 'Completed', color: 'text-green-500' },
  failed: { label: 'Failed', color: 'text-red-500' },
  skipped: { label: 'Skipped', color: 'text-yellow-500' },
};
