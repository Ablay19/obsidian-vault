package main

import (
	"encoding/gob"
	"log"
	"os"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

var wac *whatsapp.Conn

func saveSession(session whatsapp.Session) {
	file, err := os.Create("whatsapp_session.gob")
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

func loadSession() (*whatsapp.Conn, error) {
	file, err := os.Open("whatsapp_session.gob")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var session whatsapp.Session
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return nil, err
	}

	conn, err := whatsapp.RestoreWithSession(session)
	return conn, err
}
