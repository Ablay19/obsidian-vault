package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"obsidian-automation/internal/ai"
	"obsidian-automation/internal/database"
	"obsidian-automation/internal/pipeline"
	"obsidian-automation/internal/state"
	"obsidian-automation/internal/status"
	"obsidian-automation/internal/util"
	"os"
	"strconv"
)

// --- Specific Command Handlers ---

const helpMessage = "🤖 Bot active! Send images/PDFs for processing.\n\nCommands:\n/process - Process staged file\n/stats - Statistics\n/last - Show last created note\n/reprocess - Reprocess last file\n/lang - Set AI language (e.g. /lang English)\n/setprovider - Set AI provider (Dynamic Menu)\n/modelinfo - Show AI model information\n/help - This message"

type startCommandHandler struct{}

func (h *startCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	msg := tgbotapi.NewMessage(message.Chat.ID, helpMessage)
	sent, _ := bot.Send(msg)
	database.SaveMessage(message.From.ID, message.Chat.ID, sent.MessageID, "out", "text", msg.Text, "")
}

type helpCommandHandler struct{}

func (h *helpCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	msg := tgbotapi.NewMessage(message.Chat.ID, helpMessage)
	sent, _ := bot.Send(msg)
	database.SaveMessage(message.From.ID, message.Chat.ID, sent.MessageID, "out", "text", msg.Text, "")
}

type langCommandHandler struct{}

func (h *langCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
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
}

type setProviderCommandHandler struct {
	aiService ai.AIServiceInterface
}

func (h *setProviderCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	if h.aiService == nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
		return
	}

	arg := message.CommandArguments()
	if arg != "" {
		if err := h.aiService.SetProvider(arg); err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ "+err.Error()))
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "✅ AI provider set to: "+arg))
		}
		return
	}

	availableProviders := h.aiService.GetAvailableProviders()
	if len(availableProviders) == 0 {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No AI providers are configured. Please check your .env file."))
		return
	}

	healthyProviders := h.aiService.GetHealthyProviders(context.Background())
	healthyMap := make(map[string]bool)
	for _, p := range healthyProviders {
		healthyMap[p] = true
	}

	currentProviderName := h.aiService.GetActiveProviderName()

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, p := range availableProviders {
		statusIcon := "❌"
		if healthyMap[p] {
			statusIcon = "🟢"
		}

		label := fmt.Sprintf("%s %s", statusIcon, p)
		if p == currentProviderName {
			label = "✅ " + p
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "setprovider:"+p),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh Status", "refresh_providers"),
	))
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Current AI provider: *%s*\n\nSelect a provider below (🟢=Healthy, ❌=Error/Expired):", currentProviderName))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	bot.Send(msg)
}

type statsCommandHandler struct {
	rcm *state.RuntimeConfigManager
}

func (h *statsCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	statsData := status.GetStats(h.rcm)
	var statsText strings.Builder
	statsText.WriteString("📊 *Bot Statistics*\n\n")
	statsText.WriteString(fmt.Sprintf("• *Total Files Processed:* %d\n", statsData.TotalFiles))
	statsText.WriteString(fmt.Sprintf("• *Total Images Processed:* %d\n", statsData.ImageFiles))
	statsText.WriteString(fmt.Sprintf("• *Total PDFs Processed:* %d\n", statsData.PDFFiles))
	statsText.WriteString(fmt.Sprintf("• *Total AI Calls:* %d\n", statsData.AICalls))
	statsText.WriteString(fmt.Sprintf("• *Last Activity:* %s\n", util.FormatTime(statsData.LastActivity)))
	msg := tgbotapi.NewMessage(message.Chat.ID, statsText.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

type lastCommandHandler struct{}

func (h *lastCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	if state.LastCreatedNote != "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Last created note: %s", state.LastCreatedNote)))
	} else {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No note has been created yet."))
	}
}

type reprocessCommandHandler struct {
	deps CommandDependencies
}

func (h *reprocessCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	if state.LastProcessedFile != "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, " reprocessing last file..."))
		fileType := "document"
		if strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".jpg") || strings.HasSuffix(strings.ToLower(state.LastProcessedFile), ".png") {
			fileType = "image"
		}
		createObsidianNote(context.Background(), bot, h.deps.AIService, message, state, state.LastProcessedFile, fileType, 0, "")
	} else {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No file to reprocess. Please send a file first."))
	}
}

type pidCommandHandler struct{}

func (h *pidCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Bot PID: %d", os.Getpid())))
}

type linkCommandHandler struct{}

func (h *linkCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	dashboardURL := os.Getenv("DASHBOARD_URL")
	if dashboardURL == "" {
		dashboardURL = "http://localhost:8080"
	}
	link := fmt.Sprintf("%s/api/v1/auth/telegram/webhook?id=%d", dashboardURL, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔗 *Link your Dashboard Account*\n\nClick the link below while logged into the web dashboard to sync your accounts:\n\n"+link)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

type serviceStatusCommandHandler struct {
	aiService ai.AIServiceInterface
	rcm       *state.RuntimeConfigManager
}

func (h *serviceStatusCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	statuses := status.GetServicesStatus(h.aiService, h.rcm)
	var sb strings.Builder
	sb.WriteString("📊 *Service Status*\n\n")
	for _, s := range statuses {
		sb.WriteString(fmt.Sprintf("• *%s:* %s\n", s.Name, s.Status))
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, sb.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

type modelInfoCommandHandler struct {
	aiService ai.AIServiceInterface
}

func (h *modelInfoCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	if h.aiService == nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
		return
	}
	infos := h.aiService.GetProvidersInfo()
	var infoText strings.Builder
	infoText.WriteString("📊 *AI Model Information*\n\n")
	for _, info := range infos {
		infoText.WriteString(fmt.Sprintf("• *Provider:* %s\n  *Model:* %s\n", info.ProviderName, info.ModelName))
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, infoText.String())
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

type pauseBotCommandHandler struct{}

func (h *pauseBotCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	status.SetPaused(true)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Bot is paused.")
	sent, _ := bot.Send(msg)
	database.SaveMessage(message.From.ID, message.Chat.ID, sent.MessageID, "out", "text", msg.Text, "")
}

type resumeBotCommandHandler struct{}

func (h *resumeBotCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	status.SetPaused(false)
	msg := tgbotapi.NewMessage(message.Chat.ID, "Bot is resumed.")
	sent, _ := bot.Send(msg)
	database.SaveMessage(message.From.ID, message.Chat.ID, sent.MessageID, "out", "text", msg.Text, "")
}

type processCommandHandler struct {
	ingestionPipeline *pipeline.Pipeline
}

func (h *processCommandHandler) Handle(bot Bot, message *tgbotapi.Message, state *UserState) {
	if !state.IsStaging {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Nothing to process. Send a file first."))
		return
	}
	statusMsg, _ := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Submitting to processing pipeline..."))

	args := message.CommandArguments()
	outputFormat := "pdf"
	if strings.Contains(args, "--output md") {
		outputFormat = "md"
	}
	gitCommit := strings.Contains(args, "--commit")

	fileBytes, err := os.ReadFile(state.PendingFile)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "❌ Failed to read staged file."))
		zap.S().Error("Read file error", "error", err)
		return
	}

	job := pipeline.Job{
		ID:           fmt.Sprintf("%d_%d", message.Chat.ID, time.Now().UnixNano()),
		Source:       "telegram",
		SourceID:     strconv.FormatInt(int64(message.MessageID), 10),
		Data:         fileBytes,
		ContentType:  pipeline.ContentTypeImage,
		ReceivedAt:   time.Now(),
		MaxRetries:   3,
		OutputFormat: outputFormat,
		GitCommit:    gitCommit,
		UserContext: pipeline.UserContext{
			UserID:   strconv.FormatInt(message.From.ID, 10),
			Language: state.Language,
		},
		Metadata: map[string]interface{}{
			"caption": state.PendingContext,
			"chat_id": message.Chat.ID,
		},
	}

	if state.PendingFileType == "pdf" {
		job.ContentType = pipeline.ContentTypePDF
	}

	if err := h.ingestionPipeline.Submit(job); err != nil {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("❌ Pipeline full/error: %v", err)))
	} else {
		bot.Request(tgbotapi.NewEditMessageText(message.Chat.ID, statusMsg.MessageID, "✅ Job queued. You will be notified when complete."))
	}

	state.IsStaging = false
	state.PendingFile = ""
	state.PendingContext = ""
}
