package dashboard

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
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
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Prompt         string `json:"prompt" binding:"required"`
		Title          string `json:"title"`
		RetentionHours int    `json:"retention_hours"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if req.RetentionHours <= 0 || req.RetentionHours > 720 { // Max 30 days
		req.RetentionHours = 72 // Default 3 days
	}

	// TODO: Implement actual video generation with AI
	// For now, return placeholder response
	c.JSON(202, gin.H{
		"message": "Video generation started",
		"job_id":  "placeholder",
		"status":  "queued",
	})
}

// DownloadVideo handles video downloads via secure token
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

	stream, err := h.videoStorage.GetVideoStream(c.Request.Context(), videoID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to access video"})
		return
	}
	defer stream.Close()

	c.DataFromReader(http.StatusOK, -1, "video/mp4", stream, nil)
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
