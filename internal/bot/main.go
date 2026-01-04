package bot

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/state" // Import the new state package
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/config" // Import app config
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Package-level variables for dependencies
var (
	db         *sql.DB
	aiService  ai.AIServiceInterface
	rcm        *state.RuntimeConfigManager
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
func Run(database *sql.DB, ais ai.AIServiceInterface, runtimeConfigManager *state.RuntimeConfigManager, wsm *ws.Manager) error {
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
		zap.S().Warn("Failed to configure Git user", "error", err)
	}
	if gitCfg.RemoteURL != "" {
		if err := gitManager.EnsureRemote(gitCfg.RemoteURL); err != nil {
			zap.S().Warn("Failed to ensure Git remote", "error", err)
		}
	}

	// Initialize Pipeline
	processor := NewBotProcessor(aiService)
	sink := NewBotSink(db, botAPI, gitManager)
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
		zap.S().Error("Error setting bot commands", "error", err)
	}

	zap.S().Info("Authorized on account", "username", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	zap.S().Info("Bot is running...")

	for update := range updates {
		if status.IsPaused() {
			time.Sleep(1 * time.Second)
			continue
		}
		zap.S().Debug("Received update", "update_id", update.UpdateID)
		go handleUpdate(bot, &update, token, commandRouter)
	}
	return nil
}

// handleUpdate processes incoming Telegram updates.
func handleUpdate(bot Bot, update *tgbotapi.Update, token string, commandRouter map[string]CommandHandler) {
	if update.CallbackQuery != nil {
		zap.S().Info("Handling callback query", "chat_id", update.CallbackQuery.Message.Chat.ID, "data", update.CallbackQuery.Data)
		handleCallbackQuery(bot, update.CallbackQuery, commandRouter)
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
			zap.S().Warn("User is already being processed, skipping duplicate content update", "user_id", userID)
			return
		}
		defer userLocks.Delete(userID)
	}

	zap.S().Info("Handling message", "chat_id", update.Message.Chat.ID, "user", update.Message.From.UserName, "text", update.Message.Text)

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
		zap.S().Info("Attempting to map user via email", "email", email, "telegram_id", update.Message.From.ID)
		err := database.LinkTelegramToEmail(update.Message.From.ID, email)
		if err != nil {
			zap.S().Error("Failed to link user", "error", err)
		} else {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Your Telegram account has been linked to your Dashboard account ("+email+")."))
		}
	}

	if update.Message.Photo != nil {
		handlePhoto(bot, update.Message, token)
	} else if update.Message.Document != nil {
		handleDocument(bot, update.Message, token)
	} else if update.Message.IsCommand() || update.Message.Text != "" {
		handleCommand(bot, update.Message, commandRouter)
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
func handleCommand(bot Bot, message *tgbotapi.Message, commandRouter map[string]CommandHandler) {
	zap.S().Info("Processing command/text", "chat_id", message.Chat.ID, "text", message.Text)
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

		zap.S().Info("Handling non-command text as AI prompt", "chat_id", message.Chat.ID, "text_len", len(message.Text))
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

// handlePhoto processes incoming photos.
func handlePhoto(bot Bot, message *tgbotapi.Message, token string) {
	database.UpsertUser(message.From)
	status.UpdateActivity()

	photo := message.Photo[len(message.Photo)-1]
	filename := downloadFile(bot, photo.FileID, "jpg", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download image."))
		return
	}

	state := getUserState(message.From.ID)
	state.PendingFile = filename
	state.PendingFileType = "image"
	state.IsStaging = true
	state.PendingContext = ""

	// Check for caption and add it as initial context
	if message.Caption != "" {
		state.PendingContext = message.Caption
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🖼️ Image staged. Send additional text context, or type /process to analyze."))
}

// handleDocument processes incoming documents.
func handleDocument(bot Bot, message *tgbotapi.Message, token string) {
	database.UpsertUser(message.From)
	status.UpdateActivity()

	doc := message.Document
	// Basic check for PDF
	if doc.MimeType != "application/pdf" && !strings.HasSuffix(strings.ToLower(doc.FileName), ".pdf") {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Only PDFs are currently supported for documents."))
		return
	}

	filename := downloadFile(bot, doc.FileID, "pdf", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download document."))
		return
	}

	state := getUserState(message.From.ID)
	state.PendingFile = filename
	state.PendingFileType = "pdf"
	state.IsStaging = true
	state.PendingContext = ""

	if message.Caption != "" {
		state.PendingContext = message.Caption
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "📄 PDF staged. Send additional text context, or type /process to analyze."))
}

// downloadFile downloads a file from Telegram.
func downloadFile(bot Bot, fileID, ext, token string) string {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		zap.S().Error("GetFile error", "error", err)
		return ""
	}

	fileURL := file.Link(token)

	resp, err := http.Get(fileURL)
	if err != nil {
		zap.S().Error("HTTP error downloading file", "error", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		zap.S().Error("Bad response status", "status", resp.StatusCode)
		return ""
	}

	// Create 'attachments' directory if not exists
	os.MkdirAll("attachments", 0755)

	filename := fmt.Sprintf("attachments/%s.%s", time.Now().Format("20060102_150405"), ext)
	out, err := os.Create(filename)
	if err != nil {
		zap.S().Error("Create file error", "error", err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		zap.S().Error("Write file error", "error", err)
		return ""
	}

	zap.S().Info("File downloaded", "path", filename)
	return filename
}
