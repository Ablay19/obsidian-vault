package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ... (struct definitions remain the same)

// LoadConfig loads the configuration from a file, environment variables, and Vault.
func LoadConfig() {
	// Load secrets from Vault (optional)
	if err := loadSecretsFromVault(); err != nil {
		slog.Warn("Could not load secrets from Vault. Falling back to environment variables: %v\n", err)
		// Continue with environment variables instead of exiting
	}

	// Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Warn("Unable to decode into struct: %v\n", err)
		os.Exit(1)
	}
	}

func loadSecretsFromFile() error {
	secretsFile := "./secrets.json"
	
	// Try to read secrets from file
	secretsData, err := os.ReadFile(secretsFile)
	if err != nil {
		slog.Warn("Could not read secrets file: %v, falling back to Vault\n", err)
		return loadSecretsFromVault()
	}
	
	var secrets map[string]interface{}
	if err := json.Unmarshal(secretsData, &secrets); err != nil {
		slog.Warn("Could not parse secrets file: %v, falling back to Vault\n", err)
		return loadSecretsFromVault()
	}
	
	// Map secrets to viper
	for key, value := range secrets {
		viper.Set(key, value.(string))
	}
	
	slog.Info("Successfully loaded secrets from file.")
	return nil
}

func LoadConfig() {
	// 1. Load secrets from file (preferred method)
	if err := loadSecretsFromFile(); err != nil {
		// Fallback to environment variables if file loading fails
		slog.Warn("Failed to load secrets from file: %v\n", err)
		// Continue with environment variables instead of exiting
	} else {
		slog.Info("Secrets loaded from secrets.json file.")
	}
	
	// 2. Unmarshal into AppConfig struct
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	// If Vault credentials not set, that's okay for development
	if vaultAddr == "" || vaultToken == "" {
		slog.Warn("Vault credentials not found, using environment variables only\n")
		return nil // Don't return error, just skip Vault loading
	}

	config := &api.Config{
		Address: vaultAddr,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	client.SetToken(vaultToken)

	// Read all secrets from the kv-v2 engine at secret/obsidian-automation
	secret, err := client.KVv2("secret").Get(context.Background(), "obsidian-automation")
	if err != nil {
		return fmt.Errorf("failed to read secrets from Vault: %w", err)
	}

	for key, value := range secret.Data {
		viper.Set(key, value)
		slog.Info("Loaded secret from Vault: %s\n", key)
	}

	slog.Info("Successfully loaded secrets from Vault.")
	return nil
}

// AppConfig is the loaded configuration.
var AppConfig Config

func init() {
	// 1. Load .env file using godotenv (actually sets OS environment variables)
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found or error loading it: %v\n", err)
	} else {
		slog.Info(".env file loaded into environment successfully")
	}

	// 2. Set Defaults
	viper.SetDefault("telegram_api_url", "https://api.telegram.org")
	viper.SetDefault("database_url", "sqlite:///app/data/bot.db")
	viper.SetDefault("database_path", "./data/bot.db")
	viper.SetDefault("redis_url", "redis://localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("ai_provider", "local")
	viper.SetDefault("local_model_path", "./models")
	viper.SetDefault("rate_limit_per_hour", 100)
	viper.SetDefault("rate_limit_per_day", 1000)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_file", "./logs/bot.log")
	viper.SetDefault("server_port", "8080")
	viper.SetDefault("server_host", "localhost")
	viper.SetDefault("max_concurrency", 10)
	viper.SetDefault("request_timeout", "30s")
	viper.SetDefault("context_max_length", 50)

	// 3. Setup environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 4. Load main config file (config.yaml)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("config.yaml not found, using defaults")
		} else {
			slog.Error("Error reading config.yaml: %v\n", err)
		}
	}

	// 6. Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Error("Unable to decode into struct: %v\n", err)
		os.Exit(1)
	}
}

type ProviderConfig struct {
	Model                string  `mapstructure:"model"`
	Latency              int     `mapstructure:"latency"`
	Accuracy             float64 `mapstructure:"accuracy"`
	MaxTokens            int     `mapstructure:"max_tokens"`
	RateLimit            int     `mapstructure:"rate_limit"`
	Concurrency          int     `mapstructure:"concurrency"`
	SupportsStreaming    bool    `mapstructure:"supports_streaming"`
	IsDefault            bool    `mapstructure:"is_default"`
	Enabled              bool    `mapstructure:"enabled"`
	Blocked              bool    `mapstructure:"blocked"`
	InputCostPerToken    float64 `mapstructure:"input_cost_per_token"`
	OutputCostPerToken   float64 `mapstructure:"output_cost_per_token"`
	MaxInputTokens       int     `mapstructure:"max_input_tokens"`
	MaxOutputTokens      int     `mapstructure:"max_output_tokens"`
	LatencyMsThreshold   int     `mapstructure:"latency_ms_threshold"`
	AccuracyPctThreshold float64 `mapstructure:"accuracy_pct_threshold"`
	SupportsVision       bool    `mapstructure:"supports_vision"`
}

// SwitchingRules defines the rules for switching between AI providers.
type SwitchingRules struct {
	DefaultProvider   string  `mapstructure:"default_provider"`
	LatencyTarget     int     `mapstructure:"latency_target"`
	ThroughputTarget  int     `mapstructure:"throughput_target"`
	AccuracyThreshold float64 `mapstructure:"accuracy_threshold"`
	RetryCount        int     `mapstructure:"retry_count"`
	RetryDelayMs      int     `mapstructure:"retry_delay_ms"`
	OnError           string  `mapstructure:"on_error"`
}

type AuthConfig struct {
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	GoogleRedirectURL  string `mapstructure:"google_redirect_url"`
	SessionSecret      string `mapstructure:"session_secret"`
}

type Config struct {
	// Telegram Bot Configuration
	TelegramBotToken string `mapstructure:"telegram_bot_token"`
	TelegramAPIURL   string `mapstructure:"telegram_api_url"`

	// Database Configuration
	DatabaseURL  string `mapstructure:"database_url"`
	DatabasePath string `mapstructure:"database_path"`

	// Redis Configuration
	RedisURL      string `mapstructure:"redis_url"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`

	// AI Configuration
	AIProvider       string                    `mapstructure:"ai_provider"`
	LocalModelPath   string                    `mapstructure:"local_model_path"`
	HuggingFaceAPI   string                    `mapstructure:"huggingface_api"`
	ReplicateAPI     string                    `mapstructure:"replicate_api"`
	OpenRouterAPI    string                    `mapstructure:"openrouter_api"`
	ProviderProfiles map[string]ProviderConfig `mapstructure:"provider_profiles"`
	SwitchingRules   SwitchingRules            `mapstructure:"switching_rules"`

	// Rate Limiting
	RateLimitPerHour int `mapstructure:"rate_limit_per_hour"`
	RateLimitPerDay  int `mapstructure:"rate_limit_per_day"`

	// Logging Configuration
	LogLevel string `mapstructure:"log_level"`
	LogFile  string `mapstructure:"log_file"`

	// Server Configuration
	ServerPort string `mapstructure:"server_port"`
	ServerHost string `mapstructure:"server_host"`

	// Performance Configuration
	MaxConcurrency   int           `mapstructure:"max_concurrency"`
	RequestTimeout   time.Duration `mapstructure:"request_timeout"`
	ContextMaxLength int           `mapstructure:"context_max_length"`

	// Auth Configuration
	Auth AuthConfig `mapstructure:"auth"`
}
