package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetPrompt(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name          string
		itemType      ItemType
		parent        string
		context       string
		language      string
		generateTasks bool
		wantErr       bool
	}{
		{
			name:          "User Story prompt with tasks",
			itemType:      UserStory,
			parent:        "FEAT-1",
			context:       "Process credit card payments",
			language:      "english",
			generateTasks: true,
			wantErr:       false,
		},
		{
			name:          "User Story prompt without tasks",
			itemType:      UserStory,
			parent:        "FEAT-1",
			context:       "Process credit card payments",
			language:      "english",
			generateTasks: false,
			wantErr:       false,
		},
		{
			name:     "Invalid type",
			itemType: "Invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.GetPrompt(tt.itemType, tt.parent, tt.context, nil, tt.language, tt.generateTasks)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Contains(t, got, "You are an Agile development expert")
			assert.Contains(t, got, "Generate a detailed, clear, and well-written User Story")
			assert.Contains(t, got, "Title: In the format \"As a [role], I want [goal]\"")
			assert.Contains(t, got, "Description: In the format \"As a [persona], I want [feature] so that [benefit]\"")
			assert.Contains(t, got, "Acceptance Criteria: Written using the Gherkin format (Given / When / Then)")
			assert.Contains(t, got, "Input parameters:")
			assert.Contains(t, got, "Parent: "+tt.parent)
			assert.Contains(t, got, "Context provided by the user: "+tt.context)
			assert.Contains(t, got, "Output language: "+tt.language)
			assert.Contains(t, got, "Generate task suggestions?: "+boolToString(tt.generateTasks))
			assert.Contains(t, got, "Output format: Return the User Story strictly in the following JSON structure:")
			assert.Contains(t, got, "\"type\": \"User Story\"")
			assert.Contains(t, got, "\"title\": \"As a [role], I want [goal]\"")
			assert.Contains(t, got, "\"description\": \"As a [persona], I want [feature] so that [benefit]\"")
			assert.Contains(t, got, "\"acceptance_criteria\": [")
			assert.Contains(t, got, "\"suggested_tasks\": [")
		})
	}
}

func TestManager_SetPrompt(t *testing.T) {
	manager := NewManager()

	// Test setting a new prompt
	newPrompt := "New prompt template"
	err := manager.SetPrompt(UserStory, newPrompt)
	assert.NoError(t, err)

	// Verify the prompt was set
	got, err := manager.GetPrompt(UserStory, "", "", nil, "english", false)
	assert.NoError(t, err)
	assert.Contains(t, got, newPrompt)

	// Test setting prompt for invalid type
	err = manager.SetPrompt("Invalid", newPrompt)
	assert.Error(t, err)
}

// boolToString converts a boolean value to its string representation ("true" or "false").
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
