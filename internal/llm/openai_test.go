package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/leocomelli/aigile/internal/prompt"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

type mockPromptManager struct {
	getPromptFunc func(prompt.ItemType, string, string, []string, string, bool) (string, error)
}

func (m *mockPromptManager) GetPrompt(itemType prompt.ItemType, parent, ctx string, criteria []string, language string, generateTasks bool) (string, error) {
	return m.getPromptFunc(itemType, parent, ctx, criteria, language, generateTasks)
}

func TestNewOpenAIProvider(t *testing.T) {
	provider := NewOpenAIProvider(Config{APIKey: "key", Model: "gpt"})
	assert.NotNil(t, provider)
	assert.Equal(t, "gpt", provider.model)
}

type mockOpenAIClient struct {
	createFunc func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.createFunc(ctx, req)
}

func TestOpenAIProvider_GenerateContent_Success(t *testing.T) {
	provider := &OpenAIProvider{
		client: &mockOpenAIClient{
			createFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
				return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{{
						Message: openai.ChatCompletionMessage{
							Content: `{"title":"T","description":"D","type":"User Story","acceptance_criteria":["A"],"suggested_tasks":["T1"]}`,
						},
					}},
				}, nil
			},
		},
		model: "gpt",
		prompts: &mockPromptManager{getPromptFunc: func(_ prompt.ItemType, _ string, _ string, _ []string, _ string, _ bool) (string, error) {
			return "prompt", nil
		}},
	}
	result, err := provider.GenerateContent(prompt.UserStory, "p", "c", []string{"a"}, "en", true)
	assert.NoError(t, err)
	assert.Equal(t, "T", result.Title)
	assert.Equal(t, "D", result.Description)
	assert.Equal(t, "User Story", result.Type)
	assert.Equal(t, []string{"A"}, result.AcceptanceCriteria)
	assert.Equal(t, []string{"T1"}, result.SuggestedTasks)
}

func TestOpenAIProvider_GenerateContent_PromptError(t *testing.T) {
	provider := &OpenAIProvider{
		client: &mockOpenAIClient{},
		model:  "gpt",
		prompts: &mockPromptManager{getPromptFunc: func(_ prompt.ItemType, _ string, _ string, _ []string, _ string, _ bool) (string, error) {
			return "", errors.New("prompt error")
		}},
	}
	result, err := provider.GenerateContent(prompt.UserStory, "p", "c", []string{"a"}, "en", true)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get prompt")
}

func TestOpenAIProvider_GenerateContent_APIError(t *testing.T) {
	provider := &OpenAIProvider{
		client: &mockOpenAIClient{
			createFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
				return openai.ChatCompletionResponse{}, errors.New("api error")
			},
		},
		model: "gpt",
		prompts: &mockPromptManager{getPromptFunc: func(_ prompt.ItemType, _ string, _ string, _ []string, _ string, _ bool) (string, error) {
			return "prompt", nil
		}},
	}
	result, err := provider.GenerateContent(prompt.UserStory, "p", "c", []string{"a"}, "en", true)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate content")
}

func TestOpenAIProvider_GenerateContent_InvalidJSON(t *testing.T) {
	provider := &OpenAIProvider{
		client: &mockOpenAIClient{
			createFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
				return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{{
						Message: openai.ChatCompletionMessage{Content: "not a json"},
					}},
				}, nil
			},
		},
		model: "gpt",
		prompts: &mockPromptManager{getPromptFunc: func(_ prompt.ItemType, _ string, _ string, _ []string, _ string, _ bool) (string, error) {
			return "prompt", nil
		}},
	}
	result, err := provider.GenerateContent(prompt.UserStory, "p", "c", []string{"a"}, "en", true)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}

func TestOpenAIProvider_GenerateContent_ValidationError(t *testing.T) {
	provider := &OpenAIProvider{
		client: &mockOpenAIClient{
			createFunc: func(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
				return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{{
						Message: openai.ChatCompletionMessage{Content: `{"title":"","description":"","type":"","acceptance_criteria":[]}`},
					}},
				}, nil
			},
		},
		model: "gpt",
		prompts: &mockPromptManager{getPromptFunc: func(_ prompt.ItemType, _ string, _ string, _ []string, _ string, _ bool) (string, error) {
			return "prompt", nil
		}},
	}
	result, err := provider.GenerateContent(prompt.UserStory, "p", "c", []string{"a"}, "en", true)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
}

func Test_cleanJSONResponse(t *testing.T) {
	json := `foo {"a":1} bar`
	out := cleanJSONResponse(json)
	assert.Equal(t, "{\"a\":1}", out)

	json = "nojson"
	out = cleanJSONResponse(json)
	assert.Equal(t, "nojson", out)
}

func Test_validateGeneratedContent(t *testing.T) {
	c := &GeneratedContent{Title: "t", Description: "d", Type: "User Story", AcceptanceCriteria: []string{"a"}}
	assert.NoError(t, validateGeneratedContent(c))

	c.Title = ""
	assert.Error(t, validateGeneratedContent(c))
	c.Title = "t"
	c.Description = ""
	assert.Error(t, validateGeneratedContent(c))
	c.Description = "d"
	c.Type = ""
	assert.Error(t, validateGeneratedContent(c))
	c.Type = "User Story"
	c.AcceptanceCriteria = nil
	assert.Error(t, validateGeneratedContent(c))
}
