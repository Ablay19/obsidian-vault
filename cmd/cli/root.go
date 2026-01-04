package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/carapace-sh/carapace"
	"obsidian-automation/cmd/cli/config"
	"obsidian-automation/cmd/cli/database"
	"obsidian-automation/cmd/cli/logger"
	"obsidian-automation/cmd/cli/tui"
	"obsidian-automation/cmd/cli/email"
	"obsidian-automation/internal/ssh"
)

var rootCmd = &cobra.Command{
	Use:   "obsidian-cli",
	Short: "A CLI for interacting with the Obsidian Automation Bot",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
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
			panic(err)
		}
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion script",
	Run: func(cmd *cobra.Command, args []string) {
		// Just generate the script to stdout
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Start the SSH server",
	Run: func(cmd *cobra.Command, args []string) {
		ssh.StartServer()
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
			fmt.Printf("Error generating key pair for %s: %v\n", username, err)
			return
		}
		fmt.Printf("Private key for user %s:\n%s\n", username, string(privateKey))
	},
}

func init() {
	carapace.Gen(rootCmd)
	config.Init()
	logger.Init()
	database.Init()
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(emailCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(addUserCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
