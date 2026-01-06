package ssh

import "gorm.io/gorm"

// User represents an SSH user in the system.
type User struct {
	gorm.Model
	Username  string `gorm:"unique;not null"`
	Password  string // Hashed password
	PublicKey string // Authorized public key
}
