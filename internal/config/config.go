package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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

func loadSecretsFromVault() error {
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
	viper.SetDefault("providers.gemini.model", "gemini-pro")
	viper.SetDefault("dashboard.port", 8080)

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
	ProviderProfiles map[string]ProviderConfig `mapstructure:"provider_profiles"`
	SwitchingRules   SwitchingRules            `mapstructure:"switching_rules"`
	WhatsApp         struct {
		AccessToken string `mapstructure:"access_token"`
		VerifyToken string `mapstructure:"verify_token"`
		AppSecret   string `mapstructure:"app_secret"`
	} `mapstructure:"whatsapp"`
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
	Vision struct {
		Enabled          bool     `mapstructure:"enabled"`
		EncoderModel     string   `mapstructure:"encoder_model"`
		FusionMethod     string   `mapstructure:"fusion_method"`
		MinConfidence    float64  `mapstructure:"min_confidence"`
		MaxImageSize     int      `mapstructure:"max_image_size"`
		SupportedFormats []string `mapstructure:"supported_formats"`
		QualityThreshold float64  `mapstructure:"quality_threshold"`
	} `mapstructure:"vision"`
	Git struct {
		UserName  string `mapstructure:"user_name"`
		UserEmail string `mapstructure:"user_email"`
		VaultPath string `mapstructure:"vault_path"`
		RemoteURL string `mapstructure:"remote_url"`
	} `mapstructure:"git"`
}
