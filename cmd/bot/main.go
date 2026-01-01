package main

import (
	"log"
	"obsidian-automation/internal/bot"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := bot.CreatePIDFile(); err != nil {
		log.Fatalf("Failed to create PID file: %v", err)
	}
	defer bot.RemovePIDFile()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		bot.RemovePIDFile()
		os.Exit(0)
	}()

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
