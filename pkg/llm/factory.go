package llm

import (
	"fmt"
	"sync"
)

var (
	factoryMu       sync.RWMutex
	clientFactories = make(map[string]Factory)
)

// Factory is a function that creates a new LLM client
type Factory func(config ClientConfig) (Client, error)

// RegisterClientFactory registers a factory for a provider
func RegisterClientFactory(provider string, factory Factory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	clientFactories[provider] = factory
}

// CreateClient creates a new LLM client based on the provider
func CreateClient(config ClientConfig) (Client, error) {
	factoryMu.RLock()
	factory, exists := clientFactories[config.Provider]
	factoryMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", config.Provider)
	}

	return factory(config)
}

// GetRegisteredProviders returns a list of registered provider names
func GetRegisteredProviders() []string {
	factoryMu.RLock()
	defer factoryMu.RUnlock()

	providers := make([]string, 0, len(clientFactories))
	for provider := range clientFactories {
		providers = append(providers, provider)
	}
	return providers
}

// IsProviderRegistered checks if a provider is registered
func IsProviderRegistered(provider string) bool {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	_, exists := clientFactories[provider]
	return exists
}
