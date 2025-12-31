package main

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MockBot struct {
	sentMessages []tgbotapi.Chattable
}

func (m *MockBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.sentMessages = append(m.sentMessages, c)
	return tgbotapi.Message{}, nil
}

func (m *MockBot) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.sentMessages = append(m.sentMessages, c)
	return nil, nil
}

func (m *MockBot) GetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	return tgbotapi.File{}, nil
}

func TestHandleLangCommand(t *testing.T) {
	// Create mock bot and AI service
	mockBot := &MockBot{}
	aiService := &AIService{
		providers: map[string]AIProvider{
			"Gemini": &MockGeminiProvider{},
		},
		activeProvider: &MockGeminiProvider{},
	}

	// Test setting language
	msg := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 123},
		From:      &tgbotapi.User{ID: 456},
		Text:      "/lang Spanish",
	}
	// Manually set the command for the test
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}

	handleCommand(mockBot, msg, aiService)

	// Check if the language was set
	state := getUserState(456)
	if state.Language != "Spanish" {
		t.Errorf("Expected language to be Spanish, but got %s", state.Language)
	}

	// Test getting language
	msg = &tgbotapi.Message{
		MessageID: 2,
		Chat:      &tgbotapi.Chat{ID: 123},
		From:      &tgbotapi.User{ID: 456},
		Text:      "/lang",
	}
	// Manually set the command for the test
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	handleCommand(mockBot, msg, aiService)

	// Check the sent message
	expectedMsg := "Current language is Spanish.\nUsage: /lang <language>"
	lastMsg := mockBot.sentMessages[len(mockBot.sentMessages)-1].(tgbotapi.MessageConfig)
	if lastMsg.Text != expectedMsg {
		t.Errorf("Expected message %q, but got %q", expectedMsg, lastMsg.Text)
	}
}

func TestHandleSetProviderCommand(t *testing.T) {
	// Create mock bot and AI service
	mockBot := &MockBot{}
	aiService := &AIService{
		providers: map[string]AIProvider{
			"Gemini": &MockGeminiProvider{},
			"Groq":   &MockGroqProvider{},
		},
		activeProvider: &MockGeminiProvider{},
	}

	// Test setting provider
	msg := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 123},
		From:      &tgbotapi.User{ID: 456},
		Text:      "/setprovider Groq",
	}
	// Manually set the command for the test
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 12}}

	handleCommand(mockBot, msg, aiService)

	// Check if the provider was set
	if aiService.GetActiveProviderName() != "Groq" {
		t.Errorf("Expected active provider to be Groq, but got %s", aiService.GetActiveProviderName())
	}

	// Check the sent message
	expectedMsg := "AI provider set to: Groq"
	lastMsg := mockBot.sentMessages[len(mockBot.sentMessages)-1].(tgbotapi.MessageConfig)
	if lastMsg.Text != expectedMsg {
		t.Errorf("Expected message %q, but got %q", expectedMsg, lastMsg.Text)
	}
}

func TestHandleLastCommand(t *testing.T) {
	// Create mock bot and AI service
	mockBot := &MockBot{}
	aiService := &AIService{
		providers: map[string]AIProvider{
			"Gemini": &MockGeminiProvider{},
		},
		activeProvider: &MockGeminiProvider{},
	}

	// Test /last when no note has been created
	msg := &tgbotapi.Message{
		MessageID: 1,
		Chat:      &tgbotapi.Chat{ID: 123},
		From:      &tgbotapi.User{ID: 456},
		Text:      "/last",
	}
	// Manually set the command for the test
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	handleCommand(mockBot, msg, aiService)

	// Check the sent message
	expectedMsg := "No note has been created yet."
	lastMsg := mockBot.sentMessages[len(mockBot.sentMessages)-1].(tgbotapi.MessageConfig)
	if lastMsg.Text != expectedMsg {
		t.Errorf("Expected message %q, but got %q", expectedMsg, lastMsg.Text)
	}

	// Set a last created note
	state := getUserState(456)
	state.LastCreatedNote = "/path/to/note.md"

	// Test /last when a note has been created
	msg = &tgbotapi.Message{
		MessageID: 2,
		Chat:      &tgbotapi.Chat{ID: 123},
		From:      &tgbotapi.User{ID: 456},
		Text:      "/last",
	}
	// Manually set the command for the test
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 5}}
	handleCommand(mockBot, msg, aiService)

	// Check the sent message
	expectedMsg = "Last created note: /path/to/note.md"
	lastMsg = mockBot.sentMessages[len(mockBot.sentMessages)-1].(tgbotapi.MessageConfig)
	if lastMsg.Text != expectedMsg {
		t.Errorf("Expected message %q, but got %q", expectedMsg, lastMsg.Text)
	}
}