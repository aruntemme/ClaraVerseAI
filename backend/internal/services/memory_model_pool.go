package services

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"claraverse/internal/config"
	"claraverse/internal/models"
)

// MemoryModelPool manages multiple models for memory operations with health tracking and failover
type MemoryModelPool struct {
	extractorModels  []ModelCandidate
	selectorModels   []ModelCandidate
	extractorIndex   int
	selectorIndex    int
	healthTracker    map[string]*ModelHealth
	mu               sync.Mutex
	chatService      *ChatService
	db               *sql.DB // Database connection for querying model_aliases
}

// ModelCandidate represents a model eligible for memory operations
type ModelCandidate struct {
	ModelID     string
	ProviderName string
	SpeedMs     int
	DisplayName string
}

// ModelHealth tracks model health and failures
type ModelHealth struct {
	FailureCount    int
	SuccessCount    int
	LastFailure     time.Time
	LastSuccess     time.Time
	IsHealthy       bool
	ConsecutiveFails int
}

const (
	// Health thresholds
	MaxConsecutiveFailures = 3
	HealthCheckCooldown    = 5 * time.Minute
	MinSuccessesToRecover  = 2
)

// NewMemoryModelPool creates a new model pool by discovering eligible models from providers
func NewMemoryModelPool(chatService *ChatService, db *sql.DB) (*MemoryModelPool, error) {
	pool := &MemoryModelPool{
		chatService:   chatService,
		db:            db,
		healthTracker: make(map[string]*ModelHealth),
	}

	// Discover models from ChatService
	if err := pool.discoverModels(); err != nil {
		log.Printf("⚠️ [MODEL-POOL] Failed to discover models: %v", err)
		log.Printf("⚠️ [MODEL-POOL] Memory services will be disabled until models with memory flags are added")
	}

	if len(pool.extractorModels) == 0 {
		log.Printf("⚠️ [MODEL-POOL] No extractor models found - memory extraction disabled")
	}

	if len(pool.selectorModels) == 0 {
		log.Printf("⚠️ [MODEL-POOL] No selector models found - memory selection disabled")
	}

	if len(pool.extractorModels) > 0 || len(pool.selectorModels) > 0 {
		log.Printf("🎯 [MODEL-POOL] Initialized with %d extractors, %d selectors",
			len(pool.extractorModels), len(pool.selectorModels))
	}

	// Return pool even if empty - allows graceful degradation
	return pool, nil
}

// discoverModels scans database for models with memory flags
func (p *MemoryModelPool) discoverModels() error {
	// First try loading from database (MySQL-first approach)
	dbModels, err := p.discoverFromDatabase()
	if err == nil && len(dbModels) > 0 {
		log.Printf("✅ [MODEL-POOL] Discovered %d models from database", len(dbModels))
		return nil
	}

	// Fallback: Load providers configuration from providers.json
	providersConfig, err := config.LoadProviders("providers.json")
	if err != nil {
		// Gracefully handle missing providers.json (expected in admin UI workflow)
		log.Printf("⚠️ [MODEL-POOL] providers.json not found or invalid: %v", err)
		log.Printf("ℹ️ [MODEL-POOL] This is normal when starting with empty database")
		return nil // Not a fatal error - just means no models configured yet
	}

	for _, providerConfig := range providersConfig.Providers {
		if !providerConfig.Enabled {
			continue
		}

		for alias, modelAlias := range providerConfig.ModelAliases {
			// Get model configuration map (convert from ModelAlias)
			modelConfig := modelAliasToMap(modelAlias)

			// Check if model supports memory extraction
			if isExtractor, ok := modelConfig["memory_extractor"].(bool); ok && isExtractor {
				candidate := ModelCandidate{
					ModelID:      alias,
					ProviderName: providerConfig.Name,
					DisplayName:  getDisplayName(modelConfig),
					SpeedMs:      getSpeedMs(modelConfig),
				}
				p.extractorModels = append(p.extractorModels, candidate)
				p.healthTracker[alias] = &ModelHealth{IsHealthy: true}

				log.Printf("✅ [MODEL-POOL] Found extractor: %s (%s) - %dms",
					alias, providerConfig.Name, candidate.SpeedMs)
			}

			// Check if model supports memory selection
			if isSelector, ok := modelConfig["memory_selector"].(bool); ok && isSelector {
				// Avoid duplicates if model is both extractor and selector
				if _, exists := p.healthTracker[alias]; !exists {
					p.healthTracker[alias] = &ModelHealth{IsHealthy: true}
				}

				candidate := ModelCandidate{
					ModelID:      alias,
					ProviderName: providerConfig.Name,
					DisplayName:  getDisplayName(modelConfig),
					SpeedMs:      getSpeedMs(modelConfig),
				}
				p.selectorModels = append(p.selectorModels, candidate)

				log.Printf("✅ [MODEL-POOL] Found selector: %s (%s) - %dms",
					alias, providerConfig.Name, candidate.SpeedMs)
			}
		}
	}

	// Sort by speed (fastest first)
	p.sortModelsBySpeed(p.extractorModels)
	p.sortModelsBySpeed(p.selectorModels)

	return nil
}

// discoverFromDatabase loads memory models from database (model_aliases table)
func (p *MemoryModelPool) discoverFromDatabase() ([]ModelCandidate, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Query for models with memory_extractor or memory_selector flags
	rows, err := p.db.Query(`
		SELECT
			a.alias_name,
			pr.name as provider_name,
			a.display_name,
			COALESCE(a.structured_output_speed_ms, 999999) as speed_ms,
			COALESCE(a.memory_extractor, 0) as memory_extractor,
			COALESCE(a.memory_selector, 0) as memory_selector
		FROM model_aliases a
		JOIN providers pr ON a.provider_id = pr.id
		WHERE a.memory_extractor = 1 OR a.memory_selector = 1
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query model_aliases: %w", err)
	}
	defer rows.Close()

	var candidates []ModelCandidate
	for rows.Next() {
		var aliasName, providerName, displayName string
		var speedMs int
		var isExtractor, isSelector int

		if err := rows.Scan(&aliasName, &providerName, &displayName, &speedMs, &isExtractor, &isSelector); err != nil {
			log.Printf("⚠️ [MODEL-POOL] Failed to scan row: %v", err)
			continue
		}

		candidate := ModelCandidate{
			ModelID:      aliasName,
			ProviderName: providerName,
			DisplayName:  displayName,
			SpeedMs:      speedMs,
		}

		if isExtractor == 1 {
			p.extractorModels = append(p.extractorModels, candidate)
			p.healthTracker[aliasName] = &ModelHealth{IsHealthy: true}
			log.Printf("✅ [MODEL-POOL] Found extractor from DB: %s (%s) - %dms", aliasName, providerName, speedMs)
		}

		if isSelector == 1 {
			// Avoid duplicates if model is both extractor and selector
			if _, exists := p.healthTracker[aliasName]; !exists {
				p.healthTracker[aliasName] = &ModelHealth{IsHealthy: true}
			}
			p.selectorModels = append(p.selectorModels, candidate)
			log.Printf("✅ [MODEL-POOL] Found selector from DB: %s (%s) - %dms", aliasName, providerName, speedMs)
		}

		candidates = append(candidates, candidate)
	}

	// Sort by speed (fastest first)
	p.sortModelsBySpeed(p.extractorModels)
	p.sortModelsBySpeed(p.selectorModels)

	return candidates, nil
}

// modelAliasToMap converts ModelAlias struct to map for easier access
func modelAliasToMap(alias models.ModelAlias) map[string]interface{} {
	m := make(map[string]interface{})

	// Set display_name
	m["display_name"] = alias.DisplayName

	// Set structured_output_speed_ms if available
	if alias.StructuredOutputSpeedMs != nil {
		m["structured_output_speed_ms"] = *alias.StructuredOutputSpeedMs
	}

	// Set memory flags if available
	if alias.MemoryExtractor != nil {
		m["memory_extractor"] = *alias.MemoryExtractor
	}
	if alias.MemorySelector != nil {
		m["memory_selector"] = *alias.MemorySelector
	}

	return m
}

// GetNextExtractor returns the next healthy extractor model using round-robin
func (p *MemoryModelPool) GetNextExtractor() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.extractorModels) == 0 {
		return "", fmt.Errorf("no extractor models available")
	}

	// Try all models in round-robin fashion
	attempts := 0
	maxAttempts := len(p.extractorModels)

	for attempts < maxAttempts {
		candidate := p.extractorModels[p.extractorIndex]
		p.extractorIndex = (p.extractorIndex + 1) % len(p.extractorModels)
		attempts++

		// Check if model is healthy
		health := p.healthTracker[candidate.ModelID]
		if health.IsHealthy {
			log.Printf("🔄 [MODEL-POOL] Selected extractor: %s (healthy)", candidate.ModelID)
			return candidate.ModelID, nil
		}

		// Check if enough time has passed since last failure (cooldown)
		if time.Since(health.LastFailure) > HealthCheckCooldown {
			log.Printf("⚡ [MODEL-POOL] Retrying extractor after cooldown: %s", candidate.ModelID)
			health.IsHealthy = true
			health.ConsecutiveFails = 0
			return candidate.ModelID, nil
		}

		log.Printf("⏭️ [MODEL-POOL] Skipping unhealthy extractor: %s (fails: %d, last: %s ago)",
			candidate.ModelID, health.ConsecutiveFails, time.Since(health.LastFailure).Round(time.Second))
	}

	// All models unhealthy - return fastest anyway as last resort
	log.Printf("⚠️ [MODEL-POOL] All extractors unhealthy, using fastest: %s", p.extractorModels[0].ModelID)
	return p.extractorModels[0].ModelID, nil
}

// GetNextSelector returns the next healthy selector model using round-robin
func (p *MemoryModelPool) GetNextSelector() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.selectorModels) == 0 {
		return "", fmt.Errorf("no selector models available")
	}

	// Try all models in round-robin fashion
	attempts := 0
	maxAttempts := len(p.selectorModels)

	for attempts < maxAttempts {
		candidate := p.selectorModels[p.selectorIndex]
		p.selectorIndex = (p.selectorIndex + 1) % len(p.selectorModels)
		attempts++

		// Check if model is healthy
		health := p.healthTracker[candidate.ModelID]
		if health.IsHealthy {
			log.Printf("🔄 [MODEL-POOL] Selected selector: %s (healthy)", candidate.ModelID)
			return candidate.ModelID, nil
		}

		// Check if enough time has passed since last failure (cooldown)
		if time.Since(health.LastFailure) > HealthCheckCooldown {
			log.Printf("⚡ [MODEL-POOL] Retrying selector after cooldown: %s", candidate.ModelID)
			health.IsHealthy = true
			health.ConsecutiveFails = 0
			return candidate.ModelID, nil
		}

		log.Printf("⏭️ [MODEL-POOL] Skipping unhealthy selector: %s (fails: %d, last: %s ago)",
			candidate.ModelID, health.ConsecutiveFails, time.Since(health.LastFailure).Round(time.Second))
	}

	// All models unhealthy - return fastest anyway as last resort
	log.Printf("⚠️ [MODEL-POOL] All selectors unhealthy, using fastest: %s", p.selectorModels[0].ModelID)
	return p.selectorModels[0].ModelID, nil
}

// MarkSuccess records a successful model call
func (p *MemoryModelPool) MarkSuccess(modelID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	health, exists := p.healthTracker[modelID]
	if !exists {
		return
	}

	health.SuccessCount++
	health.LastSuccess = time.Now()
	health.ConsecutiveFails = 0

	// Restore health after consecutive successes
	if !health.IsHealthy && health.SuccessCount >= MinSuccessesToRecover {
		health.IsHealthy = true
		log.Printf("💚 [MODEL-POOL] Model recovered: %s (successes: %d)", modelID, health.SuccessCount)
	}
}

// MarkFailure records a failed model call
func (p *MemoryModelPool) MarkFailure(modelID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	health, exists := p.healthTracker[modelID]
	if !exists {
		return
	}

	health.FailureCount++
	health.ConsecutiveFails++
	health.LastFailure = time.Now()

	// Mark unhealthy after consecutive failures
	if health.ConsecutiveFails >= MaxConsecutiveFailures {
		health.IsHealthy = false
		log.Printf("💔 [MODEL-POOL] Model marked unhealthy: %s (consecutive fails: %d, total fails: %d)",
			modelID, health.ConsecutiveFails, health.FailureCount)
	} else {
		log.Printf("⚠️ [MODEL-POOL] Model failure: %s (consecutive: %d/%d)",
			modelID, health.ConsecutiveFails, MaxConsecutiveFailures)
	}
}

// GetStats returns current pool statistics
func (p *MemoryModelPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	healthyExtractors := 0
	healthySelectors := 0

	for _, model := range p.extractorModels {
		if p.healthTracker[model.ModelID].IsHealthy {
			healthyExtractors++
		}
	}

	for _, model := range p.selectorModels {
		if p.healthTracker[model.ModelID].IsHealthy {
			healthySelectors++
		}
	}

	return map[string]interface{}{
		"total_extractors":   len(p.extractorModels),
		"healthy_extractors": healthyExtractors,
		"total_selectors":    len(p.selectorModels),
		"healthy_selectors":  healthySelectors,
	}
}

// Helper functions

func getDisplayName(modelConfig map[string]interface{}) string {
	if name, ok := modelConfig["display_name"].(string); ok {
		return name
	}
	return "Unknown"
}

func getSpeedMs(modelConfig map[string]interface{}) int {
	if speed, ok := modelConfig["structured_output_speed_ms"].(float64); ok {
		return int(speed)
	}
	if speed, ok := modelConfig["structured_output_speed_ms"].(int); ok {
		return speed
	}
	return 999999 // Default to slow if not specified
}

func (p *MemoryModelPool) sortModelsBySpeed(models []ModelCandidate) {
	// Simple bubble sort (fine for small arrays)
	n := len(models)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if models[j].SpeedMs > models[j+1].SpeedMs {
				models[j], models[j+1] = models[j+1], models[j]
			}
		}
	}
}
