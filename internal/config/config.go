package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
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
	Git struct {
		UserName  string `mapstructure:"user_name"`
		UserEmail string `mapstructure:"user_email"`
		VaultPath string `mapstructure:"vault_path"`
		RemoteURL string `mapstructure:"remote_url"`
	} `mapstructure:"git"`
}

// AppConfig is the loaded configuration.
var AppConfig Config

// LoadConfig loads the configuration from a file and environment variables.
func LoadConfig() {
	// 1. Load .env file using godotenv (actually sets OS environment variables)
	if err := godotenv.Load(); err != nil {
		slog.Debug("No .env file found or error loading it", "error", err)
	} else {
		slog.Info(".env file loaded into environment successfully")
	}

	// 2. Set Defaults
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
	viper.SetDefault("dashboard.port", 8080)
	viper.SetDefault("auth.session_secret", "change-me-to-something-very-secure")
	viper.SetDefault("git.user_name", "Obsidian Bot")
	viper.SetDefault("git.user_email", "bot@obsidian.internal")
	viper.SetDefault("git.vault_path", "vault")

	// 3. Setup environment variable support (Viper will now see the variables set by godotenv)
	viper.AutomaticEnv()
	viper.BindEnv("TURSO_DATABASE_URL")
	viper.BindEnv("TURSO_AUTH_TOKEN")
	viper.BindEnv("GEMINI_API_KEYS")
	viper.BindEnv("GEMINI_API_KEY")
	viper.BindEnv("GROQ_API_KEY")
	viper.BindEnv("HUGGINGFACE_API_KEY")
	viper.BindEnv("HF_TOKEN")
	viper.BindEnv("OPENROUTER_API_KEY")
	viper.BindEnv("TELEGRAM_BOT_TOKEN")
	viper.BindEnv("ENVIRONMENT_MODE")
	viper.BindEnv("BACKEND_HOST")
	viper.BindEnv("ENVIRONMENT_ISOLATION_ENABLED")
	viper.BindEnv("AI_ENABLED")
	viper.BindEnv("auth.google_client_id", "GOOGLE_CLIENT_ID")
	viper.BindEnv("auth.google_client_secret", "GOOGLE_CLIENT_SECRET")
	viper.BindEnv("auth.google_redirect_url", "GOOGLE_REDIRECT_URL")
	viper.BindEnv("git.user_name", "GIT_USER_NAME")
	viper.BindEnv("git.user_email", "GIT_USER_EMAIL")
	viper.BindEnv("git.vault_path", "GIT_VAULT_PATH")
	viper.BindEnv("git.remote_url", "GIT_REMOTE_URL")

	// 4. Load main config file (config.yaml)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("config.yaml not found, using defaults")
		} else {
			slog.Error("Error reading config.yaml", "error", err)
		}
	}

	// 5. Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Error("Unable to decode into struct", "error", err)
		os.Exit(1)
	}
}