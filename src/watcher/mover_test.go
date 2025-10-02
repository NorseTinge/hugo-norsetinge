package watcher

import (
	"os"
	"path/filepath"
	"testing"

	"norsetinge/common"
	"norsetinge/config"
)

func TestMoveArticle(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "NorseTinge")

	// Create folders
	folders := []string{"kladde", "udgiv", "udgivet", "afvist", "afventer-rettelser", "opdater"}
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(basePath, folder), 0755); err != nil {
			t.Fatalf("Failed to create folder: %v", err)
		}
	}

	// Create test article in kladde
	testFile := filepath.Join(basePath, "kladde", "test.md")
	articleContent := `---
title: "Test Article"
author: "TB"
status:
  draft: 1
  revision: 0
  publish: 0
  published: 0
  rejected: 0
  update: 0
---

Test content
`
	if err := os.WriteFile(testFile, []byte(articleContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup mover
	cfg := &config.Config{
		Dropbox: config.DropboxConfig{
			BasePath:       basePath,
			FolderLanguage: "da",
		},
	}

	mover, err := NewMover(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create mover: %v", err)
	}

	// Test: Article should stay in kladde (status is draft)
	article, _ := common.ParseArticle(testFile)
	if err := mover.MoveArticle(article); err != nil {
		t.Errorf("MoveArticle failed: %v", err)
	}

	if article.FilePath != testFile {
		t.Errorf("File should not move, expected %s, got %s", testFile, article.FilePath)
	}

	// Change status to publish
	article.UpdateStatus("publish")
	if err := article.WriteFrontmatter(); err != nil {
		t.Fatalf("Failed to write frontmatter: %v", err)
	}

	// Re-parse and move
	article, _ = common.ParseArticle(article.FilePath)
	if err := mover.MoveArticle(article); err != nil {
		t.Errorf("MoveArticle failed: %v", err)
	}

	expectedPath := filepath.Join(basePath, "udgiv", "test.md")
	if article.FilePath != expectedPath {
		t.Errorf("Expected file at %s, got %s", expectedPath, article.FilePath)
	}

	// Verify file exists at new location
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File does not exist at target location")
	}

	// Verify file no longer exists at old location
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File still exists at old location")
	}
}

func TestGetFolderForStatus(t *testing.T) {
	cfg := &config.Config{
		Dropbox: config.DropboxConfig{
			BasePath:       "/test/base",
			FolderLanguage: "da",
		},
	}

	mover, err := NewMover(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create mover: %v", err)
	}

	tests := []struct {
		status   string
		expected string
	}{
		{"draft", "/test/base/kladde"},
		{"publish", "/test/base/udgiv"},
		{"published", "/test/base/udgivet"},
		{"rejected", "/test/base/afvist"},
		{"revision", "/test/base/afventer-rettelser"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			folder, err := mover.GetFolderForStatus(tt.status)
			if err != nil {
				t.Errorf("GetFolderForStatus failed: %v", err)
			}
			if folder != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, folder)
			}
		})
	}
}

func TestGetAllMonitoredFolders(t *testing.T) {
	cfg := &config.Config{
		Dropbox: config.DropboxConfig{
			BasePath:       "/test",
			FolderLanguage: "da",
		},
	}

	mover, err := NewMover(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create mover: %v", err)
	}

	folders, err := mover.GetAllMonitoredFolders()
	if err != nil {
		t.Fatalf("GetAllMonitoredFolders failed: %v", err)
	}

	if len(folders) != 6 {
		t.Errorf("Expected 6 folders, got %d", len(folders))
	}
}
