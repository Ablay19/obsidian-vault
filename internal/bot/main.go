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
	"obsidian-automation/internal/rag"
	"obsidian-automation/internal/state" // Import the new state package
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/telemetry"
	"obsidian-automation/internal/vectorstore"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func InitializeWebhookManager() error {
	return nil
}

func InitializeSecurityManager() error {
	return nil
}

// Package-level variables for dependencies
var (
	aiService         ai.AIServiceInterface
	rcm               *state.RuntimeConfigManager
	ingestionPipeline *pipeline.Pipeline
	wsManager         *ws.Manager
	globalRAGChain    *rag.RAGChain
	userLocks         sync.Map
	gitManager        *git.Manager
	startTime         = time.Now()
	botPaused         = false

	// Usage statistics counters
	statsMutex              sync.RWMutex
	totalMessages           int64
	totalCommands           int64
	totalImagesProcessed    int64
	totalDocumentsProcessed int64
	providerUsage           map[string]int64
	commandUsage            map[string]int64
	responseTimes           []time.Duration
)

// Run initializes and starts the bot.
func Run(db *sql.DB, ais ai.AIServiceInterface, runtimeConfigManager *state.RuntimeConfigManager, wsm *ws.Manager, vectorStore vectorstore.VectorStore) error {
	// Set global vector store for RAG functionality
	globalVectorStore = vectorStore

	// Initialize RAG chain
	retriever := rag.NewVectorRetriever(vectorStore, nil, 5, 0.1) // topK=5, threshold=0.1
	llm := &rag.AIServiceLLM{AIService: ais, ModelName: ais.GetActiveProviderName()}
	ragChain, err := rag.NewRAGChain(retriever, llm)
	if err != nil {
		telemetry.Error("Failed to initialize RAG chain: " + err.Error())
	} else {
		globalRAGChain = ragChain
	}

	// Initialize user state persistence
	if err := InitUserStatePersistence(); err != nil {
		telemetry.Warn("Failed to initialize user state persistence: " + err.Error())
	}

	// Check external binary dependencies first
	if err := ValidateBinaries(); err != nil {
		telemetry.Error("Binary validation failed: " + err.Error())
		telemetry.Warn("Bot will start with limited functionality. Some features may not work properly.")
	}
	aiService = ais
	rcm = runtimeConfigManager // Assign to package-level variable
	wsManager = wsm

	// Initialize usage statistics
	statsMutex.Lock()
	providerUsage = make(map[string]int64)
	commandUsage = make(map[string]int64)
	responseTimes = make([]time.Duration, 0)
	statsMutex.Unlock()

	// Initialize webhook manager
	if err := InitializeWebhookManager(); err != nil {
		telemetry.Warn("Failed to initialize webhook manager: " + err.Error())
	}

	// Initialize security manager
	if err := InitializeSecurityManager(); err != nil {
		telemetry.Warn("Failed to initialize security manager: " + err.Error())
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		telemetry.Fatal("TELEGRAM_BOT_TOKEN is not set. Bot cannot start.")
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	bot, err := NewBot(token, nil) // TODO: pass logger
	if err != nil {
		return err
	}

	// Initialize Git Manager
	gitCfg := config.AppConfig.Git
	gitManager = git.NewManager(gitCfg.VaultPath)
	if err := gitManager.ConfigureUser(gitCfg.UserName, gitCfg.UserEmail); err != nil {
		telemetry.Warn("Failed to configure Git user: " + err.Error())
	}
	if gitCfg.RemoteURL != "" {
		if err := gitManager.EnsureRemote(gitCfg.RemoteURL); err != nil {
			telemetry.Warn("Failed to ensure Git remote: " + err.Error())
		}
	}

	// Initialize Pipeline
	processor := NewBotProcessor(aiService)
	sink := NewBotSink(database.Client.DB, bot.api, gitManager)       // Use database.Client.DB
	ingestionPipeline = pipeline.NewPipeline(3, 100, processor, sink) // 3 workers, buffer 100
	ingestionPipeline.Start(context.Background())
	defer ingestionPipeline.Stop()

	// Create Command Context
	cmdCtx := &CommandContext{
		Bot:               bot,
		AIService:         aiService,
		RCM:               rcm,
		IngestionPipeline: ingestionPipeline,
		GitManager:        gitManager,
	}

	// Initialize Command Registry
	registry := NewCommandRegistry()
	SetupCommands(registry)

	// Initialize State Machine and Command Handler Manager
	messageProcessor := NewMessageProcessor(aiService)
	stagingStateMachine := NewStagingStateMachine(messageProcessor)
	commandHandlerManager := &CommandHandlerManager{
		stateMachine:  stagingStateMachine,
		commandRouter: registry.commands,
		cmdCtx:        cmdCtx,
	}
	fileHandler := NewFileHandler(stagingStateMachine)

	commands := registry.GetBotCommands()
	if len(commands) > 0 {
		if _, err := bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
			telemetry.Error("Error setting bot commands: " + err.Error())
		} else {
			telemetry.Info(fmt.Sprintf("Successfully registered %d bot commands", len(commands)))
		}
	} else {
		telemetry.Warn("No bot commands registered - check command setup")
	}

	telemetry.Info("Authorized on account: " + bot.api.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	telemetry.Info("Bot is running...")

	for update := range updates {
		if status.IsPaused() {
			time.Sleep(1 * time.Second)
			continue
		}
		telemetry.Debug("Received update")
		go handleUpdate(context.Background(), bot, &update, token, commandHandlerManager, fileHandler, ais.(*ai.AIService)) // Pass context
	}
	return nil
}

// handleUpdate processes incoming Telegram updates.
func handleUpdate(ctx context.Context, bot Bot, update *tgbotapi.Update, token string, commandHandlerManager *CommandHandlerManager, fileHandler *FileHandler, aiService *ai.AIService) {
	if update.CallbackQuery != nil {
		telemetry.Info("Handling callback query")
		handleCallbackQuery(ctx, bot, update.CallbackQuery, commandHandlerManager, aiService)
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
			telemetry.Warn("User is already being processed, skipping duplicate content update for user: " + fmt.Sprintf("%d", userID))
			return
		}
		defer userLocks.Delete(userID)
	}

	telemetry.Info("Handling message from user: " + update.Message.From.UserName)

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
		telemetry.Info("Attempting to map user via email: " + email)
		err := database.LinkTelegramToEmail(ctx, update.Message.From.ID, email) // Pass context
		if err != nil {
			telemetry.Error("Failed to link user: " + err.Error())
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

func handleCallbackQuery(ctx context.Context, bot Bot, callback *tgbotapi.CallbackQuery, commandHandlerManager *CommandHandlerManager, aiService *ai.AIService) {
	// Always answer the callback query to stop the loading spinner
	defer bot.Request(tgbotapi.NewCallback(callback.ID, ""))

	if callback.Data == "refresh_providers" {
		// Re-trigger handleCommand for setprovider to refresh the list
		msg := callback.Message
		msg.From = callback.From // Crucial: preserve the user who clicked
		msg.Text = "/setprovider"
		if handler, ok := commandHandlerManager.commandRouter["setprovider"]; ok {
			handler.Handle(ctx, msg, GetUserState(msg.From.ID), commandHandlerManager.cmdCtx)
		}
		return
	}
	if strings.HasPrefix(callback.Data, "setprovider_") {
		provider := strings.TrimPrefix(callback.Data, "setprovider_")

		// Get user state
		userState := GetUserState(callback.From.ID)

		// Validate provider
		supportedProviders := []string{"gemini", "google", "deepseek", "groq", "cloudflare", "openrouter", "replicate", "together", "huggingface"}
		supported := false
		for _, p := range supportedProviders {
			if provider == p {
				supported = true
				break
			}
		}

		if supported {
			userState.UpdateProvider(provider)

			// Set the provider on the AI service
			if err := aiService.SetProvider(normalizeProviderName(provider)); err != nil {
				// Edit the message to show error
				editMsg := tgbotapi.NewEditMessageText(
					callback.Message.Chat.ID,
					callback.Message.MessageID,
					fmt.Sprintf("❌ Provider set to %s but AI service error: %v", provider, err),
				)
				bot.Send(editMsg)
				return
			}

			// Trigger webhook event
			go TriggerProviderEvent(callback.From.ID, "", provider)

			// Edit the message to show success and remove keyboard (no notification to user)
			editMsg := tgbotapi.NewEditMessageText(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				fmt.Sprintf("✅ AI Provider set to: **%s**", strings.Title(provider)),
			)
			editMsg.ParseMode = tgbotapi.ModeMarkdown

			// Remove the inline keyboard
			editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
			}

			bot.Send(editMsg)
		} else {
			// Edit message to show invalid provider
			editMsg := tgbotapi.NewEditMessageText(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				fmt.Sprintf("❌ Invalid provider: %s", provider),
			)
			bot.Send(editMsg)
		}
	}

	// Handle mode selection callbacks
	if strings.HasPrefix(callback.Data, "mode_") {
		mode := strings.TrimPrefix(callback.Data, "mode_")
		userState := GetUserState(callback.From.ID)

		// Store processing mode in user state (you can extend this later)
		userState.UpdateProvider(mode) // Temporarily using provider field for mode

		// Update message and remove keyboard
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fmt.Sprintf("🎯 Processing mode set to: **%s**", strings.Title(mode)),
		)
		editMsg.ParseMode = tgbotapi.ModeMarkdown
		editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}
		bot.Send(editMsg)
		return
	}

	// Handle bot instance selection callbacks
	if strings.HasPrefix(callback.Data, "bot_") {
		botInstance := strings.TrimPrefix(callback.Data, "bot_")

		// For now, just show selection (you can implement actual bot switching later)
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			fmt.Sprintf("🔄 Switched to bot instance: **%s**", strings.Title(strings.Replace(botInstance, "_", " ", -1))),
		)
		editMsg.ParseMode = tgbotapi.ModeMarkdown
		editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}
		bot.Send(editMsg)
		return
	}
}

// handleCommand processes text messages and commands.
func handleCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, commandHandlerManager *CommandHandlerManager) {
	state := GetUserState(message.From.ID) // Get user state for language
	err := commandHandlerManager.HandleCommand(ctx, bot, message, state)
	if err != nil {
		telemetry.Error("Failed to handle command: " + err.Error())
	}
}

// handlePhoto processes incoming photos.
func handlePhoto(ctx context.Context, bot Bot, message *tgbotapi.Message, token string, fileHandler *FileHandler) {
	status.UpdateActivity()

	photo := message.Photo[len(message.Photo)-1]
	filename := downloadFile(ctx, bot, photo.FileID, "jpg", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download image."))
		return
	}

	// Add file to pending batch for user
	userState := GetUserState(message.From.ID)
	userState.AddPendingFile(filename, "image", message.Caption)

	// Send confirmation
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🖼️ Image added to batch (%d files pending). Send more files or use /batch to process all simultaneously.", len(userState.GetPendingFiles()))))

	// Legacy file handler for compatibility
	err := fileHandler.HandlePhoto(ctx, bot, message, filename)
	if err != nil {
		telemetry.Error("Failed to handle photo: " + err.Error())
	}
}

// handleDocument processes incoming documents.
func handleDocument(ctx context.Context, bot Bot, message *tgbotapi.Message, token string, fileHandler *FileHandler) {
	status.UpdateActivity()

	doc := message.Document
	// Basic check for PDF
	if doc.MimeType != "application/pdf" && !strings.HasSuffix(strings.ToLower(doc.FileName), ".pdf") {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Only PDFs are currently supported for documents."))
		return
	}

	filename := downloadFile(ctx, bot, doc.FileID, "pdf", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to download document."))
		return
	}

	// Add file to pending batch for user
	userState := GetUserState(message.From.ID)
	userState.AddPendingFile(filename, "pdf", message.Caption)

	// Send confirmation
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("📄 PDF added to batch (%d files pending). Send more files or use /batch to process all simultaneously.", len(userState.GetPendingFiles()))))

	// Legacy file handler for compatibility
	err := fileHandler.HandleDocument(ctx, bot, message, filename)
	if err != nil {
		telemetry.Error("Failed to handle document: " + err.Error())
	}
}

// downloadFile downloads a file from Telegram.
func downloadFile(ctx context.Context, bot Bot, fileID, ext, token string) string { // Pass context
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		telemetry.Error("GetFile error: " + err.Error())
		return ""
	}

	fileURL := file.Link(token)

	resp, err := http.Get(fileURL)
	if err != nil {
		telemetry.Error("HTTP error downloading file: " + err.Error())
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		telemetry.Error("Bad response status: " + fmt.Sprintf("%d", resp.StatusCode))
		return ""
	}

	// Create 'attachments' directory if not exists
	os.MkdirAll("attachments", 0755)

	filename := fmt.Sprintf("attachments/%s.%s", time.Now().Format("20060102_150405"), ext)
	out, err := os.Create(filename)
	if err != nil {
		telemetry.Error("Create file error: " + err.Error())
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		telemetry.Error("Write file error: " + err.Error())
		return ""
	}

	telemetry.Info("File downloaded: " + filename)
	return filename
}

// Statistics tracking functions

// TrackMessage increments the message counter
func TrackMessage() {
	statsMutex.Lock()
	totalMessages++
	statsMutex.Unlock()
}

// TrackCommand increments the command counter and tracks which command was used
func TrackCommand(command string) {
	statsMutex.Lock()
	totalCommands++
	commandUsage[command]++
	statsMutex.Unlock()
}

// TrackImageProcessed increments the image processing counter
func TrackImageProcessed() {
	statsMutex.Lock()
	totalImagesProcessed++
	statsMutex.Unlock()
}

// TrackDocumentProcessed increments the document processing counter
func TrackDocumentProcessed() {
	statsMutex.Lock()
	totalDocumentsProcessed++
	statsMutex.Unlock()
}

// TrackProviderUsage tracks which AI provider was used
func TrackProviderUsage(provider string) {
	statsMutex.Lock()
	providerUsage[provider]++
	statsMutex.Unlock()
}

// TrackResponseTime records a response time for performance monitoring
func TrackResponseTime(duration time.Duration) {
	statsMutex.Lock()
	// Keep only last 100 response times to avoid memory bloat
	if len(responseTimes) >= 100 {
		responseTimes = responseTimes[1:]
	}
	responseTimes = append(responseTimes, duration)
	statsMutex.Unlock()
}

// GetStats returns current usage statistics
func GetStats() (messages, commands, images, docs int64, provUsage map[string]int64, cmdUsage map[string]int64, avgResponse time.Duration) {
	statsMutex.RLock()
	defer statsMutex.RUnlock()

	messages = totalMessages
	commands = totalCommands
	images = totalImagesProcessed
	docs = totalDocumentsProcessed
	provUsage = providerUsage
	cmdUsage = commandUsage

	if len(responseTimes) > 0 {
		var total time.Duration
		for _, d := range responseTimes {
			total += d
		}
		avgResponse = total / time.Duration(len(responseTimes))
	}

	return
}
