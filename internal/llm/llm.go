package llm

import (
	"github.com/leocomelli/aigile/internal/prompt"
)

// LLMProvider is the interface for LLM providers
type LLMProvider interface {
	GenerateContent(itemType prompt.ItemType, parent, context string, criteria []string, language string, generateTasks bool) (*GeneratedContent, error)
}

// GeneratedContent represents the structured output from the LLM
type GeneratedContent struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	SuggestedTasks     []string `json:"suggested_tasks"`
	Type               string   `json:"type"`
}

// Config represents the configuration for the LLM provider
type Config struct {
	Provider string
	APIKey   string
	Model    string
	Endpoint string // For Azure OpenAI
}
