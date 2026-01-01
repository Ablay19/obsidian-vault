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

	if _, err := database.CheckExistingInstance(db); err != nil {
		log.Fatalf("Failed to check for existing instance: %v", err)
	}

	if err := database.AddInstance(db); err != nil {
		log.Fatalf("Failed to add instance to database: %v", err)
	}
	defer database.RemoveInstance(db)

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
		database.RemoveInstance(db)
		os.Exit(0)
	}()

	if err := bot.Run(db, aiService); err != nil {
		log.Fatal(err)
	}
}
