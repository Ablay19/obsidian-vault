package bot

import (
	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/git"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/state"
)

type CommandDependencies struct {
	AIService         ai.AIServiceInterface
	RCM               *state.RuntimeConfigManager
	IngestionPipeline *pipeline.Pipeline
	GitManager        *git.Manager
}

func SetupCommands(registry *CommandRegistry) {
	registry.Register("start", &startCommandHandler{}, "Start the bot")
	registry.Register("help", &helpCommandHandler{}, "Show help message")
	registry.Register("lang", &langCommandHandler{}, "Set AI language")
	registry.Register("setprovider", &setProviderCommandHandler{}, "Set AI provider (Dynamic Menu)")
	registry.Register("stats", &statsCommandHandler{}, "Show usage statistics")
	registry.Register("last", &lastCommandHandler{}, "Show last created note")
	registry.Register("reprocess", &reprocessCommandHandler{}, "Reprocess last sent file")
	registry.Register("pid", &pidCommandHandler{}, "Show the process ID of the bot instance")
	registry.Register("link", &linkCommandHandler{}, "Link your Dashboard Account")
	registry.Register("service_status", &serviceStatusCommandHandler{}, "Show service health")
	registry.Register("modelinfo", &modelInfoCommandHandler{}, "Show AI model information")
	registry.Register("pause_bot", &pauseBotCommandHandler{}, "Pause the bot")
	registry.Register("resume_bot", &resumeBotCommandHandler{}, "Resume the bot")
	registry.Register("process", &processCommandHandler{}, "Process staged file")
}
