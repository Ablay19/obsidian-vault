package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

var config *CLIConfig
var queueMgr *QueueManager
var logger *slog.Logger
var wac *whatsapp.Conn

func main() {
	// Set default logger to JSON format
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("Starting WhatsApp CLI")

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	config = cfg

	// Initialize queue manager
	if err := initQueueManager(config); err != nil {
		logger.Warn("Failed to initialize queue manager", "error", err)
		logger.Info("Continuing without queuing functionality")
	}

	// Initialize automation system
	initAutomation()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	logger.Info("Processing command", "command", command)

	switch command {
	case "login":
		handleLogin()
	case "send":
		handleSend()
	case "receive":
		handleReceive()
	case "status":
		handleStatus()
	case "queue":
		handleQueue()
	case "schedule":
		handleSchedule()
	case "telegram":
		handleTelegram()
	case "ai":
		handleAI()
	case "services":
		handleServices()
	case "media":
		handleMedia()
	case "template":
		handleTemplate()
	case "bulk":
		handleBulk()
	case "automation":
		handleAutomation(os.Args[1:])
	default:
		logger.Error("Unknown command", "command", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("WhatsApp CLI Tool with RabbitMQ Queuing")
	fmt.Println("Usage:")
	fmt.Println("  whatsapp-cli login           - Login with QR code")
	fmt.Println("  whatsapp-cli send <jid> <msg> - Send message")
	fmt.Println("  whatsapp-cli queue <jid> <msg> - Queue message for sending")
	fmt.Println("  whatsapp-cli receive         - Listen for messages")
	fmt.Println("  whatsapp-cli status          - Check connection status")
	fmt.Println("  whatsapp-cli schedule <jid> <msg> <delay> - Schedule delayed message")
	fmt.Println("  whatsapp-cli telegram <cmd>  - Interact with Telegram bot")
	fmt.Println("  whatsapp-cli ai <cmd>        - Direct AI interactions")
	fmt.Println("  whatsapp-cli services <cmd>  - Manage project services")
	fmt.Println("  whatsapp-cli media <cmd>     - Media processing operations")
	fmt.Println("  whatsapp-cli template <cmd>  - Message template management")
	fmt.Println("  whatsapp-cli bulk <cmd>      - Bulk messaging operations")
	fmt.Println("  whatsapp-cli automation <cmd> - Automation rule management")
	fmt.Println("  whatsapp-cli logout          - Logout and clear session")
	fmt.Println()
	fmt.Println("JID format: 1234567890@s.whatsapp.net")
	fmt.Println("Session is saved automatically after login.")
	fmt.Println("Advanced features: templates, bulk messaging, rich formatting.")
	fmt.Println("Integrated with all project services for unified control.")
}

func handleLogin() {
	logger.Info("Starting WhatsApp login process")

	var err error
	wac, err = whatsapp.NewConn(20)
	if err != nil {
		logger.Error("Failed to create WhatsApp connection", "error", err)
		os.Exit(1)
	}

	qrChan := make(chan string)
	go func() {
		qr := <-qrChan
		logger.Info("QR code generated", "qr", qr)
		fmt.Printf("QR Code: %s\n", qr)
		fmt.Println("Scan this QR code with WhatsApp on your phone.")
	}()

	session, err := wac.Login(qrChan)
	if err != nil {
		logger.Error("Failed to login to WhatsApp", "error", err)
		os.Exit(1)
	}

	// Save session
	saveSession(session)
	logger.Info("Login successful", "session_saved", true)
	fmt.Println("Login successful!")
}

func handleSend() {
	logger.Info("Processing send command", "args_count", len(os.Args))

	if len(os.Args) < 4 {
		logger.Error("Insufficient arguments for send command")
		fmt.Println("Usage: whatsapp-cli send <jid> <message>")
		os.Exit(1)
	}

	jid := os.Args[2]
	message := strings.Join(os.Args[3:], " ")

	logger.Info("Sending message", "jid", jid, "message_length", len(message))

	// Load session if not connected
	conn, err := loadSession()
	if err != nil {
		logger.Error("No saved session available", "error", err)
		fmt.Println("No saved session. Please run 'login' first.")
		os.Exit(1)
	}
	wac = conn

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: jid,
		},
		Text: message,
	}

	_, err = wac.Send(msg)
	if err != nil {
		logger.Error("Failed to send message", "error", err, "jid", jid)
		os.Exit(1)
	}

	logger.Info("Message sent successfully", "jid", jid)
	fmt.Println("Message sent successfully!")
}

func handleReceive() {
	logger.Info("Starting message receiver")

	// Load session if not connected
	conn, err := loadSession()
	if err != nil {
		logger.Error("No saved session available for receiving", "error", err)
		fmt.Println("No saved session. Please run 'login' first.")
		os.Exit(1)
	}
	wac = conn

	// Add handler for incoming messages
	wac.AddHandler(WAHandler{})

	logger.Info("Message receiver started, waiting for messages")
	fmt.Println("Listening for messages... Press Ctrl+C to stop")

	// Wait indefinitely
	select {}
}

func handleStatus() {
	if _, err := os.Stat(config.WhatsApp.SessionFile); os.IsNotExist(err) {
		fmt.Println("Status: Not logged in (no session file)")
		return
	}

	if wac == nil {
		conn, err := loadSession()
		if err != nil {
			fmt.Printf("Status: Session exists but failed to load: %v\n", err)
			return
		}
		wac = conn
	}

	queueStatus := "without queuing"
	if queueMgr != nil {
		queueStatus = "with queuing"
	}

	fmt.Printf("Status: Connected and ready (%s)\n", queueStatus)
	fmt.Printf("RabbitMQ: %s\n", func() string {
		if queueMgr != nil {
			return "connected"
		}
		return "disconnected"
	}())
}

func handleQueue() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: whatsapp-cli queue <jid> <message>")
		os.Exit(1)
	}

	if queueMgr == nil {
		fmt.Println("Error: Queue manager not available. Check RabbitMQ connection.")
		os.Exit(1)
	}

	jid := os.Args[2]
	message := strings.Join(os.Args[3:], " ")

	err := queueMessage(jid, message, 1)
	if err != nil {
		logger.Error("Failed to queue message", "error", err)
		os.Exit(1)
	}

	logger.Info("Message queued successfully", "jid", jid)
	fmt.Println("Message queued successfully for", jid)
}

func handleSchedule() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: whatsapp-cli schedule <jid> <message> <delay>")
		fmt.Println("Delay format: 30s, 5m, 1h, etc.")
		os.Exit(1)
	}

	jid := os.Args[2]
	message := os.Args[3]
	delayStr := os.Args[4]

	delay, err := time.ParseDuration(delayStr)
	if err != nil {
		logger.Error("Invalid delay format", "error", err)
		os.Exit(1)
	}

	// For now, simple implementation - could be enhanced with proper scheduling
	go func() {
		time.Sleep(delay)
		queueMessage(jid, message, 1)
	}()

	logger.Info("Message scheduled", "jid", jid, "delay", delay.String())
	fmt.Printf("Message scheduled for %s in %s\n", jid, delay)
}

func handleTelegram() {
	if len(os.Args) < 3 {
		fmt.Println("Telegram commands:")
		fmt.Println("  whatsapp-cli telegram send <chat_id> <message>  - Send to Telegram")
		fmt.Println("  whatsapp-cli telegram status                  - Check Telegram bot status")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "send":
		if len(os.Args) < 5 {
			fmt.Println("Usage: whatsapp-cli telegram send <chat_id> <message>")
			return
		}
		chatID := os.Args[3]
		message := strings.Join(os.Args[4:], " ")
		err := sendToTelegram(chatID, message)
		if err != nil {
			logger.Error("Failed to send to Telegram", "error", err)
			fmt.Printf("Failed to send to Telegram: %v\n", err)
		} else {
			fmt.Println("Message sent to Telegram")
		}
	case "status":
		status, err := getTelegramStatus()
		if err != nil {
			logger.Error("Failed to get Telegram status", "error", err)
			fmt.Printf("Failed to get Telegram status: %v\n", err)
		} else {
			fmt.Printf("Telegram Status: %s\n", status)
		}
	default:
		logger.Error("Unknown telegram command", "command", subCmd)
		fmt.Printf("Unknown telegram command: %s\n", subCmd)
	}
}

func handleAI() {
	if len(os.Args) < 3 {
		fmt.Println("AI commands:")
		fmt.Println("  whatsapp-cli ai ask <prompt>     - Get AI response")
		fmt.Println("  whatsapp-cli ai models           - List available models")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "ask":
		if len(os.Args) < 4 {
			fmt.Println("Usage: whatsapp-cli ai ask <prompt>")
			return
		}
		prompt := strings.Join(os.Args[3:], " ")
		response, err := queryAI(prompt)
		if err != nil {
			logger.Error("AI query failed", "error", err)
			fmt.Printf("AI Error: %v\n", err)
		} else {
			fmt.Printf("AI Response: %s\n", response)
		}
	case "models":
		models, err := listAIModels()
		if err != nil {
			logger.Error("Failed to list AI models", "error", err)
			fmt.Printf("Failed to list models: %v\n", err)
		} else {
			fmt.Println("Available AI Models:")
			for _, model := range models {
				fmt.Printf("  - %s\n", model)
			}
		}
	default:
		logger.Error("Unknown AI command", "command", subCmd)
		fmt.Printf("Unknown AI command: %s\n", subCmd)
	}
}

func handleServices() {
	if len(os.Args) < 3 {
		fmt.Println("Service commands:")
		fmt.Println("  whatsapp-cli services list      - List all services")
		fmt.Println("  whatsapp-cli services status    - Show service health")
		fmt.Println("  whatsapp-cli services restart <name> - Restart a service")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "list":
		services, err := listServices()
		if err != nil {
			logger.Error("Failed to list services", "error", err)
			fmt.Printf("Failed to list services: %v\n", err)
		} else {
			fmt.Println("Project Services:")
			for _, svc := range services {
				fmt.Printf("  - %s (%s)\n", svc.Name, svc.Status)
			}
		}
	case "status":
		status, err := getServicesStatus()
		if err != nil {
			logger.Error("Failed to get services status", "error", err)
			fmt.Printf("Failed to get status: %v\n", err)
		} else {
			fmt.Printf("Services Status: %s\n", status)
		}
	case "restart":
		if len(os.Args) < 4 {
			fmt.Println("Usage: whatsapp-cli services restart <service_name>")
			return
		}
		serviceName := os.Args[3]
		err := restartService(serviceName)
		if err != nil {
			logger.Error("Failed to restart service", "service", serviceName, "error", err)
			fmt.Printf("Failed to restart service: %v\n", err)
		} else {
			logger.Info("Service restarted", "service", serviceName)
			fmt.Printf("Service %s restarted\n", serviceName)
		}
	default:
		logger.Error("Unknown services command", "command", subCmd)
		fmt.Printf("Unknown services command: %s\n", subCmd)
	}
}

func handleMedia() {
	if len(os.Args) < 3 {
		fmt.Println("Media commands:")
		fmt.Println("  whatsapp-cli media upload <file>   - Upload and process media")
		fmt.Println("  whatsapp-cli media status          - Check media processing status")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "upload":
		if len(os.Args) < 4 {
			fmt.Println("Usage: whatsapp-cli media upload <file_path>")
			return
		}
		filePath := os.Args[3]
		result, err := uploadMedia(filePath)
		if err != nil {
			logger.Error("Failed to upload media", "file", filePath, "error", err)
			fmt.Printf("Failed to upload media: %v\n", err)
		} else {
			logger.Info("Media uploaded successfully", "file", filePath)
			fmt.Printf("Media uploaded: %s\n", result)
		}
	case "status":
		status, err := getMediaStatus()
		if err != nil {
			logger.Error("Failed to get media status", "error", err)
			fmt.Printf("Failed to get media status: %v\n", err)
		} else {
			fmt.Printf("Media Status: %s\n", status)
		}
	default:
		logger.Error("Unknown media command", "command", subCmd)
		fmt.Printf("Unknown media command: %s\n", subCmd)
	}
}

func handleLogout() {
	logger.Info("Processing logout command")

	if wac != nil {
		wac.Disconnect()
		wac = nil
	}
	// Remove session file
	os.Remove(config.WhatsApp.SessionFile)
	logger.Info("Logged out and session cleared")
	fmt.Println("Logged out and session cleared!")
}

func handleAutomation(args []string) {
	if len(args) < 2 {
		fmt.Println("Automation commands:")
		fmt.Println("  whatsapp-cli automation list     - List automation rules")
		fmt.Println("  whatsapp-cli automation enable <id> - Enable rule")
		fmt.Println("  whatsapp-cli automation disable <id> - Disable rule")
		fmt.Println("  whatsapp-cli automation context <jid> - Show conversation context")
		return
	}

	subCmd := args[1]
	switch subCmd {
	case "list":
		fmt.Println("Automation Rules:")
		for _, rule := range rules {
			status := "disabled"
			if rule.Enabled {
				status = "enabled"
			}
			fmt.Printf("  - %s (%s): %s\n", rule.ID, status, rule.Description)
		}
	case "enable":
		if len(args) < 3 {
			fmt.Println("Usage: whatsapp-cli automation enable <rule_id>")
			return
		}
		ruleID := args[2]
		for i, rule := range rules {
			if rule.ID == ruleID {
				rules[i].Enabled = true
				fmt.Printf("Rule '%s' enabled\n", ruleID)
				saveAutomationRules(rules)
				return
			}
		}
		fmt.Printf("Rule '%s' not found\n", ruleID)
	case "disable":
		if len(args) < 3 {
			fmt.Println("Usage: whatsapp-cli automation disable <rule_id>")
			return
		}
		ruleID := args[2]
		for i, rule := range rules {
			if rule.ID == ruleID {
				rules[i].Enabled = false
				fmt.Printf("Rule '%s' disabled\n", ruleID)
				saveAutomationRules(rules)
				return
			}
		}
		fmt.Printf("Rule '%s' not found\n", ruleID)
	case "context":
		if len(args) < 3 {
			fmt.Println("Usage: whatsapp-cli automation context <jid>")
			return
		}
		jid := args[2]
		if ctx, exists := conversations[jid]; exists {
			fmt.Printf("Conversation context for %s:\n", jid)
			for _, msg := range ctx.Messages {
				fmt.Printf("  [%s] %s: %s\n", msg.Time.Format("15:04"), msg.Role, msg.Content)
			}
		} else {
			fmt.Printf("No conversation context found for %s\n", jid)
		}
	default:
		fmt.Printf("Unknown automation command: %s\n", subCmd)
	}
}

func handleTemplate() {
	if len(os.Args) < 3 {
		fmt.Println("Template commands:")
		fmt.Println("  whatsapp-cli template list         - List available templates")
		fmt.Println("  whatsapp-cli template use <name> <jid> [vars...] - Use template")
		fmt.Println("  whatsapp-cli template create <name> <content> - Create template")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "list":
		templates, err := loadTemplates()
		if err != nil {
			logger.Error("Failed to load templates", "error", err)
			fmt.Printf("Failed to load templates: %v\n", err)
			return
		}
		fmt.Println("Available Templates:")
		for _, tmpl := range templates {
			fmt.Printf("  - %s: %s\n", tmpl.Name, tmpl.Content)
		}
	case "use":
		if len(os.Args) < 5 {
			fmt.Println("Usage: whatsapp-cli template use <name> <jid> [key=value ...]")
			return
		}
		templateName := os.Args[3]
		jid := os.Args[4]

		// Parse custom variables
		customVars := make(map[string]string)
		for _, arg := range os.Args[5:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				customVars[parts[0]] = parts[1]
			}
		}

		templates, err := loadTemplates()
		if err != nil {
			logger.Error("Failed to load templates", "error", err)
			fmt.Printf("Failed to load templates: %v\n", err)
			return
		}

		var selectedTemplate *MessageTemplate
		for _, tmpl := range templates {
			if tmpl.Name == templateName {
				selectedTemplate = &tmpl
				break
			}
		}

		if selectedTemplate == nil {
			fmt.Printf("Template '%s' not found\n", templateName)
			return
		}

		message := renderTemplate(*selectedTemplate, customVars)
		err = sendMessage(jid, message)
		if err != nil {
			logger.Error("Failed to send template message", "error", err, "jid", jid)
			fmt.Printf("Failed to send message: %v\n", err)
		} else {
			logger.Info("Template message sent", "template", templateName, "jid", jid)
			fmt.Println("Template message sent successfully!")
		}
	case "create":
		if len(os.Args) < 5 {
			fmt.Println("Usage: whatsapp-cli template create <name> <content>")
			return
		}
		name := os.Args[3]
		content := strings.Join(os.Args[4:], " ")

		templates, err := loadTemplates()
		if err != nil {
			logger.Error("Failed to load templates", "error", err)
			return
		}

		newTemplate := MessageTemplate{
			Name:    name,
			Content: content,
		}
		templates = append(templates, newTemplate)

		err = saveTemplates(templates)
		if err != nil {
			logger.Error("Failed to save templates", "error", err)
			fmt.Printf("Failed to save template: %v\n", err)
		} else {
			logger.Info("Template created", "name", name)
			fmt.Printf("Template '%s' created successfully!\n", name)
		}
	default:
		logger.Error("Unknown template command", "command", subCmd)
		fmt.Printf("Unknown template command: %s\n", subCmd)
	}
}

func handleBulk() {
	if len(os.Args) < 3 {
		fmt.Println("Bulk messaging commands:")
		fmt.Println("  whatsapp-cli bulk list          - List queued bulk messages")
		fmt.Println("  whatsapp-cli bulk add <jid> <message> - Add message to bulk queue")
		fmt.Println("  whatsapp-cli bulk send          - Send all queued bulk messages")
		fmt.Println("  whatsapp-cli bulk clear         - Clear bulk message queue")
		return
	}

	subCmd := os.Args[2]
	switch subCmd {
	case "list":
		messages, err := loadBulkMessages()
		if err != nil {
			logger.Error("Failed to load bulk messages", "error", err)
			fmt.Printf("Failed to load bulk messages: %v\n", err)
			return
		}
		fmt.Printf("Queued Bulk Messages: %d\n", len(messages))
		for i, msg := range messages {
			fmt.Printf("  %d. %s -> %s\n", i+1, msg.JID, msg.Message)
		}
	case "add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: whatsapp-cli bulk add <jid> <message>")
			return
		}
		jid := os.Args[3]
		message := strings.Join(os.Args[4:], " ")

		messages, err := loadBulkMessages()
		if err != nil {
			logger.Error("Failed to load bulk messages", "error", err)
			return
		}

		newMessage := BulkMessage{
			JID:     jid,
			Message: message,
		}
		messages = append(messages, newMessage)

		err = saveBulkMessages(messages)
		if err != nil {
			logger.Error("Failed to save bulk messages", "error", err)
			fmt.Printf("Failed to save bulk message: %v\n", err)
		} else {
			logger.Info("Bulk message added", "jid", jid)
			fmt.Println("Bulk message added successfully!")
		}
	case "send":
		messages, err := loadBulkMessages()
		if err != nil {
			logger.Error("Failed to load bulk messages", "error", err)
			fmt.Printf("Failed to load bulk messages: %v\n", err)
			return
		}

		if len(messages) == 0 {
			fmt.Println("No bulk messages to send")
			return
		}

		sent := 0
		failed := 0

		for _, msg := range messages {
			err := sendMessage(msg.JID, msg.Message)
			if err != nil {
				logger.Error("Failed to send bulk message", "error", err, "jid", msg.JID)
				failed++
			} else {
				sent++
			}
			// Small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)
		}

		logger.Info("Bulk send completed", "sent", sent, "failed", failed)
		fmt.Printf("Bulk send completed: %d sent, %d failed\n", sent, failed)

		// Clear sent messages (in real implementation, might want to keep failed ones)
		if failed == 0 {
			saveBulkMessages([]BulkMessage{})
		}
	case "clear":
		err := saveBulkMessages([]BulkMessage{})
		if err != nil {
			logger.Error("Failed to clear bulk messages", "error", err)
			fmt.Printf("Failed to clear bulk messages: %v\n", err)
		} else {
			logger.Info("Bulk messages cleared")
			fmt.Println("Bulk messages cleared!")
		}
	default:
		logger.Error("Unknown bulk command", "command", subCmd)
		fmt.Printf("Unknown bulk command: %s\n", subCmd)
	}
}

type WAHandler struct{}

func (h WAHandler) HandleError(err error) {
	logger.Error("WhatsApp error", "error", err)
}

func (h WAHandler) HandleTextMessage(msg whatsapp.TextMessage) {
	logger.Info("Received text message", "from", msg.Info.RemoteJid, "text", msg.Text)

	// Process through automation system
	processIncomingMessage(msg.Info.RemoteJid, msg.Text)

	// Publish to incoming queue
	if queueMgr != nil {
		incomingData := map[string]interface{}{
			"jid":       msg.Info.RemoteJid,
			"message":   msg.Text,
			"timestamp": msg.Info.Timestamp,
		}
		data, _ := json.Marshal(incomingData)
		routingKey := fmt.Sprintf("%s.%s", config.RabbitMQ.Queues.Incoming, msg.Info.RemoteJid)
		queueMgr.PublishMessage(routingKey, string(data))
	}
}
