package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"
)

var goApps = []types.GoApplication{
	{
		ID:          "goapp-001",
		Name:        "api-gateway",
		Version:     "1.0.0",
		Description: "Main API gateway service for routing requests",
		ModulePath:  "github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway",
		EntryPoint:  "cmd/main.go",
		Port:        8080,
		Database:    types.DatabaseConfig{},
		APIs: []types.APIEndpoint{
			{
				Path:   "/health",
				Method: "GET",
			},
		},
		Dependencies: []types.GoDependency{},
		Resources: types.ResourceLimits{
			CPU:     "500m",
			Memory:  "512Mi",
			Storage: "1Gi",
		},
		Status:    "production",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	},
}

func ListGoApps(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"applications": goApps,
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   len(goApps),
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GoAppDetail(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Path[len("/api/v1/go-applications/"):]
	for _, app := range goApps {
		if app.ID == appID {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(app)
			return
		}
	}
	writeError(w, http.StatusNotFound, "Application not found")
}

func ListPipelines(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"pipelines": []types.DeploymentPipeline{},
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   0,
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func PipelineDetail(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotFound, "Pipeline not found")
}

func ListSharedPackages(w http.ResponseWriter, r *http.Request) {
	packages := []types.SharedPackage{
		{
			ID:        "pkg-001",
			Name:      "shared-types",
			Version:   "1.0.0",
			Type:      "types",
			Languages: []string{"go", "typescript"},
			Exports:   []string{"types.go", "index.ts"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:        "pkg-002",
			Name:      "api-contracts",
			Version:   "1.0.0",
			Type:      "contracts",
			Languages: []string{"openapi"},
			Exports:   []string{"openapi.yaml"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		{
			ID:        "pkg-003",
			Name:      "communication",
			Version:   "1.0.0",
			Type:      "communication",
			Languages: []string{"go", "typescript"},
			Exports:   []string{"client.go", "client.ts"},
			Status:    "published",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
	response := map[string]interface{}{
		"packages": packages,
		"pagination": types.Pagination{
			Limit:   20,
			Offset:  0,
			Total:   len(packages),
			HasNext: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
