package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"obsidian-automation/internal/config"
)

type SSHManager struct {
	logger    *zap.Logger
	webServer *gin.Engine
	apiServer *http.Server
}

func NewSSHManager() (*SSHManager, error) {
	logger, _ := zap.NewProduction()

	// Create web server for SSH management API
	webServer := gin.New()
	webServer.Use(gin.Logger(), gin.Recovery())

	// Setup routes for SSH management
	setupSSHRoutes(webServer, logger)

	return &SSHManager{
		logger:    logger,
		webServer: webServer,
	}, nil
}

func setupSSHRoutes(r *gin.Engine, logger *zap.Logger) {
	// SSH management API routes
	r.GET("/api/ssh/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "running",
			"message":   "SSH Management API is active",
			"timestamp": time.Now().UTC(),
		})
	})

	r.GET("/api/ssh/users", func(c *gin.Context) {
		// Return mock SSH users for now
		users := []gin.H{
			{
				"username":   "admin",
				"last_login": time.Now().Add(-2 * time.Hour),
				"active":     true,
			},
			{
				"username":   "developer",
				"last_login": time.Now().Add(-24 * time.Hour),
				"active":     false,
			},
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	})

	r.POST("/api/ssh/restart", func(c *gin.Context) {
		logger.Info("SSH service restart requested")
		c.JSON(http.StatusOK, gin.H{
			"message":   "SSH service restart initiated",
			"timestamp": time.Now().UTC(),
		})
	})

	r.GET("/api/ssh/logs", func(c *gin.Context) {
		// Return mock logs for now
		logs := []string{
			"[2024-01-07 10:15:30] SSH server started on port 2222",
			"[2024-01-07 10:16:45] User admin connected from 192.168.1.100",
			"[2024-01-07 10:20:12] User admin disconnected",
			"[2024-01-07 10:25:18] SSH health check passed",
		}
		c.JSON(http.StatusOK, gin.H{"logs": logs})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "ssh-manager",
			"timestamp": time.Now().UTC(),
		})
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", gin.H{
			"title":   "SSH Management Service",
			"message": "SSH Management API is running",
			"endpoints": []string{
				"/api/ssh/status",
				"/api/ssh/users",
				"/api/ssh/restart",
				"/api/ssh/logs",
				"/health",
			},
		})
	})
}

func (s *SSHManager) Start() error {
	// Start API server
	s.apiServer = &http.Server{
		Addr:         ":8081",
		Handler:      s.webServer,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		s.logger.Info("Starting SSH Management API server on :8081")
		if err := s.apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API server error", zap.Error(err))
		}
	}()

	s.logger.Info("SSH Management Service started successfully")
	s.logger.Info("SSH Management API listening on :8081")
	s.logger.Info("SSH daemon would be running on :2222 (simulated)")

	return nil
}

func (s *SSHManager) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down SSH Management Service...")

	// Shutdown API server
	if s.apiServer != nil {
		if err := s.apiServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down API server", zap.Error(err))
		} else {
			s.logger.Info("API server shutdown complete")
		}
	}

	s.logger.Info("SSH Management Service shutdown complete")
	return nil
}

func setupGracefulShutdown(manager *SSHManager) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		manager.logger.Info("Received shutdown signal")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := manager.Shutdown(ctx); err != nil {
			manager.logger.Error("Error during shutdown", zap.Error(err))
		}
		os.Exit(0)
	}()
}

func main() {
	// Load configuration
	config.LoadConfig()

	// Create and start SSH Manager
	manager, err := NewSSHManager()
	if err != nil {
		log.Fatalf("Failed to create SSH Manager: %v", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown(manager)

	// Start server
	if err := manager.Start(); err != nil {
		log.Fatalf("Failed to start SSH Manager: %v", err)
	}

	// Wait for shutdown signal
	select {}
}
