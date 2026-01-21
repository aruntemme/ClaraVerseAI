package services

import (
	"claraverse/internal/models"
	"log"
	"sync"
)

// ImageProviderConfig holds the configuration for an image generation provider
type ImageProviderConfig struct {
	Name         string
	BaseURL      string
	APIKey       string
	DefaultModel string
	Favicon      string
}

// ImageProviderService manages image generation providers
type ImageProviderService struct {
	providers []ImageProviderConfig
	mutex     sync.RWMutex
}

var (
	imageProviderInstance *ImageProviderService
	imageProviderOnce     sync.Once
)

// GetImageProviderService returns the singleton image provider service
func GetImageProviderService() *ImageProviderService {
	imageProviderOnce.Do(func() {
		imageProviderInstance = &ImageProviderService{
			providers: make([]ImageProviderConfig, 0),
		}
	})
	return imageProviderInstance
}

// LoadFromProviders loads image providers from the providers config
// This is called during provider sync
func (s *ImageProviderService) LoadFromProviders(providers []models.ProviderConfig) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clear existing providers
	s.providers = make([]ImageProviderConfig, 0)

	for _, p := range providers {
		// Only load enabled providers with image_only flag
		if p.Enabled && p.ImageOnly {
			config := ImageProviderConfig{
				Name:         p.Name,
				BaseURL:      p.BaseURL,
				APIKey:       p.APIKey,
				DefaultModel: p.DefaultModel,
				Favicon:      p.Favicon,
			}
			s.providers = append(s.providers, config)
			log.Printf("🎨 [IMAGE-PROVIDER] Loaded image provider: %s (model: %s)", p.Name, p.DefaultModel)
		}
	}

	log.Printf("🎨 [IMAGE-PROVIDER] Total image providers loaded: %d", len(s.providers))
}

// GetProvider returns the first enabled image provider
// Returns nil if no image providers are configured
func (s *ImageProviderService) GetProvider() *ImageProviderConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.providers) == 0 {
		return nil
	}

	// Return the first provider (could be enhanced to support multiple providers)
	return &s.providers[0]
}

// GetAllProviders returns all configured image providers
func (s *ImageProviderService) GetAllProviders() []ImageProviderConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a copy to prevent external modification
	result := make([]ImageProviderConfig, len(s.providers))
	copy(result, s.providers)
	return result
}

// HasProvider checks if any image provider is configured
func (s *ImageProviderService) HasProvider() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.providers) > 0
}
