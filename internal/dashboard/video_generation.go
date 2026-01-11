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

type VideoGenerationService struct {
	aiService    ai.AIServiceInterface
	videoStorage database.VideoStorage
	workDir      string
}

func NewVideoGenerationService(aiService ai.AIServiceInterface, videoStorage database.VideoStorage) *VideoGenerationService {
	workDir := os.Getenv("VIDEO_WORK_DIR")
	if workDir == "" {
		workDir = "/tmp/video-generation"
	}

	os.MkdirAll(workDir, 0755)

	return &VideoGenerationService{
		aiService:    aiService,
		videoStorage: videoStorage,
		workDir:      workDir,
	}
}

func (s *VideoGenerationService) GenerateVideoFromPrompt(ctx context.Context, userID, prompt, title string, retentionHours int) (*database.VideoMetadata, error) {
	success := false

	telemetry.Info("Starting video generation for user: " + userID)

	defer func() {
		status := "success"
		if !success {
			status = "failed"
		}
		telemetry.Info("Video generation finished for user " + userID + ": status=" + status)
	}()

	manimCode, err := s.generateManimCode(ctx, prompt, title)
	if err != nil {
		telemetry.Error("Video generation failed at AI code generation for user " + userID + ": " + err.Error())
		return nil, fmt.Errorf("failed to generate manim code: %w", err)
	}

	jobDir := filepath.Join(s.workDir, fmt.Sprintf("job_%d_%s", time.Now().Unix(), userID))
	os.MkdirAll(jobDir, 0755)
	defer os.RemoveAll(jobDir)

	codeFile := filepath.Join(jobDir, "scene.py")
	if err := os.WriteFile(codeFile, []byte(manimCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write manim code: %w", err)
	}

	videoPath, err := s.renderVideo(jobDir, codeFile)
	if err != nil {
		telemetry.Error("Video generation failed at rendering for user " + userID + ": " + err.Error())
		return nil, fmt.Errorf("failed to render video: %w", err)
	}

	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		telemetry.Error("Video generation failed at file read for user " + userID + ": " + err.Error())
		return nil, fmt.Errorf("failed to read generated video: %w", err)
	}

	metadata, err := s.videoStorage.StoreVideo(ctx, userID, title, prompt, videoData, retentionHours)
	if err != nil {
		telemetry.Error("Video generation failed at storage for user " + userID + ": " + err.Error())
		return nil, fmt.Errorf("failed to store video: %w", err)
	}

	success = true
	telemetry.Info("Video generation completed successfully for user " + userID)
	return metadata, nil
}

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

	if !strings.Contains(code, "from manim import") && !strings.Contains(code, "import manim") {
		return "", fmt.Errorf("generated code does not appear to be valid Manim code")
	}

	if !strings.Contains(code, "class") || !strings.Contains(code, "Scene") {
		return "", fmt.Errorf("generated code does not contain a Scene class")
	}

	return code, nil
}

func (s *VideoGenerationService) renderVideo(jobDir, codeFile string) (string, error) {
	imageName := "obsidian-manim:latest"

	if _, err := exec.LookPath("docker"); err != nil {
		telemetry.Warn("Docker not found, falling back to direct execution")
		return s.renderVideoDirect(jobDir, codeFile)
	}

	if err := s.buildDockerImage(jobDir, imageName); err != nil {
		telemetry.Warn("Failed to build Docker image, falling back to direct execution: " + err.Error())
		return s.renderVideoDirect(jobDir, codeFile)
	}

	containerName := fmt.Sprintf("manim-render-%d", time.Now().Unix())

	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/workspace", jobDir),
		"-w", "/workspace",
		imageName,
		"python", "-m", "manim", "-pql", filepath.Base(codeFile), "Scene")

	output, err := cmd.CombinedOutput()
	if err != nil {
		telemetry.Error("Docker Manim rendering failed: " + string(output))
		return "", fmt.Errorf("docker manim rendering failed: %s", string(output))
	}

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

func (s *VideoGenerationService) renderVideoDirect(jobDir, codeFile string) (string, error) {
	cmd := exec.Command("python", "-m", "manim", "-pql", filepath.Base(codeFile), "Scene")
	cmd.Dir = jobDir

	cmd.Env = append(os.Environ(),
		"PYTHONPATH=/usr/local/lib/python3.9/site-packages",
		"MANIM_CONFIG_DIR="+jobDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		telemetry.Error("Direct Manim rendering failed: " + string(output))
		return "", fmt.Errorf("manim rendering failed: %s", string(output))
	}

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

func (s *VideoGenerationService) buildDockerImage(jobDir, imageName string) error {
	cmd := exec.Command("docker", "images", "-q", imageName)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return nil
	}

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

	telemetry.Info("Docker image built successfully: " + imageName)
	return nil
}
