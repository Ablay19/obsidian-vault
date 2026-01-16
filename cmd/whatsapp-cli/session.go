package main

import (
	"encoding/gob"
	"log"
	"os"

	whatsapp "github.com/Rhymen/go-whatsapp"
)

func saveSession(session whatsapp.Session) {
	file, err := os.Create(config.WhatsApp.SessionFile)
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
	file, err := os.Open(config.WhatsApp.SessionFile)
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

	conn, err := whatsapp.NewConn(20)
	if err != nil {
		return nil, err
	}
	// Restore session
	_, err = conn.RestoreWithSession(session)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}
}
	defer file.Close()

	var session whatsapp.Session
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return nil, err
	}

	conn, err := whatsapp.NewConn(20)
	if err != nil {
		return nil, err
	}
	// Restore session
	_, err = conn.RestoreWithSession(session)
	if err != nil {
		return nil, err
	}
	return conn, nil
	return conn, err
}
