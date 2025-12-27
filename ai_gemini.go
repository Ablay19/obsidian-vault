package main

import (
"fmt"
"log"
"os/exec"
"strings"
)

// Generate summary using Gemini CLI
func generateSummaryGemini(text string, category string) string {
if len(text) < 50 {
return "Text too short for summary"
}

// Limit text for performance
if len(text) > 3000 {
text = text[:3000]
}

prompt := fmt.Sprintf(`Summarize this %s content in 2-3 concise sentences. Focus on key concepts and actionable information. Be direct and clear.

Content:
%s`, category, text)

// Use gemini CLI
cmd := exec.Command("gemini", "chat", prompt)
output, err := cmd.Output()

if err != nil {
log.Printf("Gemini error: %v", err)
return "Summary unavailable"
}

summary := strings.TrimSpace(string(output))

// Clean up common artifacts
summary = strings.ReplaceAll(summary, "```", "")
summary = strings.TrimSpace(summary)

if len(summary) > 500 {
summary = summary[:500] + "..."
}

return summary
}

// Extract key topics using Gemini
func extractKeyTopicsGemini(text string) []string {
if len(text) < 50 {
return []string{}
}

if len(text) > 2000 {
text = text[:2000]
}

prompt := fmt.Sprintf(`Extract 3-5 key topics or concepts from this text. Return ONLY a comma-separated list of topics, nothing else. No explanations.

Text:
%s

Topics (comma-separated):`, text)

cmd := exec.Command("gemini", "chat", prompt)
output, err := cmd.Output()

if err != nil {
log.Printf("Gemini topics error: %v", err)
return []string{}
}

response := strings.TrimSpace(string(output))
topics := strings.Split(response, ",")

var cleaned []string
for _, topic := range topics {
topic = strings.TrimSpace(topic)
topic = strings.Trim(topic, ".-\"'")

// Remove common artifacts
if strings.Contains(strings.ToLower(topic), "here are") {
continue
}
if strings.Contains(topic, ":") {
parts := strings.Split(topic, ":")
if len(parts) > 1 {
topic = strings.TrimSpace(parts[1])
}
}

if len(topic) > 0 && len(topic) < 50 {
cleaned = append(cleaned, strings.ToLower(topic))
}
}

if len(cleaned) > 5 {
cleaned = cleaned[:5]
}

return cleaned
}

// Enhanced classification using Gemini
func classifyWithGemini(text string) string {
if len(text) < 30 {
return "general"
}

if len(text) > 1000 {
text = text[:1000]
}

prompt := fmt.Sprintf(`Classify this text into ONE category only. Choose from: physics, math, chemistry, biology, admin, document, note, general.

Return ONLY the category name, nothing else.

Text:
%s

Category:`, text)

cmd := exec.Command("gemini", "chat", prompt)
output, err := cmd.Output()

if err != nil {
return "general"
}

category := strings.TrimSpace(strings.ToLower(string(output)))

// Validate category
validCategories := map[string]bool{
"physics": true, "math": true, "chemistry": true, 
"biology": true, "admin": true, "document": true,
"note": true, "general": true,
}

if validCategories[category] {
return category
}

return "general"
}

// Generate smart questions about the content
func generateQuestions(text string) []string {
if len(text) < 100 {
return []string{}
}

if len(text) > 2000 {
text = text[:2000]
}

prompt := fmt.Sprintf(`Generate 3 important questions that could be answered by this text. Return ONLY the questions, one per line, numbered 1., 2., 3.

Text:
%s

Questions:`, text)

cmd := exec.Command("gemini", "chat", prompt)
output, err := cmd.Output()

if err != nil {
return []string{}
}

response := string(output)
lines := strings.Split(response, "\n")

var questions []string
for _, line := range lines {
line = strings.TrimSpace(line)

// Remove numbering
line = strings.TrimPrefix(line, "1.")
line = strings.TrimPrefix(line, "2.")
line = strings.TrimPrefix(line, "3.")
line = strings.TrimPrefix(line, "4.")
line = strings.TrimPrefix(line, "5.")
line = strings.TrimSpace(line)

if len(line) > 10 && len(line) < 200 && strings.Contains(line, "?") {
questions = append(questions, line)
}
}

if len(questions) > 3 {
questions = questions[:3]
}

return questions
}

// Check if Gemini CLI is available
func isGeminiAvailable() bool {
cmd := exec.Command("gemini", "chat", "test")
err := cmd.Run()
return err == nil
}
