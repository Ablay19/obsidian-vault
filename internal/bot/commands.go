package bot

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"obsidian-automation/internal/telemetry"
)

func TriggerProviderEvent(userID int64, oldProvider, newProvider string) {
	telemetry.Info("Provider changed", "user_id", userID, "old_provider", oldProvider, "new_provider", newProvider)
}

func TriggerMessageEvent(message *tgbotapi.Message, response string) {
	telemetry.Info("Message response generated", "user_id", message.From.ID, "chat_id", message.Chat.ID)
}

func SetupCommands(registry *CommandRegistry) {
	// Register command handlers
	registry.Register("setprovider", &setProviderCommandHandler{}, "Set AI provider")
	registry.Register("clear", &clearCommandHandler{}, "Clear conversation history")
	registry.Register("stats", &statsCommandHandler{}, "Show bot usage statistics")
	registry.Register("webhook", &webhookCommandHandler{}, "Manage webhooks for external integrations")
	registry.Register("security", &securityCommandHandler{}, "Manage security settings")
	registry.Register("process", &processCommandHandler{}, "Process staged file with AI")
	registry.Register("reprocess", &reprocessCommandHandler{}, "Reprocess last file")
	registry.Register("batch", &batchCommandHandler{}, "Process all pending files simultaneously")
	registry.Register("last", &lastCommandHandler{}, "Show last created note")
	registry.Register("help", &helpCommandHandler{}, "Show available commands")
	registry.Register("mode", &modeCommandHandler{}, "Select processing mode")
	registry.Register("bots", &botsCommandHandler{}, "Select bot instance")
}

// Command handler types
type startCommandHandler struct{}
type helpCommandHandler struct{}
type langCommandHandler struct{}
type setProviderCommandHandler struct{}
type pidCommandHandler struct{}
type linkCommandHandler struct{}
type serviceStatusCommandHandler struct{}
type modelInfoCommandHandler struct{}
type pauseBotCommandHandler struct{}
type resumeBotCommandHandler struct{}
type ragCommandHandler struct{}
type modeCommandHandler struct{}
type botsCommandHandler struct{}

// Implement Handle methods
func (h *startCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is running. Use /help for commands."))
	return err
}

func (h *helpCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	helpText := `🤖 **Obsidian Bot Commands**

📁 **File Processing:**
/process - Process single staged file with AI (multi-strategy)
/batch - Process ALL pending files simultaneously with progress
/reprocess - Reprocess last file
/last - Show last created note

🤖 **AI Configuration:**
/setprovider - Change AI provider (Gemini, Google, DeepSeek, Groq, Cloudflare, etc.)
/mode - Select processing mode (Fast/Quality/Conservative/Experimental)
/bots - Choose bot instance (Main/Test/Dev/Backup)

📊 **Information:**
/stats - Show bot usage statistics

🔗 **Integrations:**
/webhook - Manage webhooks
/security - Manage security settings

🧪 **Debug & Testing:**
Test files available in debug/templates/
Run: ./debug/test_bot.sh [files|commands|scenarios]

💡 **How to use:**
1. Send multiple images/PDFs (auto-staged)
2. Type /batch for parallel processing with progress bars
3. Or use /process for single files with multi-strategy OCR
4. Try /setprovider, /mode, /bots for interactive features
5. Receive AI-powered Obsidian notes with proper embeddings

✨ **New Features:**
• Multi-strategy OCR (7+ algorithms to avoid bad results)
• Progress bars for batch processing
• Interactive keyboards (no typing required)
• AI-powered embeddings for RAG
• User state persistence across restarts
• Ready debug templates for testing

📝 Notes saved to: vault/Inbox/`

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, helpText))
	return err
}

func (h *setProviderCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	telemetry.Info("SetProvider command called", "args", message.CommandArguments(), "chat_id", message.Chat.ID)

	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		// Create inline keyboard with provider options
		var keyboard [][]tgbotapi.InlineKeyboardButton
		supportedProviders := []string{"gemini", "google", "deepseek", "groq", "cloudflare", "openrouter", "replicate", "together", "huggingface"}

		// Create rows of 2 buttons each
		for i := 0; i < len(supportedProviders); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			for j := 0; j < 2 && i+j < len(supportedProviders); j++ {
				provider := supportedProviders[i+j]
				button := tgbotapi.NewInlineKeyboardButtonData(
					strings.Title(provider),
					fmt.Sprintf("setprovider_%s", provider),
				)
				row = append(row, button)
			}
			keyboard = append(keyboard, row)
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("🤖 **Current AI Provider:** %s\n\nSelect a new provider:", strings.Title(state.Provider)))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

		_, err := cmdCtx.Bot.Send(msg)
		return err
	}

	// Handle text-based provider setting
	supportedProviders := []string{"gemini", "groq", "cloudflare", "openrouter", "replicate", "together", "huggingface"}
	providerLower := strings.ToLower(args)

	supported := false
	for _, provider := range supportedProviders {
		if providerLower == provider {
			supported = true
			break
		}
	}

	if supported {
		telemetry.Info("Setting provider", "provider", providerLower, "normalized", normalizeProviderName(providerLower))
		oldProvider := state.Provider
		state.UpdateProvider(providerLower)
		// Set the provider on the AI service
		if err := cmdCtx.AIService.SetProvider(normalizeProviderName(providerLower)); err != nil {
			telemetry.Error("Failed to set provider on AI service", "error", err)
			_, sendErr := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Provider set but AI service error: %v", err)))
			return sendErr
		}
		telemetry.Info("Provider set successfully", "provider", providerLower)

		// Trigger webhook event
		go TriggerProviderEvent(message.From.ID, oldProvider, providerLower)

		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Provider set to: %s", providerLower)))
		return err
	}

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Unsupported provider: %s\nSupported: gemini, groq, cloudflare, openrouter, replicate, together, huggingface", args)))
	return err
}

type clearCommandHandler struct{}

func (h *clearCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Clear conversation history
	state.ClearConversationHistory()

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🧹 Conversation history has been cleared. Starting fresh!"))
	return err
}

type statsCommandHandler struct{}

func (h *statsCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Get current stats
	messages, commands, images, docs, provUsage, cmdUsage, avgResponse := GetStats()

	// Calculate uptime
	uptime := time.Since(startTime)
	uptimeStr := formatDuration(uptime)

	// Build stats message
	statsMsg := "📊 **Bot Usage Statistics**\n\n"
	statsMsg += "📈 **Overall Metrics**\n"
	statsMsg += fmt.Sprintf("├─ Total Messages: %d\n", messages)
	statsMsg += fmt.Sprintf("├─ Total Commands: %d\n", commands)
	statsMsg += fmt.Sprintf("├─ Images Processed: %d\n", images)
	statsMsg += fmt.Sprintf("└─ Documents Processed: %d\n\n", docs)

	statsMsg += "⏱️ **Performance**\n"
	statsMsg += fmt.Sprintf("├─ Uptime: %s\n", uptimeStr)
	statsMsg += fmt.Sprintf("└─ Avg Response Time: %s\n\n", avgResponse.Round(time.Millisecond))

	if len(provUsage) > 0 {
		statsMsg += "🤖 **Provider Usage**\n"
		for provider, count := range provUsage {
			statsMsg += fmt.Sprintf("├─ %s: %d\n", strings.Title(provider), count)
		}
		statsMsg += "\n"
	}

	if len(cmdUsage) > 0 {
		statsMsg += "⌨️ **Command Usage**\n"
		for cmd, count := range cmdUsage {
			statsMsg += fmt.Sprintf("├─ /%s: %d\n", cmd, count)
		}
	}

	// Add user stats
	statsMsg += "\n👤 **Your Session**\n"
	statsMsg += fmt.Sprintf("├─ Messages: %d\n", state.MessageCount)
	statsMsg += fmt.Sprintf("├─ Provider: %s\n", strings.Title(state.Provider))
	statsMsg += fmt.Sprintf("└─ Language: %s\n", state.Language)

	msg := tgbotapi.NewMessage(message.Chat.ID, statsMsg)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := cmdCtx.Bot.Send(msg)
	return err
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 60
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// normalizeProviderName converts provider names to the canonical form used by AI service
func normalizeProviderName(provider string) string {
	switch strings.ToLower(provider) {
	case "gemini", "google":
		return "Gemini"
	case "groq":
		return "Groq"
	case "cloudflare":
		return "Cloudflare"
	case "openrouter":
		return "OpenRouter"
	case "replicate":
		return "Replicate"
	case "together":
		return "Together"
	case "huggingface", "hugging face":
		return "Hugging Face"
	default:
		return strings.Title(strings.ToLower(provider)) // Capitalize first letter
	}
}

func (h *serviceStatusCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	status := "Service Status\n\n"

	// Bot status
	status += "Bot: Online\n"

	// RAG System status
	if globalRAGChain != nil {
		status += "RAG: Active\n"
	} else {
		status += "RAG: Not initialized\n"
	}

	// Memory status
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memUsageMB := float64(memStats.Alloc) / 1024 / 1024
	status += fmt.Sprintf("Memory: %.1f MB\n", memUsageMB)

	// Goroutine status
	goroutines := runtime.NumGoroutine()
	status += fmt.Sprintf("Goroutines: %d\n", goroutines)

	// Uptime status
	uptime := time.Since(startTime)
	status += fmt.Sprintf("Uptime: %s\n", uptime.Round(time.Minute))

	// System load (basic check)
	if runtime.NumCPU() > 0 {
		status += fmt.Sprintf("CPU Cores: %d\n", runtime.NumCPU())
	}

	// Last check timestamp
	status += fmt.Sprintf("\nLast Check: %s", time.Now().Format("2006-01-02 15:04:05"))

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, status))
	return err
}

func (h *modelInfoCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	info := "AI Model Info\n\n"

	// Get user state using our management system
	userState := GetUserState(message.From.ID)

	info += fmt.Sprintf("Provider: %s\n", userState.Provider)

	// Model information based on provider
	switch userState.Provider {
	case "openai":
		info += "Model: gpt-4-turbo\n"
		info += "Capabilities: Text generation, analysis, RAG\n"
		info += "Speed: Fast\n"
	case "anthropic":
		info += "Model: claude-3-sonnet\n"
		info += "Capabilities: Text generation, analysis, RAG\n"
		info += "Speed: Fast\n"
	case "google":
		info += "Model: gemini-pro\n"
		info += "Capabilities: Text generation, analysis, RAG\n"
		info += "Speed: Fast\n"
	case "local":
		info += "Model: Local LLM\n"
		info += "Capabilities: Text generation, analysis, RAG\n"
		info += "Speed: Variable\n"
	default:
		info += "Model: Unknown\n"
		info += "Capabilities: Basic\n"
		info += "Speed: Unknown\n"
	}

	// RAG System info
	if globalRAGChain != nil {
		info += "\nRAG: Active\n"
		info += "Document Processing: Enabled\n"
		info += "Context Retrieval: Active\n"
	} else {
		info += "\nRAG: Inactive\n"
	}

	// Language settings
	info += fmt.Sprintf("\nLanguage: %s\n", userState.Language)

	// Performance notes
	info += "\nPerformance:\n"
	info += "Response time varies\n"
	info += "RAG queries may be slower\n"

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, info))
	return err
}

func (h *pauseBotCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Check if bot is already paused
	if botPaused {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is already paused"))
		return err
	}

	// Set the global pause flag
	botPaused = true

	// Send confirmation
	msg := "Bot Paused\n"
	msg += "Bot will not respond to messages.\n"
	msg += "Use /resume_bot to restart."

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg))
	return err
}

func (h *resumeBotCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	// Check if bot is already running
	if !botPaused {
		_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Bot is already running"))
		return err
	}

	// Clear the global pause flag
	botPaused = false

	// Send confirmation
	msg := "Bot Resumed\n"
	msg += "Bot will respond to messages.\n"
	msg += "Use /pause_bot to pause."

	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg))
	return err
}

type webhookCommandHandler struct{}

func (h *webhookCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	msg := "Webhook Management\n\nWebhooks are configured via external integrations."
	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg))
	return err
}

type securityCommandHandler struct{}

func (h *securityCommandHandler) Handle(ctx context.Context, message *tgbotapi.Message, state *UserState, cmdCtx *CommandContext) error {
	msg := "Security Settings\n\nConfigure bot security options here."
	_, err := cmdCtx.Bot.Send(tgbotapi.NewMessage(message.Chat.ID, msg))
	return err
}
