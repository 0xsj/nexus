package adapters

import (
	"fmt"
	"sync"

	"github.com/0xsj/nexus/platform/internal/integration/domain"
)

// ============================================================================
// Adapter Registry Implementation
// ============================================================================

// Registry is the default implementation of ProviderAdapterRegistry.
type Registry struct {
	mu       sync.RWMutex
	adapters map[domain.ProviderType]domain.ProviderAdapter
}

// NewRegistry creates a new adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[domain.ProviderType]domain.ProviderAdapter),
	}
}

// Register registers a provider adapter.
func (r *Registry) Register(adapter domain.ProviderAdapter) error {
	if adapter == nil {
		return fmt.Errorf("adapter cannot be nil")
	}

	providerType := adapter.GetProviderType()
	if !providerType.IsValid() {
		return fmt.Errorf("invalid provider type: %s", providerType)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[providerType]; exists {
		return fmt.Errorf("adapter already registered for provider: %s", providerType)
	}

	r.adapters[providerType] = adapter
	return nil
}

// Get retrieves a provider adapter by type.
func (r *Registry) Get(providerType domain.ProviderType) (domain.ProviderAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[providerType]
	if !exists {
		return nil, fmt.Errorf("no adapter registered for provider: %s", providerType)
	}

	return adapter, nil
}

// Has checks if an adapter is registered for the given provider type.
func (r *Registry) Has(providerType domain.ProviderType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.adapters[providerType]
	return exists
}

// GetAll returns all registered adapters.
func (r *Registry) GetAll() []domain.ProviderAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters := make([]domain.ProviderAdapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}

	return adapters
}

// GetSupportedProviders returns a list of all supported provider types.
func (r *Registry) GetSupportedProviders() []domain.ProviderType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]domain.ProviderType, 0, len(r.adapters))
	for providerType := range r.adapters {
		providers = append(providers, providerType)
	}

	return providers
}

// ============================================================================
// Helper Functions
// ============================================================================

// RegisterDefaultAdapters registers all default provider adapters.
func RegisterDefaultAdapters(config *AdapterConfig) (*Registry, error) {
	registry := NewRegistry()

	// GitHub adapter
	if config.GitHub != nil {
		githubAdapter := NewGitHubAdapter(config.GitHub)
		if err := registry.Register(githubAdapter); err != nil {
			return nil, fmt.Errorf("failed to register GitHub adapter: %w", err)
		}
	}

	// LinkedIn adapter
	if config.LinkedIn != nil {
		linkedInAdapter := NewLinkedInAdapter(config.LinkedIn)
		if err := registry.Register(linkedInAdapter); err != nil {
			return nil, fmt.Errorf("failed to register LinkedIn adapter: %w", err)
		}
	}

	// Coursera adapter
	if config.Coursera != nil {
		courseraAdapter := NewCourseraAdapter(config.Coursera)
		if err := registry.Register(courseraAdapter); err != nil {
			return nil, fmt.Errorf("failed to register Coursera adapter: %w", err)
		}
	}

	// Twitter adapter
	if config.Twitter != nil {
		twitterAdapter := NewTwitterAdapter(config.Twitter)
		if err := registry.Register(twitterAdapter); err != nil {
			return nil, fmt.Errorf("failed to register Twitter adapter: %w", err)
		}
	}

	// Google adapter
	if config.Google != nil {
		googleAdapter := NewGoogleAdapter(config.Google)
		if err := registry.Register(googleAdapter); err != nil {
			return nil, fmt.Errorf("failed to register Google adapter: %w", err)
		}
	}

	// AWS adapter
	if config.AWS != nil {
		awsAdapter := NewAWSAdapter(config.AWS)
		if err := registry.Register(awsAdapter); err != nil {
			return nil, fmt.Errorf("failed to register AWS adapter: %w", err)
		}
	}

	return registry, nil
}

// ============================================================================
// Configuration
// ============================================================================

// AdapterConfig holds configuration for all provider adapters.
type AdapterConfig struct {
	GitHub   *GitHubConfig
	LinkedIn *LinkedInConfig
	Coursera *CourseraConfig
	Twitter  *TwitterConfig
	Google   *GoogleConfig
	AWS      *AWSConfig
}

// GitHubConfig holds GitHub adapter configuration.
type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// LinkedInConfig holds LinkedIn adapter configuration.
type LinkedInConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// CourseraConfig holds Coursera adapter configuration.
type CourseraConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// TwitterConfig holds Twitter adapter configuration.
type TwitterConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	BearerToken  string // For API v2
}

// GoogleConfig holds Google adapter configuration.
type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// AWSConfig holds AWS adapter configuration.
type AWSConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Region       string
}
