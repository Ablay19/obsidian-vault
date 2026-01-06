package whatsapp

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

// LocalMediaStorage implements MediaStorage interface using local filesystem
type LocalMediaStorage struct {
	basePath string
	logger   *otelzap.Logger
}

// NewLocalMediaStorage creates a new local media storage
func NewLocalMediaStorage(basePath string, logger *otelzap.Logger) MediaStorage {
	return &LocalMediaStorage{
		basePath: basePath,
		logger:   logger,
	}
}

// Store stores media data to local filesystem
func (lms *LocalMediaStorage) Store(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error) {
	// Create directory structure if it doesn't exist
	datePath := time.Now().Format("2006/01/02")
	fullPath := filepath.Join(lms.basePath, "whatsapp", datePath)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename with hash to avoid conflicts
	extension := lms.getExtensionFromMimeType(mimeType)
	hash := sha256.Sum256(data)
	hashStr := fmt.Sprintf("%x", hash)[:16]
	filename := fmt.Sprintf("%s_%s%s", mediaID, hashStr, extension)

	localPath := filepath.Join(fullPath, filename)

	// Write file
	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	lms.logger.Info("Media stored successfully",
		zap.String("media_id", mediaID),
		zap.String("local_path", localPath),
		zap.Int("size", len(data)))

	return localPath, nil
}

// getExtensionFromMimeType returns file extension from MIME type
func (lms *LocalMediaStorage) getExtensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	case "audio/mpeg":
		return ".mp3"
	case "audio/mp4":
		return ".m4a"
	case "video/mp4":
		return ".mp4"
	case "video/quicktime":
		return ".mov"
	default:
		return ".bin"
	}
}

// InMemoryMediaStorage implements MediaStorage interface using in-memory storage
type InMemoryMediaStorage struct {
	storage map[string]StoredMedia
	logger  *otelzap.Logger
}

type StoredMedia struct {
	ID        string    `json:"id"`
	MimeType  string    `json:"mime_type"`
	Data      []byte    `json:"data"`
	StoredAt  time.Time `json:"stored_at"`
	LocalPath string    `json:"local_path"`
}

// NewInMemoryMediaStorage creates a new in-memory media storage
func NewInMemoryMediaStorage(logger *otelzap.Logger) MediaStorage {
	return &InMemoryMediaStorage{
		storage: make(map[string]StoredMedia),
		logger:  logger,
	}
}

// Store stores media data in memory
func (ims *InMemoryMediaStorage) Store(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error) {
	// Generate a mock local path for consistency
	localPath := fmt.Sprintf("/tmp/whatsapp_media/%s_%s", mediaID, time.Now().Format("20060102150405"))

	stored := StoredMedia{
		ID:        mediaID,
		MimeType:  mimeType,
		Data:      data,
		StoredAt:  time.Now(),
		LocalPath: localPath,
	}

	ims.storage[mediaID] = stored

	ims.logger.Info("Media stored in memory",
		zap.String("media_id", mediaID),
		zap.String("mime_type", mimeType),
		zap.Int("size", len(data)))

	return localPath, nil
}

// Get retrieves stored media data (helper method for testing)
func (ims *InMemoryMediaStorage) Get(mediaID string) (*StoredMedia, bool) {
	stored, exists := ims.storage[mediaID]
	return &stored, exists
}

// Clear clears all stored media (helper method for testing)
func (ims *InMemoryMediaStorage) Clear() {
	ims.storage = make(map[string]StoredMedia)
}
