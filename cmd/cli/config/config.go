package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("cli")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			fmt.Println("Config file not found. Using default values.")
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}
}
