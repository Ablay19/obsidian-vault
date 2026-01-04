package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"obsidian-automation/internal/telemetry" // Add telemetry import

	"golang.org/x/crypto/ssh"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	telemetry.ZapLogger.Sugar().Debug("InitDB called")
	var err error
	DB, err = gorm.Open(sqlite.Open("ssh_users.db"), &gorm.Config{})
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorf("Failed to connect SSH user database: %v", err)
		panic(fmt.Errorf("failed to connect SSH user database: %s", err))
	}
	telemetry.ZapLogger.Sugar().Debugf("SSH DB initialized: %p", DB)

	// AutoMigrate will create/update tables for the User model
	telemetry.ZapLogger.Sugar().Debug("Running AutoMigrate for User model")
	DB.AutoMigrate(&User{})
	telemetry.ZapLogger.Sugar().Debug("AutoMigrate completed")
}

func GenerateKeyPair(username string) (privateKey []byte, err error) {
	telemetry.ZapLogger.Sugar().Debugf("GenerateKeyPair called for %s, DB state: %p", username, DB)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	privateKey = pem.EncodeToMemory(
		&pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	publicKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

		user := User{Username: username, PublicKey: string(ssh.MarshalAuthorizedKey(publicKey))}

		if DB == nil {

			telemetry.ZapLogger.Sugar().Error("GORM DB is nil in GenerateKeyPair!")

			return nil, fmt.Errorf("SSH database is not initialized")

		}

	

		var existingUser User

		result := DB.Where("username = ?", username).First(&existingUser)

	

		if result.Error == gorm.ErrRecordNotFound {

			// User does not exist, create new

			if err := DB.Create(&user).Error; err != nil {

				return nil, fmt.Errorf("failed to create SSH user: %w", err)

			}

			telemetry.ZapLogger.Sugar().Infow("Created new SSH user", "username", username)

		} else if result.Error != nil {

			// Other database error

			return nil, fmt.Errorf("database error checking for existing user: %w", result.Error)

		} else {

			// User exists, update public key

			if err := DB.Model(&existingUser).Update("public_key", user.PublicKey).Error; err != nil {

				return nil, fmt.Errorf("failed to update public key for existing user: %w", err)

			}

			telemetry.ZapLogger.Sugar().Infow("Updated public key for existing SSH user", "username", username)

		}

		return privateKey, nil
}
