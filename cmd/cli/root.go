package main 

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/carapace-sh/carapace"
	"obsidian-automation/cmd/cli/config"
	"obsidian-automation/cmd/cli/database"
	"obsidian-automation/cmd/cli/tui"
	"obsidian-automation/cmd/cli/email"
	"obsidian-automation/internal/ssh"
	"obsidian-automation/internal/telemetry" // Add telemetry import
)

var rootCmd = &cobra.Command{
	Use:   "obsidian-cli",
	Short: "A CLI for interacting with the Obsidian Automation Bot",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("### rootCmd.PersistentPreRunE called ###") // Debug print
		// Initialize SSH DB for any command that might need it, or move to specific commands
		// if only a subset requires it. For now, assume global.
		ssh.InitDB() 
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Do something
	},
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Run the TUI",
	Run: func(cmd *cobra.Command, args []string) {
		tui.Run()
	},
}

var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "Send a test email",
	Run: func(cmd *cobra.Command, args []string) {
		to := []string{"test@example.com"}
		subject := "Test email"
		body := "This is a test email from the Obsidian CLI."
		err := email.Send(to, subject, body)
		if err != nil {
			telemetry.ZapLogger.Sugar().Errorw("Error sending test email", "error", err)
			return
		}
		telemetry.ZapLogger.Sugar().Infow("Test email sent successfully", "to", to, "subject", subject)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion script",
	Run: func(cmd *cobra.Command, args []string) {
		// Just generate the script to stdout
	},
}

var addUserCmd = &cobra.Command{
	Use:   "add-user [username]",
	Short: "Add a new SSH user and generate a key pair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]
		privateKey, err := ssh.GenerateKeyPair(username)
		if err != nil {
			telemetry.ZapLogger.Sugar().Errorw("Error generating key pair for user", "username", username, "error", err)
			return
		}
		telemetry.ZapLogger.Sugar().Infow("Private key generated for user", "username", username, "private_key", string(privateKey))
	},
}

func init() {
	// Initialize telemetry logger as early as possible
	if _, err := telemetry.Init("obsidian-cli"); err != nil {
		fmt.Printf("Failed to initialize telemetry: %v\n", err) // Use fmt.Printf here as logger might not be fully ready
		os.Exit(1)
	}

	carapace.Gen(rootCmd)
	config.Init()
	// logger.Init() // Removed old logger init
	database.Init()
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(emailCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(addUserCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		telemetry.ZapLogger.Sugar().Fatalw("CLI execution failed", "error", err)
	}
}