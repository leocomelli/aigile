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
}

type GitHubProvider struct {
	issues IssuesService
	owner  string
	repo   string
}

type GitHubConfig struct {
	Token string
	Owner string
	Repo  string
}

func NewGitHubProvider(config GitHubConfig) *GitHubProvider {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubProvider{
		issues: client.Issues,
		owner:  config.Owner,
		repo:   config.Repo,
	}
}

func (p *GitHubProvider) CreateIssue(title, description string, labels []string) error {
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
		return fmt.Errorf("failed to create issue (status: %s, body: %s): %w", resp.Status, bodyStr, err)
	}

	slog.Info("issue created", "number", createdIssue.GetNumber(), "url", createdIssue.GetHTMLURL())
	return nil
}
