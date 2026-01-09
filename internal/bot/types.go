package bot

import (
	"context"
	"sort"

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

type UserState struct {
	Language          string
	LastProcessedFile string
	LastCreatedNote   string
	PendingFile       string
	PendingFileType   string
	PendingContext    string
	IsStaging         bool
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
