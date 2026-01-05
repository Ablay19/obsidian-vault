package bot

import (
	"context"
	"fmt"
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/telemetry"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StateMachine defines the interface for user state management
type StateMachine interface {
	ProcessMessage(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error
	ProcessCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState, commandRouter map[string]CommandHandler) error
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
	telemetry.ZapLogger.Sugar().Infow("Processing non-command text as AI prompt", "chat_id", message.Chat.ID, "text_len", len(message.Text))

	// Send thinking message
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Thinking..."))

	var responseText strings.Builder
	systemPrompt := fmt.Sprintf("Respond in %s. Output your response as valid HTML, with proper headings, paragraphs, and LaTeX formulas using MathJax syntax.", state.Language)

	reqCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	req := &ai.RequestModel{
		SystemPrompt: systemPrompt,
		UserPrompt:   message.Text,
		Temperature:  0.7,
	}

	err := mp.aiService.Chat(reqCtx, req, func(chunk string) {
		responseText.WriteString(chunk)
	})

	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Sorry, I had trouble thinking: "+err.Error()))
		return fmt.Errorf("AI chat failed: %w", err)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText.String())
	msg.ParseMode = tgbotapi.ModeHTML
	sentMsg, err := bot.Send(msg)
	if err == nil {
		database.SaveMessage(ctx, message.From.ID, message.Chat.ID, sentMsg.MessageID, "out", "text", responseText.String(), "")
	}

	return err
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
func (sm *StagingStateMachine) ProcessCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState, commandRouter map[string]CommandHandler) error {
	if handler, ok := commandRouter[message.Command()]; ok {
		handler.Handle(bot, message, state)
		return nil
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Unknown command. Use /help to see available commands."))
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
}

// HandleCommand processes commands and text messages
func (chm *CommandHandlerManager) HandleCommand(ctx context.Context, bot Bot, message *tgbotapi.Message, state *UserState) error {
	telemetry.ZapLogger.Sugar().Infow("Processing command/text", "chat_id", message.Chat.ID, "text", message.Text)

	// Upsert user
	database.UpsertUser(ctx, message.From)

	if !message.IsCommand() {
		return chm.stateMachine.ProcessMessage(ctx, bot, message, state)
	}

	return chm.stateMachine.ProcessCommand(ctx, bot, message, state, chm.commandRouter)
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

	state := getUserState(message.From.ID)
	fh.stateMachine.StageFile(state, filename, "image", message.Caption)

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🖼️ Image staged. Send additional text context, or type /process to analyze."))
	return nil
}

// HandleDocument processes incoming documents
func (fh *FileHandler) HandleDocument(ctx context.Context, bot Bot, message *tgbotapi.Message, filename string) error {
	database.UpsertUser(ctx, message.From)

	state := getUserState(message.From.ID)
	fh.stateMachine.StageFile(state, filename, "pdf", message.Caption)

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "📄 PDF staged. Send additional text context, or type /process to analyze."))
	return nil
}
