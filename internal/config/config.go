package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config holds the configuration for the application.
type Config struct {
	Providers struct {
		Gemini struct {
			Model string `mapstructure:"model"`
		} `mapstructure:"gemini"`
		Groq struct {
			Model string `mapstructure:"model"`
		} `mapstructure:"groq"`
	} `mapstructure:"providers"`
	Classification struct {
		Patterns map[string][]string `mapstructure:"patterns"`
	} `mapstructure:"classification"`
	LanguageDetection struct {
		FrenchWords []string `mapstructure:"french_words"`
	} `mapstructure:"language_detection"`
}

// AppConfig is the loaded configuration.
var AppConfig Config

// LoadConfig loads the configuration from a file and environment variables.
func LoadConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // or yml, json, toml
	viper.AddConfigPath(".")      // look for config in the working directory
	viper.AutomaticEnv()          // read in environment variables that match

	// Set defaults
	viper.SetDefault("providers.gemini.model", "gemini-pro")
	viper.SetDefault("providers.groq.model", "llama3-70b")
	viper.SetDefault("classification.patterns", map[string][]string{
		"physics":   {"force", "energy", "mass", "velocity", "acceleration"},
		"math":      {"equation", "function", "derivative", "integral", "matrix"},
		"chemistry": {"molecule", "atom", "reaction", "chemical"},
		"admin":     {"invoice", "contract", "form", "certificate"},
	})
	viper.SetDefault("language_detection.french_words", []string{"le", "la", "de", "et", "un"})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("Config file not found, using defaults and environment variables.")
		} else {
			// Config file was found but another error was produced
			log.Fatalf("Fatal error config file: %s \n", err)
		}
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}
