package cmd

import (
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mauritania-cli",
		Short: "CLI for remote development through Mauritanian network services",
		Long: `A command-line interface that enables remote development and project management
through Mauritanian network provider services, including social media APIs
and SM APOS Shipper for regions with limited direct internet access.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mauritania-cli.yaml)")
	cmd.PersistentFlags().Bool("json", false, "output logs in JSON format")

	// Add subcommands
	cmd.AddCommand(newSendCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newResultCmd())
	cmd.AddCommand(newQueueCmd())
	cmd.AddCommand(newRoutesCmd())
	cmd.AddCommand(newShipperCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newServerCmd())
	cmd.AddCommand(newLogsCmd())
	cmd.AddCommand(newSecurityCmd())
	cmd.AddCommand(newTUICmd())
	cmd.AddCommand(newWhatsAppCmd())
	cmd.AddCommand(newTUICmd())

	// Add Carapace completion
	carapace.Gen(cmd)

	return cmd
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
