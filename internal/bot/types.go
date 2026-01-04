package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot interfaces for mocking
type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

type UserState struct {
	Language          string
	LastProcessedFile string
	LastCreatedNote   string
	PendingFile       string
	PendingFileType   string
	PendingContext    string
	IsStaging         bool
}

// CommandHandler interface for command handling
type CommandHandler interface {
	Handle(bot Bot, message *tgbotapi.Message, state *UserState)
}
