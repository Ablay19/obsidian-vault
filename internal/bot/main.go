package bot

import (
	"database/sql"
	"fmt"
	"log"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/status"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package-level variables for dependencies
var (
	db         *sql.DB
	aiService  *ai.AIService
	userStates = make(map[int64]*UserState)
	stateMutex sync.RWMutex
)

type UserState struct {
	Language          string
	LastProcessedFile string
	LastCreatedNote   string
}

func getUserState(userID int64) *UserState {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if state, exists := userStates[userID]; exists {
		return state
	}
	state := &UserState{Language: "English"}
	userStates[userID] = state
	return state
}

// Bot interfaces for mocking
type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

type TelegramBot struct {
	*tgbotapi.BotAPI
}

func (t *TelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) { return t.BotAPI.Send(c) }
func (t *TelegramBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) { return t.BotAPI.Request(c) }
func (t *TelegramBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) { return t.BotAPI.GetFile(config) }

// Run initializes and starts the bot.
func Run(database *sql.DB, ais *ai.AIService) error {
	db = database
	aiService = ais

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	bot := &TelegramBot{botAPI}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "help", Description: "Show help message"},
		{Command: "stats", Description: "Show usage statistics"},
		{Command: "lang", Description: "Set AI language"},
		{Command: "last", Description: "Show last created note"},
		{Command: "reprocess", Description: "Reprocess last sent file"},
		{Command: "setprovider", Description: "Set AI provider (Gemini, Groq)"},
		{Command: "service_status", Description: "Show service status"},
		{Command: "pause_bot", Description: "Pause the bot"},
		{Command: "resume_bot", Description: "Resume the bot"},
	}
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		log.Printf("Error setting bot commands: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	log.Println("Bot is running...")

	for update := range updates {
		if status.IsPaused() {
			time.Sleep(1 * time.Second)
			continue
		}
		if update.Message == nil {
			continue
		}
		if update.Message.Photo != nil {
			go handlePhoto(bot, update.Message, token)
		} else if update.Message.Document != nil {
			go handleDocument(bot, update.Message, token)
		} else if update.Message.IsCommand() || update.Message.Text != "" {
			go handleCommand(bot, update.Message)
		}
	}
	return nil
}

// handleCommand processes text messages and commands.
func handleCommand(bot Bot, message *tgbotapi.Message) {
	if !message.IsCommand() {
		// Handle non-command text messages as a general AI prompt
		return
	}

	switch message.Command() {
	case "start", "help":
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Send an image or PDF to start. Use /lang to set language, /stats for stats."))
	case "pause_bot":
		status.SetPaused(true)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is paused."))
	case "resume_bot":
		status.SetPaused(false)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is resumed."))
	case "service_status":
		statuses := status.GetServicesStatus(aiService, db)
		var sb strings.Builder
		sb.WriteString("📊 *Service Status*\n\n")
		for _, s := range statuses {
			sb.WriteString(fmt.Sprintf("• *%s:* %s\n", s.Name, s.Status))
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

// handlePhoto processes incoming photos.
func handlePhoto(bot Bot, message *tgbotapi.Message, token string) {
	status.UpdateActivity()
	statusMsg, _ := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Processing image..."))
	photo := message.Photo[len(message.Photo)-1]
	filename := downloadFile(bot, photo.FileID, "jpg", token)
	if filename != "" {
		createObsidianNote(filename, "image", message, bot, message.Chat.ID, statusMsg.MessageID)
	}
}

// handleDocument processes incoming documents.
func handleDocument(bot Bot, message *tgbotapi.Message, token string) {
	status.UpdateActivity()
	// ... document handling logic
}

// downloadFile downloads a file from Telegram.
func downloadFile(bot Bot, fileID, ext, token string) string {
	// ... download logic
	return ""
}

// createObsidianNote orchestrates the whole process.
func createObsidianNote(filePath, fileType string, message *tgbotapi.Message, bot Bot, chatID int64, messageID int) {
	// ... note creation logic using package-level aiService
}