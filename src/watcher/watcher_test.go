package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"norsetinge/src/config"
)

func TestWatcher(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "NorseTinge")

	// Create all required folders
	folders := []string{"kladde", "udgiv", "udgivet", "afvist", "afventer-rettelser", "opdater"}
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(basePath, folder), 0755); err != nil {
			t.Fatalf("Failed to create folder: %v", err)
		}
	}

	// Setup config
	cfg := &config.Config{
		Dropbox: config.DropboxConfig{
			BasePath:       basePath,
			FolderLanguage: "da",
		},
	}

	// Create watcher
	w, err := NewWatcher(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Stop()

	// Start watching
	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create test file
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

	// Write file and wait for event
	if err := os.WriteFile(testFile, []byte(articleContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait for event (could be Create or Write depending on OS)
	select {
	case event := <-w.Events():
		if event.Type != EventCreated && event.Type != EventModified {
			t.Errorf("Expected EventCreated or EventModified, got %v", event.Type)
		}
		if event.FilePath != testFile {
			t.Errorf("Expected filepath %s, got %s", testFile, event.FilePath)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}

	// Modify file
	modifiedContent := `---
title: "Test Article"
author: "TB"
status:
  draft: 1
  revision: 0
  publish: 1
  published: 0
  rejected: 0
  update: 0
---

Modified content
`

	if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for modify event
	select {
	case event := <-w.Events():
		if event.Type != EventModified {
			t.Errorf("Expected EventModified, got %v", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for modify event")
	}

	// Give the mover time to move the file
	time.Sleep(1 * time.Second)

	// Verify file was moved to udgiv folder
	expectedPath := filepath.Join(basePath, "udgiv", "test.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("File was not moved to udgiv folder")
	}
}

func TestWatcherIgnoresNonMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "NorseTinge")

	// Create all required folders
	folders := []string{"kladde", "udgiv", "udgivet", "afvist", "afventer-rettelser", "opdater"}
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(basePath, folder), 0755); err != nil {
			t.Fatalf("Failed to create folder: %v", err)
		}
	}

	cfg := &config.Config{
		Dropbox: config.DropboxConfig{
			BasePath:       basePath,
			FolderLanguage: "da",
		},
	}

	w, err := NewWatcher(cfg, "")
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Stop()

	if err := w.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Create non-markdown file
	testFile := filepath.Join(basePath, "kladde", "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should NOT receive event
	select {
	case event := <-w.Events():
		t.Errorf("Should not receive event for non-markdown file, got: %v", event)
	case <-time.After(1 * time.Second):
		// Expected - no event received
	}
}
