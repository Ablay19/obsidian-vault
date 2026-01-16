package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type CLIConfig struct {
	RabbitMQ QueueConfig `mapstructure:"rabbitmq"`
	WhatsApp struct {
		SessionFile  string `mapstructure:"session_file"`
		AutoRetry    bool   `mapstructure:"auto_retry"`
		MediaTimeout string `mapstructure:"media_timeout"`
	} `mapstructure:"whatsapp"`
	AI struct {
		Enabled bool     `mapstructure:"enabled"`
		Queue   string   `mapstructure:"queue"`
		Models  []string `mapstructure:"models"`
	} `mapstructure:"ai"`
}

func loadConfig() (*CLIConfig, error) {
	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	configPath := filepath.Join(configDir, "whatsapp-cli")

	// Create config directory if it doesn't exist
	os.MkdirAll(configPath, 0755)

	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("rabbitmq.url", "amqp://localhost:5672")
	viper.SetDefault("rabbitmq.exchange", "whatsapp-exchange")
	viper.SetDefault("rabbitmq.queues.incoming", "whatsapp.incoming")
	viper.SetDefault("rabbitmq.queues.outgoing", "whatsapp.outgoing")
	viper.SetDefault("rabbitmq.queues.media", "whatsapp.media")
	viper.SetDefault("rabbitmq.queues.ai", "whatsapp.ai")
	viper.SetDefault("rabbitmq.queues.notifications", "whatsapp.notifications")
	viper.SetDefault("rabbitmq.queues.system", "whatsapp.system")

	viper.SetDefault("whatsapp.session_file", "whatsapp_session.gob")
	viper.SetDefault("whatsapp.auto_retry", true)
	viper.SetDefault("whatsapp.media_timeout", "30s")

	viper.SetDefault("ai.enabled", false)
	viper.SetDefault("ai.queue", "whatsapp.ai")
	viper.SetDefault("ai.models", []string{"gpt-3.5-turbo"})

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, use defaults and create it
		if err := viper.WriteConfigAs(filepath.Join(configPath, "config.yaml")); err != nil {
			log.Printf("Warning: Could not write default config: %v", err)
		}
	}

	var config CLIConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initQueueManager(config *CLIConfig) error {
	qm, err := NewQueueManager(config.RabbitMQ)
	if err != nil {
		return err
	}
	queueMgr = qm

	// Start background consumers
	go startQueueConsumers(config)

	return nil
}

func startQueueConsumers(config *CLIConfig) {
	// Consumer for outgoing messages
	queueMgr.ConsumeMessages(config.RabbitMQ.Queues.Outgoing, handleOutgoingMessage)

	// Consumer for AI responses
	if config.AI.Enabled {
		queueMgr.ConsumeMessages(config.AI.Queue, handleAIResponse)
	}

	// Consumer for system events
	queueMgr.ConsumeMessages(config.RabbitMQ.Queues.System, handleSystemEvent)

	log.Println("Queue consumers started")
}
