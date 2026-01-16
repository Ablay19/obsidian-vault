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
	case "logout":
		handleLogout()
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
	fmt.Println("  whatsapp-cli logout          - Logout and clear session")
	fmt.Println()
	fmt.Println("JID format: 1234567890@s.whatsapp.net")
	fmt.Println("Session is saved automatically after login.")
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

type WAHandler struct{}

func (h WAHandler) HandleError(err error) {
	logger.Error("WhatsApp error", "error", err)
}

func (h WAHandler) HandleTextMessage(msg whatsapp.TextMessage) {
	logger.Info("Received text message", "from", msg.Info.RemoteJid, "text", msg.Text)

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

	if strings.HasPrefix(msg.Text, "/ask") {
		prompt := strings.TrimPrefix(msg.Text, "/ask ")
		logger.Info("Processing AI ask command", "prompt", prompt)

		if config.AI.Enabled && queueMgr != nil {
			// Queue for AI processing
			aiRequest := map[string]interface{}{
				"jid":   msg.Info.RemoteJid,
				"query": prompt,
				"model": config.AI.Models[0],
			}
			data, _ := json.Marshal(aiRequest)
			queueMgr.PublishMessage(config.AI.Queue, string(data))
			logger.Info("Queued AI request", "jid", msg.Info.RemoteJid)
		} else {
			logger.Warn("AI not available or disabled")
		}
	} else if strings.ToLower(msg.Text) == "ping" {
		logger.Info("Responding to ping", "jid", msg.Info.RemoteJid)
	}
}

func (h WAHandler) HandleImageMessage(msg whatsapp.ImageMessage) {
	sender := strings.TrimSuffix(msg.Info.RemoteJid, "@s.whatsapp.net")
	logger.Info("Image received", "from", sender, "caption", msg.Caption)

	// Queue for media processing
	if queueMgr != nil {
		mediaData := map[string]interface{}{
			"type":     "image",
			"sender":   msg.Info.RemoteJid,
			"caption":  msg.Caption,
			"filename": "image.jpg",
		}
		data, _ := json.Marshal(mediaData)
		queueMgr.PublishMessage(config.RabbitMQ.Queues.Media, string(data))
	}
}

func (h WAHandler) HandleDocumentMessage(msg whatsapp.DocumentMessage) {
	sender := strings.TrimSuffix(msg.Info.RemoteJid, "@s.whatsapp.net")
	logger.Info("Document received", "from", sender, "filename", msg.FileName)

	// Queue for media processing
	if queueMgr != nil {
		mediaData := map[string]interface{}{
			"type":     "document",
			"sender":   msg.Info.RemoteJid,
			"filename": msg.FileName,
		}
		data, _ := json.Marshal(mediaData)
		queueMgr.PublishMessage(config.RabbitMQ.Queues.Media, string(data))
	}
}
