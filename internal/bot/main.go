package bot

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/config" // Import app config
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/state" // Import the new state package
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/telemetry"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package-level variables for dependencies
var (
	aiService         ai.AIServiceInterface
	rcm               *state.RuntimeConfigManager
	ingestionPipeline *pipeline.Pipeline
	gitManager        *git.Manager
	wsManager         *ws.Manager
	userStates        = make(map[int64]*UserState)
	stateMutex        sync.RWMutex
	userLocks         sync.Map // Per-user processing lock
)

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
func Run(db *sql.DB, ais ai.AIServiceInterface, runtimeConfigManager *state.RuntimeConfigManager, wsm *ws.Manager) error {
	// Check external binary dependencies first
	if err := ValidateBinaries(); err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Binary validation failed", "error", err)
		// Log the error but continue - bot will run with limited functionality
		telemetry.ZapLogger.Sugar().Warn("Bot will start with limited functionality. Some features may not work properly.")
	}
	aiService = ais
	rcm = runtimeConfigManager // Assign to package-level variable
	wsManager = wsm

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		telemetry.ZapLogger.Sugar().Fatal("TELEGRAM_BOT_TOKEN is not set. Bot cannot start.")
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
		telemetry.ZapLogger.Sugar().Warnw("Failed to configure Git user", "error", err)
	}
	if gitCfg.RemoteURL != "" {
		if err := gitManager.EnsureRemote(gitCfg.RemoteURL); err != nil {
			telemetry.ZapLogger.Sugar().Warnw("Failed to ensure Git remote", "error", err)
		}
	}

	// Initialize Pipeline
	processor := NewBotProcessor(aiService)
	sink := NewBotSink(database.Client.DB, botAPI, gitManager)        // Use database.Client.DB
	ingestionPipeline = pipeline.NewPipeline(3, 100, processor, sink) // 3 workers, buffer 100
	ingestionPipeline.Start(context.Background())
	defer ingestionPipeline.Stop()

	// Initialize Command Router
	commandRouter := NewCommandRouter(CommandDependencies{
		AIService:         aiService,
		RCM:               rcm,
		IngestionPipeline: ingestionPipeline,
		GitManager:        gitManager,
	})

	// Initialize State Machine and Command Handler Manager
	messageProcessor := NewMessageProcessor(aiService)
	stagingStateMachine := NewStagingStateMachine(messageProcessor)
	commandHandlerManager := &CommandHandlerManager{
		stateMachine:  stagingStateMachine,
		commandRouter: commandRouter,
	}
	fileHandler := NewFileHandler(stagingStateMachine)

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
		{Command: "modelinfo", Description: "Show AI model information"},
		{Command: "process", Description: "Process staged file"},
	}
	if _, err := bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Error setting bot commands", "error", err)
	}

	telemetry.ZapLogger.Sugar().Infow("Authorized on account", "username", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	telemetry.ZapLogger.Sugar().Info("Bot is running...")

	for update := range updates {
		if status.IsPaused() {
			time.Sleep(1 * time.Second)
			continue
		}
		telemetry.ZapLogger.Sugar().Debugw("Received update", "update_id", update.UpdateID)
		go handleUpdate(context.Background(), bot, &update, token, commandHandlerManager, fileHandler) // Pass context
	}
	return nil
}

// handleUpdate processes incoming Telegram updates.
func handleUpdate(ctx context.Context, bot Bot, update *tgbotapi.Update, token string, commandHandlerManager *CommandHandlerManager, fileHandler *FileHandler) {
	if update.CallbackQuery != nil {
		telemetry.ZapLogger.Sugar().Infow("Handling callback query", "chat_id", update.CallbackQuery.Message.Chat.ID, "data", update.CallbackQuery.Data)
		handleCallbackQuery(bot, update.CallbackQuery, commandHandlerManager.GetCommandRouter())
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
			telemetry.ZapLogger.Sugar().Warnw("User is already being processed, skipping duplicate content update", "user_id", userID)
			return
		}
		defer userLocks.Delete(userID)
	}

	telemetry.ZapLogger.Sugar().Infow("Handling message", "chat_id", update.Message.Chat.ID, "user", update.Message.From.UserName, "text", update.Message.Text)

	// Save incoming message to history
	contentType := "text"
	if update.Message.Photo != nil {
		contentType = "image"
	} else if update.Message.Document != nil {
		contentType = "pdf"
	}
	database.SaveMessage(ctx, update.Message.From.ID, update.Message.Chat.ID, update.Message.MessageID, "in", contentType, update.Message.Text, "") // Pass context

	// User Mapping Protocol: Email-based
	// Check for email in message text to link accounts
	emailRegex := regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	email := emailRegex.FindString(strings.ToLower(update.Message.Text))
	if email != "" {
		telemetry.ZapLogger.Sugar().Infow("Attempting to map user via email", "email", email, "telegram_id", update.Message.From.ID)
		err := database.LinkTelegramToEmail(ctx, update.Message.From.ID, email) // Pass context
		if err != nil {
			telemetry.ZapLogger.Sugar().Errorw("Failed to link user", "error", err)
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Your Telegram account has been linked to your Dashboard account ("+email+")."))
		}
	}

	if update.Message.Photo != nil {
		handlePhoto(ctx, bot, update.Message, token, fileHandler) // Pass context
	} else if update.Message.Document != nil {
		handleDocument(ctx, bot, update.Message, token, fileHandler) // Pass context
	} else if update.Message.IsCommand() || update.Message.Text != "" {
		handleCommand(ctx, bot, update.Message, commandHandlerManager) // Pass context
	}
}

func handleCallbackQuery(bot Bot, callback *tgbotapi.CallbackQuery, commandRouter map[string]CommandHandler) {
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
func handleCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, commandHandlerManager *CommandHandlerManager) {
	state := getUserState(message.From.ID) // Get user state for language
	err := commandHandlerManager.HandleCommand(ctx, bot, message, state)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to handle command", "error", err, "command", message.Text)
	}
}

// handlePhoto processes incoming photos.
func handlePhoto(ctx context.Context, bot Bot, message *tgbotapi.Message, token string, fileHandler *FileHandler) { // Pass context
	status.UpdateActivity()

	photo := message.Photo[len(message.Photo)-1]
	filename := downloadFile(ctx, bot, photo.FileID, "jpg", token) // Pass context
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download image."))
		return
	}

	err := fileHandler.HandlePhoto(ctx, bot, message, filename)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to handle photo", "error", err)
	}
}

// handleDocument processes incoming documents.
func handleDocument(ctx context.Context, bot Bot, message *tgbotapi.Message, token string, fileHandler *FileHandler) { // Pass context
	status.UpdateActivity()

	doc := message.Document
	// Basic check for PDF
	if doc.MimeType != "application/pdf" && !strings.HasSuffix(strings.ToLower(doc.FileName), ".pdf") {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Only PDFs are currently supported for documents."))
		return
	}

	filename := downloadFile(ctx, bot, doc.FileID, "pdf", token) // Pass context
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download document."))
		return
	}

	err := fileHandler.HandleDocument(ctx, bot, message, filename)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Failed to handle document", "error", err)
	}
}

// downloadFile downloads a file from Telegram.
func downloadFile(ctx context.Context, bot Bot, fileID, ext, token string) string { // Pass context
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("GetFile error", "error", err)
		return ""
	}

	fileURL := file.Link(token)

	resp, err := http.Get(fileURL)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("HTTP error downloading file", "error", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		telemetry.ZapLogger.Sugar().Errorw("Bad response status", "status", resp.StatusCode)
		return ""
	}

	// Create 'attachments' directory if not exists
	os.MkdirAll("attachments", 0755)

	filename := fmt.Sprintf("attachments/%s.%s", time.Now().Format("20060102_150405"), ext)
	out, err := os.Create(filename)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Create file error", "error", err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		telemetry.ZapLogger.Sugar().Errorw("Write file error", "error", err)
		return ""
	}

	telemetry.ZapLogger.Sugar().Infow("File downloaded", "path", filename)
	return filename
}
