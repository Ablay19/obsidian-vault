package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	attachmentsDir = "./attachments"
	vaultDir       = "./vault"
)

type UserState struct {
	Language          string
	LastProcessedFile string
	LastCreatedNote   string
}

var (
	userStates = make(map[int64]*UserState)
	stateMutex sync.RWMutex
)

func getUserState(userID int64) *UserState {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	if state, exists := userStates[userID]; exists {
		return state
	}

	// Create new state with defaults
	state := &UserState{
		Language: "French",
	}
	userStates[userID] = state
	return state
}

// Bot interface for testing
type Bot interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error)
}

// Production adapter
type TelegramBot struct {
	*tgbotapi.BotAPI
}

func (t *TelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return t.BotAPI.Send(c)
}

func (t *TelegramBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return t.BotAPI.Request(c)
}

func (t *TelegramBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return t.BotAPI.GetFile(config)
}

func main() {
	startHealthServer()
	stats.load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	ctx := context.Background()
	aiService := NewAIService(ctx)

	os.MkdirAll(attachmentsDir, 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Inbox"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Attachments"), 0755)

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot := &TelegramBot{botAPI}

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start the bot and see help"},
		{Command: "help", Description: "Show help message"},
		{Command: "stats", Description: "Show usage statistics"},
		{Command: "lang", Description: "Set AI language (e.g. /lang English)"},
		{Command: "last", Description: "Show last created note"},
		{Command: "reprocess", Description: "Reprocess last sent file"},
		{Command: "switchkey", Description: "Switch to next Gemini API key"},
		{Command: "setprovider", Description: "Set AI provider (e.g. /setprovider Groq)"},
	}
	config := tgbotapi.NewSetMyCommands(commands...)
	_, err = bot.Request(config)
	if err != nil {
		log.Printf("Error setting bot commands: %v", err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Send images or PDFs...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Photo != nil {
			go handlePhoto(bot, update.Message, aiService, token)
		}

		if update.Message.Document != nil {
			go handleDocument(bot, update.Message, aiService, token)
		}

		if update.Message.IsCommand() || update.Message.Text != "" {
			go handleCommand(bot, update.Message, aiService)
		}
	}
}

func handleCommand(bot Bot, message *tgbotapi.Message, aiService *AIService) {
	state := getUserState(message.From.ID)

	if !message.IsCommand() {
		if aiService == nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
			return
		}

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Thinking..."))

		var responseText string
		var mu sync.Mutex
		var lastEdit time.Time

		streamCallback := func(chunk string) {
			mu.Lock()
			responseText += chunk
			mu.Unlock()

			if time.Since(lastEdit) > 1*time.Second {
				lastEdit = time.Now()
			}
		}

		prompt := fmt.Sprintf("Respond in %s. User message: %s", state.Language, message.Text)
		fullResponse, err := aiService.GenerateContent(context.Background(), prompt, nil, ModelFlashSearch, streamCallback)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Sorry, I had trouble thinking: "+err.Error()))
			return
		}

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, fullResponse))
		return
	}

	switch message.Command() {
	case "start", "help":
		msg := tgbotapi.NewMessage(message.Chat.ID,
			"🤖 Bot active! Send images/PDFs for processing.\n\nCommands:\n/stats - Statistics\n/last - Show last created note\n/reprocess - Reprocess last file\n/lang - Set AI language (e.g. /lang English)\n/switchkey - Switch to next API key (Gemini only)\n/setprovider - Set AI provider (e.g. /setprovider Groq)\n/help - This message")
		bot.Send(msg)

	case "stats":
		stats.mu.Lock()
		statsText := fmt.Sprintf("📊 Stats\n\nTotal: %d\nImages: %d\nPDFs: %d\n\nCategories:\n",
			stats.TotalFiles, stats.ImageCount, stats.PDFCount)
		for cat, count := range stats.Categories {
			statsText += fmt.Sprintf("• %s: %d\n", cat, count)
		}
		stats.mu.Unlock()
		msg := tgbotapi.NewMessage(message.Chat.ID, statsText)
		bot.Send(msg)

	case "last":
		var text string
		if state.LastCreatedNote == "" {
			text = "No note has been created yet."
		} else {
			text = "Last created note: " + state.LastCreatedNote
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		bot.Send(msg)

	case "reprocess":
		if state.LastProcessedFile == "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "No file has been processed yet."))
			return
		}

		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Reprocessing..."))

		fileType := ""
		if strings.HasSuffix(state.LastProcessedFile, ".jpg") {
			fileType = "image"
		} else if strings.HasSuffix(state.LastProcessedFile, ".pdf") {
			fileType = "pdf"
		}

		dummyMessage := &tgbotapi.Message{Caption: "Reprocessed", From: message.From}
		createObsidianNote(state.LastProcessedFile, fileType, dummyMessage, aiService, bot, message.Chat.ID, 0, "")

	case "lang":
		lang := message.CommandArguments()
		if lang == "" {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Current language is "+state.Language+".\nUsage: /lang <language>"))
		} else {
			state.Language = lang
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Language set to: "+lang))
		}

	case "switchkey":
		if aiService != nil {
			if geminiProvider, ok := aiService.activeProvider.(*GeminiProvider); ok {
				nextKeyIndex := geminiProvider.SwitchKey()
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Switched to API Key #%d", nextKeyIndex+1)))
			} else {
				bot.Send(tgbotapi.NewMessage(message.Chat.ID, "The /switchkey command is only available for the Gemini provider."))
			}
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
		}

	case "setprovider":
		if aiService == nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, "AI service is not available."))
			return
		}

		providerName := message.CommandArguments()
		if providerName == "" {
			currentProvider := aiService.GetActiveProviderName()
			availableProviders := aiService.GetAvailableProviders()
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Current provider is %s.\nAvailable providers: %s\nUsage: /setprovider <provider>", currentProvider, strings.Join(availableProviders, ", "))))
			return
		}

		err := aiService.SetProvider(providerName)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Error setting provider: %s", err)))
		} else {
			bot.Send(tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("AI provider set to: %s", providerName)))
		}
	}
}

func handlePhoto(bot Bot, message *tgbotapi.Message, aiService *AIService, token string) {
	state := getUserState(message.From.ID)
	updateActivity()
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Processing image..."))

	photos := message.Photo
	if len(photos) == 0 {
		return
	}

	photo := photos[len(photos)-1]
	filename := downloadFile(bot, photo.FileID, "jpg", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Download failed."))
		return
	}

	if isDuplicate(filename) {
		os.Remove(filename)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Duplicate file detected."))
		return
	}

	state.LastProcessedFile = filename
	createObsidianNote(filename, "image", message, aiService, bot, message.Chat.ID, 0, token)
}

func handleDocument(bot Bot, message *tgbotapi.Message, aiService *AIService, token string) {
	state := getUserState(message.From.ID)
	updateActivity()

	doc := message.Document
	if doc == nil {
		return
	}

	if doc.MimeType != "application/pdf" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ PDFs only")
		bot.Send(msg)
		return
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "🤖 Processing PDF..."))

	filename := downloadFile(bot, doc.FileID, "pdf", token)
	if filename == "" {
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Download failed."))
		return
	}

	if isDuplicate(filename) {
		os.Remove(filename)
		bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⚠️ Duplicate file detected."))
		return
	}

	state.LastProcessedFile = filename
	createObsidianNote(filename, "pdf", message, aiService, bot, message.Chat.ID, 0, token)
}

func downloadFile(bot Bot, fileID, ext, token string) string {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("GetFile error: %v", err)
		return ""
	}

	resp, err := http.Get(file.Link(token))
	if err != nil {
		log.Printf("HTTP error: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Bad response: %d", resp.StatusCode)
		return ""
	}

	filename := fmt.Sprintf("%s.%s", time.Now().Format("20060102_150405"), ext)
	fullPath := filepath.Join(attachmentsDir, filename)

	os.MkdirAll(attachmentsDir, 0755)

	out, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Create error: %v", err)
		return ""
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("Write error: %v", err)
		return ""
	}

	log.Printf("File saved: %s", fullPath)
	return fullPath
}

func createObsidianNote(filePath, fileType string, message *tgbotapi.Message, aiService *AIService, bot Bot, chatID int64, messageID int, token string) {
	state := getUserState(message.From.ID)
	var responseText string
	var mu sync.Mutex
	var lastEdit time.Time

	streamCallback := func(chunk string) {
		mu.Lock()
		responseText += chunk
		mu.Unlock()

		if time.Since(lastEdit) > 1*time.Second {
			lastEdit = time.Now()
		}
	}

	processed := processFileWithAI(filePath, fileType, aiService, streamCallback, state.Language)
	stats.recordFile(fileType, processed.Category)

	finalText := processed.Summary
	if finalText == "" {
		finalText = "Processing complete."
	}
	bot.Send(tgbotapi.NewMessage(chatID, finalText+"\n\n✅ Note created."))

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	baseName := filepath.Base(filePath)

	caption := message.Caption
	if caption == "" {
		caption = "No caption"
	}

	tagsStr := strings.Join(processed.Tags, ", ")
	if tagsStr == "" {
		tagsStr = "unprocessed"
	}

	content := fmt.Sprintf(`---
source: telegram-bot
type: %s
category: %s
date: %s
language: %s
confidence: %.2f
tags: [%s]
ai_provider: %s
---

# %s - %s

**Received:** %s  
**Caption:** %s  
**Classification:** %s (%.0f%%)  
**Language:** %s

## File

![[%s]]

## AI Summary

%s

## Key Topics

%s

## Extracted Content

%s

## Review Questions

%s

## Notes

(Add your notes here)

---
*AI-powered by %s*
`,
		fileType,
		processed.Category,
		timestamp,
		processed.Language,
		processed.Confidence,
		tagsStr,
		processed.AIProvider,
		strings.Title(processed.Category),
		fileType,
		timestamp,
		caption,
		processed.Category,
		processed.Confidence*100,
		processed.Language,
		baseName,
		processed.Summary,
		strings.Join(processed.Topics, ", "),
		formatExtractedText(processed.Text),
		formatQuestions(processed.Questions),
		processed.AIProvider)

	noteName := fmt.Sprintf("%s_%s_%s.md",
		time.Now().Format("20060102_150405"), processed.Category, fileType)
	notePath := filepath.Join(vaultDir, "Inbox", noteName)

	os.WriteFile(notePath, []byte(content), 0644)
	log.Printf("Created note: %s", notePath)

	// Convert markdown to HTML
	htmlContent := convertMarkdownToHTML(content)

	// Convert HTML to PDF
	pdfData, err := convertHTMLToPDF(htmlContent)
	if err != nil {
		log.Printf("Error converting to PDF: %v", err)
		// Send the markdown file as a fallback
		doc := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FilePath(notePath))
		bot.Send(doc)
		return
	}

	// Send the PDF file
	pdfFile := tgbotapi.FileBytes{
		Name:  strings.Replace(noteName, ".md", ".pdf", 1),
		Bytes: pdfData,
	}
	doc := tgbotapi.NewDocument(message.Chat.ID, pdfFile)
	bot.Send(doc)

	state.LastCreatedNote = notePath
}

func formatExtractedText(text string) string {
	if len(text) == 0 {
		return "*No text extracted*"
	}
	if len(text) > 1000 {
		text = text[:1000] + "\n\n*(truncated)*"
	}
	return "```\n" + text + "\n```"
}
func formatQuestions(questions []string) string {
	if len(questions) == 0 {
		return "*No questions generated*"
	}

	result := ""
	for i, q := range questions {
		result += fmt.Sprintf("%d. %s\n", i+1, q)
	}
	return result
}
