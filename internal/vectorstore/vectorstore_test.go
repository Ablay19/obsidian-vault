package vectorstore

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing
type MockVectorDB struct {
	mock.Mock
}

func (m *MockVectorDB) Store(ctx context.Context, vectors []Vector) error {
	args := m.Called(ctx, vectors)
	return args.Error(0)
}

func (m *MockVectorDB) Search(ctx context.Context, query Vector, limit int, threshold float64) ([]SearchResult, error) {
	args := m.Called(ctx, query, limit, threshold)
	return args.Get(0).([]SearchResult), args.Error(1)
}

func (m *MockVectorDB) Delete(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockVectorDB) Update(ctx context.Context, vectors []Vector) error {
	args := m.Called(ctx, vectors)
	return args.Error(0)
}

func (m *MockVectorDB) BatchStore(ctx context.Context, vectors []Vector, batchSize int) error {
	args := m.Called(ctx, vectors, batchSize)
	return args.Error(0)
}

func (m *MockVectorDB) GetStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

type MockEmbedder struct {
	mock.Mock
}

func (m *MockEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).([][]float64), args.Error(1)
}

func (m *MockEmbedder) EmbedSingle(ctx context.Context, text string) ([]float64, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]float64), args.Error(1)
}

func (m *MockEmbedder) Dimension() int {
	args := m.Called()
	return args.Int(0)
}

func setupTestVectorStore(t *testing.T) (*VectorStore, *MockVectorDB, *MockEmbedder) {
	mockDB := &MockVectorDB{}
	mockEmbedder := &MockEmbedder{}

	vs := NewVectorStore(mockDB, mockEmbedder)

	return vs, mockDB, mockEmbedder
}

func TestNewVectorStore(t *testing.T) {
	mockDB := &MockVectorDB{}
	mockEmbedder := &MockEmbedder{}

	vs := NewVectorStore(mockDB, mockEmbedder)

	assert.NotNil(t, vs)
	assert.Equal(t, mockDB, vs.db)
	assert.Equal(t, mockEmbedder, vs.embedder)
}

func TestVectorStore_AddDocument(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	doc := Document{
		ID:       "test_doc",
		Content:  "This is a test document",
		Metadata: map[string]interface{}{"source": "test"},
	}

	// Mock embedder
	expectedEmbedding := []float64{0.1, 0.2, 0.3}
	mockEmbedder.On("EmbedSingle", mock.Anything, "This is a test document").Return(expectedEmbedding, nil)

	// Mock DB
	expectedVector := Vector{
		ID:        "test_doc",
		Values:    expectedEmbedding,
		Metadata:  doc.Metadata,
		Timestamp: time.Now(),
	}
	mockDB.On("Store", mock.Anything, mock.AnythingOfType("[]vectorstore.Vector")).Return(nil)

	err := vs.AddDocument(context.Background(), doc)

	assert.NoError(t, err)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_AddDocuments_Batch(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	docs := []Document{
		{ID: "doc1", Content: "Document 1", Metadata: map[string]interface{}{"type": "test"}},
		{ID: "doc2", Content: "Document 2", Metadata: map[string]interface{}{"type": "test"}},
	}

	// Mock embedder
	embeddings := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
	}
	mockEmbedder.On("Embed", mock.Anything, []string{"Document 1", "Document 2"}).Return(embeddings, nil)

	// Mock DB
	mockDB.On("BatchStore", mock.Anything, mock.AnythingOfType("[]vectorstore.Vector"), 10).Return(nil)

	err := vs.AddDocuments(context.Background(), docs, 10)

	assert.NoError(t, err)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_Search(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	query := "test query"

	// Mock embedder
	queryEmbedding := []float64{0.5, 0.6}
	mockEmbedder.On("EmbedSingle", mock.Anything, query).Return(queryEmbedding, nil)

	// Mock DB search
	expectedResults := []SearchResult{
		{ID: "doc1", Score: 0.95, Metadata: map[string]interface{}{"content": "matching doc"}},
	}
	mockDB.On("Search", mock.Anything, mock.AnythingOfType("vectorstore.Vector"), 5, 0.7).Return(expectedResults, nil)

	results, err := vs.Search(context.Background(), query, 5, 0.7)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_SimilaritySearch(t *testing.T) {
	vs, mockDB, _ := setupTestVectorStore(t)

	queryVector := Vector{Values: []float64{0.1, 0.2, 0.3}}

	expectedResults := []SearchResult{
		{ID: "doc1", Score: 0.9},
		{ID: "doc2", Score: 0.8},
	}

	mockDB.On("Search", mock.Anything, queryVector, 10, 0.5).Return(expectedResults, nil)

	results, err := vs.SimilaritySearch(context.Background(), queryVector, 10, 0.5)

	assert.NoError(t, err)
	assert.Equal(t, expectedResults, results)

	mockDB.AssertExpectations(t)
}

func TestVectorStore_DeleteDocuments(t *testing.T) {
	vs, mockDB, _ := setupTestVectorStore(t)

	docIDs := []string{"doc1", "doc2", "doc3"}

	mockDB.On("Delete", mock.Anything, docIDs).Return(nil)

	err := vs.DeleteDocuments(context.Background(), docIDs)

	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_UpdateDocument(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	doc := Document{
		ID:       "test_doc",
		Content:  "Updated content",
		Metadata: map[string]interface{}{"updated": true},
	}

	// Mock embedder
	embedding := []float64{0.7, 0.8, 0.9}
	mockEmbedder.On("EmbedSingle", mock.Anything, "Updated content").Return(embedding, nil)

	// Mock DB update
	mockDB.On("Update", mock.Anything, mock.AnythingOfType("[]vectorstore.Vector")).Return(nil)

	err := vs.UpdateDocument(context.Background(), doc)

	assert.NoError(t, err)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_GetStats(t *testing.T) {
	vs, mockDB, _ := setupTestVectorStore(t)

	expectedStats := map[string]interface{}{
		"total_vectors": 1000,
		"dimensions":    384,
		"index_size":    "50MB",
	}

	mockDB.On("GetStats", mock.Anything).Return(expectedStats, nil)

	stats, err := vs.GetStats(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)

	mockDB.AssertExpectations(t)
}

func TestVectorStore_BatchEmbedAndStore(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	texts := []string{"Text 1", "Text 2", "Text 3"}
	metadata := []map[string]interface{}{
		{"id": "1"},
		{"id": "2"},
		{"id": "3"},
	}

	// Mock embedder
	embeddings := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
		{0.5, 0.6},
	}
	mockEmbedder.On("Embed", mock.Anything, texts).Return(embeddings, nil)

	// Mock DB
	mockDB.On("BatchStore", mock.Anything, mock.AnythingOfType("[]vectorstore.Vector"), 5).Return(nil)

	err := vs.BatchEmbedAndStore(context.Background(), texts, metadata, 5)

	assert.NoError(t, err)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_FilteredSearch(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	query := "test query"
	filter := map[string]interface{}{
		"category": "important",
		"status":   "active",
	}

	// Mock embedder
	queryEmbedding := []float64{0.1, 0.2}
	mockEmbedder.On("EmbedSingle", mock.Anything, query).Return(queryEmbedding, nil)

	// Mock DB search
	results := []SearchResult{
		{ID: "doc1", Score: 0.9, Metadata: filter},
	}
	mockDB.On("Search", mock.Anything, mock.AnythingOfType("vectorstore.Vector"), 5, 0.8).Return(results, nil)

	// For filtered search, we would need to filter results after DB search
	// This is a simplified version
	filteredResults, err := vs.Search(context.Background(), query, 5, 0.8)

	assert.NoError(t, err)
	assert.NotEmpty(t, filteredResults)

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestVectorStore_HealthCheck(t *testing.T) {
	vs, mockDB, _ := setupTestVectorStore(t)

	// Mock DB stats call as health check
	mockDB.On("GetStats", mock.Anything).Return(map[string]interface{}{"healthy": true}, nil)

	healthy := vs.HealthCheck(context.Background())

	assert.True(t, healthy)

	mockDB.AssertExpectations(t)
}

func TestVectorStore_HealthCheck_Unhealthy(t *testing.T) {
	vs, mockDB, _ := setupTestVectorStore(t)

	// Mock DB failure
	mockDB.On("GetStats", mock.Anything).Return(nil, assert.AnError)

	healthy := vs.HealthCheck(context.Background())

	assert.False(t, healthy)

	mockDB.AssertExpectations(t)
}

func TestVector_Validate(t *testing.T) {
	tests := []struct {
		name      string
		vector    Vector
		shouldErr bool
	}{
		{
			name: "valid vector",
			vector: Vector{
				ID:     "test_id",
				Values: []float64{0.1, 0.2, 0.3},
			},
			shouldErr: false,
		},
		{
			name: "empty ID",
			vector: Vector{
				Values: []float64{0.1, 0.2, 0.3},
			},
			shouldErr: true,
		},
		{
			name: "empty values",
			vector: Vector{
				ID: "test_id",
			},
			shouldErr: true,
		},
		{
			name: "zero values",
			vector: Vector{
				ID:     "test_id",
				Values: []float64{},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vector.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDocument_Validate(t *testing.T) {
	tests := []struct {
		name      string
		doc       Document
		shouldErr bool
	}{
		{
			name: "valid document",
			doc: Document{
				ID:      "test_id",
				Content: "Test content",
			},
			shouldErr: false,
		},
		{
			name: "empty ID",
			doc: Document{
				Content: "Test content",
			},
			shouldErr: true,
		},
		{
			name: "empty content",
			doc: Document{
				ID: "test_id",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSearchResult_Compare(t *testing.T) {
	results := []SearchResult{
		{ID: "doc1", Score: 0.8},
		{ID: "doc2", Score: 0.95},
		{ID: "doc3", Score: 0.6},
	}

	// Sort by score (higher first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	assert.Equal(t, "doc2", results[0].ID) // Highest score
	assert.Equal(t, "doc1", results[1].ID)
	assert.Equal(t, "doc3", results[2].ID) // Lowest score
}

func TestVectorStore_ConcurrentOperations(t *testing.T) {
	vs, mockDB, mockEmbedder := setupTestVectorStore(t)

	// Setup mocks for concurrent calls
	mockEmbedder.On("EmbedSingle", mock.Anything, mock.AnythingOfType("string")).Return([]float64{0.1, 0.2}, nil)
	mockDB.On("Store", mock.Anything, mock.AnythingOfType("[]vectorstore.Vector")).Return(nil)

	// Run concurrent operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			doc := Document{
				ID:      fmt.Sprintf("doc_%d", id),
				Content: fmt.Sprintf("Content %d", id),
			}
			err := vs.AddDocument(context.Background(), doc)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	mockEmbedder.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
