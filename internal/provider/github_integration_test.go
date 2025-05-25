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

	t.Logf("Testing with: owner=%s, repo=%s, token_length=%d", owner, repo, len(token))
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

	provider := NewGitHubProvider(config)
	require.NotNil(t, provider)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	title := fmt.Sprintf("[TEST] Integration Test Issue - %s", timestamp)
	description := "This is an integration test issue created by the test suite.\nIt should be automatically created in the repository."
	labels := []string{"test", "integration"}

	t.Logf("Creating issue: title=%s, owner=%s, repo=%s", title, owner, repo)
	err := provider.CreateIssue(title, description, labels)
	if err != nil {
		t.Fatalf("Failed to create issue: %v\nPlease verify:\n1. The token has 'repo' scope\n2. The repository exists and is accessible\n3. The owner/repo combination is correct", err)
	}
}
