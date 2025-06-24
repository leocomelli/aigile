// Package provider implementa integrações com serviços externos como GitHub e Azure DevOps.
package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// IssuesService interface for GitHub Issues API.
type IssuesService interface {
	Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

// RepositoriesService interface for GitHub Repositories API.
type RepositoriesService interface {
	Get(ctx context.Context, owner string, repo string) (*github.Repository, *github.Response, error)
}

// GitHubProvider provides methods to interact with GitHub Issues and Projects.
type GitHubProvider struct {
	issues IssuesService
	repos  RepositoriesService
	owner  string
	repo   string
	client *github.Client
}

// GitHubConfig holds the configuration for the GitHub provider.
type GitHubConfig struct {
	Token string
	Owner string
	Repo  string
}

// ProjectInfo holds information about a GitHub Project v2.
type ProjectInfo struct {
	ProjectNumber int    // The project number (visible in the project URL)
	ProjectOwner  string // The owner of the project (user or organization)
	ProjectID     string // The project's node ID
}

// GraphQL queries/mutations as constants for clarity and reuse.
const (
	queryProjectV2ByName = `query($owner: String!) {
		repositoryOwner(login: $owner) {
			... on User {
				projectsV2(first: 100) {
					nodes { id number title }
					totalCount
				}
			}
			... on Organization {
				projectsV2(first: 100) {
					nodes { id number title }
					totalCount
				}
			}
		}
	}`

	queryIssueNodeID = `query($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			issue(number: $number) { id number title }
		}
	}`

	mutationAddProjectV2ItemByID = `mutation($projectId: ID!, $contentId: ID!) {
		addProjectV2ItemById(input: {projectId: $projectId, contentId: $contentId}) {
			item { id content { ... on Issue { number title } } }
		}
	}`
)

// NewGitHubProvider creates a new GitHubProvider with the given configuration.
func NewGitHubProvider(config GitHubConfig) (*GitHubProvider, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	provider := &GitHubProvider{
		issues: client.Issues,
		repos:  client.Repositories,
		owner:  config.Owner,
		repo:   config.Repo,
		client: client,
	}

	return provider, nil
}

// CreateIssue creates a new issue in the configured GitHub repository and optionally adds it to a project.
func (p *GitHubProvider) CreateIssue(title, description string, labels []string, project *ProjectInfo) (Issue, error) {
	ctx := context.Background()

	issue := &github.IssueRequest{
		Title:  &title,
		Body:   &description,
		Labels: &labels,
	}

	createdIssue, resp, err := p.issues.Create(ctx, p.owner, p.repo, issue)
	if err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
		bodyStr := string(bodyBytes)
		return nil, fmt.Errorf("failed to create issue (status: %s, body: %s): %w", resp.Status, bodyStr, err)
	}

	slog.Info("issue created", "number", createdIssue.GetNumber(), "url", createdIssue.GetHTMLURL())

	// If project info is provided, add the issue to the project
	if project != nil {
		if err := p.addIssueToProject(ctx, createdIssue, project); err != nil {
			slog.Warn("failed to add issue to project", "error", err)
		}
	}

	return createdIssue, nil
}

// GetProjectByName fetches project information using the project name.
func (p *GitHubProvider) GetProjectByName(ctx context.Context, projectName string) (*ProjectInfo, error) {
	slog.Debug("searching for project", "name", projectName, "owner", p.owner)

	vars := map[string]interface{}{"owner": p.owner}
	req, err := p.client.NewRequest("POST", "graphql", map[string]interface{}{
		"query":     queryProjectV2ByName,
		"variables": vars,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	var result struct {
		Data struct {
			RepositoryOwner struct {
				ProjectsV2 struct {
					Nodes []struct {
						ID     string `json:"id"`
						Number int    `json:"number"`
						Title  string `json:"title"`
					} `json:"nodes"`
					TotalCount int `json:"totalCount"`
				} `json:"projectsV2"`
			} `json:"repositoryOwner"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	resp, err := p.client.Do(ctx, req, &result)
	if err != nil {
		if resp != nil && resp.Body != nil {
			defer func() {
				if cerr := resp.Body.Close(); cerr != nil {
					slog.Warn("failed to close response body", "error", cerr)
				}
			}()
			if resp.StatusCode != 200 {
				bodyBytes, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("failed to get projects (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
			}
		}
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get projects (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
	}

	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			slog.Error("graphql error", "message", err.Message)
		}
		return nil, fmt.Errorf("graphql errors occurred")
	}

	slog.Debug("found projects", "total_count", result.Data.RepositoryOwner.ProjectsV2.TotalCount)

	for _, project := range result.Data.RepositoryOwner.ProjectsV2.Nodes {
		slog.Debug("checking project", "title", project.Title, "number", project.Number)
		if project.Title == projectName {
			slog.Info("found project", "title", project.Title, "number", project.Number)
			return &ProjectInfo{
				ProjectID:     project.ID,
				ProjectNumber: project.Number,
			}, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", projectName)
}

// addIssueToProject adds an existing issue to a GitHub Project v2 using addProjectV2ItemById.
func (p *GitHubProvider) addIssueToProject(ctx context.Context, issue *github.Issue, project *ProjectInfo) error {
	slog.Debug("adding issue to project",
		"issue_number", issue.GetNumber(),
		"project_number", project.ProjectNumber,
		"project_id", project.ProjectID,
		"owner", p.owner,
		"repo", p.repo)

	// 1. Buscar node_id da issue
	vars := map[string]interface{}{"owner": p.owner, "repo": p.repo, "number": issue.GetNumber()}
	req, err := p.client.NewRequest("POST", "graphql", map[string]interface{}{
		"query":     queryIssueNodeID,
		"variables": vars,
	})
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request for issue: %w", err)
	}

	var issueResult struct {
		Data struct {
			Repository struct {
				Issue struct {
					ID     string `json:"id"`
					Number int    `json:"number"`
					Title  string `json:"title"`
				} `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	resp, err := p.client.Do(ctx, req, &issueResult)
	if err != nil {
		if resp != nil && resp.Body != nil {
			if resp.StatusCode != 200 {
				bodyBytes, _ := io.ReadAll(resp.Body)
				if cerr := resp.Body.Close(); cerr != nil {
					slog.Warn("failed to close response body", "error", cerr)
				}
				return fmt.Errorf("failed to get issue (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
			}
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Warn("failed to close response body", "error", cerr)
			}
		}
		return fmt.Errorf("failed to execute GraphQL request for issue: %w", err)
	}

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
		return fmt.Errorf("failed to get issue (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
	}

	if len(issueResult.Errors) > 0 {
		for _, err := range issueResult.Errors {
			slog.Error("graphql error", "message", err.Message)
		}
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
		return fmt.Errorf("graphql errors occurred while getting issue")
	}

	slog.Debug("got issue details",
		"issue_id", issueResult.Data.Repository.Issue.ID,
		"issue_number", issueResult.Data.Repository.Issue.Number,
		"issue_title", issueResult.Data.Repository.Issue.Title)

	// 2. Adicionar ao projeto
	varsMutation := map[string]interface{}{"projectId": project.ProjectID, "contentId": issueResult.Data.Repository.Issue.ID}
	req, err = p.client.NewRequest("POST", "graphql", map[string]interface{}{
		"query":     mutationAddProjectV2ItemByID,
		"variables": varsMutation,
	})
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request for adding to project: %w", err)
	}

	var mutationResult struct {
		Data struct {
			AddProjectV2ItemByID struct {
				Item struct {
					ID      string `json:"id"`
					Content struct {
						Number int    `json:"number"`
						Title  string `json:"title"`
					} `json:"content"`
				} `json:"item"`
			} `json:"addProjectV2ItemById"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	resp, err = p.client.Do(ctx, req, &mutationResult)
	if err != nil {
		if resp == nil || resp.Body == nil {
			return fmt.Errorf("failed to execute GraphQL request for adding to project: %w", err)
		}
		if resp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			if cerr := resp.Body.Close(); cerr != nil {
				slog.Warn("failed to close response body", "error", cerr)
			}
			return fmt.Errorf("failed to add issue to project (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
		}
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
		return fmt.Errorf("failed to execute GraphQL request for adding to project: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("response or response body is nil after GraphQL request for adding to project")
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
		return fmt.Errorf("failed to add issue to project (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
	}

	if len(mutationResult.Errors) > 0 {
		for _, err := range mutationResult.Errors {
			slog.Error("graphql error", "message", err.Message)
		}
		return fmt.Errorf("graphql errors occurred while adding to project")
	}

	slog.Info("issue added to project",
		"issue_number", issueResult.Data.Repository.Issue.Number,
		"project_number", project.ProjectNumber,
		"project_item_id", mutationResult.Data.AddProjectV2ItemByID.Item.ID,
		"issue_title", mutationResult.Data.AddProjectV2ItemByID.Item.Content.Title)
	return nil
}

// AddSubIssue adds sub-issue to a parent issue using the GitHub REST API.
func (p *GitHubProvider) AddSubIssue(parentNumber int, childID int64) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/sub_issues", p.owner, p.repo, parentNumber)
	slog.Debug("adding sub-issues", "url", url, "parent_number", parentNumber, "child_id", childID)
	body := map[string]interface{}{
		"sub_issue_id": childID,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal sub-issues body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create sub-issues request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("GITHUB_TOKEN")))
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute sub-issues request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("failed to close response body", "error", cerr)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add sub-issues (status: %d, body: %s)", resp.StatusCode, string(respBody))
	}
	return nil
}
