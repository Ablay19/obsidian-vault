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

func NewCommandRouter(deps CommandDependencies) map[string]CommandHandler {
	return map[string]CommandHandler{
		"start":          &startCommandHandler{},
		"help":           &helpCommandHandler{},
		"lang":           &langCommandHandler{},
		"setprovider":    &setProviderCommandHandler{aiService: deps.AIService},
		"stats":          &statsCommandHandler{rcm: deps.RCM},
		"last":           &lastCommandHandler{},
		"reprocess":      &reprocessCommandHandler{deps: deps},
		"pid":            &pidCommandHandler{},
		"link":           &linkCommandHandler{},
		"service_status": &serviceStatusCommandHandler{aiService: deps.AIService, rcm: deps.RCM},
		"modelinfo":      &modelInfoCommandHandler{aiService: deps.AIService},
		"pause_bot":      &pauseBotCommandHandler{},
		"resume_bot":     &resumeBotCommandHandler{},
		"process":        &processCommandHandler{ingestionPipeline: deps.IngestionPipeline},
	}
}
