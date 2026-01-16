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

	session, err := wac.Login(qrChan)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	// Save session
	saveSession(session)
	fmt.Println("Login successful!")
}

func handleSend() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: whatsapp-cli send <jid> <message>")
		os.Exit(1)
	}

	jid := os.Args[2]
	message := strings.Join(os.Args[3:], " ")

	if wac == nil {
		loadSession()
	}

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
		loadSession()
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
	}
	// Remove session file
	os.Remove("session.gob")
	fmt.Println("Logged out successfully!")
}

type messageHandler struct{}

func (messageHandler) HandleError(err error) {
	log.Printf("Error: %v", err)
}

func (messageHandler) HandleTextMessage(message whatsapp.TextMessage) {
	fmt.Printf("Received from %s: %s\n", message.Info.RemoteJid, message.Text)
}
