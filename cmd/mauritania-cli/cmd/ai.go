package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"obsidian-automation/cmd/mauritania-cli/internal/ai"
	"obsidian-automation/cmd/mauritania-cli/internal/utils"
)

var aiService *ai.AIManimService
var sessionManager *ai.AISessionManager

// initAIService initializes the AI service
func initAIService() {
	if aiService == nil {
		config := &utils.Config{} // Would load from actual config
		logger := utils.NewLogger("ai-cli")
		aiService = ai.NewAIManimService(config, logger)
		sessionManager = ai.NewAISessionManager(logger)
	}
}

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered educational content generation",
	Long:  `Generate mathematical animations and educational videos using AI`,
}

var mathCmd = &cobra.Command{
	Use:   "math [problem]",
	Short: "Generate mathematical animation video",
	Long: `Create an animated video explanation for a mathematical problem or concept.
Examples:
  mauritania-cli ai math "Explain the Pythagorean theorem"
  mauritania-cli ai math "Show me how derivatives work"`,
	Args: cobra.MinimumNArgs(1),
	Run:  runMathCommand,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check AI job status",
	Long:  `Show the status of AI video generation jobs`,
	Run:   runStatusCommand,
}

func init() {
	initAIService()

	// Add AI subcommands
	aiCmd.AddCommand(mathCmd)
	aiCmd.AddCommand(statusCmd)

	// Math command flags
	mathCmd.Flags().String("quality", "medium", "Video quality (low, medium, high)")
	mathCmd.Flags().String("format", "mp4", "Video format (mp4, webm)")
	mathCmd.Flags().Int("max-duration", 300, "Maximum video duration in seconds")
}

func runMathCommand(cmd *cobra.Command, args []string) {
	initAIService()

	problem := strings.Join(args, " ")

	// Validate problem
	if len(problem) < 10 {
		fmt.Println("Error: Problem description too short (minimum 10 characters)")
		return
	}

	if len(problem) > 2000 {
		fmt.Println("Error: Problem description too long (maximum 2000 characters)")
		return
	}

	// Get options
	quality, _ := cmd.Flags().GetString("quality")
	format, _ := cmd.Flags().GetString("format")
	maxDuration, _ := cmd.Flags().GetInt("max-duration")

	// Validate quality
	if quality != "low" && quality != "medium" && quality != "high" {
		fmt.Println("Error: Quality must be low, medium, or high")
		return
	}

	// Validate format
	if format != "mp4" && format != "webm" {
		fmt.Println("Error: Format must be mp4 or webm")
		return
	}

	fmt.Printf("🎬 Generating mathematical animation for: %s\n", problem)
	fmt.Printf("Quality: %s, Format: %s, Max Duration: %d seconds\n\n", quality, format, maxDuration)

	// Create video options
	options := ai.VideoOptions{
		Quality:            quality,
		Format:             format,
		MaxDurationSeconds: maxDuration,
	}

	// Generate video
	job, err := aiService.GenerateVideo(cmd.Context(), problem, options)
	if err != nil {
		fmt.Printf("❌ Failed to start video generation: %v\n", err)
		return
	}

	fmt.Printf("✅ Video generation started!\n")
	fmt.Printf("Job ID: %s\n", job.JobID)
	fmt.Printf("Status: %s\n", job.Status)
	fmt.Printf("Created: %s\n\n", job.CreatedAt.Format(time.RFC3339))

	// Wait for completion with progress updates
	fmt.Println("Waiting for completion...")
	status, err := aiService.WaitForCompletion(cmd.Context(), job.JobID, 10*time.Minute)
	if err != nil {
		fmt.Printf("❌ Video generation failed: %v\n", err)
		return
	}

	if status.Status == "completed" {
		fmt.Printf("✅ Video generation completed!\n")
		// Try to get video URL from the job
		if job.VideoURL != "" {
			fmt.Printf("📹 Video URL: %s\n", job.VideoURL)
		} else {
			fmt.Printf("📹 Video URL: [Check job status for download link]\n")
		}
		fmt.Printf("📊 Final Status: %s\n", status.Status)
		fmt.Printf("⏱️  Processing Time: %s\n", time.Since(job.CreatedAt))
	} else {
		fmt.Printf("❌ Video generation ended with status: %s\n", status.Status)
		if status.Error != "" {
			fmt.Printf("Error: %s\n", status.Error)
		}
	}
}

func runStatusCommand(cmd *cobra.Command, args []string) {
	initAIService()

	fmt.Println("🤖 AI Job Status")
	fmt.Println("================")

	// Get session stats
	stats := sessionManager.GetStats()
	fmt.Printf("Total Jobs: %d\n", stats["total_jobs"])
	fmt.Printf("Pending: %d\n", stats["pending"])
	fmt.Printf("Processing: %d\n", stats["processing"])
	fmt.Printf("Completed: %d\n", stats["completed"])
	fmt.Printf("Failed: %d\n\n", stats["failed"])

	// Show recent jobs (placeholder)
	fmt.Println("Recent Jobs:")
	fmt.Println("(No jobs found - AI service integration pending)")

	// Check AI service health
	err := aiService.IsHealthy(cmd.Context())
	if err != nil {
		fmt.Printf("⚠️  AI Service Status: Unhealthy (%v)\n", err)
	} else {
		fmt.Println("✅ AI Service Status: Healthy")
	}
}
