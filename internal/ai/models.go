package ai

// RequestModel represents a standardized request to an AI provider.
type RequestModel struct {
	SystemPrompt string   `json:"system_prompt,omitempty"`
	UserPrompt   string   `json:"user_prompt"`
	Images       []string `json:"images,omitempty"`     // Base64 encoded or URLs
	ImageData    []byte   `json:"image_data,omitempty"` // Raw bytes
	Model        string   `json:"model,omitempty"`
	Temperature  float64  `json:"temperature,omitempty"`
	MaxTokens    int      `json:"max_tokens,omitempty"`
	JSONMode     bool     `json:"json_mode,omitempty"` // Request JSON output
	RequestID    string   `json:"request_id,omitempty"`
}

// ResponseModel represents a standardized response from an AI provider.
type ResponseModel struct {
	Content      string            `json:"content"`
	Usage        TokenUsage        `json:"usage,omitempty"`
	Cached       bool              `json:"cached,omitempty"`
	ProviderInfo ModelInfo         `json:"provider_info"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// StreamResponse represents a chunk of a streamed response.
type StreamResponse struct {
	Content string `json:"content,omitempty"`
	Error   error  `json:"error,omitempty"`
	Done    bool   `json:"done,omitempty"`
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// AnalysisResult represents the structured output expected from document analysis.
type AnalysisResult struct {
	Category  string   `json:"category"`
	Topics    []string `json:"topics"`
	Questions []string `json:"questions"`
}
