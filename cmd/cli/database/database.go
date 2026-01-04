package database

import (
	"fmt"

	"github.com/spf13/viper"
	sqlite "github.com/glebarez/sqlite" // Use glebarez/sqlite
	"gorm.io/gorm"
	. "obsidian-automation/cmd/cli/database/models"
)

var DB *gorm.DB

func Init() {
	dbType := viper.GetString("database.type")
	dbPath := viper.GetString("database.path")

	var err error
	if dbType == "sqlite" {
		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			panic(fmt.Errorf("failed to connect database: %s", err))
		}
		DB.AutoMigrate(&User{})
	} else {
		panic(fmt.Errorf("unsupported database type: %s", dbType))
	}
}
