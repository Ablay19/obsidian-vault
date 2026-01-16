package main

import (
	"fmt"
	"log"
	"strings"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

type messageHandler struct{}

func (messageHandler) HandleError(err error) {
	log.Printf("WhatsApp error: %v", err)
}

func (messageHandler) HandleTextMessage(message whatsapp.TextMessage) {
	sender := message.Info.RemoteJid
	if strings.HasSuffix(sender, "@s.whatsapp.net") {
		sender = strings.TrimSuffix(sender, "@s.whatsapp.net")
	}

	fmt.Printf("📱 [%s]: %s\n", sender, message.Text)

	// Auto-reply for demo
	if strings.ToLower(message.Text) == "ping" {
		reply := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: message.Info.RemoteJid,
			},
			Text: "Pong! WhatsApp CLI is working.",
		}
		wac.Send(reply)
	}
}

func (messageHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	sender := strings.TrimSuffix(message.Info.RemoteJid, "@s.whatsapp.net")
	fmt.Printf("📷 Image received from %s\n", sender)
}

func (messageHandler) HandleDocumentMessage(message whatsapp.DocumentMessage) {
	sender := strings.TrimSuffix(message.Info.RemoteJid, "@s.whatsapp.net")
	fmt.Printf("📄 Document received from %s: %s\n", sender, message.FileName)
}
