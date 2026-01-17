package utils

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ErrorAlertManager manages error alerts and notifications
type ErrorAlertManager struct {
	metricsCollector *ErrorMetricsCollector
	alertRules       []AlertRule
	activeAlerts     map[string]*Alert
	alertChannels    []AlertChannel
	mutex            sync.RWMutex
	logger           *slog.Logger
}

// AlertRule defines when to trigger an alert
type AlertRule struct {
	ID             string
	Name           string
	Service        string
	ErrorType      string
	Threshold      AlertThreshold
	CooldownPeriod time.Duration
	Severity       AlertSeverity
	Enabled        bool
}

// AlertThreshold defines the conditions for triggering an alert
type AlertThreshold struct {
	ErrorRateThreshold      float64       // errors per minute
	ErrorCountThreshold     int64         // total errors in time window
	ConsecutiveFailureCount int           // consecutive failures
	TimeWindow              time.Duration // time window for counting
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity int

const (
	AlertSeverityLow AlertSeverity = iota
	AlertSeverityMedium
	AlertSeverityHigh
	AlertSeverityCritical
)

func (as AlertSeverity) String() string {
	switch as {
	case AlertSeverityLow:
		return "low"
	case AlertSeverityMedium:
		return "medium"
	case AlertSeverityHigh:
		return "high"
	case AlertSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Alert represents an active alert
type Alert struct {
	ID          string
	RuleID      string
	Service     string
	ErrorType   string
	Severity    AlertSeverity
	Message     string
	TriggeredAt time.Time
	LastSeenAt  time.Time
	ResolvedAt  *time.Time
	Count       int64
	Metadata    map[string]interface{}
}

// AlertChannel defines how alerts are delivered
type AlertChannel interface {
	SendAlert(alert *Alert) error
	ResolveAlert(alertID string) error
}

// SlackAlertChannel sends alerts to Slack
type SlackAlertChannel struct {
	WebhookURL string
}

// EmailAlertChannel sends alerts via email
type EmailAlertChannel struct {
	SMTPHost  string
	SMTPPort  int
	Username  string
	Password  string
	FromEmail string
	ToEmails  []string
}

// NewErrorAlertManager creates a new error alert manager
func NewErrorAlertManager(metricsCollector *ErrorMetricsCollector, logger *slog.Logger) *ErrorAlertManager {
	return &ErrorAlertManager{
		metricsCollector: metricsCollector,
		alertRules:       make([]AlertRule, 0),
		activeAlerts:     make(map[string]*Alert),
		alertChannels:    make([]AlertChannel, 0),
		logger:           logger,
	}
}

// AddAlertRule adds a new alert rule
func (eam *ErrorAlertManager) AddAlertRule(rule AlertRule) {
	eam.mutex.Lock()
	defer eam.mutex.Unlock()

	eam.alertRules = append(eam.alertRules, rule)
	eam.logger.Info("Added alert rule", "rule_id", rule.ID, "service", rule.Service)
}

// RemoveAlertRule removes an alert rule
func (eam *ErrorAlertManager) RemoveAlertRule(ruleID string) {
	eam.mutex.Lock()
	defer eam.mutex.Unlock()

	for i, rule := range eam.alertRules {
		if rule.ID == ruleID {
			eam.alertRules = append(eam.alertRules[:i], eam.alertRules[i+1:]...)
			eam.logger.Info("Removed alert rule", "rule_id", ruleID)
			break
		}
	}
}

// AddAlertChannel adds an alert channel
func (eam *ErrorAlertManager) AddAlertChannel(channel AlertChannel) {
	eam.mutex.Lock()
	defer eam.mutex.Unlock()

	eam.alertChannels = append(eam.alertChannels, channel)
}

// CheckAlerts checks all alert rules and triggers alerts if necessary
func (eam *ErrorAlertManager) CheckAlerts() {
	eam.mutex.RLock()
	rules := make([]AlertRule, len(eam.alertRules))
	copy(rules, eam.alertRules)
	eam.mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		eam.checkRule(&rule)
	}
}

// checkRule checks a single alert rule
func (eam *ErrorAlertManager) checkRule(rule *AlertRule) {
	metrics := eam.metricsCollector.GetMetrics(rule.Service, rule.ErrorType)
	if metrics == nil {
		return
	}

	alertKey := rule.ID
	shouldAlert := eam.evaluateThreshold(metrics, &rule.Threshold)

	eam.mutex.Lock()
	defer eam.mutex.Unlock()

	if shouldAlert {
		alert, exists := eam.activeAlerts[alertKey]
		if !exists {
			// Create new alert
			alert = &Alert{
				ID:          generateAlertID(),
				RuleID:      rule.ID,
				Service:     rule.Service,
				ErrorType:   rule.ErrorType,
				Severity:    rule.Severity,
				Message:     eam.createAlertMessage(rule, metrics),
				TriggeredAt: time.Now(),
				LastSeenAt:  time.Now(),
				Count:       1,
				Metadata: map[string]interface{}{
					"error_rate":      metrics.ErrorRate,
					"total_count":     metrics.TotalCount,
					"last_error_time": metrics.LastErrorTime,
				},
			}
			eam.activeAlerts[alertKey] = alert

			// Send alert to all channels
			eam.sendAlertToChannels(alert)
			eam.logger.Warn("Alert triggered",
				"alert_id", alert.ID,
				"rule_id", rule.ID,
				"service", rule.Service,
				"severity", rule.Severity.String())

		} else {
			// Update existing alert
			alert.LastSeenAt = time.Now()
			alert.Count++
			alert.Metadata["error_rate"] = metrics.ErrorRate
			alert.Metadata["total_count"] = metrics.TotalCount
			alert.Metadata["last_error_time"] = metrics.LastErrorTime
		}
	} else {
		// Check if we should resolve an existing alert
		if alert, exists := eam.activeAlerts[alertKey]; exists {
			// Check if alert should be resolved based on cooldown period
			if time.Since(alert.LastSeenAt) > rule.CooldownPeriod {
				alert.ResolvedAt = &time.Time{}
				*alert.ResolvedAt = time.Now()

				// Send resolution to channels
				eam.resolveAlertInChannels(alert.ID)
				eam.logger.Info("Alert resolved",
					"alert_id", alert.ID,
					"rule_id", rule.ID,
					"service", rule.Service)

				delete(eam.activeAlerts, alertKey)
			}
		}
	}
}

// evaluateThreshold evaluates if the threshold conditions are met
func (eam *ErrorAlertManager) evaluateThreshold(metrics *ErrorMetrics, threshold *AlertThreshold) bool {
	// Check error rate threshold
	if threshold.ErrorRateThreshold > 0 && metrics.ErrorRate >= threshold.ErrorRateThreshold {
		return true
	}

	// Check error count threshold
	if threshold.ErrorCountThreshold > 0 && metrics.TotalCount >= threshold.ErrorCountThreshold {
		return true
	}

	// Check time window - if errors are happening too frequently
	if threshold.TimeWindow > 0 {
		timeSinceFirst := time.Since(metrics.FirstErrorTime)
		if timeSinceFirst <= threshold.TimeWindow && metrics.TotalCount >= int64(threshold.ConsecutiveFailureCount) {
			return true
		}
	}

	return false
}

// createAlertMessage creates a human-readable alert message
func (eam *ErrorAlertManager) createAlertMessage(rule *AlertRule, metrics *ErrorMetrics) string {
	return fmt.Sprintf("🚨 Alert: %s in service %s\nError Rate: %.2f/min\nTotal Errors: %d\nLast Error: %s",
		rule.Name,
		rule.Service,
		metrics.ErrorRate,
		metrics.TotalCount,
		metrics.LastErrorTime.Format(time.RFC3339))
}

// sendAlertToChannels sends an alert to all configured channels
func (eam *ErrorAlertManager) sendAlertToChannels(alert *Alert) {
	for _, channel := range eam.alertChannels {
		go func(ch AlertChannel, a *Alert) {
			if err := ch.SendAlert(a); err != nil {
				eam.logger.Error("Failed to send alert to channel", "error", err, "alert_id", a.ID)
			}
		}(channel, alert)
	}
}

// resolveAlertInChannels resolves an alert in all configured channels
func (eam *ErrorAlertManager) resolveAlertInChannels(alertID string) {
	for _, channel := range eam.alertChannels {
		go func(ch AlertChannel, id string) {
			if err := ch.ResolveAlert(id); err != nil {
				eam.logger.Error("Failed to resolve alert in channel", "error", err, "alert_id", id)
			}
		}(channel, alertID)
	}
}

// GetActiveAlerts returns all active alerts
func (eam *ErrorAlertManager) GetActiveAlerts() map[string]*Alert {
	eam.mutex.RLock()
	defer eam.mutex.RUnlock()

	alerts := make(map[string]*Alert)
	for k, v := range eam.activeAlerts {
		alerts[k] = v
	}
	return alerts
}

// GetAlertRules returns all alert rules
func (eam *ErrorAlertManager) GetAlertRules() []AlertRule {
	eam.mutex.RLock()
	defer eam.mutex.RUnlock()

	rules := make([]AlertRule, len(eam.alertRules))
	copy(rules, eam.alertRules)
	return rules
}

// StartMonitoring starts the alert monitoring routine
func (eam *ErrorAlertManager) StartMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			eam.CheckAlerts()
		}
	}()

	eam.logger.Info("Started error alert monitoring", "interval", interval)
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// SlackAlertChannel implementation
func (sac *SlackAlertChannel) SendAlert(alert *Alert) error {
	// Implementation would send HTTP POST to Slack webhook
	// For now, just log
	fmt.Printf("SLACK ALERT: %s\n", alert.Message)
	return nil
}

func (sac *SlackAlertChannel) ResolveAlert(alertID string) error {
	// Implementation would send resolution message to Slack
	fmt.Printf("SLACK RESOLVE: Alert %s resolved\n", alertID)
	return nil
}

// EmailAlertChannel implementation
func (eac *EmailAlertChannel) SendAlert(alert *Alert) error {
	// Implementation would send email via SMTP
	fmt.Printf("EMAIL ALERT: %s\n", alert.Message)
	return nil
}

func (eac *EmailAlertChannel) ResolveAlert(alertID string) error {
	// Implementation would send resolution email
	fmt.Printf("EMAIL RESOLVE: Alert %s resolved\n", alertID)
	return nil
}
