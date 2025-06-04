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
			Epic: `You are creating an Epic to organize and group high-level functionalities in an agile backlog. The Epic should reflect a strategic business goal or a major value area.

Parent: {{.Parent}}

Generate a JSON response with the following structure:
{
    "type": "Epic",
    "title": "A clear and concise title that reflects the strategic goal",
    "description": "A broad description explaining the business objective and value",
    "parent": "{{.Parent}}",
    "additional_info": "Additional context including target users, timeline, and potential features"
}

Guidelines for the content:
- Title: Should be clear, concise, and reflect the strategic goal
- Description: Focus on business objective, value proposition, and high-level scope
- Additional Info: Include target users, timeline, and potential features/stories
- Business value and impact: Must be clearly stated

IMPORTANT: Generate the content in {{.Language}}

Context: {{.Context}}
Additional Requirements:
{{.Criteria}}`,

			Feature: `You are creating a Feature in an agile backlog. The feature should represent a valuable business functionality derived from an Epic.

Generate a JSON response with the following structure:
{
    "type": "Feature",
    "title": "A clear and descriptive title for the feature",
    "description": "Detailed description of the functionality and its business value",
    "parent": "{{.Parent}}",
    "acceptance_criteria": [
        "Criterion 1: Focus on business/functional requirements",
        "Criterion 2: Include performance or scalability requirements if applicable"
    ],
    "additional_info": "Technical considerations, integration points, and dependencies"
}

Guidelines for the content:
- Title: Should be clear and describe the feature's main functionality
- Description: Explain the functionality, business value, and target users
- Acceptance Criteria: List high-level validation criteria
- Additional Info: Include technical considerations and dependencies

IMPORTANT: Generate the content in {{.Language}}

Context: {{.Context}}`,

			UserStory: `You are a Product Owner writing a user story for an agile team. The story should represent a real user need based on a specific business context.

Parent Feature: {{.Parent}}

Generate a JSON response with the following structure:
{
    "type": "User Story",
    "title": "As a [role], I want [goal]",
    "description": "As a [persona], I want [feature] so that [benefit]",
    "parent": "{{.Parent}}",
    "acceptance_criteria": [
        "Given [initial context] When [action] Then [outcome]"
    ],
    "additional_info": "Important context, business rules, or technical constraints"
}

Guidelines for the content:
- Title: Follow the exact format "As a [role], I want [goal]"
- Description: Use the format "As a [persona], I want [feature] so that [benefit]"
- Acceptance Criteria: Use Given/When/Then format
- Additional Info: Include scope, constraints, and business rules

IMPORTANT: Generate the content in {{.Language}}

Context: {{.Context}}
Additional Criteria to Consider: {{.Criteria}}`,

			Task: `You are a developer breaking down a user story into technical tasks. The task should represent a clear and objective technical action.

Generate a JSON response with the following structure:
{
    "type": "Task",
    "title": "Clear and actionable technical title",
    "description": "Detailed technical description of the implementation",
    "parent": "{{.Parent}}",
    "additional_info": "Technical details including dependencies, testing requirements, and constraints"
}

Guidelines for the content:
- Title: Should be clear, actionable, and technical
- Description: Include specific implementation details and approach
- Additional Info: Include dependencies, testing requirements, and technical constraints

IMPORTANT: Generate the content in {{.Language}}

Context: {{.Context}}`,
		},
	}
}

// GetPrompt returns the prompt for the given item type
func (m *Manager) GetPrompt(itemType ItemType, parent, context string, criteria []string, language string) (string, error) {
	promptTemplate, ok := m.prompts[itemType]
	if !ok {
		return "", fmt.Errorf("invalid item type: %s", itemType)
	}

	// Replace template variables
	prompt := strings.ReplaceAll(promptTemplate, "{{.Parent}}", parent)
	prompt = strings.ReplaceAll(prompt, "{{.Context}}", context)
	prompt = strings.ReplaceAll(prompt, "{{.Criteria}}", strings.Join(criteria, ", "))
	prompt = strings.ReplaceAll(prompt, "{{.Language}}", language)

	// Add common instructions for JSON output
	prompt += "\n\nIMPORTANT:\n" +
		"1. Provide the response in valid JSON format only\n" +
		"2. Do not include any explanations or additional text outside the JSON structure\n" +
		"3. Ensure all JSON fields are properly escaped\n" +
		"4. Keep the response focused and concise"

	return prompt, nil
}

// SetPrompt allows customizing the prompt for a specific item type
func (m *Manager) SetPrompt(itemType ItemType, prompt string) error {
	if !itemType.IsValid() {
		return fmt.Errorf("invalid item type: %s", itemType)
	}
	m.prompts[itemType] = prompt
	return nil
}
