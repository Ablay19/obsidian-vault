package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	. "obsidian-automation/internal/ssh/models"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("ssh_users.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect SSH user database: %s", err))
	}

	// AutoMigrate will create/update tables for the User model
	DB.AutoMigrate(&User{})
}

func GenerateKeyPair(username string) (privateKey []byte, err error) {
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
	if err := DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return privateKey, nil
}
