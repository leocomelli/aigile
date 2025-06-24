package provider

import (
	"context"
	"fmt"
)

// Provider is the interface for issue providers (GitHub, Console, etc).
type Provider interface {
	CreateIssue(title, description string, labels []string, project *ProjectInfo) (Issue, error)
	AddSubIssue(parentNumber int, childID int64) error
	GetProjectByName(ctx context.Context, projectName string) (*ProjectInfo, error)
}

// Issue is the interface for issue objects returned by providers.
type Issue interface {
	GetNumber() int
	GetID() int64
	GetHTMLURL() string
}

// ConsoleProvider implements a provider that prints issues to the console instead of creating them externally.
type ConsoleProvider struct{}

// NewConsoleProvider creates a new ConsoleProvider.
func NewConsoleProvider() *ConsoleProvider {
	return &ConsoleProvider{}
}

// ConsoleIssue is a struct to mimic the GitHub Issue for compatibility.
type ConsoleIssue struct {
	title       string
	description string
	labels      []string
}

// GetNumber returns the issue number (always 0 for ConsoleIssue).
func (i *ConsoleIssue) GetNumber() int { return 0 }

// GetID returns the issue ID (always 0 for ConsoleIssue).
func (i *ConsoleIssue) GetID() int64 { return 0 }

// GetHTMLURL returns the issue URL (always empty for ConsoleIssue).
func (i *ConsoleIssue) GetHTMLURL() string { return "" }

// GetTitle returns the issue title.
func (i *ConsoleIssue) GetTitle() string { return i.title }

// GetBody returns the issue description.
func (i *ConsoleIssue) GetBody() string { return i.description }

// GetLabels returns the issue labels.
func (i *ConsoleIssue) GetLabels() []string { return i.labels }

// CreateIssue prints the issue data to the console and returns a ConsoleIssue.
func (p *ConsoleProvider) CreateIssue(title, description string, labels []string, project *ProjectInfo) (Issue, error) {
	fmt.Println("\n[CONSOLE PROVIDER] Issue Preview:")
	fmt.Println("Title:", title)
	fmt.Println("Labels:", labels)
	fmt.Println("Description:\n" + description)
	if project != nil {
		fmt.Printf("Project: %v\n", project)
	}
	return &ConsoleIssue{title: title, description: description, labels: labels}, nil
}

// AddSubIssue is a no-op for the console provider.
func (p *ConsoleProvider) AddSubIssue(parentNumber int, childID int64) error {
	fmt.Printf("[CONSOLE PROVIDER] Would link sub-issue %d to parent %d\n", childID, parentNumber)
	return nil
}

// GetProjectByName is a no-op for the console provider.
func (p *ConsoleProvider) GetProjectByName(_ context.Context, _ string) (*ProjectInfo, error) {
	return nil, nil
}
