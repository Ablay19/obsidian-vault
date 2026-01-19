package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

// AIManimService handles communication with Cloudflare AI Manim workers
type AIManimService struct {
	config     *utils.Config
	httpClient *http.Client
	baseURL    string
	apiKey     string
	logger     *utils.Logger
}

// VideoOptions represents video generation options (aligned with worker types)
type VideoOptions struct {
	Quality            string `json:"quality,omitempty"`     // low, medium, high, ultra
	Format             string `json:"format,omitempty"`      // mp4, webm
	MaxDurationSeconds int    `json:"maxDuration,omitempty"` // seconds
}

// VideoJob represents a video generation job (aligned with worker ProcessingJob)
type VideoJob struct {
	JobID                 string     `json:"jobId"`
	SessionID             string     `json:"sessionId"`
	Problem               string     `json:"problem"`
	SubmissionType        string     `json:"submissionType"` // "problem" | "direct_code"
	Platform              string     `json:"platform"`       // "telegram" | "whatsapp" | "web"
	Status                string     `json:"status"`         // ProcessingStatus from worker
	StatusMessage         string     `json:"statusMessage,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	StartedAt             *time.Time `json:"startedAt,omitempty"`
	CompletedAt           *time.Time `json:"completedAt,omitempty"`
	VideoURL              string     `json:"videoUrl,omitempty"`
	AIProviderUsed        string     `json:"aiProviderUsed,omitempty"`
	RenderDurationSeconds *int       `json:"renderDurationSeconds,omitempty"`
	VideoSizeBytes        *int64     `json:"videoSizeBytes,omitempty"`
	Error                 string     `json:"error,omitempty"`
	RetryCount            int        `json:"retryCount,omitempty"`
}

// JobStatus represents the current status of a job
type JobStatus struct {
	JobID       string     `json:"jobId"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"` // 0-100
	ETA         int        `json:"eta"`      // seconds remaining
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// NewAIManimService creates a new AI Manim service client
func NewAIManimService(config *utils.Config, logger *utils.Logger) *AIManimService {
	return &AIManimService{
		config: config,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: config.AI.ManimWorkerURL,
		apiKey:  config.AI.APIKey,
		logger:  logger,
	}
}

// GenerateVideo starts a video generation job for a mathematical problem
func (s *AIManimService) GenerateVideo(ctx context.Context, problem string, options VideoOptions) (*VideoJob, error) {
	requestData := map[string]interface{}{
		"problem": problem,
		"options": options,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/generate", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.logger.Error("AI service error", fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var job VideoJob
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	s.logger.Info("Started video generation job", "jobID", job.JobID, "problem", problem[:50]+"...")
	return &job, nil
}

// GetJobStatus retrieves the current status of a video generation job
func (s *AIManimService) GetJobStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	url := fmt.Sprintf("%s/api/job/%s/status", s.baseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("job %s not found", jobID)
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var status JobStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &status, nil
}

// DownloadVideo downloads a completed video file
func (s *AIManimService) DownloadVideo(ctx context.Context, jobID string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/job/%s/download", s.baseURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	videoData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video data: %w", err)
	}

	s.logger.Info("Downloaded video", "jobID", jobID, "bytes", len(videoData))
	return videoData, nil
}

// IsHealthy checks if the AI service is available
func (s *AIManimService) IsHealthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

// WaitForCompletion polls job status until completion or timeout
func (s *AIManimService) WaitForCompletion(ctx context.Context, jobID string, timeout time.Duration) (*JobStatus, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Until(deadline)):
			return nil, fmt.Errorf("timeout waiting for job %s completion", jobID)
		case <-ticker.C:
			status, err := s.GetJobStatus(ctx, jobID)
			if err != nil {
				s.logger.Warn("Failed to get job status", "error", err)
				continue
			}

			if status.Status == "completed" {
				return status, nil
			} else if status.Status == "failed" {
				return status, fmt.Errorf("job failed: %s", status.Error)
			}

			s.logger.Debug("Job status update", "jobID", jobID, "status", status.Status, "progress", status.Progress)
		}
	}
}
