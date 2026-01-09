package dashboard

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/telemetry"
)

// VideoGenerationService handles the generation of educational videos using Manim
type VideoGenerationService struct {
	aiService    ai.AIServiceInterface
	videoStorage database.VideoStorage
	workDir      string
}

// NewVideoGenerationService creates a new video generation service
func NewVideoGenerationService(aiService ai.AIServiceInterface, videoStorage database.VideoStorage) *VideoGenerationService {
	workDir := os.Getenv("VIDEO_WORK_DIR")
	if workDir == "" {
		workDir = "/tmp/video-generation"
	}

	// Ensure work directory exists
	os.MkdirAll(workDir, 0755)

	return &VideoGenerationService{
		aiService:    aiService,
		videoStorage: videoStorage,
		workDir:      workDir,
	}
}

// GenerateVideoFromPrompt generates a video from a text prompt
func (s *VideoGenerationService) GenerateVideoFromPrompt(ctx context.Context, userID, prompt, title string, retentionHours int) (*database.VideoMetadata, error) {
	startTime := time.Now()
	success := false
	errorStage := ""

	telemetry.ZapLogger.Sugar().Infow("Starting video generation",
		"user_id", userID,
		"prompt_length", len(prompt),
		"title", title)

	// Track video generation attempts
	defer func() {
		duration := time.Since(startTime)
		status := "success"
		if !success {
			status = "failed"
		}

		telemetry.ZapLogger.Sugar().Infow("Video generation finished",
			"user_id", userID,
			"duration_ms", duration.Milliseconds(),
			"status", status,
			"error_stage", errorStage)
	}()

	// Generate Manim Python code from prompt
	manimCode, err := s.generateManimCode(ctx, prompt, title)
	if err != nil {
		errorStage = "ai_code_generation"
		telemetry.ZapLogger.Sugar().Errorw("Video generation failed at AI code generation",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds())
		return nil, fmt.Errorf("failed to generate manim code: %w", err)
	}

	// Create temporary directory for this generation job
	jobDir := filepath.Join(s.workDir, fmt.Sprintf("job_%d_%s", time.Now().Unix(), userID))
	os.MkdirAll(jobDir, 0755)
	defer os.RemoveAll(jobDir) // Clean up after generation

	// Write Manim code to file
	codeFile := filepath.Join(jobDir, "scene.py")
	if err := os.WriteFile(codeFile, []byte(manimCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write manim code: %w", err)
	}

	// Render video using Manim
	videoPath, err := s.renderVideo(jobDir, codeFile)
	if err != nil {
		errorStage = "video_rendering"
		telemetry.ZapLogger.Sugar().Errorw("Video generation failed at rendering",
			"user_id", userID,
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds())
		return nil, fmt.Errorf("failed to render video: %w", err)
	}

	// Read video file
	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		errorStage = "video_read"
		telemetry.ZapLogger.Sugar().Errorw("Video generation failed at file read",
			"user_id", userID,
			"error", err,
			"video_path", videoPath,
			"duration_ms", time.Since(startTime).Milliseconds())
		return nil, fmt.Errorf("failed to read generated video: %w", err)
	}

	videoSizeMB := float64(len(videoData)) / (1024 * 1024)

	// Store video in database
	metadata, err := s.videoStorage.StoreVideo(ctx, userID, title, prompt, videoData, retentionHours)
	if err != nil {
		errorStage = "video_storage"
		telemetry.ZapLogger.Sugar().Errorw("Video generation failed at storage",
			"user_id", userID,
			"error", err,
			"video_size_mb", videoSizeMB,
			"duration_ms", time.Since(startTime).Milliseconds())
		return nil, fmt.Errorf("failed to store video: %w", err)
	}

	// Mark as successful
	success = true

	telemetry.ZapLogger.Sugar().Infow("Video generation completed successfully",
		"video_id", metadata.ID,
		"user_id", userID,
		"video_size_mb", videoSizeMB,
		"duration_ms", time.Since(startTime).Milliseconds())
	return metadata, nil
}

// generateManimCode uses AI to generate Manim Python code from a prompt
func (s *VideoGenerationService) generateManimCode(ctx context.Context, prompt, title string) (string, error) {
	systemPrompt := `You are an expert at creating educational videos using Manim (Mathematical Animation Engine).

Your task is to generate Python code using Manim Community v0.18.0+ that creates an engaging educational video based on the user's prompt.

Requirements:
- Use Manim Community syntax (Scene, construct() method, etc.)
- Create educational content with clear explanations
- Use appropriate animations for the subject matter
- Include text explanations and mathematical notations when relevant
- Keep videos concise (30-90 seconds)
- Use high-quality animations with proper timing
- Include a title scene and conclusion
- Handle both mathematical and general educational topics

The code should be complete, runnable Python code that imports from manim and defines a Scene class.

Return ONLY the Python code, no explanations or markdown formatting.`

	userPrompt := fmt.Sprintf("Create a Manim video about: %s\n\nTitle: %s", prompt, title)

	req := &ai.RequestModel{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		MaxTokens:    2000,
		Temperature:  0.7,
	}

	var generatedCode strings.Builder
	err := s.aiService.Chat(ctx, req, func(chunk string) {
		generatedCode.WriteString(chunk)
	})

	if err != nil {
		return "", fmt.Errorf("AI chat failed: %w", err)
	}

	code := strings.TrimSpace(generatedCode.String())

	// Basic validation - ensure it looks like Python code
	if !strings.Contains(code, "from manim import") && !strings.Contains(code, "import manim") {
		return "", fmt.Errorf("generated code does not appear to be valid Manim code")
	}

	if !strings.Contains(code, "class") || !strings.Contains(code, "Scene") {
		return "", fmt.Errorf("generated code does not contain a Scene class")
	}

	return code, nil
}

// renderVideo executes Manim in a Docker container to render the video
func (s *VideoGenerationService) renderVideo(jobDir, codeFile string) (string, error) {
	// Build Docker image name
	imageName := "obsidian-manim:latest"

	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		telemetry.ZapLogger.Sugar().Warnw("Docker not found, falling back to direct execution", "error", err)
		return s.renderVideoDirect(jobDir, codeFile)
	}

	// Build Docker image if it doesn't exist
	if err := s.buildDockerImage(jobDir, imageName); err != nil {
		telemetry.ZapLogger.Sugar().Warnw("Failed to build Docker image, falling back to direct execution", "error", err)
		return s.renderVideoDirect(jobDir, codeFile)
	}

	// Run Manim in Docker container
	containerName := fmt.Sprintf("manim-render-%d", time.Now().Unix())

	// Mount job directory as volume and run manim
	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", jobDir),
		"-w", "/workspace",
		imageName,
		"python", "-m", "manim", "-pql", filepath.Base(codeFile), "Scene")

	output, err := cmd.CombinedOutput()
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Docker Manim rendering failed",
			"error", err,
			"output", string(output),
			"job_dir", jobDir,
			"container", containerName)
		return "", fmt.Errorf("docker manim rendering failed: %s", string(output))
	}

	// Find the generated video file
	videoPath := filepath.Join(jobDir, "media", "videos", "scene", "1080p60", "Scene.mp4")

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		// Try alternative paths
		alternatives := []string{
			filepath.Join(jobDir, "media", "videos", "Scene.mp4"),
			filepath.Join(jobDir, "Scene.mp4"),
		}

		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				videoPath = alt
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("video file not found in expected locations")
		}
	}

	return videoPath, nil
}

// renderVideoDirect executes Manim directly (fallback when Docker is not available)
func (s *VideoGenerationService) renderVideoDirect(jobDir, codeFile string) (string, error) {
	// Change to job directory
	cmd := exec.Command("python", "-m", "manim", "-pql", filepath.Base(codeFile), "Scene")
	cmd.Dir = jobDir

	// Set up environment
	cmd.Env = append(os.Environ(),
		"PYTHONPATH=/usr/local/lib/python3.9/site-packages",
		"MANIM_CONFIG_DIR="+jobDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Direct Manim rendering failed",
			"error", err,
			"output", string(output),
			"job_dir", jobDir)
		return "", fmt.Errorf("manim rendering failed: %s", string(output))
	}

	// Find the generated video file (same logic as Docker version)
	videoPath := filepath.Join(jobDir, "media", "videos", "scene", "1080p60", "Scene.mp4")

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		alternatives := []string{
			filepath.Join(jobDir, "media", "videos", "Scene.mp4"),
			filepath.Join(jobDir, "Scene.mp4"),
		}

		found := false
		for _, alt := range alternatives {
			if _, err := os.Stat(alt); err == nil {
				videoPath = alt
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("video file not found in expected locations")
		}
	}

	return videoPath, nil
}

// buildDockerImage builds the Manim Docker image if it doesn't exist
func (s *VideoGenerationService) buildDockerImage(jobDir, imageName string) error {
	// Check if image already exists
	cmd := exec.Command("docker", "images", "-q", imageName)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// Image exists
		return nil
	}

	// Build the image
	dockerfilePath := filepath.Join(jobDir, "..", "..", "docker", "manim.Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("dockerfile not found at %s", dockerfilePath)
	}

	cmd = exec.Command("docker", "build", "-f", dockerfilePath, "-t", imageName, ".")
	cmd.Dir = filepath.Dir(dockerfilePath)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build docker image: %s", string(output))
	}

	telemetry.ZapLogger.Sugar().Infow("Docker image built successfully", "image", imageName)
	return nil
}
