package vectorstore

import (
	"context"
	"math"
	"sort"
	"sync"
)

type Document struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
	Vector   []float32
}

type VectorStore interface {
	AddDocuments(ctx context.Context, docs []Document) error
	SimilaritySearch(ctx context.Context, queryVector []float32, k int) ([]Document, error)
	Delete(ctx context.Context, ids []string) error
}

// Simple in-memory vector store for initial RAG implementation
// TODO: Replace with Pinecone/Weaviate for production
type MemoryVectorStore struct {
	documents map[string]Document
	mu        sync.RWMutex
}

func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		documents: make(map[string]Document),
	}
}

func (m *MemoryVectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, doc := range docs {
		m.documents[doc.ID] = doc
	}

	return nil
}

func (m *MemoryVectorStore) SimilaritySearch(ctx context.Context, queryVector []float32, k int) ([]Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type scoredDoc struct {
		doc   Document
		score float32
	}

	var scoredDocs []scoredDoc
	for _, doc := range m.documents {
		if len(doc.Vector) != len(queryVector) {
			continue
		}

		score := cosineSimilarity(queryVector, doc.Vector)
		scoredDocs = append(scoredDocs, scoredDoc{doc: doc, score: score})
	}

	// Sort by score descending
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].score > scoredDocs[j].score
	})

	// Return top k
	result := make([]Document, 0, k)
	for i, scored := range scoredDocs {
		if i >= k {
			break
		}
		result = append(result, scored.doc)
	}

	return result, nil
}

func (m *MemoryVectorStore) Delete(ctx context.Context, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range ids {
		delete(m.documents, id)
	}

	return nil
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
