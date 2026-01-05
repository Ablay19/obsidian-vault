// R2 Media Storage Interface for Obsidian Bot
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
)

// R2MediaStorage interface for Cloudflare R2 integration
type R2MediaStorage struct {
	bucket    string
	cdnDomain string
	publicURL string
	apiToken  string
	accountID string
}

type R2Config struct {
	AccountID    string
	APIToken     string
	BucketName   string
	CustomDomain string
}

func NewR2MediaStorage(cfg *R2Config) *R2MediaStorage {
	return &R2MediaStorage{
		bucket:    cfg.BucketName,
		cdnDomain: cfg.CustomDomain,
		publicURL: fmt.Sprintf("https://%s", cfg.CustomDomain),
		apiToken:  cfg.APIToken,
		accountID: cfg.AccountID,
	}
}

// StoreWhatsAppMedia stores WhatsApp media with automatic organization and CDN URL generation
func (r *R2MediaStorage) StoreWhatsAppMedia(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error) {
	// Generate hash for deduplication (maintain existing logic)
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	// Generate organized path with date
	now := time.Now()
	datePath := now.Format("2006/01/02")

	// Determine file extension from MIME type
	ext := getFileExtension(mimeType)
	filename := fmt.Sprintf("%s_%s%s", mediaID, hashStr[:8], ext)
	key := fmt.Sprintf("whatsapp/%s/%s", datePath, filename)

	log.Printf("Would upload to R2: key=%s, size=%d bytes", key, len(data))
	log.Printf("Generated CDN URL: https://%s/%s", r.cdnDomain, key)

	// TODO: Implement actual R2 upload using HTTP API
	// For now, return placeholder URL
	publicURL := fmt.Sprintf("https://%s/%s", r.publicURL, key)
	return publicURL, nil
}

// StoreSSHKey stores SSH public keys securely
func (r *R2MediaStorage) StoreSSHKey(ctx context.Context, username string, publicKey string) (string, error) {
	key := fmt.Sprintf("ssh/keys/%s.pub", username)

	log.Printf("Would store SSH key: key=%s, length=%d", key, len(publicKey))

	// TODO: Implement actual R2 upload
	publicURL := fmt.Sprintf("%s/%s", r.publicURL, key)
	return publicURL, nil
}

// GetMediaURL returns the public URL for a stored media file
func (r *R2MediaStorage) GetMediaURL(path string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, strings.TrimPrefix(path, "/"))
}

// DeleteMedia removes a media file from storage
func (r *R2MediaStorage) DeleteMedia(ctx context.Context, path string) error {
	key := strings.TrimPrefix(path, "/")

	log.Printf("Would delete from R2: key=%s", key)
	// TODO: Implement actual R2 deletion
	return nil
}

// ListMediaByPrefix lists all media files with a given prefix
func (r *R2MediaStorage) ListMediaByPrefix(ctx context.Context, prefix string) ([]MediaInfo, error) {
	log.Printf("Would list R2 objects with prefix: %s", prefix)

	// TODO: Implement actual R2 listing
	return []MediaInfo{}, nil
}

// GetStorageStats returns storage usage statistics
func (r *R2MediaStorage) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	log.Printf("Would get R2 storage stats for bucket: %s", r.bucket)

	// TODO: Implement actual R2 stats
	return &StorageStats{
		TotalObjects:  0,
		TotalSize:     0,
		WhatsappFiles: 0,
		SSHFiles:      0,
	}, nil
}

// Helper functions
func getFileExtension(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"video/mp4":       ".mp4",
		"video/quicktime": ".mov",
		"audio/mpeg":      ".mp3",
		"audio/wav":       ".wav",
		"audio/ogg":       ".ogg",
		"application/pdf": ".pdf",
		"text/plain":      ".txt",
		"application/zip": ".zip",
	}

	if ext, exists := mimeToExt[mimeType]; exists {
		return ext
	}

	return ".bin" // fallback
}

type MediaInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	PublicURL    string    `json:"publicUrl"`
}

type StorageStats struct {
	TotalObjects  int   `json:"totalObjects"`
	TotalSize     int64 `json:"totalSize"`
	WhatsappFiles int   `json:"whatsappFiles"`
	SSHFiles      int   `json:"sshFiles"`
}

// Migration helper for existing local storage
func (r *R2MediaStorage) MigrateFromLocalStorage(localPath string) error {
	log.Printf("Starting migration from local storage: %s", localPath)

	// This would scan local storage and upload to R2
	// Implementation depends on current local storage structure
	return nil
}
