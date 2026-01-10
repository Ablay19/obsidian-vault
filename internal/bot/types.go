package bot

import (
	"context"
	"sort"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/state"
)

// Bot interfaces for mocking
type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

type ConversationMessage struct {
	Role    string // "user" or "assistant"
	Content string
	Time    time.Time
}

type UserState struct {
	Language            string
	Provider            string
	MessageCount        int64
	LastActivity        time.Time
	LinkToken           string
	LinkExpiry          time.Time
	IsAdmin             bool
	LastProcessedFile   string
	LastCreatedNote     string
	PendingFile         string
	PendingFileType     string
	PendingContext      string
	IsStaging           bool
	ConversationHistory []ConversationMessage
	mu                  sync.RWMutex
}

// UserStates manages all user states
type UserStates struct {
	states map[int64]*UserState
	mu     sync.RWMutex
}

var userStates = &UserStates{
	states: make(map[int64]*UserState),
}

// GetUserState returns the current user state, creating one if needed
func GetUserState(userID int64) *UserState {
	userStates.mu.Lock()
	defer userStates.mu.Unlock()

	if state, exists := userStates.states[userID]; exists {
		return state
	}

	// Create new user state with defaults
	state := &UserState{
		Language:     "English",
		Provider:     "gemini",
		MessageCount: 0,
		LastActivity: time.Now(),
		IsAdmin:      false,
	}

	userStates.states[userID] = state
	return state
}

// UpdateLanguage updates the user's language preference
func (s *UserState) UpdateLanguage(lang string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Language = lang
	s.LastActivity = time.Now()
}

// UpdateProvider updates the user's AI provider preference
func (s *UserState) UpdateProvider(provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Provider = provider
	s.LastActivity = time.Now()
}

// IncrementMessageCount increments the user's message counter
func (s *UserState) IncrementMessageCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessageCount++
	s.LastActivity = time.Now()
}

// AddToConversation adds a message to the conversation history
func (s *UserState) AddToConversation(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ConversationHistory = append(s.ConversationHistory, ConversationMessage{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})

	// Keep only last 20 messages to avoid token limits
	if len(s.ConversationHistory) > 20 {
		s.ConversationHistory = s.ConversationHistory[len(s.ConversationHistory)-20:]
	}

	s.LastActivity = time.Now()
}

// GetConversationHistory returns the recent conversation history (last 10 messages)
func (s *UserState) GetConversationHistory() []ConversationMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return last 10 messages
	history := s.ConversationHistory
	if len(history) > 10 {
		history = history[len(history)-10:]
	}

	return history
}

// ClearConversationHistory clears the conversation history
func (s *UserState) ClearConversationHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ConversationHistory = nil
	s.LastActivity = time.Now()
}

// GetStats returns user statistics
func (s *UserState) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"language":      s.Language,
		"provider":      s.Provider,
		"message_count": s.MessageCount,
		"last_activity": s.LastActivity,
		"is_admin":      s.IsAdmin,
	}
}

// CommandContext holds common dependencies for command handlers
type CommandContext struct {
	Bot               Bot
	AIService         ai.AIServiceInterface
	RCM               *state.RuntimeConfigManager
	IngestionPipeline *pipeline.Pipeline
	GitManager        *git.Manager
}

// CommandRegistry manages command registration and descriptions
type CommandRegistry struct {
	commands     map[string]CommandHandler
	descriptions map[string]string
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands:     make(map[string]CommandHandler),
		descriptions: make(map[string]string),
	}
}

func (r *CommandRegistry) Register(name string, handler CommandHandler, desc string) {
	r.commands[name] = handler
	r.descriptions[name] = desc
}

func (r *CommandRegistry) GetHandler(name string) (CommandHandler, bool) {
	handler, ok := r.commands[name]
	return handler, ok
}

func (r *CommandRegistry) GetDescriptions() map[string]string {
	return r.descriptions
}

func (r *CommandRegistry) GetBotCommands() []tgbotapi.BotCommand {
	var cmds []tgbotapi.BotCommand
	for name, desc := range r.descriptions {
		cmds = append(cmds, tgbotapi.BotCommand{Command: name, Description: desc})
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Command < cmds[j].Command
	})
	return cmds
}

// CommandHandler interface for command handling
type CommandHandler interface {
	Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error
}
