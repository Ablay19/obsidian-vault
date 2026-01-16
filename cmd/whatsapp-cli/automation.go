package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type AutomationRule struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Trigger     RuleTrigger  `json:"trigger"`
	Actions     []RuleAction `json:"actions"`
	Enabled     bool         `json:"enabled"`
	Priority    int          `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
}

type RuleTrigger struct {
	Type       string                 `json:"type"` // message, time, event
	Pattern    string                 `json:"pattern,omitempty"`
	Schedule   string                 `json:"schedule,omitempty"` // cron expression
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

type RuleAction struct {
	Type   string                 `json:"type"` // reply, forward, webhook, command
	Config map[string]interface{} `json:"config"`
}

type ConversationContext struct {
	JID         string                 `json:"jid"`
	Messages    []ContextMessage       `json:"messages"`
	Context     map[string]interface{} `json:"context"`
	LastUpdated time.Time              `json:"last_updated"`
}

type ContextMessage struct {
	Role    string    `json:"role"` // user, assistant
	Content string    `json:"content"`
	Time    time.Time `json:"time"`
}

var (
	rules         []AutomationRule
	conversations map[string]*ConversationContext
)

func initAutomation() {
	rules = loadAutomationRules()
	conversations = loadConversationContexts()

	// Setup default automation rules
	setupDefaultRules()

	logger.Info("Automation system initialized", "rules_count", len(rules))
}

func setupDefaultRules() {
	defaultRules := []AutomationRule{
		{
			ID:          "ping-responder",
			Name:        "Ping Responder",
			Description: "Automatically respond to ping messages",
			Trigger: RuleTrigger{
				Type:    "message",
				Pattern: "^ping$",
			},
			Actions: []RuleAction{
				{
					Type: "reply",
					Config: map[string]interface{}{
						"message": "Pong! 🤖 WhatsApp automation is active!",
					},
				},
			},
			Enabled:  true,
			Priority: 1,
		},
		{
			ID:          "help-request",
			Name:        "Help Request Handler",
			Description: "Respond to help requests",
			Trigger: RuleTrigger{
				Type:    "message",
				Pattern: "(?i)(help|support|assist)",
			},
			Actions: []RuleAction{
				{
					Type: "reply",
					Config: map[string]interface{}{
						"message": "Hi! I'm here to help. You can ask me questions or use commands like /ask, /status, etc.",
					},
				},
			},
			Enabled:  true,
			Priority: 2,
		},
		{
			ID:          "daily-reminder",
			Name:        "Daily Reminder",
			Description: "Send daily reminders",
			Trigger: RuleTrigger{
				Type:     "time",
				Schedule: "0 9 * * *", // 9 AM daily
			},
			Actions: []RuleAction{
				{
					Type: "broadcast",
					Config: map[string]interface{}{
						"message":  "Good morning! 🌅 Don't forget to check your messages.",
						"contacts": []string{"saved_contacts"}, // Would be dynamic
					},
				},
			},
			Enabled:  false, // Disabled by default
			Priority: 3,
		},
	}

	for _, rule := range defaultRules {
		if !ruleExists(rule.ID) {
			rules = append(rules, rule)
			if rule.Trigger.Type == "time" && rule.Enabled {
				setupCronJob(rule)
			}
		}
	}

	saveAutomationRules(rules)
}

func ruleExists(id string) bool {
	for _, rule := range rules {
		if rule.ID == id {
			return true
		}
	}
	return false
}

func setupCronJob(rule AutomationRule) {
	// Would add to cron scheduler if available
	logger.Info("Would schedule automation rule", "rule_id", rule.ID, "schedule", rule.Trigger.Schedule)
}

func processIncomingMessage(jid, message string) {
	logger.Info("Processing incoming message for automation", "jid", jid, "message_length", len(message))

	// Update conversation context
	updateConversationContext(jid, "user", message)

	// Check automation rules
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if matchesTrigger(rule.Trigger, message, jid) {
			logger.Info("Rule triggered", "rule_id", rule.ID, "rule_name", rule.Name, "jid", jid)
			executeRuleActions(rule, map[string]interface{}{
				"jid":     jid,
				"message": message,
			})
			break // Execute only highest priority matching rule
		}
	}

	// Check for AI integration
	if config.AI.Enabled && strings.HasPrefix(message, "/ask") {
		handleAIRequest(jid, message)
	}
}

func matchesTrigger(trigger RuleTrigger, message, jid string) bool {
	switch trigger.Type {
	case "message":
		if trigger.Pattern != "" {
			matched, _ := regexp.MatchString(trigger.Pattern, message)
			return matched
		}
		return strings.Contains(strings.ToLower(message), strings.ToLower(trigger.Pattern))
	case "time":
		return false // Handled by cron
	default:
		return false
	}
}

func executeRuleActions(rule AutomationRule, context map[string]interface{}) {
	for _, action := range rule.Actions {
		switch action.Type {
		case "reply":
			if msg, ok := action.Config["message"].(string); ok && context != nil {
				if jid, exists := context["jid"].(string); exists {
					sendMessage(jid, msg)
					updateConversationContext(jid, "assistant", msg)
				}
			}
		case "forward":
			// Forward to another platform
			if platform, ok := action.Config["platform"].(string); ok {
				if msg, exists := context["message"].(string); exists {
					handleCrossPlatformMessage(platform, msg)
				}
			}
		case "webhook":
			// Trigger webhook
			if event, ok := action.Config["event"].(string); ok {
				publishToWebhook(event, context)
			}
		case "broadcast":
			// Send to multiple contacts
			if msg, ok := action.Config["message"].(string); ok {
				if contacts, ok := action.Config["contacts"].([]interface{}); ok {
					for _, contact := range contacts {
						if jid, ok := contact.(string); ok {
							sendMessage(jid, msg)
						}
					}
				}
			}
		}
	}
}

func handleAIRequest(jid, message string) {
	prompt := strings.TrimPrefix(message, "/ask")
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		sendMessage(jid, "Please provide a question after /ask")
		return
	}

	// Call AI service
	response, err := queryAI(prompt)
	if err != nil {
		logger.Error("AI request failed", "error", err, "jid", jid)
		sendMessage(jid, "Sorry, I couldn't process your request right now.")
		return
	}

	// Send response
	sendMessage(jid, fmt.Sprintf("🤖 %s", response))

	// Update conversation context
	updateConversationContext(jid, "user", prompt)
	updateConversationContext(jid, "assistant", response)
}

func updateConversationContext(jid, role, content string) {
	if conversations[jid] == nil {
		conversations[jid] = &ConversationContext{
			JID:         jid,
			Messages:    []ContextMessage{},
			Context:     make(map[string]interface{}),
			LastUpdated: time.Now(),
		}
	}

	ctx := conversations[jid]
	ctx.Messages = append(ctx.Messages, ContextMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})

	// Keep only last 10 messages
	if len(ctx.Messages) > 10 {
		ctx.Messages = ctx.Messages[len(ctx.Messages)-10:]
	}

	ctx.LastUpdated = time.Now()
}

func loadAutomationRules() []AutomationRule {
	// Load from config file
	return []AutomationRule{}
}

func saveAutomationRules(rules []AutomationRule) {
	// Save to config file
}

func loadConversationContexts() map[string]*ConversationContext {
	return make(map[string]*ConversationContext)
}

func saveConversationContexts() {
	// Save contexts to persistent storage
}
