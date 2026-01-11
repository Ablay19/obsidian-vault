package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"obsidian-automation/internal/ai"
)

// AIServiceLLM wraps the aiService to implement LangChain's LLM interface
type AIServiceLLM struct {
	AIService ai.AIServiceInterface
	ModelName string
}

func (l *AIServiceLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	// Use the aiService to generate content
	req := &ai.RequestModel{
		UserPrompt:  prompt,
		Temperature: 0.5,
	}
	var response strings.Builder
	err := l.AIService.Chat(ctx, req, func(chunk string) {
		response.WriteString(chunk)
	})
	if err != nil {
		return "", err
	}
	return response.String(), nil
}

func (l *AIServiceLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Simplified implementation for LangChain compatibility
	prompt := ""
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				prompt += textPart.Text
			}
		}
	}
	content, err := l.Call(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: content,
			},
		},
	}, nil
}

// RAGChain implements Retrieval-Augmented Generation
type RAGChain struct {
	retriever Retriever
	llm       llms.Model
	prompt    prompts.PromptTemplate
}

// NewRAGChain creates a new RAG chain
func NewRAGChain(retriever Retriever, llm llms.Model) (*RAGChain, error) {
	// Default RAG prompt template
	template := `You are a helpful assistant with access to relevant documents.
Use the following pieces of context to answer the question at the end.
If you don't know the answer, just say that you don't know, don't try to make up an answer.

Context:
{{.context}}

Question: {{.question}}

Answer:`

	prompt := prompts.NewPromptTemplate(template, []string{"context", "question"})

	return &RAGChain{
		retriever: retriever,
		llm:       llm,
		prompt:    prompt,
	}, nil
}

// Call executes the RAG chain
func (r *RAGChain) Call(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	question, ok := inputs["question"].(string)
	if !ok {
		return nil, fmt.Errorf("question input must be a string")
	}

	// Retrieve relevant documents
	documents, err := r.retriever.GetRelevantDocuments(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// Format context from retrieved documents
	contextParts := make([]string, len(documents))
	for i, doc := range documents {
		score := doc.Metadata["similarity_score"]
		contextParts[i] = fmt.Sprintf("[Source %d, Relevance: %.2f]\n%s", i+1, score, doc.PageContent)
	}
	context := strings.Join(contextParts, "\n\n")

	// Prepare prompt inputs
	promptInputs := map[string]any{
		"context":  context,
		"question": question,
	}

	// Generate response using LangChain
	llmChain := chains.NewLLMChain(r.llm, r.prompt)
	result, err := chains.Call(ctx, llmChain, promptInputs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Add source information to the result
	if len(documents) > 0 {
		sources := make([]map[string]interface{}, len(documents))
		for i, doc := range documents {
			sources[i] = map[string]interface{}{
				"content":          doc.PageContent,
				"similarity_score": doc.Metadata["similarity_score"],
				"metadata":         doc.Metadata,
			}
		}
		result["sources"] = sources
	}

	return result, nil
}

// Query is a convenience method for simple question answering
func (r *RAGChain) Query(ctx context.Context, question string) (string, error) {
	inputs := map[string]any{"question": question}
	result, err := r.Call(ctx, inputs)
	if err != nil {
		return "", err
	}

	answer, ok := result["text"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected result format")
	}

	return answer, nil
}

// QueryWithSources returns the answer along with source documents
func (r *RAGChain) QueryWithSources(ctx context.Context, question string) (string, []schema.Document, error) {
	inputs := map[string]any{"question": question}
	result, err := r.Call(ctx, inputs)
	if err != nil {
		return "", nil, err
	}

	answer, ok := result["text"].(string)
	if !ok {
		return "", nil, fmt.Errorf("unexpected result format")
	}

	sources, err := r.retriever.GetRelevantDocuments(ctx, question)
	if err != nil {
		return answer, nil, fmt.Errorf("failed to get sources: %w", err)
	}

	return answer, sources, nil
}
