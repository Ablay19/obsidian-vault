package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/state" // Import the new state package
	"obsidian-automation/internal/status"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Package-level variables for dependencies
var (
	db         *sql.DB
	aiService  *ai.AIService
	rcm        *state.RuntimeConfigManager // Add package-level RCM
	userStates = make(map[int64]*UserState)
	stateMutex sync.RWMutex
	isBusy     atomic.Bool
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
func (t *TelegramBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return t.BotAPI.Request(c)
}
func (t *TelegramBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return t.BotAPI.GetFile(config)
}

// Run initializes and starts the bot.
func Run(database *sql.DB, ais *ai.AIService, runtimeConfigManager *state.RuntimeConfigManager) error {
	db = database
	aiService = ais
	rcm = runtimeConfigManager // Assign to package-level variable

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
		{Command: "pid", Description: "Show the process ID of the bot instance"},
		{Command: "setprovider", Description: "Set AI provider (Gemini, Groq)"},
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
		if status.IsPaused() || isBusy.Load() {
			time.Sleep(1 * time.Second)
			continue
		}
		if update.Message == nil {
			continue
		}
		go handleUpdate(bot, &update, token)
	}
	return nil
}

func handleUpdate(bot Bot, update *tgbotapi.Update, token string) {
	isBusy.Store(true)
	defer isBusy.Store(false)

	if update.Message.Photo != nil {
		handlePhoto(bot, update.Message, token)
	} else if update.Message.Document != nil {
		handleDocument(bot, update.Message, token)
	} else if update.Message.IsCommand() || update.Message.Text != "" {
		handleCommand(bot, update.Message)
	}
}

// handleCommand processes text messages and commands.
func handleCommand(bot Bot, message *tgbotapi.Message) {
	database.UpsertUser(message.From)
	state := getUserState(message.From.ID) // Get user state for language

	if !message.IsCommand() {
		// Handle non-command text messages as a general AI prompt
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Thinking..."))

		var responseText strings.Builder
		writer := &responseText

		systemPrompt := fmt.Sprintf("Respond in %s. Output your response as valid HTML, with proper headings, paragraphs, and LaTeX formulas using MathJax syntax.", state.Language)

		err := aiService.Process(context.Background(), writer, systemPrompt, message.Text, nil) // Assuming Process takes context, writer, system, prompt, images
		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Sorry, I had trouble thinking: "+err.Error()))
			return
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, responseText.String())
		msg.ParseMode = tgbotapi.ModeHTML
		bot.Send(msg)
		return
	}

	switch message.Command() {
	case "start", "help":
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Bot active! Send images/PDFs for processing.\n\nCommands:\n/stats - Statistics\n/last - Show last created note\n/reprocess - Reprocess last file\n/lang - Set AI language (e.g. /lang English)\n/switchkey - Switch to next API key (Gemini only)\n/setprovider - Set AI provider (e.g. /setprovider Groq)\n/modelinfo - Show AI model information\n/help - This message"))
	case "pause_bot":
		status.SetPaused(true)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is paused."))
	case "resume_bot":
		status.SetPaused(false)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is resumed."))
	case "service_status":
		statuses := status.GetServicesStatus(aiService, rcm)
		var sb strings.Builder
		sb.WriteString("📊 *Service Status*\n\n")
		for _, s := range statuses {
			sb.WriteString(fmt.Sprintf("• *%s:* %s\n", s.Name, s.Status))
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "pid":
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Bot PID: %d", os.Getpid())))
	case "reprocess":
		state := getUserState(message.From.ID)
		if state.LastProcessedFile != "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, " reprocessing last file..."))
			// Determine file type (simple guess based on extension, or store in UserState)
			fileType := "document" // Default to document for now, needs refinement
			if strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".jpg") || strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".png") {
				fileType = "image"
			}
			createObsidianNote(state.LastProcessedFile, fileType, message, bot, message.Chat.ID, 0) // messageID 0 as it's not a direct reply
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No file to reprocess. Please send a file first."))
		}
	case "modelinfo": // New command handler
		if aiService == nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
			return
		}
		infos := aiService.GetProvidersInfo() // Assuming aiService.GetProvidersInfo() exists and returns []ai.ModelInfo
		var infoText strings.Builder
		infoText.WriteString("📊 *AI Model Information*\n\n")
		for _, info := range infos {
			infoText.WriteString(fmt.Sprintf("• *Provider:* %s\n  *Model:* %s\n", info.ProviderName, info.ModelName))
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, infoText.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "lang":
		args := message.CommandArguments()
		if args == "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Current language is %s. Usage: /lang <English|French>", state.Language)))
			return
		}
		newLang := strings.Title(strings.ToLower(args))
		if newLang == "English" || newLang == "French" {
			state.Language = newLang
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Language set to %s.", newLang)))
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Unsupported language. Please use English or French."))
		}
	case "setprovider":
		if aiService == nil || len(aiService.GetAvailableProviders()) == 0 {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available or no providers are configured."))
			return
		}
		providerName := message.CommandArguments()
		if providerName == "" {
			var currentProviderName string
			activeProvider, _, _ := aiService.GetActiveProvider(context.Background())
			if activeProvider != nil {
				currentProviderName = activeProvider.GetModelInfo().ProviderName
			} else {
				currentProviderName = "None"
			}
			availableProviders := aiService.GetAvailableProviders()
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Current provider is %s.\nAvailable providers: %s\nUsage: /setprovider <provider>", currentProviderName, strings.Join(availableProviders, ", "))))
			return
		}

		if err := aiService.SetProvider(providerName); err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Failed to set AI provider: %v", err)))
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("AI provider set to: %s", providerName)))
		}
	case "stats": // Handler for /stats command
		statsData := status.GetStats() // Assuming status.GetStats() returns relevant statistics
		var statsText strings.Builder
		statsText.WriteString("📊 *Bot Statistics*\n\n")
		statsText.WriteString(fmt.Sprintf("• *Total Files Processed:* %d\n", statsData.TotalFiles))
		statsText.WriteString(fmt.Sprintf("• *Total Images Processed:* %d\n", statsData.ImageFiles))
		statsText.WriteString(fmt.Sprintf("• *Total PDFs Processed:* %d\n", statsData.PDFFiles))
		statsText.WriteString(fmt.Sprintf("• *Total AI Calls:* %d\n", statsData.AICalls))
		statsText.WriteString(fmt.Sprintf("• *Last Activity:* %s\n", formatTime(statsData.LastActivity))) // Assuming formatTime function is available
		msg := tgbotapi.NewMessage(message.Chat.ID, statsText.String())
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "last": // Handler for /last command
		state := getUserState(message.From.ID)
		if state.LastCreatedNote != "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Last created note: %s", state.LastCreatedNote)))
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No note has been created yet."))
		}
	default:
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Use /help to see available commands."))
	}
}

// handlePhoto processes incoming photos.
func handlePhoto(bot Bot, message *tgbotapi.Message, token string) {
	database.UpsertUser(message.From)
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
	database.UpsertUser(message.From)
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
	state := getUserState(message.From.ID)
	updateStatus := func(status string) {
		if messageID != 0 {
			bot.Send(tgbotapi.NewEditMessageText(chatID, messageID, status))
		}
	}

	streamCallback := func(chunk string) {
		// This could be used to stream the response to the user in real-time
	}

	content := processFileWithAI(filePath, fileType, aiService, streamCallback, state.Language, updateStatus)

	if content.Category == "unprocessed" || content.Category == "error" {
		bot.Send(tgbotapi.NewMessage(chatID, "Could not process the file."))
		return
	}

	// Create note content
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# %s\n\n", time.Now().Format("2006-01-02_15-04-05")))
	builder.WriteString(fmt.Sprintf("**Category:** %s\n", content.Category))
	builder.WriteString(fmt.Sprintf("**AI Provider:** %s\n", content.AIProvider))
	builder.WriteString(fmt.Sprintf("**Tags:** #%s\n\n", strings.Join(content.Tags, " #")))

	if content.Summary != "" {
		builder.WriteString("## Summary\n")
		builder.WriteString(content.Summary + "\n\n")
	}
	if len(content.Topics) > 0 {
		builder.WriteString("## Key Topics\n")
		for _, topic := range content.Topics {
			builder.WriteString(fmt.Sprintf("- %s\n", topic))
		}
		builder.WriteString("\n")
	}
	if len(content.Questions) > 0 {
		builder.WriteString("## Review Questions\n")
		for _, q := range content.Questions {
			builder.WriteString(fmt.Sprintf("- %s\n", q))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("## Extracted Text\n")
	builder.WriteString("```\n")
	builder.WriteString(content.Text)
	builder.WriteString("\n```\n")

	// Save the note
	noteFilename := fmt.Sprintf("%s_%s.md", time.Now().Format("20060102_150405"), content.Category)
	notePath := filepath.Join("vault", "Inbox", noteFilename)
	err := os.WriteFile(notePath, []byte(builder.String()), 0644)
	if err != nil {
		slog.Error("Error writing note file", "error", err)
		bot.Send(tgbotapi.NewMessage(chatID, "Error saving the note."))
		return
	}

	// Save to database
	hash, err := getFileHash(filePath)
	if err != nil {
		slog.Error("Error getting file hash", "error", err)
	} else {
		err := SaveProcessed(hash, content.Category, content.Text, content.Summary, content.Topics, content.Questions, content.AIProvider, message.From.ID)
		if err != nil {
			slog.Error("Error saving processed file to DB", "error", err)
		}
	}

	// Organize the note
	organizeNote(notePath, content.Category)

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Note '%s' created successfully!", noteFilename)))
	state.LastCreatedNote = noteFilename
	state.LastProcessedFile = filePath
}

// formatTime formats a time.Time object into a human-readable string.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "--"
	}
	diff := time.Since(t)

	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return t.Format("Jan 02, 2006 15:04 MST")
}
