package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArticle(t *testing.T) {
	// Create temp test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	content := `---
title: "Test Article"
author: "TB (twisted brain)"
status:
  draft: 1
  revision: 0
  publish: 0
  published: 0
  rejected: 0
  update: 0
description: "Test description"
tags: ["test", "example"]
---

This is the article content.

## Heading

More content here.
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse article
	article, err := ParseArticle(testFile)
	if err != nil {
		t.Fatalf("ParseArticle failed: %v", err)
	}

	// Validate fields
	if article.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", article.Title)
	}

	if article.Author != "TB (twisted brain)" {
		t.Errorf("Expected author 'TB (twisted brain)', got '%s'", article.Author)
	}

	if article.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", article.Description)
	}

	if len(article.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(article.Tags))
	}

	if article.Status.Draft != 1 {
		t.Errorf("Expected draft status 1, got %d", article.Status.Draft)
	}
}

func TestGetCurrentStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "draft only",
			status:   Status{Draft: 1},
			expected: "draft",
		},
		{
			name:     "draft then publish",
			status:   Status{Draft: 1, Publish: 1},
			expected: "publish",
		},
		{
			name:     "full workflow to published",
			status:   Status{Draft: 1, Revision: 1, Publish: 1, Published: 1},
			expected: "published",
		},
		{
			name:     "update after published",
			status:   Status{Draft: 1, Publish: 1, Published: 1, Update: 1},
			expected: "update",
		},
		{
			name:     "rejected",
			status:   Status{Draft: 1, Publish: 1, Rejected: 1},
			expected: "rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article := &Article{Status: tt.status}
			result := article.GetCurrentStatus()
			if result != tt.expected {
				t.Errorf("Expected status '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	article := &Article{
		Status: Status{Draft: 1},
	}

	// Update to publish (preserves history)
	if err := article.UpdateStatus("publish"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	if article.Status.Draft != 1 {
		t.Error("Draft status should be preserved")
	}

	if article.Status.Publish != 1 {
		t.Error("Publish status should be set")
	}

	if article.GetCurrentStatus() != "publish" {
		t.Errorf("Current status should be 'publish', got '%s'", article.GetCurrentStatus())
	}
}
