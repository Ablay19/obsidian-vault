package types

import (
	"log/slog"
	"os"
)

type WorkerModule struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	EntryPoint   string            `json:"entryPoint"`
	Dependencies []string          `json:"dependencies"`
	Environment  map[string]string `json:"environment"`
	Routes       []RouteMapping    `json:"routes"`
	Permissions  []string          `json:"permissions"`
	Resources    ResourceLimits    `json:"resources"`
	Status       string            `json:"status"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type RouteMapping struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Target  string            `json:"target"`
	Timeout int               `json:"timeout"`
	Headers map[string]string `json:"headers"`
}

type ResourceLimits struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Storage string `json:"storage"`
}

type GoApplication struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Description  string         `json:"description"`
	ModulePath   string         `json:"modulePath"`
	EntryPoint   string         `json:"entryPoint"`
	Port         int            `json:"port"`
	Database     DatabaseConfig `json:"database"`
	APIs         []APIEndpoint  `json:"apis"`
	Dependencies []GoDependency `json:"dependencies"`
	Resources    ResourceLimits `json:"resources"`
	Status       string         `json:"status"`
	CreatedAt    string         `json:"createdAt"`
	UpdatedAt    string         `json:"updatedAt"`
}

type DatabaseConfig struct {
	Type       string `json:"type"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Name       string `json:"name"`
	Migrations string `json:"migrations"`
}

type APIEndpoint struct {
	Path           string          `json:"path"`
	Method         string          `json:"method"`
	Description    string          `json:"description"`
	Authentication bool            `json:"authentication"`
	RateLimit      RateLimitConfig `json:"rateLimit"`
}

type RateLimitConfig struct {
	Requests int    `json:"requests"`
	Window   string `json:"window"`
}

type GoDependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SharedPackage struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         string            `json:"type"`
	Languages    []string          `json:"languages"`
	GoModule     *GoPackageConfig  `json:"goModule,omitempty"`
	NpmPackage   *NpmPackageConfig `json:"npmPackage,omitempty"`
	Dependencies []string          `json:"dependencies"`
	Exports      []string          `json:"exports"`
	Status       string            `json:"status"`
	CreatedAt    string            `json:"createdAt"`
	UpdatedAt    string            `json:"updatedAt"`
}

type GoPackageConfig struct {
	ModulePath string   `json:"modulePath"`
	GoVersion  string   `json:"goVersion"`
	Imports    []string `json:"imports"`
}

type NpmPackageConfig struct {
	PackageName string            `json:"packageName"`
	Main        string            `json:"main"`
	Types       string            `json:"types"`
	Scripts     map[string]string `json:"scripts"`
}

type APIGateway struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Routes      []Route         `json:"routes"`
	Middlewares []Middleware    `json:"middlewares"`
	RateLimit   RateLimitConfig `json:"rateLimit"`
	CORS        CORSConfig      `json:"cors"`
	Auth        AuthConfig      `json:"auth"`
}

type Route struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Target  string      `json:"target"`
	Timeout int         `json:"timeout"`
	Retry   RetryConfig `json:"retry"`
}

type Middleware struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type RetryConfig struct {
	MaxAttempts int    `json:"maxAttempts"`
	Backoff     int    `json:"backoff"`
	Strategy    string `json:"strategy"`
}

type CORSConfig struct {
	AllowedOrigins []string `json:"allowedOrigins"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedHeaders []string `json:"allowedHeaders"`
	MaxAge         int      `json:"maxAge"`
}

type AuthConfig struct {
	Enabled   bool   `json:"enabled"`
	Type      string `json:"type"`
	RateLimit bool   `json:"rateLimit"`
}

type DeploymentPipeline struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	ComponentID    string               `json:"componentId"`
	ComponentType  string               `json:"componentType"`
	Triggers       []TriggerConfig      `json:"triggers"`
	Stages         []PipelineStage      `json:"stages"`
	Environment    EnvironmentConfig    `json:"environment"`
	Notifications  []NotificationConfig `json:"notifications"`
	LastDeployment *DeploymentStatus    `json:"lastDeployment,omitempty"`
	CreatedAt      string               `json:"createdAt"`
	UpdatedAt      string               `json:"updatedAt"`
}

type TriggerConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type PipelineStage struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Commands  []string `json:"commands"`
	Timeout   int      `json:"timeout"`
	OnFailure string   `json:"onFailure"`
}

type EnvironmentConfig struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Variables map[string]string `json:"variables"`
}

type NotificationConfig struct {
	Type   string   `json:"type"`
	Target string   `json:"target"`
	Events []string `json:"events"`
}

type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasNext bool `json:"hasNext"`
}

type DeploymentStatus struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime,omitempty"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

type APIResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	TimeFormat string `json:"timeFormat"`
}

func NewStructuredLogger(component string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	logger = logger.With("component", component)
	return logger
}

func LogError(logger *slog.Logger, err error, message string, args ...any) {
	logger.Error(message, append([]any{"error", err}, args...)...)
}

func LogInfo(logger *slog.Logger, message string, args ...any) {
	logger.Info(message, args...)
}

func LogWarn(logger *slog.Logger, message string, args ...any) {
	logger.Warn(message, args...)
}

func LogDebug(logger *slog.Logger, message string, args ...any) {
	logger.Debug(message, args...)
}
