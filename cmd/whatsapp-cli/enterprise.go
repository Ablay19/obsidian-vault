package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Tenant struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	APIKey      string                 `json:"api_key"`
	CreatedAt   time.Time              `json:"created_at"`
	Settings    map[string]interface{} `json:"settings"`
	Permissions map[string]bool        `json:"permissions"`
}

type AuditLog struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Data      string    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address"`
}

type HealthStatus struct {
	Service   string                 `json:"service"`
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Uptime    time.Time              `json:"uptime"`
	Metrics   map[string]interface{} `json:"metrics"`
	LastCheck time.Time              `json:"last_check"`
}

var (
	tenants       = make(map[string]*Tenant)
	auditLogs     = make([]AuditLog, 0)
	healthStatus  = make(map[string]*HealthStatus)
	encryptionKey []byte
)

func loadEncryptionKey() []byte {
	if keyStr := os.Getenv("ENCRYPTION_KEY"); len(keyStr) >= 32 {
		return []byte(keyStr[:32])
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err == nil && logger != nil {
		logger.Warn("ENCRYPTION_KEY not set or too short, using generated key")
	}
	return bytes
}

func initEnterprise() {
	encryptionKey = loadEncryptionKey()
	loadTenants()
	loadAuditLogs()
	initHealthChecks()

	logger.Info("Enterprise features initialized", "tenants_loaded", len(tenants))
}

func loadTenants() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	tenantFile := filepath.Join(configDir, "whatsapp-cli", "tenants.json")

	if _, err := os.Stat(tenantFile); os.IsNotExist(err) {
		createDefaultTenant()
		saveTenants()
		return
	}

	data, err := os.ReadFile(tenantFile)
	if err != nil {
		logger.Error("Failed to read tenants file", "error", err)
		createDefaultTenant()
		return
	}

	var tenantList []Tenant
	if err := json.Unmarshal(data, &tenantList); err != nil {
		logger.Error("Failed to unmarshal tenants, falling back to default tenant", "error", err)
		createDefaultTenant()
		return
	}

	for _, tenant := range tenantList {
		t := tenant
		tenants[t.ID] = &t
	}
}

func createDefaultTenant() {
	defaultTenant := &Tenant{
		ID:        "default",
		Name:      "Default Organization",
		APIKey:    generateAPIKey(),
		CreatedAt: time.Now(),
		Settings: map[string]interface{}{
			"max_messages_per_day": 1000,
			"ai_enabled":           true,
		},
		Permissions: map[string]bool{
			"send_messages":    true,
			"manage_templates": true,
			"view_analytics":   true,
		},
	}
	tenants["default"] = defaultTenant
}

func saveTenants() {
	var tenantList []Tenant
	for _, tenant := range tenants {
		tenantList = append(tenantList, *tenant)
	}

	data, err := json.MarshalIndent(tenantList, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal tenants", "error", err)
		return
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	tenantFile := filepath.Join(configDir, "whatsapp-cli", "tenants.json")

	os.MkdirAll(filepath.Dir(tenantFile), 0755)
	if err := os.WriteFile(tenantFile, data, 0644); err != nil {
		logger.Error("Failed to save tenants", "error", err)
	}
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		logger.Error("Failed to generate random bytes for API key", "error", err)
		return ""
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func authenticateTenant(apiKey string) (*Tenant, error) {
	for _, tenant := range tenants {
		if tenant.APIKey == apiKey {
			return tenant, nil
		}
	}
	return nil, fmt.Errorf("invalid API key")
}

func authorizeAction(tenant *Tenant, action string) bool {
	if tenant.ID == "default" {
		return true // Default tenant has all permissions
	}
	return tenant.Permissions[action]
}

func logAudit(tenantID, userID, action, resource, data, ipAddress string) {
	entry := AuditLog{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		TenantID:  tenantID,
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Data:      data,
		Timestamp: time.Now(),
		IPAddress: ipAddress,
	}

	auditLogs = append(auditLogs, entry)

	// Keep only last 1000 entries
	if len(auditLogs) > 1000 {
		auditLogs = auditLogs[len(auditLogs)-1000:]
	}

	logger.Info("Audit log entry created",
		"tenant_id", tenantID,
		"action", action,
		"resource", resource)
}

func encryptData(data string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptData(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func initHealthChecks() {
	// WhatsApp service health
	healthStatus["whatsapp"] = &HealthStatus{
		Service:   "whatsapp",
		Status:    "unknown",
		Version:   "1.0.0",
		Uptime:    time.Now(),
		Metrics:   make(map[string]interface{}),
		LastCheck: time.Now(),
	}

	// Queue manager health
	healthStatus["queue"] = &HealthStatus{
		Service:   "queue",
		Status:    "unknown",
		Version:   "1.0.0",
		Uptime:    time.Now(),
		Metrics:   make(map[string]interface{}),
		LastCheck: time.Now(),
	}

	// AI service health
	healthStatus["ai"] = &HealthStatus{
		Service:   "ai",
		Status:    "unknown",
		Version:   "1.0.0",
		Uptime:    time.Now(),
		Metrics:   make(map[string]interface{}),
		LastCheck: time.Now(),
	}

	// Start health check goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			updateHealthStatus()
		}
	}()
}

func updateHealthStatus() {
	// Update WhatsApp status
	if wac != nil {
		healthStatus["whatsapp"].Status = "healthy"
		healthStatus["whatsapp"].Metrics["connected"] = true
	} else {
		healthStatus["whatsapp"].Status = "disconnected"
		healthStatus["whatsapp"].Metrics["connected"] = false
	}

	// Update queue status
	if queueMgr != nil {
		healthStatus["queue"].Status = "healthy"
		healthStatus["queue"].Metrics["initialized"] = true
	} else {
		healthStatus["queue"].Status = "unavailable"
		healthStatus["queue"].Metrics["initialized"] = false
	}

	// Update AI status
	if config.AI.Enabled {
		healthStatus["ai"].Status = "enabled"
		healthStatus["ai"].Metrics["configured"] = true
	} else {
		healthStatus["ai"].Status = "disabled"
		healthStatus["ai"].Metrics["configured"] = false
	}

	for _, status := range healthStatus {
		status.LastCheck = time.Now()
	}
}

func loadAuditLogs() {
	// In a real implementation, load from database
	// For now, keep in memory
}

func saveAuditLogs() {
	// In a real implementation, save to database
	// For now, just log summary
	logger.Info("Audit logs summary", "total_entries", len(auditLogs))
}
