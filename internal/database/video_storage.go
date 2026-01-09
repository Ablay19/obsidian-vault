package database

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

type VideoMetadata struct {
	ID               string
	UserID           string
	Title            string
	Description      string
	OriginalPrompt   string
	VideoFormat      string
	FileSizeBytes    int64
	ChunkCount       int
	CreatedAt        time.Time
	ExpiresAt        time.Time
	ProcessingStatus string
	ErrorMessage     string
	DownloadToken    string
	RetentionHours   int
}

type VideoChunk struct {
	VideoID    string
	ChunkIndex int
	ChunkData  []byte
	ChunkHash  string
}

type VideoChunk struct {
	VideoID    string `gorm:"primaryKey"`
	ChunkIndex int    `gorm:"primaryKey"`
	ChunkData  []byte `gorm:"type:BLOB"`
	ChunkHash  string
}

type VideoStorage interface {
	StoreVideo(ctx context.Context, userID, title, prompt string, videoData []byte, retentionHours int) (*VideoMetadata, error)
	GetVideo(ctx context.Context, videoID string) ([]byte, error)
	GetVideoStream(ctx context.Context, videoID string) (io.ReadCloser, error)
	GetUserVideos(ctx context.Context, userID string, limit, offset int) ([]VideoMetadata, error)
	DeleteVideo(ctx context.Context, videoID string) error
	CleanupExpired(ctx context.Context) (int, error)
	ValidateToken(ctx context.Context, token string) (string, error)
}

type DatabaseVideoStorage struct {
	db        *sql.DB
	chunkSize int
}

func NewVideoStorage(db *sql.DB) VideoStorage {
	return &DatabaseVideoStorage{
		db:        db,
		chunkSize: 1024 * 1024, // 1MB chunks
	}
}

func (s *DatabaseVideoStorage) StoreVideo(ctx context.Context, userID, title, prompt string, videoData []byte, retentionHours int) (*VideoMetadata, error) {
	videoID := generateVideoID()
	downloadToken := generateSecureToken()

	expiresAt := time.Now().Add(time.Duration(retentionHours) * time.Hour)

	// Store metadata
	metadata := VideoMetadata{
		ID:             videoID,
		UserID:         userID,
		Title:          title,
		OriginalPrompt: prompt,
		VideoFormat:    "mp4",
		FileSizeBytes:  int64(len(videoData)),
		CreatedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		DownloadToken:  downloadToken,
		RetentionHours: retentionHours,
	}

	// Chunk and store video data
	chunks := s.chunkVideo(videoData)
	metadata.ChunkCount = len(chunks)

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&metadata).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i, chunk := range chunks {
		chunkHash := hashChunk(chunk)
		chunkRecord := VideoChunk{
			VideoID:    videoID,
			ChunkIndex: i,
			ChunkData:  chunk,
			ChunkHash:  chunkHash,
		}
		if err := tx.Create(&chunkRecord).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (s *DatabaseVideoStorage) GetVideo(ctx context.Context, videoID string) ([]byte, error) {
	var chunks []VideoChunk
	if err := s.db.Where("video_id = ?", videoID).Order("chunk_index").Find(&chunks).Error; err != nil {
		return nil, err
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	// Reassemble chunks
	var videoData []byte
	for _, chunk := range chunks {
		videoData = append(videoData, chunk.ChunkData...)
	}

	return videoData, nil
}

func (s *DatabaseVideoStorage) GetVideoStream(ctx context.Context, videoID string) (io.ReadCloser, error) {
	return &videoReader{
		storage: s,
		videoID: videoID,
	}, nil
}

func (s *DatabaseVideoStorage) GetUserVideos(ctx context.Context, userID string, limit, offset int) ([]VideoMetadata, error) {
	var videos []VideoMetadata
	query := s.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&videos).Error; err != nil {
		return nil, err
	}

	return videos, nil
}

func (s *DatabaseVideoStorage) DeleteVideo(ctx context.Context, videoID string) error {
	return s.db.Where("id = ?", videoID).Delete(&VideoMetadata{}).Error
}

func (s *DatabaseVideoStorage) CleanupExpired(ctx context.Context) (int, error) {
	result := s.db.Where("expires_at < ?", time.Now()).Delete(&VideoMetadata{})
	return int(result.RowsAffected), result.Error
}

func (s *DatabaseVideoStorage) ValidateToken(ctx context.Context, token string) (string, error) {
	var video VideoMetadata
	if err := s.db.Where("download_token = ? AND expires_at > ?", token, time.Now()).First(&video).Error; err != nil {
		return "", err
	}
	return video.ID, nil
}

func (s *DatabaseVideoStorage) chunkVideo(data []byte) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += s.chunkSize {
		end := i + s.chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

type videoReader struct {
	storage      *DatabaseVideoStorage
	videoID      string
	currentChunk int
	chunks       []VideoChunk
}

func (r *videoReader) Read(p []byte) (n int, err error) {
	if r.chunks == nil {
		// Load chunks on first read
		if err := r.storage.db.Where("video_id = ?", r.videoID).Order("chunk_index").Find(&r.chunks).Error; err != nil {
			return 0, err
		}
		if len(r.chunks) == 0 {
			return 0, io.EOF
		}
	}

	if r.currentChunk >= len(r.chunks) {
		return 0, io.EOF
	}

	chunk := r.chunks[r.currentChunk]
	data := chunk.ChunkData

	if len(data) == 0 {
		r.currentChunk++
		return 0, nil
	}

	n = copy(p, data)
	r.chunks[r.currentChunk].ChunkData = data[n:] // Advance in chunk

	if len(r.chunks[r.currentChunk].ChunkData) == 0 {
		r.currentChunk++
	}

	return n, nil
}

func (r *videoReader) Close() error {
	r.chunks = nil
	return nil
}

func generateVideoID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashChunk(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
