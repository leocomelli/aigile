package provider

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestConsoleProvider_CreateIssue(t *testing.T) {
	provider := NewConsoleProvider()
	output := captureStdout(func() {
		issue, err := provider.CreateIssue("Test Title", "Test Description", []string{"bug", "feature"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if issue.GetTitle() != "Test Title" {
			t.Errorf("expected title 'Test Title', got '%s'", issue.GetTitle())
		}
		if issue.GetBody() != "Test Description" {
			t.Errorf("expected body 'Test Description', got '%s'", issue.GetBody())
		}
		if len(issue.GetLabels()) != 2 {
			t.Errorf("expected 2 labels, got %d", len(issue.GetLabels()))
		}
	})
	if !strings.Contains(output, "[CONSOLE PROVIDER] Issue Preview:") {
		t.Errorf("expected output to contain '[CONSOLE PROVIDER] Issue Preview:', got %s", output)
	}
}

func TestConsoleProvider_CreateIssue_WithProject(t *testing.T) {
	provider := NewConsoleProvider()
	project := &ProjectInfo{ProjectNumber: 1, ProjectOwner: "owner", ProjectID: "id"}
	output := captureStdout(func() {
		_, err := provider.CreateIssue("Title", "Desc", []string{"label"}, project)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "Project: &{1 owner id}") {
		t.Errorf("expected output to contain project info, got %s", output)
	}
}

func TestConsoleProvider_AddSubIssue(t *testing.T) {
	provider := NewConsoleProvider()
	output := captureStdout(func() {
		err := provider.AddSubIssue(1, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(output, "Would link sub-issue 2 to parent 1") {
		t.Errorf("expected output to contain sub-issue link info, got %s", output)
	}
}

func TestConsoleProvider_GetProjectByName(t *testing.T) {
	provider := NewConsoleProvider()
	project, err := provider.GetProjectByName(context.Background(), "any")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project != nil {
		t.Errorf("expected nil project, got %v", project)
	}
}

func TestConsoleIssue_Methods(t *testing.T) {
	issue := &ConsoleIssue{title: "t", description: "d", labels: []string{"a"}}
	if issue.GetNumber() != 0 {
		t.Errorf("expected number 0, got %d", issue.GetNumber())
	}
	if issue.GetID() != 0 {
		t.Errorf("expected id 0, got %d", issue.GetID())
	}
	if issue.GetHTMLURL() != "" {
		t.Errorf("expected empty url, got %s", issue.GetHTMLURL())
	}
	if issue.GetTitle() != "t" {
		t.Errorf("expected title 't', got '%s'", issue.GetTitle())
	}
	if issue.GetBody() != "d" {
		t.Errorf("expected body 'd', got '%s'", issue.GetBody())
	}
	if len(issue.GetLabels()) != 1 || issue.GetLabels()[0] != "a" {
		t.Errorf("expected labels ['a'], got %v", issue.GetLabels())
	}
}
