package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/bot"
	"obsidian-automation/internal/config"
	"obsidian-automation/internal/dashboard"
	"obsidian-automation/internal/database"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.LoadConfig()

	db := database.OpenDB()
	defer db.Close()

	database.ApplySchemaAndMigrations(db)

	ctx := context.Background()
	aiService := ai.NewAIService(ctx)

	router := http.NewServeMux()
	bot.StartHealthServer(router)
	dash := dashboard.NewDashboard(aiService, db)
	dash.RegisterRoutes(router)

	dashboardPort := config.AppConfig.Dashboard.Port
	go func() {
		addr := fmt.Sprintf(":%d", dashboardPort)
		log.Printf("Starting HTTP server on %s for health and dashboard...", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// database.RemoveInstance(db) // Removed as PID lock is no longer used
		os.Exit(0)
	}()

	if err := bot.Run(db, aiService); err != nil {
		log.Fatal(err)
	}
}
