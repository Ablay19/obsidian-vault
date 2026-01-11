package rag

import (
	"context"
	"fmt"
	"math"

	"github.com/tmc/langchaingo/schema"
	"obsidian-automation/internal/vectorstore"
)

// Retriever defines the interface for document retrieval
type Retriever interface {
	GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error)
}

// VectorRetriever implements Retriever using vector similarity search
type VectorRetriever struct {
	store     vectorstore.VectorStore
	embedder  Embedder
	topK      int
	threshold float32
}

type Embedder interface {
	EmbedText(text string) ([]float32, error)
}

// SimpleEmbedder provides basic text embedding (placeholder for production models)
type SimpleEmbedder struct{}

func (e *SimpleEmbedder) EmbedText(text string) ([]float32, error) {
	// TODO: Replace with proper embedding model (OpenAI, Sentence Transformers, etc.)
	return generateSimpleEmbedding(text), nil
}

// NewVectorRetriever creates a new vector retriever
func NewVectorRetriever(store vectorstore.VectorStore, embedder Embedder, topK int, threshold float32) *VectorRetriever {
	if embedder == nil {
		embedder = &SimpleEmbedder{}
	}
	return &VectorRetriever{
		store:     store,
		embedder:  embedder,
		topK:      topK,
		threshold: threshold,
	}
}

// GetRelevantDocuments retrieves documents relevant to the query
func (r *VectorRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	// Generate embedding for the query
	queryVector, err := r.embedder.EmbedText(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search for similar documents
	results, err := r.store.SimilaritySearch(ctx, queryVector, r.topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	// Convert to LangChain document format and filter by threshold
	var documents []schema.Document
	for _, result := range results {
		// Calculate similarity score
		docVector := result.Vector
		if len(docVector) != len(queryVector) {
			continue
		}

		similarity := cosineSimilarity(queryVector, docVector)
		if similarity < r.threshold {
			continue
		}

		// Convert metadata to map[string]interface{}
		metadata := make(map[string]interface{})
		for k, v := range result.Metadata {
			metadata[k] = v
		}
		metadata["similarity_score"] = similarity

		doc := schema.Document{
			PageContent: result.Content,
			Metadata:    metadata,
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

// generateSimpleEmbedding creates a basic embedding vector from text
// TODO: Replace with proper embedding model (OpenAI, Sentence Transformers, etc.)
func generateSimpleEmbedding(text string) []float32 {
	// Simple hash-based embedding for demonstration
	// In production, use proper embedding models
	const embeddingDim = 384 // Common embedding dimension
	embedding := make([]float32, embeddingDim)

	// Simple hash function to generate pseudo-random values
	for i, char := range text {
		hash := int(char) * (i + 1)
		idx := hash % embeddingDim
		embedding[idx] += float32(hash%100) / 100.0
	}

	// Normalize the embedding
	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}
