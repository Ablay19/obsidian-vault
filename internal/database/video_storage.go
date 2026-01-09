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

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert metadata
	_, err = tx.Exec(`
		INSERT INTO videos (id, user_id, title, description, original_prompt, video_format, file_size_bytes, chunk_count, created_at, expires_at, processing_status, download_token, retention_hours)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metadata.ID, metadata.UserID, metadata.Title, metadata.Description, metadata.OriginalPrompt,
		metadata.VideoFormat, metadata.FileSizeBytes, metadata.ChunkCount, metadata.CreatedAt, metadata.ExpiresAt,
		metadata.ProcessingStatus, metadata.DownloadToken, metadata.RetentionHours)
	if err != nil {
		return nil, err
	}

	// Insert chunks
	for i, chunk := range chunks {
		chunkHash := hashChunk(chunk)
		_, err = tx.Exec(`
			INSERT INTO video_chunks (video_id, chunk_index, chunk_data, chunk_hash)
			VALUES (?, ?, ?, ?)`,
			videoID, i, chunk, chunkHash)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (s *DatabaseVideoStorage) GetVideo(ctx context.Context, videoID string) ([]byte, error) {
	rows, err := s.db.Query("SELECT chunk_data FROM video_chunks WHERE video_id = ? ORDER BY chunk_index", videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videoData []byte
	for rows.Next() {
		var chunkData []byte
		if err := rows.Scan(&chunkData); err != nil {
			return nil, err
		}
		videoData = append(videoData, chunkData...)
	}

	if len(videoData) == 0 {
		return nil, fmt.Errorf("video not found")
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
	query := "SELECT id, user_id, title, description, original_prompt, video_format, file_size_bytes, chunk_count, created_at, expires_at, processing_status, error_message, download_token, retention_hours FROM videos WHERE user_id = ? ORDER BY created_at DESC"
	args := []interface{}{userID}

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []VideoMetadata
	for rows.Next() {
		var v VideoMetadata
		err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.OriginalPrompt, &v.VideoFormat,
			&v.FileSizeBytes, &v.ChunkCount, &v.CreatedAt, &v.ExpiresAt, &v.ProcessingStatus, &v.ErrorMessage,
			&v.DownloadToken, &v.RetentionHours)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}

	return videos, nil
}

func (s *DatabaseVideoStorage) DeleteVideo(ctx context.Context, videoID string) error {
	// Delete chunks first
	_, err := s.db.Exec("DELETE FROM video_chunks WHERE video_id = ?", videoID)
	if err != nil {
		return err
	}
	// Delete metadata
	_, err = s.db.Exec("DELETE FROM videos WHERE id = ?", videoID)
	return err
}

func (s *DatabaseVideoStorage) CleanupExpired(ctx context.Context) (int, error) {
	result, err := s.db.Exec("DELETE FROM videos WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

func (s *DatabaseVideoStorage) ValidateToken(ctx context.Context, token string) (string, error) {
	var videoID string
	err := s.db.QueryRow("SELECT id FROM videos WHERE download_token = ? AND expires_at > ?", token, time.Now()).Scan(&videoID)
	if err != nil {
		return "", err
	}
	return videoID, nil
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
		rows, err := r.storage.db.Query("SELECT chunk_data FROM video_chunks WHERE video_id = ? ORDER BY chunk_index", r.videoID)
		if err != nil {
			return 0, err
		}
		defer rows.Close()

		r.chunks = []VideoChunk{}
		for rows.Next() {
			var chunkData []byte
			if err := rows.Scan(&chunkData); err != nil {
				return 0, err
			}
			r.chunks = append(r.chunks, VideoChunk{ChunkData: chunkData})
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
