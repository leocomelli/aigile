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

func (m *mockIssuesService) Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error) {
	args := m.Called(ctx, owner, repo, number, issue)
	return args.Get(0).(*github.Issue), args.Get(1).(*github.Response), args.Error(2)
}

type mockRepositoriesService struct {
	mock.Mock
}

func (m *mockRepositoriesService) Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error) {
	args := m.Called(ctx, owner, repo)
	return args.Get(0).(*github.Repository), args.Get(1).(*github.Response), args.Error(2)
}

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestGitHubProvider_CreateIssue_Success(t *testing.T) {
	// Arrange
	mockIssues := new(mockIssuesService)
	client := github.NewClient(nil)
	provider := &GitHubProvider{
		issues: mockIssues,
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
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
	createdIssue, err := provider.CreateIssue("Test Issue", "Test Description", []string{"bug"}, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdIssue)
	assert.Equal(t, issueNumber, *createdIssue.Number)
	mockIssues.AssertExpectations(t)
}

func TestGitHubProvider_CreateIssue_WithProject(t *testing.T) {
	// Arrange
	mockIssues := new(mockIssuesService)
	client := github.NewClient(nil)
	provider := &GitHubProvider{
		issues: mockIssues,
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
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
		mock.Anything,
	).Return(expectedIssue, mockResponse, nil)

	project := &ProjectInfo{
		ProjectNumber: 1,
		ProjectOwner:  "testowner",
	}

	// Act
	createdIssue, err := provider.CreateIssue("Test Issue", "Test Description", []string{"bug"}, project)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdIssue)
	assert.Equal(t, issueNumber, *createdIssue.Number)
	mockIssues.AssertExpectations(t)
}

func TestGitHubProvider_CreateIssue_Error(t *testing.T) {
	// Arrange
	mockIssues := new(mockIssuesService)
	client := github.NewClient(nil)
	provider := &GitHubProvider{
		issues: mockIssues,
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
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
	createdIssue, err := provider.CreateIssue("", "Test Description", []string{"bug"}, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, createdIssue)
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
	provider, err := NewGitHubProvider(config)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "testowner", provider.owner)
	assert.Equal(t, "testrepo", provider.repo)
	assert.NotNil(t, provider.issues)
	assert.NotNil(t, provider.repos)
	assert.NotNil(t, provider.client)
}

type mockTransport struct {
	mock *mockHTTPClient
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.mock.Do(req)
}
