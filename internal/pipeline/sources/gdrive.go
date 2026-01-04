package sources

import (
	"context"
	"obsidian-automation/internal/pipeline"
	"time"

	"go.uber.org/zap"
)

// GoogleDriveSource is a connector for Google Drive.
// It uses the authenticated user's client to poll for new files.
type GoogleDriveSource struct {
	folderID string
	interval time.Duration
}

func NewGoogleDriveSource(folderID string) *GoogleDriveSource {
	return &GoogleDriveSource{
		folderID: folderID,
		interval: 1 * time.Minute,
	}
}

func (s *GoogleDriveSource) Name() string {
	return "gdrive"
}

func (s *GoogleDriveSource) Start(ctx context.Context, jobChan chan<- pipeline.Job) error {
	zap.S().Info("Starting Google Drive Source", "folder_id", s.folderID)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.poll(ctx, jobChan)
		}
	}
}

func (s *GoogleDriveSource) poll(ctx context.Context, jobChan chan<- pipeline.Job) {
	// Placeholder for actual GDrive polling logic using internal/gcp or official library.
	// This would:
	// 1. List files in s.folderID using authenticated client (from internal/auth).
	// 2. Check if file is new (state tracking).
	// 3. Download file content.
	// 4. Create pipeline.Job.
	// 5. jobChan <- job
	zap.S().Debug("Polling Google Drive (stub)")
}
