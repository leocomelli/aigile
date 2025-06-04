package provider

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// IssuesService interface for GitHub Issues API
type IssuesService interface {
	Create(ctx context.Context, owner string, repo string, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
	Edit(ctx context.Context, owner string, repo string, number int, issue *github.IssueRequest) (*github.Issue, *github.Response, error)
}

// RepositoriesService interface for GitHub Repositories API
type RepositoriesService interface {
	Get(ctx context.Context, owner string, repo string) (*github.Repository, *github.Response, error)
}

type GitHubProvider struct {
	issues IssuesService
	repos  RepositoriesService
	owner  string
	repo   string
	client *github.Client
}

type GitHubConfig struct {
	Token string
	Owner string
	Repo  string
}

type ProjectInfo struct {
	ProjectNumber int    // The project number (visible in the project URL)
	ProjectOwner  string // The owner of the project (user or organization)
	ProjectID     string // The project's node ID
}

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

func (p *GitHubProvider) CreateIssue(title, description string, labels []string, project *ProjectInfo) (*github.Issue, error) {
	ctx := context.Background()

	issue := &github.IssueRequest{
		Title:  &title,
		Body:   &description,
		Labels: &labels,
	}

	createdIssue, resp, err := p.issues.Create(ctx, p.owner, p.repo, issue)
	if err != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
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

// GetProjectByName fetches project information using the project name
func (p *GitHubProvider) GetProjectByName(ctx context.Context, projectName string) (*ProjectInfo, error) {
	slog.Debug("searching for project", "name", projectName, "owner", p.owner)

	// GraphQL query to get project information
	query := fmt.Sprintf(`
		query {
			repositoryOwner(login: "%s") {
				... on User {
					projects(first: 100, orderBy: {field: TITLE, direction: ASC}) {
						nodes {
							id
							number
							title
						}
						totalCount
					}
				}
				... on Organization {
					projects(first: 100, orderBy: {field: TITLE, direction: ASC}) {
						nodes {
							id
							number
							title
						}
						totalCount
					}
				}
			}
		}
	`, p.owner)

	req, err := p.client.NewRequest("POST", "graphql", map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	var result struct {
		Data struct {
			RepositoryOwner struct {
				Projects struct {
					Nodes []struct {
						ID     string `json:"id"`
						Number int    `json:"number"`
						Title  string `json:"title"`
					} `json:"nodes"`
					TotalCount int `json:"totalCount"`
				} `json:"projects"`
			} `json:"repositoryOwner"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	resp, err := p.client.Do(ctx, req, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get project (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
	}

	// Log GraphQL errors if any
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			slog.Error("graphql error", "message", err.Message)
		}
		return nil, fmt.Errorf("graphql errors occurred")
	}

	slog.Debug("found projects", "total_count", result.Data.RepositoryOwner.Projects.TotalCount)

	// Check if project was found
	for _, project := range result.Data.RepositoryOwner.Projects.Nodes {
		slog.Debug("checking project", "title", project.Title, "number", project.Number)
		if project.Title == projectName {
			return &ProjectInfo{
				ProjectNumber: project.Number,
				ProjectOwner:  p.owner,
				ProjectID:     project.ID,
			}, nil
		}
	}

	return nil, fmt.Errorf("project '%s' not found (searched in %s's projects)", projectName, p.owner)
}

func (p *GitHubProvider) addIssueToProject(ctx context.Context, issue *github.Issue, project *ProjectInfo) error {
	// The mutation to add an item to a project
	mutation := fmt.Sprintf(`
		mutation {
			addProjectItemById(input: {
				projectId: "%s"
				contentId: "%s"
			}) {
				item {
					id
				}
			}
		}
	`, project.ProjectID, issue.GetNodeID())

	// Execute the GraphQL mutation
	req, err := p.client.NewRequest("POST", "graphql", map[string]interface{}{
		"query": mutation,
	})
	if err != nil {
		return fmt.Errorf("failed to create GraphQL request: %w", err)
	}

	resp, err := p.client.Do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add issue to project (status: %d, body: %s)", resp.StatusCode, string(bodyBytes))
	}

	slog.Info("issue added to project", "issue_number", issue.GetNumber(), "project_number", project.ProjectNumber)
	return nil
}
