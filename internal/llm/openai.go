package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client  *openai.Client
	model   string
	prompts *prompt.Manager
}

func NewOpenAIProvider(config Config) *OpenAIProvider {
	client := openai.NewClient(config.APIKey)
	return &OpenAIProvider{
		client:  client,
		model:   config.Model,
		prompts: prompt.NewManager(),
	}
}

func (p *OpenAIProvider) GenerateContent(itemType prompt.ItemType, parent, ctx string, criteria []string, language string) (*GeneratedContent, error) {
	// Get the appropriate prompt for the item type
	promptText, err := p.prompts.GetPrompt(itemType, parent, ctx, criteria, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	resp, err := p.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: p.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert in agile methodologies and software development. Your task is to generate high-quality agile artifacts in JSON format.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: promptText,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Clean up the response to ensure it's valid JSON
	content := cleanJSONResponse(resp.Choices[0].Message.Content)

	// Parse the JSON response
	var result GeneratedContent
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Validate the required fields
	if err := validateGeneratedContent(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// cleanJSONResponse removes any non-JSON content from the response
func cleanJSONResponse(content string) string {
	// Find the first '{' and last '}'
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")

	if start == -1 || end == -1 || end < start {
		return content // Return original if no valid JSON markers found
	}

	// Extract only the JSON part
	return content[start : end+1]
}

// validateGeneratedContent ensures all required fields are present
func validateGeneratedContent(content *GeneratedContent) error {
	if content.Title == "" {
		return fmt.Errorf("title is required")
	}
	if content.Description == "" {
		return fmt.Errorf("description is required")
	}
	if content.Type == "" {
		return fmt.Errorf("type is required")
	}
	return nil
}
