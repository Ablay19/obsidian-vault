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
func LoadConfig() error {
	// 1. Load secrets from file (preferred method)
	if err := loadSecretsFromFile(); err != nil {
		// Fallback to environment variables if file loading fails
		slog.Warn("Failed to load secrets from file", "error", err)
		// Continue with environment variables instead of exiting
	} else {
		slog.Info("Secrets loaded from secrets.json file.")
	}

	// 2. Try to load from Vault if credentials are available
	if err := loadSecretsFromVault(); err != nil {
		slog.Warn("Could not load secrets from Vault. Using environment variables only", "error", err)
		// Continue with environment variables instead of exiting
	}

	// 3. Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Error("Unable to decode into struct", "error", err)
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func loadSecretsFromFile() error {
	secretsFile := "./secrets.json"

	// Try to read secrets from file
	secretsData, err := os.ReadFile(secretsFile)
	if err != nil {
		return fmt.Errorf("could not read secrets file: %w", err)
	}

	var secrets map[string]interface{}
	if err := json.Unmarshal(secretsData, &secrets); err != nil {
		return fmt.Errorf("could not parse secrets file: %w", err)
	}

	// Map secrets to viper
	for key, value := range secrets {
		if strValue, ok := value.(string); ok {
			viper.Set(key, strValue)
		}
	}

	slog.Info("Successfully loaded secrets from file.")
	return nil
}

func loadSecretsFromVault() error {
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	// If Vault credentials not set, skip Vault loading
	if vaultAddr == "" || vaultToken == "" {
		return fmt.Errorf("vault credentials not found")
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
		slog.Info("Loaded secret from Vault", "key", key)
	}

	slog.Info("Successfully loaded secrets from Vault", "count", len(secret.Data))
	return nil
}

// AppConfig is the loaded configuration.
var AppConfig Config

func init() {
	// 1. Load .env file using godotenv (actually sets OS environment variables)
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found or error loading it", "error", err)
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
			slog.Error("Error reading config.yaml", "error", err)
		}
	}

	// 6. Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		slog.Error("Unable to decode into struct", "error", err)
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

type WhatsAppConfig struct {
	AccessToken string `mapstructure:"access_token"`
	AppSecret   string `mapstructure:"app_secret"`
	VerifyToken string `mapstructure:"verify_token"`
}

type VisionConfig struct {
	Enabled       bool    `mapstructure:"enabled"`
	EncoderModel  string  `mapstructure:"encoder_model"`
	FusionMethod  string  `mapstructure:"fusion_method"`
	MinConfidence float64 `mapstructure:"min_confidence"`
}

type GitConfig struct {
	VaultPath  string `mapstructure:"vault_path"`
	UserName   string `mapstructure:"user_name"`
	UserEmail  string `mapstructure:"user_email"`
	RemoteURL  string `mapstructure:"remote_url"`
	AutoCommit bool   `mapstructure:"auto_commit"`
	AutoPush   bool   `mapstructure:"auto_push"`
}

type ClassificationConfig struct {
	Enabled  bool                `mapstructure:"enabled"`
	Patterns map[string][]string `mapstructure:"patterns"`
}

type LanguageDetectionConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	FrenchWords  []string `mapstructure:"french_words"`
	SpanishWords []string `mapstructure:"spanish_words"`
}

type DashboardConfig struct {
	Port       string `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	EnableAuth bool   `mapstructure:"enable_auth"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

type IsDevelopmentConfig struct {
	Enabled bool `mapstructure:"enabled"`
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

	// WhatsApp Configuration
	WhatsApp WhatsAppConfig `mapstructure:"whatsapp"`

	// Vision Configuration
	Vision VisionConfig `mapstructure:"vision"`

	// Git Configuration
	Git GitConfig `mapstructure:"git"`

	// Classification Configuration
	Classification ClassificationConfig `mapstructure:"classification"`

	// Language Detection Configuration
	LanguageDetection LanguageDetectionConfig `mapstructure:"language_detection"`

	// Dashboard Configuration
	Dashboard DashboardConfig `mapstructure:"dashboard"`

	// IsDevelopment Configuration
	IsDevelopment IsDevelopmentConfig `mapstructure:"is_development"`
}
