package whatsapp

// Message types
const (
	MessageTypeText     = "text"
	MessageTypeImage    = "image"
	MessageTypeAudio    = "audio"
	MessageTypeVideo    = "video"
	MessageTypeDocument = "document"
)

// Config holds WhatsApp service configuration
type Config struct {
	AccessToken   string `json:"access_token"`
	VerifyToken   string `json:"verify_token"`
	AppSecret     string `json:"app_secret"`
	WebhookURL    string `json:"webhook_url"`
	PhoneNumberID string `json:"phone_number_id"`
}

// Message represents a WhatsApp message
type Message struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	Timestamp int64                  `json:"timestamp"`
}

// WebhookPayload represents the complete webhook payload from WhatsApp
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

// Entry represents a webhook entry
type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

// Change represents a change within an entry
type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

// Value represents the value in a change
type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Messages         []Message `json:"messages,omitempty"`
}

// Metadata represents message metadata
type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// Media represents media information
type Media struct {
	ID        string `json:"id"`
	MimeType  string `json:"mime_type"`
	Sha256    string `json:"sha256"`
	FileSize  int64  `json:"file_size"`
	LocalPath string `json:"local_path"`
	URL       string `json:"url"`
}
