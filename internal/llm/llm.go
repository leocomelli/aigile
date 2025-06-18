// Package llm provides interfaces and implementations for Large Language Model providers.
package llm

import (
	"github.com/leocomelli/aigile/internal/prompt"
)

// Provider defines the interface for Large Language Model providers used to generate content.
type Provider interface {
	GenerateContent(itemType prompt.ItemType, parent, context string, criteria []string, language string, generateTasks bool) (*GeneratedContent, error)
}

// GeneratedContent represents the structured output returned by the LLM provider.
type GeneratedContent struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	SuggestedTasks     []string `json:"suggested_tasks"`
	Type               string   `json:"type"`
}

// Config holds the configuration parameters for the LLM provider.
type Config struct {
	Provider string
	APIKey   string
	Model    string
	Endpoint string // For Azure OpenAI
}
