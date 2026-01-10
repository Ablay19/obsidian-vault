package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"obsidian-automation/internal/telemetry"
	"os"

	sqlite "github.com/glebarez/sqlite" // Use glebarez/sqlite
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	telemetry.Debug("InitDB called")
	var err error

	// Revert to local SQLite
	DB, err = gorm.Open(sqlite.Open("ssh_users.db"), &gorm.Config{})
	telemetry.Info("Connecting SSH DB to local SQLite: ssh_users.db")

	if err != nil {
		telemetry.Error("Failed to connect SSH user database: " + err.Error())
		os.Exit(1)
	}
	telemetry.Debug("SSH DB initialized")

	// AutoMigrate will create/update tables for the User model
	telemetry.Debug("Running AutoMigrate for User model")
	DB.AutoMigrate(&User{})
	telemetry.Debug("AutoMigrate completed")
}

func GenerateKeyPair(username string) (privateKey []byte, err error) {
	telemetry.Debug("GenerateKeyPair called for " + username)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	privateKey = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	publicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	user := User{Username: username, PublicKey: string(ssh.MarshalAuthorizedKey(publicKey))}

	if DB == nil {

		telemetry.Error("GORM DB is nil in GenerateKeyPair!")

		return nil, fmt.Errorf("SSH database is not initialized")

	}

	var existingUser User

	result := DB.Where("username = ?", username).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {

		// User does not exist, create new

		if err := DB.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create SSH user: %w", err)
		}
		telemetry.Info("Created new SSH user: " + username)

	} else if result.Error != nil {

		// Other database error

		return nil, fmt.Errorf("database error checking for existing user: %w", result.Error)

	} else {

		// User exists, update public key

		if err := DB.Model(&existingUser).Update("public_key", user.PublicKey).Error; err != nil {
			return nil, fmt.Errorf("failed to update public key for existing user: %w", err)
		}
		telemetry.Info("Updated public key for existing SSH user: " + username)

	}

	return privateKey, nil
}
