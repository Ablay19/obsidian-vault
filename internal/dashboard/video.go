package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/telemetry"
)

type VideoHandler struct {
	videoStorage database.VideoStorage
	aiService    ai.AIServiceInterface
}

func NewVideoHandler(videoStorage database.VideoStorage, aiService ai.AIServiceInterface) *VideoHandler {
	return &VideoHandler{
		videoStorage: videoStorage,
		aiService:    aiService,
	}
}

func (h *VideoHandler) RegisterRoutes(router *gin.Engine) {
	video := router.Group("/api/videos")
	{
		video.POST("/generate", h.GenerateVideo)
		video.GET("/download/:token", h.DownloadVideo)
		video.GET("", h.ListVideos)
	}
}

// GenerateVideo handles video generation requests
func (h *VideoHandler) GenerateVideo(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		telemetry.ZapLogger.Sugar().Warnw("Video generation attempted without authentication")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Prompt         string `json:"prompt" binding:"required"`
		Title          string `json:"title"`
		RetentionHours int    `json:"retention_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		telemetry.ZapLogger.Sugar().Warnw("Invalid video generation request",
			"user_id", userID,
			"error", err)
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if req.RetentionHours <= 0 || req.RetentionHours > 720 { // Max 30 days
		req.RetentionHours = 72 // Default 3 days
	}

	telemetry.ZapLogger.Sugar().Infow("Video generation request received",
		"user_id", userID,
		"prompt_length", len(req.Prompt),
		"title", req.Title,
		"retention_hours", req.RetentionHours)

	// Create video generation service
	videoGenService := NewVideoGenerationService(h.aiService, h.videoStorage)

	// Generate video asynchronously
	go func() {
		ctx := context.Background() // TODO: Use request context with timeout
		metadata, err := videoGenService.GenerateVideoFromPrompt(ctx, userID, req.Prompt, req.Title, req.RetentionHours)

		if err != nil {
			// TODO: Store error status in database and notify user via websocket
			telemetry.ZapLogger.Sugar().Errorw("Video generation failed",
				"user_id", userID,
				"prompt_truncated", func() string {
					if len(req.Prompt) > 100 {
						return req.Prompt[:100] + "..."
					}
					return req.Prompt
				}(),
				"error", err)
			return
		}

		// TODO: Notify user via websocket that video is ready
		telemetry.ZapLogger.Sugar().Infow("Video generation completed successfully",
			"video_id", metadata.ID,
			"user_id", userID,
			"video_size_bytes", metadata.FileSizeBytes)
	}()

	c.JSON(202, gin.H{
		"message": "Video generation started",
		"status":  "processing",
	})
}

// DownloadVideo handles video streaming via secure token with range request support
func (h *VideoHandler) DownloadVideo(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "Token required"})
		return
	}

	videoID, err := h.videoStorage.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(404, gin.H{"error": "Video not found or expired"})
		return
	}

	// Get full video data for range request processing
	videoData, err := h.videoStorage.GetVideo(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to access video"})
		return
	}

	videoSize := int64(len(videoData))

	// Handle range requests for efficient video streaming
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		h.serveRangeRequest(c, videoData, videoSize, rangeHeader)
		return
	}

	// Serve full video
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Length", fmt.Sprintf("%d", videoSize))
	c.Header("Accept-Ranges", "bytes")
	c.Data(http.StatusOK, "video/mp4", videoData)
}

// serveRangeRequest handles HTTP range requests for video streaming
func (h *VideoHandler) serveRangeRequest(c *gin.Context, videoData []byte, videoSize int64, rangeHeader string) {
	// Parse range header (format: "bytes=start-end")
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		c.JSON(400, gin.H{"error": "Invalid range header"})
		return
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		c.JSON(400, gin.H{"error": "Invalid range format"})
		return
	}

	startStr, endStr := parts[0], parts[1]

	var start, end int64

	// Parse start position
	if startStr == "" {
		start = 0
	} else {
		var err error
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil || start < 0 {
			c.JSON(400, gin.H{"error": "Invalid start position"})
			return
		}
	}

	// Parse end position
	if endStr == "" {
		end = videoSize - 1
	} else {
		var err error
		end, err = strconv.ParseInt(endStr, 10, 64)
		if err != nil || end >= videoSize {
			end = videoSize - 1
		}
	}

	// Validate range
	if start > end || start >= videoSize {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", videoSize))
		c.JSON(416, gin.H{"error": "Range not satisfiable"})
		return
	}

	// Calculate content length
	contentLength := end - start + 1

	// Set response headers
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Length", fmt.Sprintf("%d", contentLength))
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, videoSize))
	c.Header("Accept-Ranges", "bytes")

	// Send partial content
	c.Data(http.StatusPartialContent, "video/mp4", videoData[start:end+1])
}

// ListVideos returns user's videos
func (h *VideoHandler) ListVideos(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	videos, err := h.videoStorage.GetUserVideos(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve videos"})
		return
	}

	c.JSON(200, gin.H{
		"videos": videos,
		"limit":  limit,
		"offset": offset,
	})
}
