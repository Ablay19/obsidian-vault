package services

import (
	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

type WorkerService struct {
	logger *types.Logger
}

func NewWorkerService(logger *types.Logger) *WorkerService {
	return &WorkerService{
		logger: logger,
	}
}

func (s *WorkerService) ListWorkers() ([]types.WorkerModule, error) {
	s.logger.Info("Listing all workers")

	workers := []types.WorkerModule{
		{
			ID:          "ai-worker-001",
			Name:        "ai-worker",
			Version:     "1.0.0",
			Description: "AI processing worker for natural language tasks",
			EntryPoint:  "src/index.ts",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
		{
			ID:          "image-worker-001",
			Name:        "image-worker",
			Version:     "1.0.0",
			Description: "Image processing worker for vision tasks",
			EntryPoint:  "src/index.ts",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	s.logger.Info("Listed workers", "count", len(workers))
	return workers, nil
}

func (s *WorkerService) GetWorker(id string) (*types.WorkerModule, error) {
	s.logger.Info("Getting worker", "id", id)

	workers := map[string]types.WorkerModule{
		"ai-worker-001": {
			ID:          "ai-worker-001",
			Name:        "ai-worker",
			Version:     "1.0.0",
			Description: "AI processing worker for natural language tasks",
			EntryPoint:  "src/index.ts",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
		"image-worker-001": {
			ID:          "image-worker-001",
			Name:        "image-worker",
			Version:     "1.0.0",
			Description: "Image processing worker for vision tasks",
			EntryPoint:  "src/index.ts",
			Status:      "active",
			CreatedAt:   "2024-01-01T00:00:00Z",
			UpdatedAt:   "2024-01-01T00:00:00Z",
		},
	}

	worker, exists := workers[id]
	if !exists {
		s.logger.Warn("Worker not found", "id", id)
		return nil, nil
	}

	s.logger.Info("Got worker", "id", id, "name", worker.Name)
	return &worker, nil
}
