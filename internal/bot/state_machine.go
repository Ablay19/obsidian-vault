package bot

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/telemetry"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StateMachine defines the interface for user state management
type StateMachine interface {
	ProcessMessage(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error
	ProcessCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState, commandRouter map[string]CommandHandler, cmdCtx *CommandContext) error
	IsStaging(state *UserState) bool
	AddContext(state *UserState, context string)
	StageFile(state *UserState, filename, fileType, caption string)
}

// MessageProcessor handles non-command messages
type MessageProcessor struct {
	aiService ai.AIServiceInterface
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(aiService ai.AIServiceInterface) *MessageProcessor {
	return &MessageProcessor{
		aiService: aiService,
	}
}

// ProcessMessage processes non-command text messages
func (mp *MessageProcessor) ProcessMessage(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error {
	telemetry.Info("Processing non-command text as AI prompt", "chat_id", message.Chat.ID, "text_len", len(message.Text))

	// Add user message to conversation history
	state.AddToConversation("user", message.Text)

	// Set user's preferred provider on AI service
	if state.Provider != "" {
		providerName := normalizeProviderName(state.Provider)
		if err := mp.aiService.SetProvider(providerName); err != nil {
			telemetry.Warn("Failed to set user provider", "provider", providerName, "error", err)
		}
	}

	// Send thinking message
	_, _ = bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Thinking..."))

	var responseText strings.Builder
	systemPrompt := fmt.Sprintf("Respond in %s. Output your response as valid HTML, with proper headings, paragraphs, and LaTeX formulas using MathJax syntax.", state.Language)

	// Build conversation context
	conversationHistory := state.GetConversationHistory()
	var conversationContext strings.Builder

	// Add system prompt
	conversationContext.WriteString(systemPrompt)
	conversationContext.WriteString("\n\n")

	// Add conversation history
	for _, msg := range conversationHistory {
		if msg.Role == "user" {
			conversationContext.WriteString("User: ")
			conversationContext.WriteString(msg.Content)
			conversationContext.WriteString("\n")
		} else if msg.Role == "assistant" {
			conversationContext.WriteString("Assistant: ")
			conversationContext.WriteString(msg.Content)
			conversationContext.WriteString("\n")
		}
	}

	// Add current user message
	conversationContext.WriteString("User: ")
	conversationContext.WriteString(message.Text)
	conversationContext.WriteString("\nAssistant: ")

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	req := &ai.RequestModel{
		SystemPrompt: systemPrompt,
		UserPrompt:   conversationContext.String(),
		Temperature:  0.7,
	}

	// Track which provider will be used
	providerToUse := state.Provider
	if providerToUse == "" {
		providerToUse = mp.aiService.GetActiveProviderName()
	}

	// Track response time
	startTime := time.Now()
	err := mp.aiService.Chat(reqCtx, req, func(chunk string) {
		responseText.WriteString(chunk)
	})
	responseTime := time.Since(startTime)
	TrackResponseTime(responseTime)

	if err != nil {
		userMsg := "Sorry, I had trouble thinking right now. Please try again later."
		if appErr, ok := err.(*ai.AppError); ok && appErr.UserMessage != "" {
			userMsg = appErr.UserMessage
		}
		_, sendErr := bot.Send(tgbotapi.NewMessage(message.Chat.ID, userMsg))
		if sendErr != nil {
			telemetry.Error("Failed to send error message", "error", sendErr)
		}
		return fmt.Errorf("AI chat failed: %w", err)
	}

	// Track provider usage
	TrackProviderUsage(providerToUse)

	// Clean the response - strip unsupported HTML tags
	cleanResponse := cleanTelegramResponse(responseText.String())

	msg := tgbotapi.NewMessage(message.Chat.ID, cleanResponse)
	msg.ParseMode = tgbotapi.ModeHTML
	sentMsg, err := bot.Send(msg)
	if err == nil {
		// Add assistant response to conversation history
		state.AddToConversation("assistant", cleanResponse)
		database.SaveMessage(ctx, message.From.ID, message.Chat.ID, sentMsg.MessageID, "out", "text", cleanResponse, "")

		// Trigger webhook event
		go TriggerMessageEvent(message, cleanResponse)
	}

	return err
}

// cleanTelegramResponse removes unsupported HTML tags from the response
func cleanTelegramResponse(response string) string {
	// Remove unsupported HTML tags (h1-h6, div, span, etc.)
	re := regexp.MustCompile(`<(/?)(h[1-6]|div|span|section|article|header|footer|main|nav|aside)(\s[^>]*)?>`)
	cleaned := re.ReplaceAllString(response, "")

	// Convert markdown headers to bold
	re = regexp.MustCompile(`(?m)^#+\s+(.+)$`)
	cleaned = re.ReplaceAllString(cleaned, "<b>$1</b>")

	// Convert markdown lists to HTML lists
	re = regexp.MustCompile(`(?m)^\s*[-*]\s+(.+)$`)
	cleaned = re.ReplaceAllString(cleaned, "• $1")

	// Clean up multiple newlines
	re = regexp.MustCompile(`\n{3,}`)
	cleaned = re.ReplaceAllString(cleaned, "\n\n")

	return strings.TrimSpace(cleaned)
}

// StagingStateMachine handles the staging state machine
type StagingStateMachine struct {
	messageProcessor *MessageProcessor
}

// NewStagingStateMachine creates a new staging state machine
func NewStagingStateMachine(messageProcessor *MessageProcessor) *StagingStateMachine {
	return &StagingStateMachine{
		messageProcessor: messageProcessor,
	}
}

// ProcessMessage implements StateMachine interface
func (sm *StagingStateMachine) ProcessMessage(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error {
	if sm.IsStaging(state) {
		sm.AddContext(state, message.Text)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "📝 Context added. Add more or type /process."))
		return nil
	}

	return sm.messageProcessor.ProcessMessage(ctx, bot, message, state)
}

// ProcessCommand implements StateMachine interface
func (sm *StagingStateMachine) ProcessCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState, commandRouter map[string]CommandHandler, cmdCtx *CommandContext) error {
	if handler, ok := commandRouter[message.Command()]; ok {
		return handler.Handle(ctx, message, state, cmdCtx)
	}

	_, err := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Use /help to see available commands."))
	if err != nil {
		return err
	}
	return fmt.Errorf("unknown command: %s", message.Command())
}

// IsStaging checks if user is in staging mode
func (sm *StagingStateMachine) IsStaging(state *UserState) bool {
	return state.IsStaging
}

// AddContext adds context to staging state
func (sm *StagingStateMachine) AddContext(state *UserState, contextText string) {
	if state.PendingContext != "" {
		state.PendingContext += "\n"
	}
	state.PendingContext += contextText
}

// StageFile stages a file for processing
func (sm *StagingStateMachine) StageFile(state *UserState, filename, fileType, caption string) {
	state.PendingFile = filename
	state.PendingFileType = fileType
	state.IsStaging = true
	state.PendingContext = caption
}

// CommandHandlerManager manages command routing and execution
type CommandHandlerManager struct {
	stateMachine  StateMachine
	commandRouter map[string]CommandHandler
	cmdCtx        *CommandContext
}

// HandleCommand processes commands and text messages
func (chm *CommandHandlerManager) HandleCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error {
	telemetry.Info("Processing command/text", "chat_id", message.Chat.ID, "text", message.Text)

	// Upsert user
	database.UpsertUser(ctx, message.From)

	// Track message
	TrackMessage()

	if !message.IsCommand() {
		return chm.stateMachine.ProcessMessage(ctx, bot, message, state)
	}

	// Track command usage
	TrackCommand(message.Command())

	return chm.stateMachine.ProcessCommand(ctx, bot, message, state, chm.commandRouter, chm.cmdCtx)
}

// GetCommandRouter returns the command router
func (chm *CommandHandlerManager) GetCommandRouter() map[string]CommandHandler {
	return chm.commandRouter
}

// FileHandler manages file processing operations
type FileHandler struct {
	stateMachine StateMachine
}

// NewFileHandler creates a new file handler
func NewFileHandler(stateMachine StateMachine) *FileHandler {
	return &FileHandler{
		stateMachine: stateMachine,
	}
}

// HandlePhoto processes incoming photos
func (fh *FileHandler) HandlePhoto(ctx context.Context, bot Bot, message *tgbotapi.Message, filename string) error {
	database.UpsertUser(ctx, message.From)

	state := GetUserState(message.From.ID)
	fh.stateMachine.StageFile(state, filename, "image", message.Caption)

	// Track image processed
	TrackImageProcessed()

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🖼️ Image staged. Send additional text context, or type /process to analyze."))
	return nil
}

// HandleDocument processes incoming documents
func (fh *FileHandler) HandleDocument(ctx context.Context, bot Bot, message *tgbotapi.Message, filename string) error {
	database.UpsertUser(ctx, message.From)

	state := GetUserState(message.From.ID)
	fh.stateMachine.StageFile(state, filename, "pdf", message.Caption)

	// Track document processed
	TrackDocumentProcessed()

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "📄 PDF staged. Send additional text context, or type /process to analyze."))
	return nil
}
