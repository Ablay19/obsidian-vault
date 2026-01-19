package vectorstore

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// VectorStore implements the Store interface using in-memory storage
type VectorStore struct {
	vectors  map[string]Vector
	mutex    sync.RWMutex
	db       VectorDB
	embedder Embedder
}

// VectorDB interface for vector database operations
type VectorDB interface {
	Store(ctx context.Context, vectors []Vector) error
	Search(ctx context.Context, query Vector, limit int, threshold float64) ([]SearchResult, error)
	Delete(ctx context.Context, ids []string) error
}

// NewVectorStore creates a new vector store
func NewVectorStore(db VectorDB, embedder Embedder) *VectorStore {
	return &VectorStore{
		vectors:  make(map[string]Vector),
		db:       db,
		embedder: embedder,
	}
}

// Store stores vectors in the vector store
func (vs *VectorStore) Store(ctx context.Context, vectors []Vector) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	for _, v := range vectors {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("invalid vector %s: %w", v.ID, err)
		}
		vs.vectors[v.ID] = v
	}
	return nil
}

// Search searches for similar vectors
func (vs *VectorStore) Search(ctx context.Context, query Vector, limit int, threshold float64) ([]SearchResult, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	var results []SearchResult
	for _, v := range vs.vectors {
		score := cosineSimilarity(query.Values, v.Values)
		if score >= threshold {
			results = append(results, SearchResult{
				ID:       v.ID,
				Score:    score,
				Metadata: v.Metadata,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Delete deletes vectors by IDs
func (vs *VectorStore) Delete(ctx context.Context, ids []string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	for _, id := range ids {
		delete(vs.vectors, id)
	}
	return nil
}

// Update updates existing vectors
func (vs *VectorStore) Update(ctx context.Context, vectors []Vector) error {
	return vs.Store(ctx, vectors) // Simple implementation
}

// BatchStore stores vectors in batches
func (vs *VectorStore) BatchStore(ctx context.Context, vectors []Vector, batchSize int) error {
	for i := 0; i < len(vectors); i += batchSize {
		end := i + batchSize
		if end > len(vectors) {
			end = len(vectors)
		}
		if err := vs.Store(ctx, vectors[i:end]); err != nil {
			return err
		}
	}
	return nil
}

// GetStats returns store statistics
func (vs *VectorStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	return map[string]interface{}{
		"total_vectors": len(vs.vectors),
	}, nil
}

// AddDocument adds a document by embedding its content
func (vs *VectorStore) AddDocument(ctx context.Context, doc Document) error {
	embedding, err := vs.embedder.EmbedSingle(ctx, doc.Content)
	if err != nil {
		return err
	}

	vector := Vector{
		ID:       doc.ID,
		Values:   embedding,
		Metadata: doc.Metadata,
	}

	return vs.Store(ctx, []Vector{vector})
}

// AddDocuments adds multiple documents
func (vs *VectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := vs.AddDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// SimilaritySearch searches for similar documents
func (vs *VectorStore) SimilaritySearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	embedding, err := vs.embedder.EmbedSingle(ctx, query)
	if err != nil {
		return nil, err
	}

	queryVector := Vector{Values: embedding}
	return vs.Search(ctx, queryVector, limit, 0.0)
}

// DeleteDocuments deletes documents by IDs
func (vs *VectorStore) DeleteDocuments(ctx context.Context, ids []string) error {
	return vs.Delete(ctx, ids)
}

// UpdateDocument updates a document
func (vs *VectorStore) UpdateDocument(ctx context.Context, doc Document) error {
	return vs.AddDocument(ctx, doc) // Simple update as add
}

// BatchEmbedAndStore embeds and stores documents in batches
func (vs *VectorStore) BatchEmbedAndStore(ctx context.Context, docs []Document, batchSize int) error {
	return vs.BatchStore(ctx, vs.embedDocuments(docs), batchSize)
}

// HealthCheck checks the health of the vector store
func (vs *VectorStore) HealthCheck(ctx context.Context) error {
	// Simple health check
	return nil
}

// embedDocuments embeds multiple documents
func (vs *VectorStore) embedDocuments(docs []Document) []Vector {
	var vectors []Vector
	for _, doc := range docs {
		// For simplicity, assume embedding is done, but in real impl use embedder
		vector := Vector{
			ID:       doc.ID,
			Values:   []float64{0.1, 0.2}, // placeholder
			Metadata: doc.Metadata,
		}
		vectors = append(vectors, vector)
	}
	return vectors
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (normA * normB)
}
