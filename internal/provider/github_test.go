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

// mockIssuesService is a mock implementation of the IssuesService interface for testing.
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

// mockHTTPClient is a mock implementation of the HTTP client for testing GraphQL requests.
type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// TestGitHubProvider_CreateIssue_Success tests successful creation of a GitHub issue.
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
	assert.Equal(t, issueNumber, createdIssue.GetNumber())
	mockIssues.AssertExpectations(t)
}

// TestGitHubProvider_CreateIssue_WithProject tests creating a GitHub issue and adding it to a project.
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
		ProjectID:     "project-node-id",
	}

	// Act
	createdIssue, err := provider.CreateIssue("Test Issue", "Test Description", []string{"bug"}, project)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, createdIssue)
	assert.Equal(t, issueNumber, createdIssue.GetNumber())
	mockIssues.AssertExpectations(t)
	// We do not test the real GraphQL call, but we ensure the flow does not break
}

// TestGitHubProvider_CreateIssue_Error tests error handling when issue creation fails.
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

// TestGitHubProvider_New tests the creation of a new GitHubProvider instance.
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

// mockTransport is a mock implementation of http.RoundTripper for testing.
type mockTransport struct {
	mock *mockHTTPClient
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.mock.Do(req)
}

// TestGitHubProvider_GetProjectByName_Success tests fetching a project by name successfully.
func TestGitHubProvider_GetProjectByName_Success(t *testing.T) {
	// Arrange
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	graphqlResponse := `{"data":{"repositoryOwner":{"projectsV2":{"nodes":[{"id":"project-id-1","number":1,"title":"Project 1"}],"totalCount":1}}}}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(graphqlResponse)),
	}
	mockClient.On("Do", mock.Anything).Return(resp, nil)

	// Act
	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Project 1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "project-id-1", project.ProjectID)
	assert.Equal(t, 1, project.ProjectNumber)
}

// TestGitHubProvider_GetProjectByName_NotFound tests error handling when the project is not found.
func TestGitHubProvider_GetProjectByName_NotFound(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	graphqlResponse := `{"data":{"repositoryOwner":{"projectsV2":{"nodes":[{"id":"project-id-1","number":1,"title":"Project 1"}],"totalCount":1}}}}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(graphqlResponse)),
	}
	mockClient.On("Do", mock.Anything).Return(resp, nil)

	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Nonexistent Project")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "project not found")
}

// TestGitHubProvider_GetProjectByName_RequestError tests error handling for request errors in GetProjectByName.
func TestGitHubProvider_GetProjectByName_RequestError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	// Em vez de retornar nil, retorne um *http.Response vazio para evitar panic
	emptyResp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockClient.On("Do", mock.Anything).Return(emptyResp, errors.New("request failed"))

	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Project 1")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "failed to execute GraphQL request")
}

// TestGitHubProvider_GetProjectByName_GraphQLError tests error handling for GraphQL errors in GetProjectByName.
func TestGitHubProvider_GetProjectByName_GraphQLError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	graphqlResponse := `{"data":{"repositoryOwner":{"projectsV2":{"nodes":[],"totalCount":0}}},"errors":[{"message":"Some GraphQL error"}]}`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(graphqlResponse)),
	}
	mockClient.On("Do", mock.Anything).Return(resp, nil)

	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Project 1")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "graphql errors occurred")
}

// TestGitHubProvider_GetProjectByName_StatusCodeNot200 tests error handling for non-200 status codes in GetProjectByName.
func TestGitHubProvider_GetProjectByName_StatusCodeNot200(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewBufferString("not found")),
	}
	mockClient.On("Do", mock.Anything).Return(resp, nil)

	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Project 1")
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "failed to get projects (status: 404, body: not found)")
}

// TestGitHubProvider_GetProjectByName_MalformedJSON tests error handling for malformed JSON responses in GetProjectByName.
func TestGitHubProvider_GetProjectByName_MalformedJSON(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("{invalid json}")),
	}
	mockClient.On("Do", mock.Anything).Return(resp, nil)

	ctx := context.Background()
	project, err := provider.GetProjectByName(ctx, "Project 1")
	assert.Error(t, err)
	assert.Nil(t, project)
}

// TestGitHubProvider_addIssueToProject_Success tests successfully adding an issue to a project.
func TestGitHubProvider_addIssueToProject_Success(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	// 1. Buscar node_id da issue
	issueNodeResponse := `{"data":{"repository":{"issue":{"id":"issue-node-id","number":1,"title":"Test Issue"}}}}`
	resp1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(issueNodeResponse)),
	}
	// 2. Adicionar ao projeto
	addProjectResponse := `{"data":{"addProjectV2ItemById":{"item":{"id":"item-id","content":{"number":1,"title":"Test Issue"}}}}}`
	resp2 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(addProjectResponse)),
	}
	mockClient.On("Do", mock.Anything).Return(resp1, nil).Once()
	mockClient.On("Do", mock.Anything).Return(resp2, nil).Once()

	issue := &github.Issue{Number: github.Int(1)}
	project := &ProjectInfo{ProjectID: "project-id", ProjectNumber: 1}

	err := provider.addIssueToProject(context.Background(), issue, project)
	assert.NoError(t, err)
}

// TestGitHubProvider_addIssueToProject_NodeIDError tests error handling when fetching the issue node ID fails.
func TestGitHubProvider_addIssueToProject_NodeIDError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	emptyResp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockClient.On("Do", mock.Anything).Return(emptyResp, errors.New("request failed")).Once()

	issue := &github.Issue{Number: github.Int(1)}
	project := &ProjectInfo{ProjectID: "project-id", ProjectNumber: 1}

	err := provider.addIssueToProject(context.Background(), issue, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute GraphQL request for issue")
}

// TestGitHubProvider_addIssueToProject_AddProjectError tests error handling when adding an issue to a project fails.
func TestGitHubProvider_addIssueToProject_AddProjectError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	// 1. Buscar node_id da issue
	issueNodeResponse := `{"data":{"repository":{"issue":{"id":"issue-node-id","number":1,"title":"Test Issue"}}}}`
	resp1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(issueNodeResponse)),
	}
	// 2. Falha ao adicionar ao projeto
	emptyResp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockClient.On("Do", mock.Anything).Return(resp1, nil).Once()
	mockClient.On("Do", mock.Anything).Return(emptyResp, errors.New("request failed")).Once()

	issue := &github.Issue{Number: github.Int(1)}
	project := &ProjectInfo{ProjectID: "project-id", ProjectNumber: 1}

	err := provider.addIssueToProject(context.Background(), issue, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute GraphQL request for adding to project")
}

// TestGitHubProvider_addIssueToProject_GraphQLError tests error handling for GraphQL errors when adding an issue to a project.
func TestGitHubProvider_addIssueToProject_GraphQLError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	// 1. Buscar node_id da issue
	issueNodeResponse := `{"data":{"repository":{"issue":{"id":"issue-node-id","number":1,"title":"Test Issue"}}},"errors":[{"message":"Some GraphQL error"}]}`
	resp1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(issueNodeResponse)),
	}
	mockClient.On("Do", mock.Anything).Return(resp1, nil).Once()

	issue := &github.Issue{Number: github.Int(1)}
	project := &ProjectInfo{ProjectID: "project-id", ProjectNumber: 1}

	err := provider.addIssueToProject(context.Background(), issue, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "graphql errors occurred while getting issue")
}

// TestGitHubProvider_addIssueToProject_StatusCodeNot200 tests error handling for non-200 status codes when adding an issue to a project.
func TestGitHubProvider_addIssueToProject_StatusCodeNot200(t *testing.T) {
	mockClient := new(mockHTTPClient)
	client := github.NewClient(&http.Client{Transport: &mockTransport{mock: mockClient}})
	provider := &GitHubProvider{
		owner:  "testowner",
		repo:   "testrepo",
		client: client,
	}

	// 1. Buscar node_id da issue
	issueNodeResponse := `{"data":{"repository":{"issue":{"id":"issue-node-id","number":1,"title":"Test Issue"}}}}`
	resp1 := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(issueNodeResponse)),
	}
	// 2. Status code diferente de 200 ao adicionar ao projeto
	resp2 := &http.Response{
		StatusCode: 403,
		Body:       io.NopCloser(bytes.NewBufferString("forbidden")),
	}
	mockClient.On("Do", mock.Anything).Return(resp1, nil).Once()
	mockClient.On("Do", mock.Anything).Return(resp2, nil).Once()

	issue := &github.Issue{Number: github.Int(1)}
	project := &ProjectInfo{ProjectID: "project-id", ProjectNumber: 1}

	err := provider.addIssueToProject(context.Background(), issue, project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add issue to project (status: 403, body: forbidden)")
}
