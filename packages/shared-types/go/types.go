package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
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
	logger.Error(message, append([]any{"error", err.Error()}, args...)...)
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

type ColoredJSONHandler struct {
	handler slog.Handler
}

func NewColoredJSONHandler(output *os.File, opts *slog.HandlerOptions) *ColoredJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ColoredJSONHandler{
		handler: slog.NewJSONHandler(output, opts),
	}
}

func (h *ColoredJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ColoredJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: r.Level,
	})
	if err := jsonHandler.Handle(ctx, r); err != nil {
		return err
	}
	colored := colorizeJSON(buf.String())
	_, err := fmt.Fprint(os.Stdout, colored)
	return err
}

func (h *ColoredJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColoredJSONHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *ColoredJSONHandler) WithGroup(name string) slog.Handler {
	return &ColoredJSONHandler{handler: h.handler.WithGroup(name)}
}

func colorizeJSON(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			str, end := extractString(s, i+1)
			result.WriteString("\033[38;5;214m\"")
			result.WriteString(str)
			result.WriteString("\"\033[0m")
			i = end
		} else if c == '{' || c == '[' {
			result.WriteString("\033[38;5;39m")
			result.WriteByte(c)
			result.WriteString("\033[0m")
		} else if c == '}' || c == ']' {
			result.WriteString("\033[38;5;39m")
			result.WriteByte(c)
			result.WriteString("\033[0m")
		} else if c == ':' {
			result.WriteString("\033[38;5;39m:\033[0m")
		} else if c == ',' {
			result.WriteString("\033[38;5;39m,\033[0m")
		} else if c == 't' && i+4 <= len(s) && s[i:i+4] == "true" {
			result.WriteString("\033[38;5;220mtrue\033[0m")
			i += 3
		} else if c == 'f' && i+5 <= len(s) && s[i:i+5] == "false" {
			result.WriteString("\033[38;5;220mfalse\033[0m")
			i += 4
		} else if c == 'n' && i+4 <= len(s) && s[i:i+4] == "null" {
			result.WriteString("\033[38;5;220mnull\033[0m")
			i += 3
		} else if (c >= '0' && c <= '9') || c == '-' {
			num, end := extractNumber(s, i)
			result.WriteString("\033[38;5;154m")
			result.WriteString(num)
			result.WriteString("\033[0m")
			i = end - 1
		} else {
			result.WriteByte(c)
		}
	}
	return result.String()
}

func extractString(s string, start int) (string, int) {
	var result strings.Builder
	for i := start; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			result.WriteByte(s[i])
			result.WriteByte(s[i+1])
			i++
		} else if s[i] == '"' {
			return result.String(), i
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String(), len(s)
}

func extractNumber(s string, start int) (string, int) {
	var result strings.Builder
	for i := start; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-' {
			result.WriteByte(c)
		} else {
			return result.String(), i
		}
	}
	return result.String(), len(s)
}

func NewColoredLogger(component string) *slog.Logger {
	handler := NewColoredJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler).With("component", component)
}

type Logger struct {
	*slog.Logger
}

func (l *Logger) Error(message string, args ...any) {
	l.Log(context.Background(), slog.LevelError, message, args...)
}

func (l *Logger) Info(message string, args ...any) {
	l.Log(context.Background(), slog.LevelInfo, message, args...)
}

func (l *Logger) Warn(message string, args ...any) {
	l.Log(context.Background(), slog.LevelWarn, message, args...)
}

func (l *Logger) Debug(message string, args ...any) {
	l.Log(context.Background(), slog.LevelDebug, message, args...)
}

func LogStructured(level string, message string, data map[string]interface{}) {
	entry := map[string]interface{}{
		"level":   level,
		"msg":     message,
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"service": "obsidian-vault",
	}
	for k, v := range data {
		entry[k] = v
	}
	jsonBytes, _ := json.Marshal(entry)
	colored := colorizeJSON(string(jsonBytes))
	fmt.Println(colored)
}
