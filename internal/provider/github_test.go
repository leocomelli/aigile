package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockIssuesService struct {
	mock.Mock
}

func (m *mockIssuesService) Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	args := m.Called(ctx, owner, repo, issue)
	return args.Get(0).(*github.Issue), args.Get(1).(*github.Response), args.Error(2)
}

func TestGitHubProvider_CreateIssue_Success(t *testing.T) {
	// Arrange
	mockIssues := new(mockIssuesService)
	provider := &GitHubProvider{
		issues: mockIssues,
		owner:  "testowner",
		repo:   "testrepo",
	}

	issueNumber := 1
	issueURL := "https://github.com/testowner/testrepo/issues/1"
	expectedIssue := &github.Issue{
		Number:  &issueNumber,
		HTMLURL: &issueURL,
	}

	mockResponse := &github.Response{
		Response: &http.Response{
			StatusCode: http.StatusCreated,
			Status:     "201 Created",
			Body:       io.NopCloser(bytes.NewBufferString("")),
		},
	}

	mockIssues.On("Create",
		mock.Anything,
		"testowner",
		"testrepo",
		mock.MatchedBy(func(issue *github.IssueRequest) bool {
			return *issue.Title == "Test Issue" &&
				*issue.Body == "Test Description" &&
				len(*issue.Labels) == 1 &&
				(*issue.Labels)[0] == "bug"
		}),
	).Return(expectedIssue, mockResponse, nil)

	// Act
	err := provider.CreateIssue("Test Issue", "Test Description", []string{"bug"})

	// Assert
	assert.NoError(t, err)
	mockIssues.AssertExpectations(t)
}

func TestGitHubProvider_CreateIssue_Error(t *testing.T) {
	// Arrange
	mockIssues := new(mockIssuesService)
	provider := &GitHubProvider{
		issues: mockIssues,
		owner:  "testowner",
		repo:   "testrepo",
	}

	errorBody := `{"message": "Validation Failed","errors": [{"resource": "Issue","field": "title","code": "missing_field"}]}`
	mockResponse := &github.Response{
		Response: &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Status:     "422 Unprocessable Entity",
			Body:       io.NopCloser(bytes.NewBufferString(errorBody)),
		},
	}

	mockIssues.On("Create",
		mock.Anything,
		"testowner",
		"testrepo",
		mock.Anything,
	).Return(&github.Issue{}, mockResponse, errors.New("validation failed"))

	// Act
	err := provider.CreateIssue("", "Test Description", []string{"bug"})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "422 Unprocessable Entity")
	assert.Contains(t, err.Error(), errorBody)
	mockIssues.AssertExpectations(t)
}

func TestGitHubProvider_New(t *testing.T) {
	// Arrange
	config := GitHubConfig{
		Token: "test-token",
		Owner: "testowner",
		Repo:  "testrepo",
	}

	// Act
	provider := NewGitHubProvider(config)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "testowner", provider.owner)
	assert.Equal(t, "testrepo", provider.repo)
	assert.NotNil(t, provider.issues)
}
