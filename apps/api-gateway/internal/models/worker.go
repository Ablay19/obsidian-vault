package models

import (
	"fmt"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var workers = []types.WorkerModule{
	{
		ID:           "worker-001",
		Name:         "ai-worker",
		Version:      "1.0.0",
		Description:  "AI processing worker for natural language tasks",
		EntryPoint:   "src/index.ts",
		Dependencies: []string{"@obsidian-vault/shared-types", "@obsidian-vault/communication"},
		Environment:  map[string]string{"NODE_ENV": "production"},
		Routes: []types.RouteMapping{
			{
				Path:    "/health",
				Method:  "GET",
				Timeout: 5000,
			},
		},
		Permissions: []string{},
		Resources: types.ResourceLimits{
			CPU:     "256m",
			Memory:  "512Mi",
			Storage: "1Gi",
		},
		Status:    "production",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	},
}

func GetAllWorkers() []types.WorkerModule {
	return workers
}

func GetWorkerByID(id string) (*types.WorkerModule, bool) {
	for i := range workers {
		if workers[i].ID == id {
			return &workers[i], true
		}
	}
	return nil, false
}

func CreateWorker(worker types.WorkerModule) error {
	worker.ID = fmt.Sprintf("worker-%03d", len(workers)+1)
	worker.Status = "development"
	worker.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	worker.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	workers = append(workers, worker)
	return nil
}

func UpdateWorker(id string, updated types.WorkerModule) (*types.WorkerModule, error) {
	for i := range workers {
		if workers[i].ID == id {
			updated.ID = id
			updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			workers[i] = updated
			return &workers[i], nil
		}
	}
	return nil, fmt.Errorf("worker not found")
}

func DeleteWorker(id string) error {
	for i := range workers {
		if workers[i].ID == id {
			workers = append(workers[:i], workers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("worker not found")
}
