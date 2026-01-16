package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Rhymen/go-whatsapp"
)

type MessageTemplate struct {
	Name    string            `json:"name"`
	Content string            `json:"content"`
	Vars    map[string]string `json:"vars,omitempty"`
}

type BulkMessage struct {
	JID      string `json:"jid"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

func loadTemplates() ([]MessageTemplate, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	templateFile := filepath.Join(configDir, "whatsapp-cli", "templates.json")

	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		// Create default templates
		defaultTemplates := []MessageTemplate{
			{
				Name:    "welcome",
				Content: "Welcome {{name}}! Thank you for joining our service.",
				Vars:    map[string]string{"name": "Customer"},
			},
			{
				Name:    "reminder",
				Content: "Hi {{name}}, this is a reminder about: {{topic}}",
				Vars:    map[string]string{"name": "User", "topic": "your appointment"},
			},
			{
				Name:    "support",
				Content: "Hello! How can we help you today? Our support team is here to assist.",
			},
		}
		data, _ := json.MarshalIndent(defaultTemplates, "", "  ")
		os.MkdirAll(filepath.Dir(templateFile), 0755)
		ioutil.WriteFile(templateFile, data, 0644)
		return defaultTemplates, nil
	}

	data, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}

	var templates []MessageTemplate
	json.Unmarshal(data, &templates)
	return templates, nil
}

func saveTemplates(templates []MessageTemplate) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	templateFile := filepath.Join(configDir, "whatsapp-cli", "templates.json")

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(templateFile), 0755)
	return ioutil.WriteFile(templateFile, data, 0644)
}

func renderTemplate(template MessageTemplate, customVars map[string]string) string {
	content := template.Content

	// Merge template vars with custom vars
	vars := make(map[string]string)
	for k, v := range template.Vars {
		vars[k] = v
	}
	for k, v := range customVars {
		vars[k] = v
	}

	// Replace variables
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	return content
}

func loadBulkMessages() ([]BulkMessage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	bulkFile := filepath.Join(configDir, "whatsapp-cli", "bulk_messages.json")

	if _, err := os.Stat(bulkFile); os.IsNotExist(err) {
		return []BulkMessage{}, nil
	}

	data, err := ioutil.ReadFile(bulkFile)
	if err != nil {
		return nil, err
	}

	var messages []BulkMessage
	json.Unmarshal(data, &messages)
	return messages, nil
}

func saveBulkMessages(messages []BulkMessage) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	bulkFile := filepath.Join(configDir, "whatsapp-cli", "bulk_messages.json")

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(bulkFile), 0755)
	return ioutil.WriteFile(bulkFile, data, 0644)
}

func sendRichMessage(jid, text string, formatting map[string]string) error {
	// Apply basic formatting (bold, italic, etc.)
	formattedText := text

	// Simple markdown-like formatting
	formattedText = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(formattedText, "*$1*") // Bold to italic as approximation
	formattedText = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(formattedText, "_$1_")     // Italic to underline

	return sendMessage(jid, formattedText)
}

func sendMessage(jid, text string) error {
	if wac == nil {
		conn, err := loadSession()
		if err != nil {
			return fmt.Errorf("no session available: %v", err)
		}
		wac = conn
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: jid,
		},
		Text: text,
	}

	_, err := wac.Send(msg)
	return err
}
