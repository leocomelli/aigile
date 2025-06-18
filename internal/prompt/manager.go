// Package prompt gerencia templates e utilitários para geração de prompts para LLMs.
package prompt

import (
	"fmt"
	"strings"
)

// Manager handles the prompts for different item types
type Manager struct {
	prompts map[ItemType]string
}

// NewManager creates a new prompt manager with default prompts
func NewManager() *Manager {
	return &Manager{
		prompts: map[ItemType]string{
			UserStory: `
You are an Agile development expert specialized in writing well-structured and detailed User Stories following all industry best practices.

Objective:
Generate a detailed, clear, and well-written User Story, following the Agile format below:

Title: In the format "As a [role], I want [goal]"
Description: In the format "As a [persona], I want [feature] so that [benefit]"
Acceptance Criteria: Written using the Gherkin format (Given / When / Then)
(Optional) Suggested tasks: A list of implementation tasks written in clear and actionable language

Input parameters:
Parent: {{.Parent}}
Context provided by the user: {{.Context}}
Output language: {{.Language}}
Generate task suggestions?: {{.GenerateTasks}}
Output format: Return the User Story strictly in the following JSON structure:
{
  "type": "User Story",
  "title": "As a [role], I want [goal]",
  "description": "As a [persona], I want [feature] so that [benefit]",
  "acceptance_criteria": [
    "Given [initial context] When [action] Then [outcome]",
    "Given [initial context] When [action] Then [outcome]"
  ],
  "suggested_tasks": [
    "Task 1",
    "Task 2"
  ]
}
Mandatory rules:
The content must follow the language defined in the {language} parameter.
If the {generate_tasks} parameter is false, the "suggested_tasks" array must be empty.
Be highly descriptive and detailed, especially in the description and acceptance_criteria fields.
Always use the provided context as the main source for generating the User Story.
Do not include any explanations, comments, or instructional text in the output. Only return the pure JSON result.
`,
		},
	}
}

// GetPrompt returns the prompt string for the given item type and context, filling in template variables.
func (m *Manager) GetPrompt(itemType ItemType, parent, context string, criteria []string, language string, generateTasks bool) (string, error) {
	promptTemplate, ok := m.prompts[itemType]
	if !ok {
		return "", fmt.Errorf("invalid item type: %s", itemType)
	}

	// Replace template variables
	prompt := strings.ReplaceAll(promptTemplate, "{{.Parent}}", parent)
	prompt = strings.ReplaceAll(prompt, "{{.Context}}", context)
	prompt = strings.ReplaceAll(prompt, "{{.Criteria}}", strings.Join(criteria, ", "))
	prompt = strings.ReplaceAll(prompt, "{{.Language}}", language)
	prompt = strings.ReplaceAll(prompt, "{{.GenerateTasks}}", fmt.Sprintf("%v", generateTasks))

	// Add common instructions for JSON output
	prompt += "\n\nIMPORTANT:\n" +
		"1. Provide the response in valid JSON format only\n" +
		"2. Do not include any explanations or additional text outside the JSON structure\n" +
		"3. Ensure all JSON fields are properly escaped\n" +
		"4. Keep the response focused and concise"

	return prompt, nil
}

// SetPrompt allows customizing the prompt template for a specific item type.
func (m *Manager) SetPrompt(itemType ItemType, prompt string) error {
	if !itemType.IsValid() {
		return fmt.Errorf("invalid item type: %s", itemType)
	}
	m.prompts[itemType] = prompt
	return nil
}
