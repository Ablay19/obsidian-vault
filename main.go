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
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	attachmentsDir = "./attachments"
	vaultDir       = "./vault"
)

func main() {
	startHealthServer()
	stats.load()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	ctx := context.Background()
	aiService := NewAIService(ctx)

	// Create directories first
	os.MkdirAll(attachmentsDir, 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Inbox"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Attachments"), 0755)
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	os.MkdirAll(attachmentsDir, 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Inbox"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "Attachments"), 0755)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is running. Send images or PDFs...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Photo != nil {
			handlePhoto(bot, update.Message, aiService)
		}

		if update.Message.Document != nil {
			handleDocument(bot, update.Message, aiService)
		}

		if update.Message.Text != "" {
			handleCommand(bot, update.Message)
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	switch message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(message.Chat.ID,
		"🤖 Bot active! Send images/PDFs.\n\nCommands:\n/stats - Statistics\n/help - This message")
		bot.Send(msg)

	case "/stats":
		stats.mu.Lock()
		statsText := fmt.Sprintf("📊 Stats\n\nTotal: %d\nImages: %d\nPDFs: %d\n\nCategories:\n",
		stats.TotalFiles, stats.ImageCount, stats.PDFCount)
		for cat, count := range stats.Categories {
			statsText += fmt.Sprintf("• %s: %d\n", cat, count)
		}
		stats.mu.Unlock()
		msg := tgbotapi.NewMessage(message.Chat.ID, statsText)
		bot.Send(msg)

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Bot running! Send images or PDFs.")
		bot.Send(msg)
	}
}

func handlePhoto(bot *tgbotapi.BotAPI, message *tgbotapi.Message, aiService *AIService) {
	updateActivity()
	photos := message.Photo
	if len(photos) == 0 {
		return
	}

	photo := photos[len(photos)-1]
	filename := downloadFile(bot, photo.FileID, "jpg")
	if filename == "" {
		return
	}

	if isDuplicate(filename) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Duplicate")
		bot.Send(msg)
		os.Remove(filename)
		return
	}

	createObsidianNote(filename, "image", message, aiService)
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ Image: %s", filepath.Base(filename)))
	bot.Send(msg)
}

func handleDocument(bot *tgbotapi.BotAPI, message *tgbotapi.Message, aiService *AIService) {
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

	filename := downloadFile(bot, doc.FileID, "pdf")
	if filename == "" {
		return
	}

	if isDuplicate(filename) {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Duplicate")
		bot.Send(msg)
		os.Remove(filename)
		return
	}

	createObsidianNote(filename, "pdf", message, aiService)
	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("✅ PDF: %s", filepath.Base(filename)))
	bot.Send(msg)
}
func downloadFile(bot *tgbotapi.BotAPI, fileID, ext string) string {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		log.Printf("GetFile error: %v", err)
		return ""
	}

	resp, err := http.Get(file.Link(bot.Token))
	if err != nil {
		log.Printf("HTTP error: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != 200 {
		log.Printf("Bad response: %d", resp.StatusCode)
		return ""
	}

	filename := fmt.Sprintf("%s.%s", time.Now().Format("20060102_150405"), ext)
	fullPath := filepath.Join(attachmentsDir, filename)

	// Ensure directory exists with correct permissions
	os.MkdirAll(attachmentsDir, 0755)

	out, err := os.Create(fullPath)
	if err != nil {
		log.Printf("Create error: %v", err)

		// Try alternative path
		fullPath = "/tmp/" + filename
		out, err = os.Create(fullPath)
		if err != nil {
			log.Printf("Fallback create error: %v", err)
			return ""
		}
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		log.Printf("Write error: %v", err)
		return ""
	}

	log.Printf("File saved: %s", fullPath)
	return fullPath
}

func createObsidianNote(filePath, fileType string, message *tgbotapi.Message, aiService *AIService) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	baseName := filepath.Base(filePath)

	caption := message.Caption
	if caption == "" {
		caption = "No caption"
	}

	processed := processFileWithAI(filePath, fileType, aiService)
	stats.recordFile(fileType, processed.Category)

	tagsStr := strings.Join(processed.Tags, ", ")
	if tagsStr == "" {
		tagsStr = "unprocessed"
	}

	content := fmt.Sprintf(`---
	source: whatsapp-telegram
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

	// Auto-organize if confidence is high
	if processed.Confidence > 0.7 && processed.Category != "general" {
		go func() {
			time.Sleep(500 * time.Millisecond)
			if err := organizeNote(notePath, processed.Category); err != nil {
				log.Printf("Organization error: %v", err)
			}
		}()
	}
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

