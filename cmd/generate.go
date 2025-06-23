// Package cmd implements the CLI commands for aigile.
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
	Long:  `Generate User Stories from an XLSX file using LLM and create them in GitHub/Azure DevOps.`,
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("file", "f", "", "Path to XLSX file or Google Sheets URL")
	generateCmd.Flags().StringP("language", "g", "english", "Language to generate the content (e.g., english, portuguese)")
	generateCmd.Flags().Bool("auto-tasks", false, "Automatically generate and create tasks for each user story")
	generateCmd.Flags().String("google-credentials-file", "", "Path to Google Service Account credentials JSON file (required for Google Sheets)")
	if err := generateCmd.MarkFlagRequired("file"); err != nil {
		panic(fmt.Sprintf("failed to mark 'file' flag as required: %v", err))
	}
}

// runGenerate is the main handler for the 'generate' command, processing the XLSX file and creating issues.
func runGenerate(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	language, _ := cmd.Flags().GetString("language")
	autoTasks, _ := cmd.Flags().GetBool("auto-tasks")
	googleCredentialsFile, _ := cmd.Flags().GetString("google-credentials-file")
	slog.Info("starting generate command", "file", filePath, "language", language, "autoTasks", autoTasks)

	var r reader.Reader
	if strings.HasPrefix(filePath, "https://docs.google.com/spreadsheets/") {
		if googleCredentialsFile == "" {
			return fmt.Errorf("google-credentials-file flag is required for Google Sheets")
		}
		r = reader.NewGoogleSheetsReader(extractSpreadsheetID(filePath), googleCredentialsFile)
	} else {
		r = reader.NewXLSXReader(filePath)
	}
	items, err := r.Read()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	slog.Debug("items read from input source", "items", items)

	// Initialize LLM provider
	llmConfig := llm.Config{
		Provider: os.Getenv("LLM_PROVIDER"),
		APIKey:   os.Getenv("LLM_API_KEY"),
		Model:    os.Getenv("LLM_MODEL"),
		Endpoint: os.Getenv("LLM_ENDPOINT"),
	}

	var llmProvider llm.Provider
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
			autoTasks,
		)
		if err != nil {
			return fmt.Errorf("failed to generate content: %w", err)
		}

		// Create issue in GitHub
		title := content.Title
		if title == "" {
			title = fmt.Sprintf("%s %s", item.Type, item.Context[:50])
		}
		title = fmt.Sprintf("[📖 User Story] %s", title)

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

		// If there are suggested tasks, create each one as an issue and collect their IDs
		var taskIDs []int64
		if autoTasks && len(content.SuggestedTasks) > 0 {
			for _, task := range content.SuggestedTasks {
				taskTitle := fmt.Sprintf("[🛠️ Task] %s", task)
				taskDescription := fmt.Sprintf("Task for User Story #%d: %s\n\n%s", createdIssue.GetNumber(), title, task)

				taskIssue, err := githubProvider.CreateIssue(taskTitle, taskDescription, []string{"Task"}, project)
				if err != nil {
					slog.Warn("failed to create task issue", "task", task, "error", err)
					continue
				}
				slog.Info("task issue created", "task", task, "number", taskIssue.GetNumber())
				if taskIssue.GetID() != 0 {
					taskIDs = append(taskIDs, taskIssue.GetID())
				}
			}
			// Add the tasks as sub-issues of the User Story
			if len(taskIDs) > 0 {
				for _, taskID := range taskIDs {
					err := githubProvider.AddSubIssue(createdIssue.GetNumber(), taskID)
					if err != nil {
						slog.Warn("failed to add sub-issue", "error", err)
					}
				}
			}
		}
	}

	return nil
}

func formatDescription(content *llm.GeneratedContent) string {
	var sb strings.Builder

	// Add description
	sb.WriteString(content.Description)
	sb.WriteString("\n\n")

	// Add acceptance criteria if available
	if len(content.AcceptanceCriteria) > 0 {
		sb.WriteString("## Acceptance Criteria\n")
		for i, c := range content.AcceptanceCriteria {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, c))
		}
		sb.WriteString("\n")
	}

	// Add suggested tasks if available
	if len(content.SuggestedTasks) > 0 {
		sb.WriteString("## Suggested Tasks\n")
		for i, task := range content.SuggestedTasks {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, task))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// extractSpreadsheetID extrai o ID da planilha de uma URL do Google Sheets.
func extractSpreadsheetID(url string) string {
	const prefix = "https://docs.google.com/spreadsheets/d/"
	if !strings.HasPrefix(url, prefix) {
		return ""
	}
	idAndRest := strings.TrimPrefix(url, prefix)
	parts := strings.SplitN(idAndRest, "/", 2)
	return parts[0]
}
