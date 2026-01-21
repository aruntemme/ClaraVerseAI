package vision

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Provider represents a minimal provider interface for vision
type Provider struct {
	ID      int
	Name    string
	BaseURL string
	APIKey  string
	Enabled bool
}

// ModelAlias represents a model alias with vision support info
type ModelAlias struct {
	DisplayName    string
	ActualModel    string
	SupportsVision *bool
}

// ProviderGetter is a function type to get provider by ID
type ProviderGetter func(id int) (*Provider, error)

// VisionModelFinder is a function type to find vision-capable models
type VisionModelFinder func() (providerID int, modelName string, err error)

// Service handles image analysis using vision-capable models
type Service struct {
	httpClient        *http.Client
	providerGetter    ProviderGetter
	visionModelFinder VisionModelFinder
	mu                sync.RWMutex
}

var (
	instance *Service
	once     sync.Once
)

// GetService returns the singleton vision service
// Note: Must call InitService first to set up dependencies
func GetService() *Service {
	return instance
}

// InitService initializes the vision service with dependencies
func InitService(providerGetter ProviderGetter, visionModelFinder VisionModelFinder) *Service {
	once.Do(func() {
		instance = &Service{
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
			providerGetter:    providerGetter,
			visionModelFinder: visionModelFinder,
		}
	})
	return instance
}

// DescribeImageRequest contains parameters for image description
type DescribeImageRequest struct {
	ImageData []byte
	MimeType  string
	Question  string // Optional question about the image
	Detail    string // "brief" or "detailed"
}

// DescribeImageResponse contains the result of image description
type DescribeImageResponse struct {
	Description string `json:"description"`
	Model       string `json:"model"`
	Provider    string `json:"provider"`
}

// DescribeImage analyzes an image and returns a text description
func (s *Service) DescribeImage(req *DescribeImageRequest) (*DescribeImageResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.visionModelFinder == nil || s.providerGetter == nil {
		return nil, fmt.Errorf("vision service not properly initialized")
	}

	log.Printf("🖼️ [VISION] Analyzing image (%d bytes, %s)", len(req.ImageData), req.MimeType)

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(req.ImageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", req.MimeType, base64Image)

	// Find a vision-capable model
	providerID, modelName, err := s.visionModelFinder()
	if err != nil {
		return nil, fmt.Errorf("no vision-capable model available: %w", err)
	}

	provider, err := s.providerGetter(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Build the prompt
	prompt := "Describe this image in detail."
	if req.Question != "" {
		prompt = req.Question
	} else if req.Detail == "brief" {
		prompt = "Briefly describe this image in 1-2 sentences."
	}

	// Build the API request
	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": prompt,
				},
				{
					"type": "image_url",
					"image_url": map[string]interface{}{
						"url":    dataURL,
						"detail": "auto",
					},
				},
			},
		},
	}

	// Detect if using OpenAI - they require max_completion_tokens instead of max_tokens
	isOpenAI := strings.Contains(strings.ToLower(provider.BaseURL), "openai.com")

	requestBody := map[string]interface{}{
		"model":    modelName,
		"messages": messages,
	}

	// Use correct token limit parameter based on provider
	if isOpenAI {
		requestBody["max_completion_tokens"] = 1000
	} else {
		requestBody["max_tokens"] = 1000
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the API call
	apiURL := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(provider.BaseURL, "/"))
	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey))

	log.Printf("🔄 [VISION] Calling %s with model %s", provider.Name, modelName)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [VISION] API error: %d - %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	// Parse response
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from vision model")
	}

	description := apiResp.Choices[0].Message.Content
	log.Printf("✅ [VISION] Image described successfully (%d chars)", len(description))

	return &DescribeImageResponse{
		Description: description,
		Model:       modelName,
		Provider:    provider.Name,
	}, nil
}
