package visualizer

import (
	"context"
	"fmt"
	"sync"
)

// MockStorage is an in-memory implementation for testing
type MockStorage struct {
	problems map[string]*Problem
	analyses map[string]*AnalysisResult
	mutex    sync.RWMutex
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		problems: make(map[string]*Problem),
		analyses: make(map[string]*AnalysisResult),
	}
}

// StoreProblem stores a problem in memory
func (m *MockStorage) StoreProblem(ctx context.Context, problem *Problem) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.problems[problem.ID] = problem
	return nil
}

// StoreAnalysis stores an analysis in memory
func (m *MockStorage) StoreAnalysis(ctx context.Context, analysis *AnalysisResult) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.analyses[analysis.ProblemID] = analysis
	return nil
}

// GetProblem retrieves a problem from memory
func (m *MockStorage) GetProblem(ctx context.Context, id string) (*Problem, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	problem, exists := m.problems[id]
	if !exists {
		return nil, fmt.Errorf("problem not found: %s", id)
	}

	return problem, nil
}

// GetAnalysis retrieves an analysis from memory
func (m *MockStorage) GetAnalysis(ctx context.Context, problemID string) (*AnalysisResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	analysis, exists := m.analyses[problemID]
	if !exists {
		return nil, fmt.Errorf("analysis not found: %s", problemID)
	}

	return analysis, nil
}

// UpdateProblem updates a problem in memory
func (m *MockStorage) UpdateProblem(ctx context.Context, problem *Problem) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	existing, exists := m.problems[problem.ID]
	if !exists {
		return fmt.Errorf("problem not found: %s", problem.ID)
	}

	existing.Title = problem.Title
	existing.Description = problem.Description
	existing.UpdatedAt = problem.UpdatedAt

	return nil
}

// GetAllProblems returns all stored problems
func (m *MockStorage) GetAllProblems(ctx context.Context) ([]*Problem, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	problems := make([]*Problem, 0, len(m.problems))
	i := 0
	for _, problem := range m.problems {
		problems[i] = problem
		i++
	}

	return problems, nil
}

// GetAllAnalyses returns all stored analyses
func (m *MockStorage) GetAllAnalyses(ctx context.Context) ([]*AnalysisResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	analyses := make([]*AnalysisResult, 0, len(m.analyses))
	i := 0
	for _, analysis := range m.analyses {
		analyses[i] = analysis
		i++
	}

	return analyses, nil
}

// Clear clears all stored data
func (m *MockStorage) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.problems = make(map[string]*Problem)
	m.analyses = make(map[string]*AnalysisResult)
}
