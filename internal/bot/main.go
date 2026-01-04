package bot

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/pipeline/sources"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/state" // Import the new state package
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/config" // Import app config
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package-level variables for dependencies
var (
	db         *sql.DB
	aiService  *ai.AIService
	rcm        *state.RuntimeConfigManager // Add package-level RCM
	ingestionPipeline *pipeline.Pipeline
	gitManager *git.Manager
	wsManager  *ws.Manager
	userStates = make(map[int64]*UserState)
	stateMutex sync.RWMutex
	userLocks  sync.Map // Per-user processing lock
)

type UserState struct {
	Language          string
	LastProcessedFile string
	LastCreatedNote   string
	PendingFile       string
	PendingFileType   string
	PendingContext    string
	IsStaging         bool
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
func (t *TelegramBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return t.BotAPI.Request(c)
}
func (t *TelegramBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return t.BotAPI.GetFile(config)
}

// Run initializes and starts the bot.
func Run(database *sql.DB, ais *ai.AIService, runtimeConfigManager *state.RuntimeConfigManager, wsm *ws.Manager) error {
	db = database
	aiService = ais
	rcm = runtimeConfigManager // Assign to package-level variable
	wsManager = wsm

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		fmt.Fprintf(os.Stderr, "FATAL: TELEGRAM_BOT_TOKEN is not set. Bot cannot start.\n")
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}
	bot := &TelegramBot{botAPI}

	// Initialize Git Manager
	gitCfg := config.AppConfig.Git
	gitManager = git.NewManager(gitCfg.VaultPath)
	if err := gitManager.ConfigureUser(gitCfg.UserName, gitCfg.UserEmail); err != nil {
		slog.Warn("Failed to configure Git user", "error", err)
	}
	if gitCfg.RemoteURL != "" {
		if err := gitManager.EnsureRemote(gitCfg.RemoteURL); err != nil {
			slog.Warn("Failed to ensure Git remote", "error", err)
		}
	}

	// Initialize Pipeline
	processor := NewBotProcessor(aiService)
	sink := NewBotSink(db, botAPI, gitManager)
	ingestionPipeline = pipeline.NewPipeline(3, 100, processor, sink) // 3 workers, buffer 100
	ingestionPipeline.Start(context.Background())
	defer ingestionPipeline.Stop()

	// Initialize WhatsApp Source
	waSource := sources.NewWhatsAppSource(wsManager)
	go func() {
		if err := waSource.Start(context.Background(), ingestionPipeline.GetJobChan()); err != nil {
			slog.Error("WhatsApp source failed to start", "error", err)
		}
	}()

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "help", Description: "Show help message"},
		{Command: "stats", Description: "Show usage statistics"},
		{Command: "lang", Description: "Set AI language"},
		{Command: "last", Description: "Show last created note"},
		{Command: "reprocess", Description: "Reprocess last sent file"},
		{Command: "pid", Description: "Show the process ID of the bot instance"},
		{Command: "setprovider", Description: "Set AI provider (Gemini, Groq)"},
		{Command: "link", Description: "Link Telegram to Dashboard account"},
		{Command: "service_status", Description: "Show service status"},
		{Command: "pause_bot", Description: "Pause the bot"},
		{Command: "resume_bot", Description: "Resume the bot"},
	}
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		slog.Error("Error setting bot commands", "error", err)
	}

	slog.Info("Authorized on account", "username", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	slog.Info("Bot is running...")

	for update := range updates {
		if status.IsPaused() {
			time.Sleep(1 * time.Second)
			continue
		}
		slog.Debug("Received update", "update_id", update.UpdateID)
		go handleUpdate(bot, &update, token)
	}
	return nil
}

func handleUpdate(bot Bot, update *tgbotapi.Update, token string) {
	if update.CallbackQuery != nil {
		slog.Info("Handling callback query", "chat_id", update.CallbackQuery.Message.Chat.ID, "data", update.CallbackQuery.Data)
		handleCallbackQuery(bot, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	// For messages, we only lock if it's NOT a command. 
	// Commands like /setprovider should always be allowed.
	if !update.Message.IsCommand() {
		userID := update.Message.From.ID
		if _, loaded := userLocks.LoadOrStore(userID, true); loaded {
			slog.Warn("User is already being processed, skipping duplicate content update", "user_id", userID)
			return
		}
		defer userLocks.Delete(userID)
	}

	slog.Info("Handling message", "chat_id", update.Message.Chat.ID, "user", update.Message.From.UserName, "text", update.Message.Text)

	// Save incoming message to history
	contentType := "text"
	if update.Message.Photo != nil {
		contentType = "image"
	} else if update.Message.Document != nil {
		contentType = "pdf"
	}
	database.SaveMessage(update.Message.From.ID, update.Message.Chat.ID, update.Message.MessageID, "in", contentType, update.Message.Text, "")

	// User Mapping Protocol: Email-based
	// Check for email in message text to link accounts
	emailRegex := regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	email := emailRegex.FindString(strings.ToLower(update.Message.Text))
	if email != "" {
		slog.Info("Attempting to map user via email", "email", email, "telegram_id", update.Message.From.ID)
		err := database.LinkTelegramToEmail(update.Message.From.ID, email)
		if err != nil {
			slog.Error("Failed to link user", "error", err)
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Your Telegram account has been linked to your Dashboard account ("+email+")."))
		}
	}

	if update.Message.Photo != nil {
		handlePhoto(bot, update.Message, token)
	} else if update.Message.Document != nil {
		handleDocument(bot, update.Message, token)
	} else if update.Message.IsCommand() || update.Message.Text != "" {
		handleCommand(bot, update.Message)
	}
}

func handleCallbackQuery(bot Bot, callback *tgbotapi.CallbackQuery) {
	// Always answer the callback query to stop the loading spinner
	defer bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	if callback.Data == "refresh_providers" {
		// Re-trigger handleCommand for setprovider to refresh the list
		msg := callback.Message
		msg.From = callback.From // Crucial: preserve the user who clicked
		msg.Text = "/setprovider"
		if handler, ok := commandRouter["setprovider"]; ok {
			handler.Handle(bot, msg, getUserState(msg.From.ID)) // Pass the state
		}
		return
	}
	if strings.HasPrefix(callback.Data, "setprovider:") {
		providerName := strings.TrimPrefix(callback.Data, "setprovider:")
		if err := aiService.SetProvider(providerName); err != nil {
			bot.Send(tgbotapi.NewMessage(callback.Message.Chat.ID, fmt.Sprintf("❌ Failed to set provider: %v", err)))
		} else {
			editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, fmt.Sprintf("✅ AI provider has been set to: *%s*", providerName))
			editMsg.ParseMode = "Markdown"
			bot.Send(editMsg)
		}
	}
}

// handleCommand processes text messages and commands.
func handleCommand(bot Bot, message *tgbotapi.Message, commandRouter map[string]CommandHandler) {
	slog.Info("Processing command/text", "chat_id", message.Chat.ID, "text", message.Text)
	database.UpsertUser(message.From)
	state := getUserState(message.From.ID) // Get user state for language

	if !message.IsCommand() {
		// New Staging Logic
		if state.IsStaging {
			if state.PendingContext != "" {
				state.PendingContext += "\n"
			}
			state.PendingContext += message.Text
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "📝 Context added. Add more or type /process."))
			return
		}

		slog.Info("Handling non-command text as AI prompt", "chat_id", message.Chat.ID, "text_len", len(message.Text))
		// Handle non-command text messages as a general AI prompt
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Thinking..."))

		var responseText strings.Builder
		// writer := &responseText // No longer using io.Writer

		systemPrompt := fmt.Sprintf("Respond in %s. Output your response as valid HTML, with proper headings, paragraphs, and LaTeX formulas using MathJax syntax.", state.Language)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		req := &ai.RequestModel{
			SystemPrompt: systemPrompt,
			UserPrompt:   message.Text,
			Temperature:  0.7,
		}

		err := aiService.Chat(ctx, req, func(chunk string) {
			responseText.WriteString(chunk)
		})

		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Sorry, I had trouble thinking: "+err.Error()))
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, responseText.String())
		msg.ParseMode = tgbotapi.ModeHTML
		sentMsg, err := bot.Send(msg)
		if err == nil {
			database.SaveMessage(message.From.ID, message.Chat.ID, sentMsg.MessageID, "out", "text", responseText.String(), "")
		}
		return
	}

	if handler, ok := commandRouter[message.Command()]; ok {
		handler.Handle(bot, message, state)
	} else {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Use /help to see available commands."))
	}
}


