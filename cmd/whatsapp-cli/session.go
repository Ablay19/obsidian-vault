package main

import (
	"encoding/gob"
	whatsapp "github.com/Rhymen/go-whatsapp"
	"log"
	"os"
)

func saveSession(session whatsapp.Session) {
	file, err := os.Create("session.gob")
	if err != nil {
		log.Printf("Failed to create session file: %v", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		log.Printf("Failed to encode session: %v", err)
	}
}

func loadSession() {
	file, err := os.Open("session.gob")
	if err != nil {
		log.Printf("Failed to open session file: %v", err)
		return
	}
	defer file.Close()

	var session whatsapp.Session
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		log.Printf("Failed to decode session: %v", err)
		return
	}

	wac, err = whatsapp.RestoreWithSession(session)
	if err != nil {
		log.Printf("Failed to restore session: %v", err)
	}
}
