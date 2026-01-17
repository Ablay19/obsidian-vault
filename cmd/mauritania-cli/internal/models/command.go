package models

import "time"

// CommandStatus represents the execution status of a command
type CommandStatus string

const (
	StatusReceived  CommandStatus = "received"
	StatusValidated CommandStatus = "validated"
	StatusQueued    CommandStatus = "queued"
	StatusExecuting CommandStatus = "executing"
	StatusCompleted CommandStatus = "completed"
	StatusFailed    CommandStatus = "failed"
	StatusResponded CommandStatus = "responded"
	StatusExpired   CommandStatus = "expired"
)

// CommandPriority defines execution priority levels
type CommandPriority string

const (
	PriorityLow    CommandPriority = "low"
	PriorityNormal CommandPriority = "normal"
	PriorityHigh   CommandPriority = "high"
	PriorityUrgent CommandPriority = "urgent"
)

// TransportType defines available transport mechanisms
type TransportType string

const (
	TransportSocialMedia TransportType = "social_media"
	TransportSMApos      TransportType = "sm_apos"
	TransportNRT         TransportType = "nrt"
	TransportDirect      TransportType = "direct"
)

// SocialMediaCommand represents a command sent via social media transport
type SocialMediaCommand struct {
	ID          string          `json:"id"`
	SenderID    string          `json:"sender_id"`
	Platform    string          `json:"platform"` // whatsapp, telegram, facebook
	Command     string          `json:"command"`
	Timestamp   time.Time       `json:"timestamp"`
	Priority    CommandPriority `json:"priority"`
	Status      CommandStatus   `json:"status"`
	TransportID string          `json:"transport_id"`
	SessionID   string          `json:"session_id"`
}

// Command alias for backward compatibility
type Command = SocialMediaCommand

// MessageResponse represents the response from sending a message via transport
type MessageResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"` // sent, delivered, failed
	Timestamp time.Time `json:"timestamp"`
}

// IncomingMessage represents a message received from a transport
type IncomingMessage struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Transport string    `json:"transport"` // whatsapp, telegram, facebook
}

// TransportStatus represents the status of a transport
type TransportStatus struct {
	Available   bool      `json:"available"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}

// RateLimit represents rate limiting information
type RateLimit struct {
	RequestsPerHour   int       `json:"requests_per_hour"`
	RequestsRemaining int       `json:"requests_remaining"`
	ResetTime         time.Time `json:"reset_time"`
	IsThrottled       bool      `json:"is_throttled"`
}

// ShipperTransport defines the interface for SM APOS Shipper transport
type ShipperTransport interface {
	// Authenticate establishes a session with the shipper service
	Authenticate(credentials map[string]string) (*ShipperSession, error)

	// ExecuteCommand sends a command for execution through the shipper
	ExecuteCommand(session *ShipperSession, command string, timeout int) (*CommandResult, error)

	// GetCommandStatus checks the status of a running command
	GetCommandStatus(session *ShipperSession, commandID string) (*ShipperCommandStatus, error)

	// CancelCommand cancels a running command
	CancelCommand(session *ShipperSession, commandID string) error

	// ListActiveSessions returns all active shipper sessions
	ListActiveSessions() ([]*ShipperSession, error)

	// CloseSession terminates a shipper session
	CloseSession(sessionID string) error
}

// ShipperCommand represents a command sent to SM APOS Shipper
type ShipperCommand struct {
	ID         string            `json:"id"`
	Command    string            `json:"command"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Priority   CommandPriority   `json:"priority"`
	Timeout    int               `json:"timeout"` // seconds
	Encrypted  bool              `json:"encrypted"`
	SessionID  string            `json:"session_id"`
	CreatedAt  time.Time         `json:"created_at"`
	Status     CommandStatus     `json:"status"`
}

// ShipperCredentials holds authentication credentials for SM APOS Shipper
type ShipperCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"` // encrypted in storage
	APIKey   string `json:"api_key,omitempty"`
	Endpoint string `json:"endpoint"`
}

// ShipperCommandStatus represents the status of a command execution in shipper
type ShipperCommandStatus struct {
	CommandID   string        `json:"command_id"`
	Status      CommandStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	QueuedAt    *time.Time    `json:"queued_at,omitempty"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Progress    int           `json:"progress"` // 0-100
	Error       string        `json:"error,omitempty"`
}

// CommandResult represents the result of command execution
type CommandResult struct {
	ID            string    `json:"id"`
	CommandID     string    `json:"command_id"`
	Status        string    `json:"status"` // success, failure, partial, timeout
	ExitCode      int       `json:"exit_code"`
	Stdout        string    `json:"stdout"`
	Stderr        string    `json:"stderr"`
	ExecutionTime int64     `json:"execution_time"` // milliseconds
	TransportUsed string    `json:"transport_used"`
	Cost          float64   `json:"cost"`
	CompletedAt   time.Time `json:"completed_at"`
}

// NetworkRoute represents available network paths
type NetworkRoute struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // social_media, sm_apos, direct, nrt
	Provider    string    `json:"provider"`
	CostPerMB   float64   `json:"cost_per_mb"`
	Bandwidth   float64   `json:"bandwidth"`   // Mbps
	Latency     int       `json:"latency"`     // ms
	Reliability float64   `json:"reliability"` // 0-100
	LastTested  time.Time `json:"last_tested"`
	IsActive    bool      `json:"is_active"`
}

// ShipperSession represents authenticated session with SM APOS Shipper
type ShipperSession struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Token        string    `json:"-"` // encrypted, not serialized
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	Permissions  []string  `json:"permissions"`
	RateLimit    RateLimit `json:"rate_limit"`
	LastActivity time.Time `json:"last_activity"`
}

// ServiceConfig represents service configuration for management
type ServiceConfig struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"` // docker, kubernetes, process
	Image        string            `json:"image,omitempty"`
	Command      []string          `json:"command,omitempty"`
	Ports        []PortMapping     `json:"ports,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Resources    ResourceLimits    `json:"resources,omitempty"`
}

// PortMapping represents port configuration
type PortMapping struct {
	Container int    `json:"container"`
	Host      int    `json:"host"`
	Protocol  string `json:"protocol"` // tcp, udp
}

// ResourceLimits represents resource constraints
type ResourceLimits struct {
	CPU    CPULimits    `json:"cpu"`
	Memory MemoryLimits `json:"memory"`
}

// CPULimits represents CPU resource limits
type CPULimits struct {
	Request string `json:"request"` // e.g., "100m", "0.1"
	Limit   string `json:"limit"`   // e.g., "500m", "0.5"
}

// MemoryLimits represents memory resource limits
type MemoryLimits struct {
	Request string `json:"request"` // e.g., "64Mi", "128Mi"
	Limit   string `json:"limit"`   // e.g., "256Mi", "512Mi"
}

// ServiceStatus represents real-time service operational status
type ServiceStatus struct {
	ServiceID     string        `json:"service_id"`
	State         string        `json:"state"` // stopped, starting, healthy, unhealthy, stopping, failed, unknown
	Health        HealthStatus  `json:"health"`
	LastSeen      time.Time     `json:"last_seen"`
	Uptime        int64         `json:"uptime"` // seconds
	ResourceUsage ResourceUsage `json:"resource_usage"`
	ErrorMessage  string        `json:"error_message,omitempty"`
	RestartCount  int           `json:"restart_count"`
}

// HealthStatus represents health check results
type HealthStatus struct {
	Status    string    `json:"status"` // healthy, unhealthy, unknown
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
}

// ResourceUsage represents current resource consumption
type ResourceUsage struct {
	CPU    string `json:"cpu"`    // percentage
	Memory string `json:"memory"` // usage string
}

// TransportMetrics represents transport performance metrics
type TransportMetrics struct {
	TransportType  string    `json:"transport_type"`
	SuccessRate    float64   `json:"success_rate"`
	AverageLatency int       `json:"average_latency"`
	TotalCost      float64   `json:"total_cost"`
	MessagesSent   int       `json:"messages_sent"`
	LastUsed       time.Time `json:"last_used"`
}

// OfflineQueue represents queued commands for offline execution
type OfflineQueue struct {
	ID            string             `json:"id"`
	Command       SocialMediaCommand `json:"command"`
	QueuedAt      time.Time          `json:"queued_at"`
	Priority      int                `json:"priority"`
	RetryCount    int                `json:"retry_count"`
	LastRetry     *time.Time         `json:"last_retry,omitempty"`
	NextRetry     time.Time          `json:"next_retry"`
	FailureReason string             `json:"failure_reason,omitempty"`
}

// CommandHistory represents audit trail of CLI operations
type CommandHistory struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	Command          string    `json:"command"`
	Args             []string  `json:"args"`
	User             string    `json:"user"`
	Environment      string    `json:"environment"`
	Status           string    `json:"status"`   // success, failure, partial
	Duration         int64     `json:"duration"` // milliseconds
	AffectedServices []string  `json:"affected_services"`
	ErrorMessage     string    `json:"error_message,omitempty"`
}
