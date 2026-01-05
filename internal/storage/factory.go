// Storage Factory for switching between local and R2 storage
package storage

import (
	"context"
	"os"
)

type StorageType string

const (
	StorageLocal StorageType = "local"
	StorageR2    StorageType = "r2"
)

type MediaStorage interface {
	StoreWhatsAppMedia(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error)
	StoreSSHKey(ctx context.Context, username string, publicKey string) (string, error)
	GetMediaURL(path string) string
	DeleteMedia(ctx context.Context, path string) error
	ListMediaByPrefix(ctx context.Context, prefix string) ([]MediaInfo, error)
	GetStorageStats(ctx context.Context) (*StorageStats, error)
}

type LocalMediaStorage struct {
	basePath string
}

func NewLocalMediaStorage(basePath string) *LocalMediaStorage {
	return &LocalMediaStorage{
		basePath: basePath,
	}
}

func (l *LocalMediaStorage) StoreWhatsAppMedia(ctx context.Context, mediaID string, mimeType string, data []byte) (string, error) {
	// Implement local storage logic
	filename := mediaID + getFileExtension(mimeType)
	return filename, nil
}

func (l *LocalMediaStorage) StoreSSHKey(ctx context.Context, username string, publicKey string) (string, error) {
	return username + ".pub", nil
}

func (l *LocalMediaStorage) GetMediaURL(path string) string {
	return "/static/" + path
}

func (l *LocalMediaStorage) DeleteMedia(ctx context.Context, path string) error {
	return os.Remove(path)
}

func (l *LocalMediaStorage) ListMediaByPrefix(ctx context.Context, prefix string) ([]MediaInfo, error) {
	return []MediaInfo{}, nil
}

func (l *LocalMediaStorage) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	return &StorageStats{}, nil
}

// Factory function to create appropriate storage backend
func NewMediaStorage(storageType StorageType, config interface{}) MediaStorage {
	switch storageType {
	case StorageR2:
		if r2Config, ok := config.(*R2Config); ok {
			return NewR2MediaStorage(r2Config)
		}
		// Fallback to local if R2 config is invalid
		fallthrough
	default:
		return NewLocalMediaStorage("./storage")
	}
}

// Get storage type from environment
func GetStorageTypeFromEnv() StorageType {
	storageType := os.Getenv("STORAGE_TYPE")
	switch storageType {
	case "r2":
		return StorageR2
	default:
		return StorageLocal
	}
}
