package llm

import (
	"github.com/leocomelli/aigile/internal/prompt"
)

// GeneratedContent represents the structured output from the LLM
type GeneratedContent struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	AdditionalInfo     string   `json:"additional_info,omitempty"`
	Parent             string   `json:"parent,omitempty"`
	Type               string   `json:"type"`
}

type LLMProvider interface {
	GenerateContent(itemType prompt.ItemType, parent, context string, criteria []string, language string) (*GeneratedContent, error)
}

type Config struct {
	Provider string
	APIKey   string
	Model    string
	Endpoint string // For Azure OpenAI
}
