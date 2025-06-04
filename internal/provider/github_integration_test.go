//go:build integration || integration_test
// +build integration integration_test

package provider

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGitHubProvider_Integration_CreateIssue(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")
	projectNumber := os.Getenv("GITHUB_PROJECT_NUMBER")

	t.Logf("Testing with: owner=%s, repo=%s, token_length=%d, project_number=%s", owner, repo, len(token), projectNumber)
	require.NotEmpty(t, token, "GITHUB_TOKEN is required")
	require.NotEmpty(t, owner, "GITHUB_OWNER is required")
	require.NotEmpty(t, repo, "GITHUB_REPO is required")

	// Configure debug logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	config := GitHubConfig{
		Token: token,
		Owner: owner,
		Repo:  repo,
	}

	provider, err := NewGitHubProvider(config)
	require.NoError(t, err, "Failed to create GitHub provider. Please verify:\n1. The token has required permissions (repo, project, read:org, read:user)\n2. The repository exists and is accessible\n3. The owner/repo combination is correct")
	require.NotNil(t, provider)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	title := fmt.Sprintf("[TEST] Integration Test Issue - %s", timestamp)
	description := "This is an integration test issue created by the test suite.\nIt should be automatically created in the repository."
	labels := []string{"test", "integration"}

	var project *ProjectInfo
	if projectNumber != "" {
		project = &ProjectInfo{
			ProjectNumber: atoi(projectNumber),
			ProjectOwner:  owner,
		}
	}

	t.Logf("Creating issue: title=%s, owner=%s, repo=%s, project=%v", title, owner, repo, project)
	createdIssue, err := provider.CreateIssue(title, description, labels, project)
	if err != nil {
		t.Fatalf("Failed to create issue: %v\nPlease verify:\n1. The token has 'repo' scope\n2. The repository exists and is accessible\n3. The owner/repo combination is correct", err)
	}

	require.NotNil(t, createdIssue)
	require.NotNil(t, createdIssue.Number)
	require.NotEmpty(t, createdIssue.GetHTMLURL())
}

// atoi converts a string to an int, panics if the string is not a valid integer
func atoi(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		panic(fmt.Sprintf("invalid integer: %s", s))
	}
	return n
}
