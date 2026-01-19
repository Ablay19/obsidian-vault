package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// AIConfig holds AI service configuration
type AIConfig struct {
	ManimWorkerURL string `mapstructure:"manim_worker_url"` // Cloudflare Worker URL
	APIKey         string `mapstructure:"api_key"`          // API key for authentication
	Timeout        int    `mapstructure:"timeout"`          // request timeout in seconds
	RetryAttempts  int    `mapstructure:"retry_attempts"`   // number of retries
	RetryDelay     int    `mapstructure:"retry_delay"`      // delay between retries in seconds
}

// Config represents the application configuration
type Config struct {
	// Database settings
	Database DatabaseConfig `mapstructure:"database"`

	// Network transport configurations
	Transports TransportConfig `mapstructure:"transports"`

	// Authentication settings
	Auth AuthConfig `mapstructure:"auth"`

	// AI service configuration
	AI AIConfig `mapstructure:"ai"`

	// Network settings
	Network NetworkConfig `mapstructure:"network"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`

	// CLI settings
	CLI CLIConfig `mapstructure:"cli"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`     // sqlite, postgres, mysql
	Path     string `mapstructure:"path"`     // for sqlite
	Host     string `mapstructure:"host"`     // for postgres/mysql
	Port     int    `mapstructure:"port"`     // for postgres/mysql
	Database string `mapstructure:"database"` // for postgres/mysql
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"` // encrypted in config file
	SSLMode  string `mapstructure:"ssl_mode"` // for postgres
}

// TransportConfig holds transport-specific configurations
type TransportConfig struct {
	SocialMedia SocialMediaConfig `mapstructure:"social_media"`
	Shipper     ShipperConfig     `mapstructure:"shipper"`
	NRT         NRTConfig         `mapstructure:"nrt"`
	Default     string            `mapstructure:"default"` // default transport
}

// SocialMediaConfig holds social media API configurations
type SocialMediaConfig struct {
	WhatsApp WhatsAppConfig `mapstructure:"whatsapp"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Facebook FacebookConfig `mapstructure:"facebook"`
}

// WhatsAppConfig holds WhatsApp configuration for WhatsMeow
type WhatsAppConfig struct {
	DatabasePath  string `mapstructure:"database_path"`  // SQLite database path for session storage
	WebhookSecret string `mapstructure:"webhook_secret"` // webhook signature verification secret
	RateLimit     int    `mapstructure:"rate_limit"`     // messages per hour
	AutoConnect   bool   `mapstructure:"auto_connect"`   // automatically connect on startup
}

// TelegramConfig holds Telegram Bot API configuration
type TelegramConfig struct {
	BotToken   string `mapstructure:"bot_token"`
	WebhookURL string `mapstructure:"webhook_url"`
	ChatID     string `mapstructure:"chat_id"` // allowed chat ID
	RateLimit  int    `mapstructure:"rate_limit"`
}

// FacebookConfig holds Facebook Messenger API configuration
type FacebookConfig struct {
	AppID       string `mapstructure:"app_id"`
	AppSecret   string `mapstructure:"app_secret"`
	AccessToken string `mapstructure:"access_token"`
	VerifyToken string `mapstructure:"verify_token"`
	PageID      string `mapstructure:"page_id"`
	RateLimit   int    `mapstructure:"rate_limit"`
}

// ShipperConfig holds SM APOS Shipper configuration
type ShipperConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	RateLimit int    `mapstructure:"rate_limit"`
	Timeout   int    `mapstructure:"timeout"` // seconds
}

// NRTConfig holds NRT routing configuration
type NRTConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
	APIKey    string   `mapstructure:"api_key"`
	RateLimit int      `mapstructure:"rate_limit"`
	Timeout   int      `mapstructure:"timeout"`
}

// AuthConfig holds authentication and security settings
type AuthConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	JWTSecret       string   `mapstructure:"jwt_secret"`
	TokenExpiry     int      `mapstructure:"token_expiry"` // hours
	AllowedUsers    []string `mapstructure:"allowed_users"`
	AllowedCommands []string `mapstructure:"allowed_commands"`
	RequireApproval bool     `mapstructure:"require_approval"`
}

// NetworkConfig holds network and connectivity settings
type NetworkConfig struct {
	Timeout         int      `mapstructure:"timeout"` // seconds
	RetryAttempts   int      `mapstructure:"retry_attempts"`
	RetryDelay      int      `mapstructure:"retry_delay"` // seconds
	OfflineMode     bool     `mapstructure:"offline_mode"`
	PreferredRoutes []string `mapstructure:"preferred_routes"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`    // DEBUG, INFO, WARN, ERROR
	File       string `mapstructure:"file"`     // log file path
	MaxSize    int    `mapstructure:"max_size"` // MB
	MaxAge     int    `mapstructure:"max_age"`  // days
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// CLIConfig holds CLI-specific settings
type CLIConfig struct {
	PromptSymbol    string `mapstructure:"prompt_symbol"`
	HistorySize     int    `mapstructure:"history_size"`
	AutoComplete    bool   `mapstructure:"auto_complete"`
	MobileOptimized bool   `mapstructure:"mobile_optimized"`
}

// ConfigManager manages application configuration
type ConfigManager struct {
	config *Config
	v      *viper.Viper
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	v := viper.New()
	cm := &ConfigManager{
		v: v,
	}

	// Set defaults
	cm.setDefaults()

	return cm
}

// setDefaults sets default configuration values
func (cm *ConfigManager) setDefaults() {
	// Database defaults
	cm.v.SetDefault("database.type", "sqlite")
	cm.v.SetDefault("database.path", "./data/mauritania-cli.db")

	// Transport defaults
	cm.v.SetDefault("transports.default", "social_media")
	cm.v.SetDefault("transports.social_media.whatsapp.rate_limit", 100)
	cm.v.SetDefault("transports.social_media.telegram.rate_limit", 30)
	cm.v.SetDefault("transports.social_media.facebook.rate_limit", 200)
	cm.v.SetDefault("transports.shipper.rate_limit", 1000)
	cm.v.SetDefault("transports.shipper.timeout", 30)
	cm.v.SetDefault("transports.nrt.rate_limit", 500)
	cm.v.SetDefault("transports.nrt.timeout", 10)

	// Auth defaults
	cm.v.SetDefault("auth.enabled", false)
	cm.v.SetDefault("auth.token_expiry", 24)
	cm.v.SetDefault("auth.require_approval", false)

	// Network defaults
	cm.v.SetDefault("network.timeout", 30)
	cm.v.SetDefault("network.retry_attempts", 3)
	cm.v.SetDefault("network.retry_delay", 5)
	cm.v.SetDefault("network.offline_mode", false)

	// Logging defaults
	cm.v.SetDefault("logging.level", "INFO")
	cm.v.SetDefault("logging.max_size", 10)
	cm.v.SetDefault("logging.max_age", 30)
	cm.v.SetDefault("logging.max_backups", 5)
	cm.v.SetDefault("logging.compress", true)

	// CLI defaults
	cm.v.SetDefault("cli.prompt_symbol", ">")
	cm.v.SetDefault("cli.history_size", 1000)
	cm.v.SetDefault("cli.auto_complete", true)
	cm.v.SetDefault("cli.mobile_optimized", true)
}

// Load loads configuration from files and environment
func (cm *ConfigManager) Load() error {
	// Set the default config file path first
	defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".mauritania-cli.toml")
	cm.v.SetConfigFile(defaultConfigPath)

	// Set config file search paths
	cm.v.SetConfigName("mauritania-cli")
	cm.v.SetConfigType("toml")

	// Add config paths in order of priority
	configPaths := []string{
		".",                   // current directory
		"./config",            // config subdirectory
		os.Getenv("HOME"),     // user home
		"/etc/mauritania-cli", // system config
	}

	for _, path := range configPaths {
		cm.v.AddConfigPath(path)
	}

	// Read config file (optional - never fails)
	if err := cm.v.ReadInConfig(); err != nil {
		fmt.Printf("Warning: Failed to read config file: %v\n", err)
	}

	// Unmarshal into config struct (also ignore errors)
	config := &Config{}
	if err := cm.v.Unmarshal(config); err != nil {
		fmt.Printf("Warning: Failed to unmarshal config: %v\n", err)
		// Use defaults if unmarshal fails
		cm.setDefaults()
		if err := cm.v.Unmarshal(config); err != nil {
			fmt.Printf("Warning: Failed to unmarshal defaults: %v\n", err)
			// If still failing, create empty config with defaults
			config = &Config{}
		}
	}
	fmt.Printf("Debug: After unmarshal, config.Auth.Enabled = %t\n", config.Auth.Enabled)

	cm.config = config

	// Validate configuration
	if err := cm.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// Validate validates the configuration and returns detailed error messages
func (cm *ConfigManager) Validate() error {
	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	var errors []string

	// Validate database configuration
	if err := cm.validateDatabaseConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate transport configuration
	if err := cm.validateTransportConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate authentication configuration
	if err := cm.validateAuthConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate network configuration
	if err := cm.validateNetworkConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate logging configuration
	if err := cm.validateLoggingConfig(); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateDatabaseConfig validates database configuration
func (cm *ConfigManager) validateDatabaseConfig() error {
	var errors []string

	db := cm.config.Database

	// Validate database type
	validTypes := []string{"sqlite", "postgres", "mysql"}
	if !contains(validTypes, db.Type) {
		errors = append(errors, fmt.Sprintf("  - database.type must be one of: %s, got: %s", strings.Join(validTypes, ", "), db.Type))
	}

	// Validate SQLite path
	if db.Type == "sqlite" {
		if db.Path == "" {
			errors = append(errors, "  - database.path is required for SQLite")
		}
		// Check if directory exists
		dir := filepath.Dir(db.Path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("  - database directory does not exist: %s", dir))
		}
	}

	// Validate PostgreSQL/MySQL settings
	if db.Type == "postgres" || db.Type == "mysql" {
		if db.Host == "" {
			errors = append(errors, "  - database.host is required for PostgreSQL/MySQL")
		}
		if db.Port <= 0 || db.Port > 65535 {
			errors = append(errors, "  - database.port must be between 1-65535")
		}
		if db.Database == "" {
			errors = append(errors, "  - database.database is required for PostgreSQL/MySQL")
		}
		if db.Username == "" {
			errors = append(errors, "  - database.username is required for PostgreSQL/MySQL")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("database configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateTransportConfig validates transport configuration
func (cm *ConfigManager) validateTransportConfig() error {
	var errors []string

	transports := cm.config.Transports

	// Validate default transport
	validDefaults := []string{"social_media", "sm_apos", "nrt"}
	if !contains(validDefaults, transports.Default) {
		errors = append(errors, fmt.Sprintf("  - transports.default must be one of: %s, got: %s", strings.Join(validDefaults, ", "), transports.Default))
	}

	// Validate WhatsApp config
	whatsapp := transports.SocialMedia.WhatsApp
	if whatsapp.RateLimit < 0 {
		errors = append(errors, "  - transports.social_media.whatsapp.rate_limit must be >= 0")
	}
	if whatsapp.WebhookSecret == "" {
		errors = append(errors, "  - transports.social_media.whatsapp.webhook_secret is required for webhook verification")
	}

	// Validate Telegram config
	telegram := transports.SocialMedia.Telegram
	if telegram.RateLimit < 0 {
		errors = append(errors, "  - transports.social_media.telegram.rate_limit must be >= 0")
	}
	if telegram.BotToken == "" {
		errors = append(errors, "  - transports.social_media.telegram.bot_token is required")
	}

	// Validate Facebook config
	facebook := transports.SocialMedia.Facebook
	if facebook.RateLimit < 0 {
		errors = append(errors, "  - transports.social_media.facebook.rate_limit must be >= 0")
	}
	if facebook.AppID == "" {
		errors = append(errors, "  - transports.social_media.facebook.app_id is required")
	}
	if facebook.AppSecret == "" {
		errors = append(errors, "  - transports.social_media.facebook.app_secret is required")
	}
	if facebook.AccessToken == "" {
		errors = append(errors, "  - transports.social_media.facebook.access_token is required")
	}

	// Validate Shipper config
	shipper := transports.Shipper
	if shipper.RateLimit < 0 {
		errors = append(errors, "  - transports.shipper.rate_limit must be >= 0")
	}
	if shipper.Timeout <= 0 {
		errors = append(errors, "  - transports.shipper.timeout must be > 0")
	}

	// Validate NRT config
	nrt := transports.NRT
	if nrt.RateLimit < 0 {
		errors = append(errors, "  - transports.nrt.rate_limit must be >= 0")
	}
	if nrt.Timeout <= 0 {
		errors = append(errors, "  - transports.nrt.timeout must be > 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf("transport configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateAuthConfig validates authentication configuration
func (cm *ConfigManager) validateAuthConfig() error {
	var errors []string

	auth := cm.config.Auth

	if auth.TokenExpiry <= 0 {
		errors = append(errors, "  - auth.token_expiry must be > 0 hours")
	}

	if len(errors) > 0 {
		return fmt.Errorf("authentication configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateNetworkConfig validates network configuration
func (cm *ConfigManager) validateNetworkConfig() error {
	var errors []string

	network := cm.config.Network

	if network.Timeout <= 0 {
		errors = append(errors, "  - network.timeout must be > 0 seconds")
	}

	if network.RetryAttempts < 0 {
		errors = append(errors, "  - network.retry_attempts must be >= 0")
	}

	if network.RetryDelay <= 0 {
		errors = append(errors, "  - network.retry_delay must be > 0 seconds")
	}

	if len(errors) > 0 {
		return fmt.Errorf("network configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func (cm *ConfigManager) validateLoggingConfig() error {
	var errors []string

	logging := cm.config.Logging

	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, logging.Level) {
		errors = append(errors, fmt.Sprintf("  - logging.level must be one of: %s, got: %s", strings.Join(validLevels, ", "), logging.Level))
	}

	if logging.MaxSize <= 0 {
		errors = append(errors, "  - logging.max_size must be > 0 MB")
	}

	if logging.MaxBackups <= 0 {
		errors = append(errors, "  - logging.max_backups must be > 0")
	}

	if logging.MaxAge <= 0 {
		errors = append(errors, "  - logging.max_age must be > 0 days")
	}

	if len(errors) > 0 {
		return fmt.Errorf("logging configuration errors:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Get returns the current configuration
func (cm *ConfigManager) Get() *Config {
	return cm.config
}

// Save saves the current configuration to file
func (cm *ConfigManager) Save() error {
	// Get the config file path
	configPath := cm.v.ConfigFileUsed()
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".mauritania-cli.toml")
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if configDir != "." && configDir != "" {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Write config directly to the known path to avoid viper path issues
	configPath = filepath.Join(os.Getenv("HOME"), ".mauritania-cli.toml")

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Get current values from viper (which includes our Set() calls)
	authEnabled := cm.v.GetBool("auth.enabled")
	requireApproval := cm.v.GetBool("auth.require_approval")
	tokenExpiry := cm.v.GetInt("auth.token_expiry")
	autoComplete := cm.v.GetBool("cli.auto_complete")
	historySize := cm.v.GetInt("cli.history_size")
	mobileOptimized := cm.v.GetBool("cli.mobile_optimized")
	promptSymbol := cm.v.GetString("cli.prompt_symbol")
	dbPath := cm.v.GetString("database.path")
	dbType := cm.v.GetString("database.type")
	logCompress := cm.v.GetBool("logging.compress")
	logLevel := cm.v.GetString("logging.level")
	logMaxAge := cm.v.GetInt("logging.max_age")
	logMaxBackups := cm.v.GetInt("logging.max_backups")
	logMaxSize := cm.v.GetInt("logging.max_size")
	offlineMode := cm.v.GetBool("network.offline_mode")
	retryAttempts := cm.v.GetInt("network.retry_attempts")
	retryDelay := cm.v.GetInt("network.retry_delay")
	networkTimeout := cm.v.GetInt("network.timeout")
	defaultTransport := cm.v.GetString("transports.default")

	content := fmt.Sprintf(`[auth]
enabled = %t
require_approval = %t
token_expiry = %d

[cli]
auto_complete = %t
history_size = %d
mobile_optimized = %t
prompt_symbol = "%s"

[database]
path = "%s"
type = "%s"

[logging]
compress = %t
level = "%s"
max_age = %d
max_backups = %d
max_size = %d

[network]
offline_mode = %t
retry_attempts = %d
retry_delay = %d
timeout = %d

[transports]
default = "%s"
`,
		authEnabled,
		requireApproval,
		tokenExpiry,
		autoComplete,
		historySize,
		mobileOptimized,
		promptSymbol,
		dbPath,
		dbType,
		logCompress,
		logLevel,
		logMaxAge,
		logMaxBackups,
		logMaxSize,
		offlineMode,
		retryAttempts,
		retryDelay,
		networkTimeout,
		defaultTransport,
	)

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write config content: %w", err)
	}

	return nil
}

// Set updates a configuration value
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.v.Set(key, value)
	// Update cached config if it exists
	if cm.config != nil {
		// Re-unmarshal to update the cached config
		_ = cm.v.Unmarshal(cm.config)
	}
}

// GetString returns a string configuration value
func (cm *ConfigManager) GetString(key string) string {
	return cm.v.GetString(key)
}

// GetInt returns an integer configuration value
func (cm *ConfigManager) GetInt(key string) int {
	return cm.v.GetInt(key)
}

// GetBool returns a boolean configuration value
func (cm *ConfigManager) GetBool(key string) bool {
	return cm.v.GetBool(key)
}

// GetStringSlice returns a string slice configuration value
func (cm *ConfigManager) GetStringSlice(key string) []string {
	return cm.v.GetStringSlice(key)
}

// IsSet checks if a configuration key is set
func (cm *ConfigManager) IsSet(key string) bool {
	return cm.v.IsSet(key)
}

// GetTransportConfig returns transport configuration for a specific type
func (cm *ConfigManager) GetTransportConfig(transportType string) interface{} {
	switch transportType {
	case "whatsapp":
		config := WhatsAppConfig{}
		cm.v.UnmarshalKey("transports.social_media.whatsapp", &config)
		return config
	case "telegram":
		config := TelegramConfig{}
		cm.v.UnmarshalKey("transports.social_media.telegram", &config)
		return config
	case "facebook":
		config := FacebookConfig{}
		cm.v.UnmarshalKey("transports.social_media.facebook", &config)
		return config
	case "shipper":
		config := ShipperConfig{}
		cm.v.UnmarshalKey("transports.shipper", &config)
		return config
	case "nrt":
		config := NRTConfig{}
		cm.v.UnmarshalKey("transports.nrt", &config)
		return config
	default:
		return nil
	}
}

// ValidateConfig validates the configuration
func (cm *ConfigManager) ValidateConfig() []string {
	var errors []string

	// Validate transport configurations
	if cm.config.Transports.SocialMedia.WhatsApp.DatabasePath == "" &&
		cm.config.Transports.SocialMedia.Telegram.BotToken == "" &&
		cm.config.Transports.SocialMedia.Facebook.AccessToken == "" {
		errors = append(errors, "at least one social media transport must be configured")
	}

	// Validate database configuration
	if cm.config.Database.Type == "sqlite" && cm.config.Database.Path == "" {
		errors = append(errors, "sqlite database path must be specified")
	}

	// Validate logging configuration
	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	levelValid := false
	for _, level := range validLevels {
		if strings.ToUpper(cm.config.Logging.Level) == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		errors = append(errors, "invalid logging level, must be one of: DEBUG, INFO, WARN, ERROR")
	}

	return errors
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(configPath string) error {
	cm := NewConfigManager()

	// Set some example values for the default config
	cm.Set("database.path", "./data/mauritania-cli.db")
	cm.Set("logging.level", "INFO")
	cm.Set("network.timeout", 30)
	cm.Set("network.retry_attempts", 3)

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set the config file path
	cm.v.SetConfigFile(configPath)

	// Save the configuration
	return cm.Save()
}
