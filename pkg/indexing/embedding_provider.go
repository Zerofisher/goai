package indexing

import (
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/openai"
)

// EinoEmbeddingProvider wraps Eino's OpenAI embedding component
type EinoEmbeddingProvider struct {
	embedder   *openai.Embedder
	dimensions int
}

// NewEinoEmbeddingProvider creates a new embedding provider using Eino's OpenAI component
func NewEinoEmbeddingProvider(ctx context.Context) (*EinoEmbeddingProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI embedder using Eino
	embedder, err := openai.NewEmbedder(ctx, &openai.EmbeddingConfig{
		APIKey: apiKey,
		Model:  "text-embedding-3-small", // Default model
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI embedder: %w", err)
	}

	return &EinoEmbeddingProvider{
		embedder:   embedder,
		dimensions: 1536, // text-embedding-3-small dimensions
	}, nil
}

// GenerateEmbedding generates a single embedding for text using Eino
func (p *EinoEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Use Eino embedder to generate embedding
	embeddings, err := p.embedder.EmbedStrings(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return embeddings[0], nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts using Eino
func (p *EinoEmbeddingProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Use Eino embedder batch functionality
	embeddings, err := p.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (p *EinoEmbeddingProvider) GetDimensions() int {
	return p.dimensions
}

// MockEmbeddingProvider provides mock embeddings for testing
type MockEmbeddingProvider struct {
	dimensions int
}

// NewMockEmbeddingProvider creates a new mock embedding provider
func NewMockEmbeddingProvider(dimensions int) *MockEmbeddingProvider {
	if dimensions <= 0 {
		dimensions = 384 // Default to 384 dimensions (common for smaller models)
	}
	
	return &MockEmbeddingProvider{
		dimensions: dimensions,
	}
}

// GenerateEmbedding generates a single embedding for text
func (p *MockEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided")
	}

	// Generate a simple hash-based embedding for consistent results
	embedding := make([]float64, p.dimensions)
	hash := p.simpleHash(text)
	
	for i := 0; i < p.dimensions; i++ {
		// Use hash to seed pseudo-random generator for consistency
		r := rand.New(rand.NewSource(int64(hash + uint32(i))))
		embedding[i] = r.NormFloat64()
	}

	// Normalize the embedding vector
	return p.normalize(embedding), nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (p *MockEmbeddingProvider) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embedding, err := p.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (p *MockEmbeddingProvider) GetDimensions() int {
	return p.dimensions
}

// simpleHash creates a simple hash of the input text
func (p *MockEmbeddingProvider) simpleHash(text string) uint32 {
	var hash uint32 = 2166136261 // FNV offset basis
	for _, b := range []byte(text) {
		hash ^= uint32(b)
		hash *= 16777619 // FNV prime
	}
	return hash
}

// normalize normalizes a vector to unit length
func (p *MockEmbeddingProvider) normalize(vec []float64) []float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	
	if norm == 0 {
		return vec
	}
	
	norm = 1.0 / (norm * norm) // Use square of norm for more variation
	for i := range vec {
		vec[i] *= norm
	}
	
	return vec
}