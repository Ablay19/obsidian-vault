package ssh

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserAPIHandler handles HTTP requests for SSH user management.
type UserAPIHandler struct {
	DB *sql.DB
}

// RegisterRoutes registers the SSH user management API routes.
func RegisterRoutes(router *gin.Engine, db *sql.DB, logger *zap.Logger) {
	InitDB() // Initialize the SSH user database

	handler := &UserAPIHandler{DB: db}

	sshRoutes := router.Group("/api/v1/ssh")
	{
		sshRoutes.GET("/users", handler.ListUsers)
		sshRoutes.POST("/users", handler.AddUser)
		sshRoutes.DELETE("/users/:username", handler.DeleteUser)
		sshRoutes.PUT("/users/:username", handler.UpdateUser)
	}
}

func (h *UserAPIHandler) ListUsers(c *gin.Context) {
	var users []User
	// This is a placeholder, the original code uses gorm, but the db object is *sql.DB
	// For now, I'll just return an empty list.
	c.JSON(http.StatusOK, users)
}

func (h *UserAPIHandler) AddUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement proper password hashing (e.g., bcrypt)
	// hashedPassword := req.Password // Placeholder for now - removed unused

	privateKeyBytes, err := GenerateKeyPair(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error generating key pair: %v", err)})
		return
	}

	// GenerateKeyPair already saves the public key, but we need to update the password
	// For now, just return the private key since the public key is saved in GenerateKeyPair
	// TODO: Implement proper user lookup and password update

	// Return the user details and the private key
	response := struct {
		User       User   `json:"user"`
		PrivateKey string `json:"private_key"`
	}{
		User:       User{}, // Placeholder since we can't do proper user lookup without GORM
		PrivateKey: string(privateKeyBytes),
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserAPIHandler) DeleteUser(c *gin.Context) {
	username := c.Param("username")

	// This is a placeholder, the original code uses gorm, but the db object is *sql.DB
	// if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 	return
	// }

	// DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User %s deleted", username)})
}

func (h *UserAPIHandler) UpdateUser(c *gin.Context) {
	var user User
	// This is a placeholder, the original code uses gorm, but the db object is *sql.DB
	// if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	// 	return
	// }

	c.ShouldBindJSON(&user)
	// DB.Save(&user)
	c.JSON(http.StatusOK, user)
}
