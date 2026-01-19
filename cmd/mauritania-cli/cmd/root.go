package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var rootCmd *cobra.Command

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:   "mauritania-cli",
			Short: "CLI for remote development through Mauritanian network services",
			Long: `A command-line interface that enables remote development and project management
through Mauritanian network provider services, including social media APIs
and SM APOS Shipper for regions with limited direct internet access.`,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return initConfig()
			},
		}

		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mauritania-cli.yaml)")
		rootCmd.PersistentFlags().Bool("json", false, "output logs in JSON format")

		// Add subcommands
		rootCmd.AddCommand(newSendCmd())
		rootCmd.AddCommand(newStatusCmd())
		rootCmd.AddCommand(newResultCmd())
		rootCmd.AddCommand(newQueueCmd())
		rootCmd.AddCommand(newRoutesCmd())
		rootCmd.AddCommand(newShipperCmd())
		rootCmd.AddCommand(newConfigCmd())
		rootCmd.AddCommand(newServerCmd())
		rootCmd.AddCommand(newLogsCmd())
		rootCmd.AddCommand(newSecurityCmd())
		rootCmd.AddCommand(newTUICmd())
		rootCmd.AddCommand(newWhatsAppCmd())
		rootCmd.AddCommand(newTestCmd())
		rootCmd.AddCommand(newCITestCmd())
		rootCmd.AddCommand(newMCPServerCmd())

		// Add new commands
		rootCmd.AddCommand(cleanupCmd)
		rootCmd.AddCommand(docsCmd)
		rootCmd.AddCommand(validateCmd)
		rootCmd.AddCommand(aiCmd)

		// Add Carapace completion
		carapace.Gen(rootCmd)
	}

	return rootCmd
}

// initConfig reads in config file and ENV variables if set
func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".mauritania-cli")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	return nil
}
