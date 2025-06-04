package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/leocomelli/aigile/internal/llm"
	"github.com/leocomelli/aigile/internal/provider"
	"github.com/leocomelli/aigile/internal/reader"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate items from XLSX file",
	Long:  `Generate Epics, Features, User Stories and Tasks from an XLSX file using LLM and create them in GitHub/Azure DevOps.`,
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("file", "f", "", "Path to XLSX file")
	generateCmd.Flags().StringP("language", "g", "english", "Language to generate the content (e.g., english, portuguese)")
	generateCmd.MarkFlagRequired("file")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	language, _ := cmd.Flags().GetString("language")
	slog.Info("starting generate command", "file", filePath, "language", language)

	// Initialize XLSX reader
	xlsxReader := reader.NewXLSXReader(filePath)
	items, err := xlsxReader.Read()
	if err != nil {
		return fmt.Errorf("failed to read XLSX file: %w", err)
	}

	// Initialize LLM provider
	llmConfig := llm.Config{
		Provider: os.Getenv("LLM_PROVIDER"),
		APIKey:   os.Getenv("LLM_API_KEY"),
		Model:    os.Getenv("LLM_MODEL"),
		Endpoint: os.Getenv("LLM_ENDPOINT"),
	}

	var llmProvider llm.LLMProvider
	switch llmConfig.Provider {
	case "openai", "":
		llmProvider = llm.NewOpenAIProvider(llmConfig)
	default:
		return fmt.Errorf("unsupported LLM provider: %s", llmConfig.Provider)
	}

	// Initialize GitHub provider
	githubConfig := provider.GitHubConfig{
		Token: os.Getenv("GITHUB_TOKEN"),
		Owner: os.Getenv("GITHUB_OWNER"),
		Repo:  os.Getenv("GITHUB_REPO"),
	}
	githubProvider, err := provider.NewGitHubProvider(githubConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub provider: %w", err)
	}

	// Process each item
	for _, item := range items {
		content, err := llmProvider.GenerateContent(
			item.Type,
			item.Parent,
			item.Context,
			item.Criteria,
			language,
		)
		if err != nil {
			return fmt.Errorf("failed to generate content: %w", err)
		}

		// Create issue in GitHub
		title := content.Title
		if title == "" {
			title = fmt.Sprintf("[%s] %s", item.Type, item.Context[:50])
		}

		// Get project info if parent is specified
		var project *provider.ProjectInfo
		if item.Parent != "" {
			slog.Debug("searching for project from parent field", "parent", item.Parent)
			var err error
			project, err = githubProvider.GetProjectByName(context.Background(), item.Parent)
			if err != nil {
				slog.Warn("failed to get project info", "parent", item.Parent, "error", err)
			} else {
				slog.Debug("project found", "number", project.ProjectNumber, "owner", project.ProjectOwner)
			}
		}

		fullDescription := formatDescription(content)
		createdIssue, err := githubProvider.CreateIssue(title, fullDescription, []string{item.Type.String()}, project)
		if err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}
		slog.Info("issue created", "type", item.Type, "title", title, "number", createdIssue.GetNumber(), "project", project)
	}

	return nil
}

func formatDescription(content *llm.GeneratedContent) string {
	var sb strings.Builder

	// Add description
	sb.WriteString(content.Description)
	sb.WriteString("\n\n")

	// Add parent reference if available
	if content.Parent != "" {
		sb.WriteString(fmt.Sprintf("Parent: %s\n\n", content.Parent))
	}

	// Add acceptance criteria if available
	if len(content.AcceptanceCriteria) > 0 {
		sb.WriteString("## Acceptance Criteria\n")
		for i, c := range content.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, c))
		}
		sb.WriteString("\n")
	}

	// Add additional info if available
	if content.AdditionalInfo != "" {
		sb.WriteString("## Additional Information\n")
		sb.WriteString(content.AdditionalInfo)
		sb.WriteString("\n")
	}

	return sb.String()
}
