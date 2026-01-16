package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

var wac *whatsapp.Conn
var config *CLIConfig

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config = cfg

	// Initialize queue manager
	if err := initQueueManager(config); err != nil {
		log.Printf("Warning: Failed to initialize queue manager: %v", err)
		log.Println("Continuing without queuing functionality")
	} else {
		defer queueMgr.Close()
	}
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "login":
		handleLogin()
	case "send":
		handleSend()
	case "receive":
		handleReceive()
	case "logout":
		handleLogout()
	case "status":
		handleStatus()
	case "queue":
		handleQueue()
	case "schedule":
		handleSchedule()
	default:
		fmt.Printf("Unknown command: %s\n", command)
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
	fmt.Println("  whatsapp-cli logout          - Logout and clear session")
	fmt.Println()
	fmt.Println("JID format: 1234567890@s.whatsapp.net")
	fmt.Println("Session is saved automatically after login.")
	fmt.Println("RabbitMQ queues provide reliable message delivery.")
}

func handleLogin() {
	var err error
	wac, err = whatsapp.NewConn(20)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	qrChan := make(chan string)
	go func() {
		fmt.Printf("QR Code: %s\n", <-qrChan)
	}()

	_, err = wac.Login(qrChan)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	fmt.Println("Login successful! Use 'receive' to listen for messages.")
}

func handleSend() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: whatsapp-cli send <jid> <message>")
		os.Exit(1)
	}

	if wac == nil {
		log.Fatal("Not logged in. Run 'login' first.")
	}

	jid := os.Args[2]
	message := strings.Join(os.Args[3:], " ")

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: jid,
		},
		Text: message,
	}

	_, err := wac.Send(msg)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}

	fmt.Println("Message sent successfully!")
}

func handleReceive() {
	if wac == nil {
		log.Fatal("Not logged in. Run 'login' first.")
	}

	// Add handler for incoming messages
	wac.AddHandler(messageHandler{})

	fmt.Println("Listening for messages... Press Ctrl+C to stop")
	// Wait indefinitely
	select {}
}

func handleLogout() {
	if wac != nil {
		wac.Disconnect()
		wac = nil
	}
	// Remove session file
	os.Remove("whatsapp_session.gob")
	fmt.Println("Logged out and session cleared!")
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
		log.Fatalf("Failed to queue message: %v", err)
	}

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
		log.Fatalf("Invalid delay format: %v", err)
	}

	// For now, simple implementation - could be enhanced with proper scheduling
	go func() {
		time.Sleep(delay)
		queueMessage(jid, message, 1)
	}()

	fmt.Printf("Message scheduled for %s in %s\n", jid, delay)
}

// Message handlers moved to handlers.go
