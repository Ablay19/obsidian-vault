package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/telemetry"
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

type PendingFile struct {
	FilePath string
	FileType string
	Caption  string
}

type UserState struct {
	Language            string
	Provider            string
	MessageCount        int64
	LastActivity        time.Time
	IsAdmin             bool
	LastProcessedFile   string
	LastCreatedNote     string
	PendingFile         string        // For backward compatibility
	PendingFileType     string        // For backward compatibility
	PendingFiles        []PendingFile // New: support multiple files
	PendingContext      string
	IsStaging           bool
	IsBatchProcessing   bool   // New: batch processing flag
	BatchJobID          string // New: track batch job
	ConversationHistory []ConversationMessage
	mu                  sync.RWMutex
}

// UserStates manages all user states with persistence
type UserStates struct {
	states    map[int64]*UserState
	stateFile string
	autoSave  bool
	mu        sync.RWMutex
}

// SaveUserStates saves all user states to a JSON file
func (us *UserStates) SaveUserStates() error {
	if us.stateFile == "" {
		return nil
	}

	us.mu.RLock()
	defer us.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(us.stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	data, err := json.MarshalIndent(us.states, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user states: %w", err)
	}

	if err := os.WriteFile(us.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write user states: %w", err)
	}

	telemetry.Info("User states saved successfully", "file", us.stateFile, "users", len(us.states))
	return nil
}

// LoadUserStates loads user states from a JSON file
func (us *UserStates) LoadUserStates() error {
	if us.stateFile == "" {
		return nil
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	data, err := os.ReadFile(us.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			telemetry.Info("No existing user state file found, starting fresh", "file", us.stateFile)
			return nil
		}
		return fmt.Errorf("failed to read user states: %w", err)
	}

	if err := json.Unmarshal(data, &us.states); err != nil {
		return fmt.Errorf("failed to unmarshal user states: %w", err)
	}

	telemetry.Info("User states loaded successfully", "file", us.stateFile, "users", len(us.states))
	return nil
}

// EnableAutoSave enables automatic saving on state changes
func (us *UserStates) EnableAutoSave() {
	us.autoSave = true
}

// SetStateFile sets the file path for persistence
func (us *UserStates) SetStateFile(filePath string) {
	us.stateFile = filePath
}

var userStates = &UserStates{
	states:    make(map[int64]*UserState),
	stateFile: ".data/user_states.json",
	autoSave:  true,
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

	// Auto-save if enabled
	if userStates.autoSave {
		go userStates.SaveUserStates()
	}
}

// UpdateProvider updates the user's AI provider preference
func (s *UserState) UpdateProvider(provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Provider = provider
	s.LastActivity = time.Now()

	// Auto-save if enabled
	if userStates.autoSave {
		go userStates.SaveUserStates()
	}
}

// IncrementMessageCount increments the user's message counter
func (s *UserState) IncrementMessageCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessageCount++
	s.LastActivity = time.Now()

	// Auto-save if enabled
	if userStates.autoSave {
		go userStates.SaveUserStates()
	}
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

	// Auto-save if enabled
	if userStates.autoSave {
		go userStates.SaveUserStates()
	}
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

	// Auto-save if enabled
	if userStates.autoSave {
		go userStates.SaveUserStates()
	}
}

// InitUserStatePersistence initializes user state persistence
func InitUserStatePersistence() error {
	// Load existing states on startup
	if err := userStates.LoadUserStates(); err != nil {
		telemetry.Warn("Failed to load user states, starting fresh: " + err.Error())
	}
	return nil
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

// AddPendingFile adds a file to the pending batch
func (s *UserState) AddPendingFile(filePath, fileType, caption string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Backward compatibility: if only one file, set legacy fields
	if len(s.PendingFiles) == 0 {
		s.PendingFile = filePath
		s.PendingFileType = fileType
	}

	s.PendingFiles = append(s.PendingFiles, PendingFile{
		FilePath: filePath,
		FileType: fileType,
		Caption:  caption,
	})
	s.LastActivity = time.Now()
}

// GetPendingFiles returns all pending files
func (s *UserState) GetPendingFiles() []PendingFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.PendingFiles
}

// ClearPendingFiles removes all pending files
func (s *UserState) ClearPendingFiles() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PendingFiles = nil
	s.PendingFile = ""
	s.PendingFileType = ""
	s.PendingContext = ""
	s.IsStaging = false
	s.LastActivity = time.Now()
}

// SetBatchProcessing sets batch processing state
func (s *UserState) SetBatchProcessing(processing bool, jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsBatchProcessing = processing
	s.BatchJobID = jobID
	s.LastActivity = time.Now()
}

// GetBatchProcessingStatus returns batch processing status
func (s *UserState) GetBatchProcessingStatus() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsBatchProcessing, s.BatchJobID
}

// CommandHandler interface for command handling
type CommandHandler interface {
	Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error
}
