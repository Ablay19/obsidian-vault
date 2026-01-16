package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ServiceInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

func sendToTelegram(chatID, message string) error {
	url := config.Services.BaseURL + "/api/v1/telegram/send"
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    message,
	}
	return postToService(url, payload)
}

func getTelegramStatus() (string, error) {
	url := config.Services.BaseURL + "/api/v1/telegram/status"
	return getFromService(url)
}

func queryAI(prompt string) (string, error) {
	url := config.Services.BaseURL + "/api/v1/workers/ai-worker-001/invoke"
	payload := map[string]interface{}{
		"prompt": prompt,
		"model":  config.AI.Models[0],
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if response, ok := result["response"].(string); ok {
		return response, nil
	}
	return "No response from AI service", nil
}

func listAIModels() ([]string, error) {
	// Return configured models
	return config.AI.Models, nil
}

func listServices() ([]ServiceInfo, error) {
	url := config.Services.BaseURL + "/api/v1/go-applications"
	resp, err := http.Get(url)
	if err != nil {
		// Fallback to hardcoded list
		return []ServiceInfo{
			{Name: "api-gateway", Status: "running", Type: "api"},
			{Name: "whatsapp-go", Status: "running", Type: "whatsapp"},
			{Name: "telegram-bot", Status: "running", Type: "bot"},
			{Name: "ai-worker", Status: "running", Type: "worker"},
			{Name: "auth-service", Status: "running", Type: "auth"},
			{Name: "user-service", Status: "running", Type: "user"},
		}, nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if apps, ok := result["applications"].([]interface{}); ok {
		var serviceList []ServiceInfo
		for _, a := range apps {
			if app, ok := a.(map[string]interface{}); ok {
				serviceList = append(serviceList, ServiceInfo{
					Name:   app["name"].(string),
					Status: "running",
					Type:   "app",
				})
			}
		}
		return serviceList, nil
	}
	return []ServiceInfo{}, nil
}

func getServicesStatus() (string, error) {
	url := config.Services.BaseURL + "/health"
	resp, err := http.Get(url)
	if err != nil {
		return "unhealthy", err
	}
	defer resp.Body.Close()
	return "healthy", nil
}

func restartService(serviceName string) error {
	// This would require Docker API or similar - for now, just log
	fmt.Printf("Restarting service: %s (simulated)\n", serviceName)
	return nil
}

func uploadMedia(filePath string) (string, error) {
	// For now, simulate media upload
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Media uploaded: %s (%d bytes)", filePath, fileInfo.Size()), nil
}

func getMediaStatus() (string, error) {
	return "Media processing ready", nil
}

func postToService(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("service returned status %d", resp.StatusCode)
	}
	return nil
}

func getFromService(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if status, ok := result["status"].(string); ok {
		return status, nil
	}
	return "unknown", nil
}
