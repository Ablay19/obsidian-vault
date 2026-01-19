package vectorstore

import (
	"context"
	"errors"
	"time"
)

// Document represents a document stored in the vector store
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Validate validates the document
func (d *Document) Validate() error {
	if d.ID == "" {
		return errors.New("document ID cannot be empty")
	}
	if d.Content == "" {
		return errors.New("document content cannot be empty")
	}
	return nil
}

// Vector represents an embedding vector
type Vector struct {
	ID        string                 `json:"id"`
	Values    []float64              `json:"values"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Validate validates the vector
func (v *Vector) Validate() error {
	if v.ID == "" {
		return errors.New("vector ID cannot be empty")
	}
	if len(v.Values) == 0 {
		return errors.New("vector values cannot be empty")
	}
	return nil
}

// SearchResult represents a search result from the vector store
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Store represents the vector storage interface
type Store interface {
	Store(ctx context.Context, vectors []Vector) error
	Search(ctx context.Context, query Vector, limit int, threshold float64) ([]SearchResult, error)
	Delete(ctx context.Context, ids []string) error
	Update(ctx context.Context, vectors []Vector) error
	BatchStore(ctx context.Context, vectors []Vector, batchSize int) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// Embedder represents text embedding interface
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
	EmbedSingle(ctx context.Context, text string) ([]float64, error)
	Dimension() int
}
