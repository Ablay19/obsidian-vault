package ai

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// AIJob represents an AI video generation job
type AIJob struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"userId"`
	Platform  string                 `json:"platform"` // telegram, whatsapp
	Problem   string                 `json:"problem"`
	Status    string                 `json:"status"`          // pending, processing, completed, failed
	JobID     string                 `json:"jobId,omitempty"` // Cloudflare job ID
	VideoURL  string                 `json:"videoUrl,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AISessionManager manages AI job sessions and tracking
type AISessionManager struct {
	jobs   map[string]*AIJob
	mutex  sync.RWMutex
	logger *utils.Logger
}

// NewAISessionManager creates a new AI session manager
func NewAISessionManager(logger *utils.Logger) *AISessionManager {
	return &AISessionManager{
		jobs:   make(map[string]*AIJob),
		logger: logger,
	}
}

// CreateJob creates a new AI job for tracking
func (sm *AISessionManager) CreateJob(userID, problem, platform, jobType string) (*AIJob, error) {
	jobID, err := sm.generateJobID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate job ID: %w", err)
	}

	job := &AIJob{
		ID:        jobID,
		UserID:    userID,
		Platform:  platform,
		Problem:   problem,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"type": jobType,
		},
	}

	sm.mutex.Lock()
	sm.jobs[jobID] = job
	sm.mutex.Unlock()

	sm.logger.Info("Created AI job", "jobID", jobID, "userID", userID, "problem", problem[:50]+"...")

	return job, nil
}

// GetJob retrieves a job by ID
func (sm *AISessionManager) GetJob(jobID string) (*AIJob, error) {
	sm.mutex.RLock()
	job, exists := sm.jobs[jobID]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	return job, nil
}

// UpdateJobStatus updates the status of an AI job
func (sm *AISessionManager) UpdateJobStatus(jobID, status string, metadata map[string]interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	job, exists := sm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}

	oldStatus := job.Status
	job.Status = status
	job.UpdatedAt = time.Now()

	// Update metadata if provided
	if metadata != nil {
		for k, v := range metadata {
			job.Metadata[k] = v
		}

		// Handle special metadata fields
		if videoURL, ok := metadata["video_url"].(string); ok {
			job.VideoURL = videoURL
		}
		if cfJobID, ok := metadata["cloudflare_job_id"].(string); ok {
			job.JobID = cfJobID
		}
		if err, ok := metadata["error"].(string); ok {
			job.Error = err
		}
	}

	sm.logger.Info("Updated job status", "jobID", jobID, "oldStatus", oldStatus, "newStatus", status)

	return nil
}

// ListUserJobs returns all jobs for a user
func (sm *AISessionManager) ListUserJobs(userID string, limit int) []*AIJob {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var userJobs []*AIJob
	count := 0

	for _, job := range sm.jobs {
		if job.UserID == userID {
			userJobs = append(userJobs, job)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(userJobs)-1; i++ {
		for j := i + 1; j < len(userJobs); j++ {
			if userJobs[i].CreatedAt.Before(userJobs[j].CreatedAt) {
				userJobs[i], userJobs[j] = userJobs[j], userJobs[i]
			}
		}
	}

	return userJobs
}

// CleanupOldJobs removes completed jobs older than the specified duration
func (sm *AISessionManager) CleanupOldJobs(maxAge time.Duration) int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	removed := 0

	for id, job := range sm.jobs {
		if job.Status == "completed" || job.Status == "failed" {
			if job.UpdatedAt.Before(cutoff) {
				delete(sm.jobs, id)
				removed++
			}
		}
	}

	if removed > 0 {
		sm.logger.Info("Cleaned up old AI jobs", "count", removed)
	}

	return removed
}

// GetStats returns session manager statistics
func (sm *AISessionManager) GetStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_jobs": len(sm.jobs),
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
	}

	for _, job := range sm.jobs {
		switch job.Status {
		case "pending":
			stats["pending"] = stats["pending"].(int) + 1
		case "processing":
			stats["processing"] = stats["processing"].(int) + 1
		case "completed":
			stats["completed"] = stats["completed"].(int) + 1
		case "failed":
			stats["failed"] = stats["failed"].(int) + 1
		}
	}

	return stats
}

// generateJobID generates a unique job ID
func (sm *AISessionManager) generateJobID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ai_" + hex.EncodeToString(bytes), nil
}
