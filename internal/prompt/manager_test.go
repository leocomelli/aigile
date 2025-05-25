package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GetPrompt(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		itemType    ItemType
		parent      string
		context     string
		criteria    []string
		language    string
		wantErr     bool
		checkPrompt func(t *testing.T, prompt string)
	}{
		{
			name:     "Epic prompt",
			itemType: Epic,
			parent:   "PROJ-1",
			context:  "Create a new payment system",
			criteria: []string{"Must support credit cards", "Must be PCI compliant"},
			language: "english",
			checkPrompt: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Parent: PROJ-1")
				assert.Contains(t, prompt, "Context: Create a new payment system")
				assert.Contains(t, prompt, "Must support credit cards")
				assert.Contains(t, prompt, "Must be PCI compliant")
				assert.Contains(t, prompt, "Business value and impact")
			},
		},
		{
			name:     "User Story prompt",
			itemType: UserStory,
			parent:   "FEAT-1",
			context:  "Process credit card payments",
			criteria: []string{"Support Visa", "Support Mastercard"},
			language: "english",
			checkPrompt: func(t *testing.T, prompt string) {
				assert.Contains(t, prompt, "Parent Feature: FEAT-1")
				assert.Contains(t, prompt, "As a [role], I want [goal]")
				assert.Contains(t, prompt, "Given/When/Then")
			},
		},
		{
			name:     "Invalid type",
			itemType: "Invalid",
			language: "english",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := manager.GetPrompt(tt.itemType, tt.parent, tt.context, tt.criteria, tt.language)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.checkPrompt(t, prompt)
		})
	}
}

func TestManager_SetPrompt(t *testing.T) {
	manager := NewManager()

	// Test setting a valid prompt
	customPrompt := "Custom prompt for {{.Context}}"
	err := manager.SetPrompt(Epic, customPrompt)
	require.NoError(t, err)

	prompt, err := manager.GetPrompt(Epic, "", "test context", nil, "english")
	require.NoError(t, err)
	assert.True(t, strings.Contains(prompt, "test context"))

	// Test setting an invalid type
	err = manager.SetPrompt("Invalid", "some prompt")
	assert.Error(t, err)
}
