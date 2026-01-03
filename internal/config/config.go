package config

import (
	"log/slog"
	"os"

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

		HuggingFace struct {
			APIKey string `mapstructure:"api_key"`
			Model  string `mapstructure:"model"`
		} `mapstructure:"huggingface"`

		OpenRouter struct {
			APIKey string `mapstructure:"api_key"`
			Model  string `mapstructure:"model"`
		} `mapstructure:"openrouter"`
	} `mapstructure:"providers"`
	Classification struct {
		Patterns map[string][]string `mapstructure:"patterns"`
	} `mapstructure:"classification"`
	LanguageDetection struct {
		FrenchWords []string `mapstructure:"french_words"`
	} `mapstructure:"language_detection"`
	Dashboard struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"dashboard"`
	Auth struct {
		GoogleClientID     string `mapstructure:"google_client_id"`
		GoogleClientSecret string `mapstructure:"google_client_secret"`
		GoogleRedirectURL  string `mapstructure:"google_redirect_url"`
		SessionSecret      string `mapstructure:"session_secret"`
	} `mapstructure:"auth"`
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
	viper.SetDefault("providers.openrouter.model", "openai/gpt-3.5-turbo")
	viper.SetDefault("classification.patterns", map[string][]string{
		"physics":   {"force", "energy", "mass", "velocity", "acceleration"},
		"math":      {"equation", "function", "derivative", "integral", "matrix"},
		"chemistry": {"molecule", "atom", "reaction", "chemical"},
		"admin":     {"invoice", "contract", "form", "certificate"},
	})
	viper.SetDefault("language_detection.french_words", []string{"le", "la", "de", "et", "un"})
	viper.SetDefault("dashboard.port", 8080) // New default for dashboard port
	viper.SetDefault("auth.session_secret", "change-me-to-something-very-secure")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			slog.Info("Config file not found, using defaults and environment variables.")
		} else {
			// Config file was found but another error was produced
			slog.Error("Fatal error config file", "error", err)
			os.Exit(1)
		}
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		slog.Error("Unable to decode into struct", "error", err)
		os.Exit(1)
	}
}
