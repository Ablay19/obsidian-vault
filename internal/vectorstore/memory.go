package vectorstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"

	_ "github.com/mattn/go-sqlite3"
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

// SQLite-backed vector store for persistent document storage
type SQLiteVectorStore struct {
	db   *sql.DB
	mu   sync.RWMutex
	path string
}

func NewSQLiteVectorStore(db *sql.DB) (*SQLiteVectorStore, error) {
	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS vector_documents (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		metadata TEXT,
		vector TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_content ON vector_documents(content);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteVectorStore{
		db: db,
	}, nil
}

func (s *SQLiteVectorStore) AddDocuments(ctx context.Context, docs []Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO vector_documents (id, content, metadata, vector)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, doc := range docs {
		metadataJSON, err := json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		vectorJSON, err := json.Marshal(doc.Vector)
		if err != nil {
			return fmt.Errorf("failed to marshal vector: %w", err)
		}

		_, err = stmt.ExecContext(ctx, doc.ID, doc.Content, string(metadataJSON), string(vectorJSON))
		if err != nil {
			return fmt.Errorf("failed to insert document: %w", err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteVectorStore) SimilaritySearch(ctx context.Context, queryVector []float32, k int) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all documents (for small datasets; in production, use indexing or approximation)
	rows, err := s.db.QueryContext(ctx, "SELECT id, content, metadata, vector FROM vector_documents")
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	type scoredDoc struct {
		doc   Document
		score float32
	}

	var scoredDocs []scoredDoc

	for rows.Next() {
		var id, content, metadataStr, vectorStr string
		if err := rows.Scan(&id, &content, &metadataStr, &vectorStr); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		var vector []float32
		if err := json.Unmarshal([]byte(vectorStr), &vector); err != nil {
			return nil, fmt.Errorf("failed to unmarshal vector: %w", err)
		}

		doc := Document{
			ID:       id,
			Content:  content,
			Metadata: metadata,
			Vector:   vector,
		}

		if len(doc.Vector) != len(queryVector) {
			continue
		}

		score := cosineSimilarity(queryVector, doc.Vector)
		scoredDocs = append(scoredDocs, scoredDoc{doc: doc, score: score})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
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

func (s *SQLiteVectorStore) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM vector_documents WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	for _, id := range ids {
		_, err = stmt.ExecContext(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete document %s: %w", id, err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteVectorStore) Close() error {
	return s.db.Close()
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
