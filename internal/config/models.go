package config

// ProviderConfig holds the configuration for a single AI provider.
type ProviderConfig struct {
	ProviderName         string  `mapstructure:"provider_name"`
	ModelName            string  `mapstructure:"model_name"`
	InputCostPerToken    float64 `mapstructure:"input_cost_per_token"`
	OutputCostPerToken   float64 `mapstructure:"output_cost_per_token"`
	MaxInputTokens       int     `mapstructure:"max_input_tokens"`
	MaxOutputTokens      int     `mapstructure:"max_output_tokens"`
	LatencyMsThreshold   int     `mapstructure:"latency_ms_threshold"`
	AccuracyPctThreshold float64 `mapstructure:"accuracy_pct_threshold"`
}
