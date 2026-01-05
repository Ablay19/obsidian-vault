package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// ... (struct definitions remain the same)

// LoadConfig loads the configuration from a file, environment variables, and Vault.
func LoadConfig() {
	// ... (existing godotenv and viper setup remains the same)

	// Load secrets from Vault
	if err := loadSecretsFromVault(); err != nil {
		zap.S().Warnw("Could not load secrets from Vault. Falling back to environment variables.", "error", err)
	}

	// ... (rest of the viper setup and unmarshalling remains the same)
}

func loadSecretsFromVault() error {
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")

	if vaultAddr == "" || vaultToken == "" {
		return fmt.Errorf("VAULT_ADDR and VAULT_TOKEN must be set")
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
		zap.S().Debugf("Loaded secret from Vault: %s", key)
	}

	zap.S().Info("Successfully loaded secrets from Vault.")
	return nil
}

// AppConfig is the loaded configuration.
var AppConfig Config

func init() {
	// 1. Load .env file using godotenv (actually sets OS environment variables)
	if err := godotenv.Load(); err != nil {
		zap.S().Debug("No .env file found or error loading it", "error", err)
	} else {
		zap.S().Info(".env file loaded into environment successfully")
	}

	// 2. Set Defaults
	viper.SetDefault("providers.gemini.model", "gemini-pro")
	// ... (rest of the defaults remain the same)

	// 3. Setup environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 4. Load main config file (config.yaml)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			zap.S().Info("config.yaml not found, using defaults")
		} else {
			zap.S().Error("Error reading config.yaml", "error", err)
		}
	}

	// 5. Load secrets from Vault
	if err := loadSecretsFromVault(); err != nil {
		zap.S().Warnw("Could not load secrets from Vault. Falling back to environment variables.", "error", err)
	}

	// 6. Unmarshal into AppConfig struct
	if err := viper.Unmarshal(&AppConfig); err != nil {
		zap.S().Fatalw("Unable to decode into struct", "error", err)
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
	Git struct {
		UserName  string `mapstructure:"user_name"`
		UserEmail string `mapstructure:"user_email"`
		VaultPath string `mapstructure:"vault_path"`
		RemoteURL string `mapstructure:"remote_url"`
	} `mapstructure:"git"`
}
