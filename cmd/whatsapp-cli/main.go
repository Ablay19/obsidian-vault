package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

var wac *whatsapp.Conn

func main() {
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
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("WhatsApp CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  whatsapp-cli login")
	fmt.Println("  whatsapp-cli send <jid> <message>")
	fmt.Println("  whatsapp-cli receive")
	fmt.Println("  whatsapp-cli logout")
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

func handleStatus() {
	if _, err := os.Stat("whatsapp_session.gob"); os.IsNotExist(err) {
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

	fmt.Println("Status: Connected and ready")
}

// Message handlers moved to handlers.go
