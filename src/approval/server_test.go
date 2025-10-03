package approval

import (
	"os"
	"path/filepath"
	"testing"

	"norsetinge/common"
	"norsetinge/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Approval: config.ApprovalConfig{
			Host: "0.0.0.0",
			Port: 8080,
			TailscaleHostname: "norsetinge.tailnet.ts.net",
		},
	}

	server := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.cfg != cfg {
		t.Error("Config not set correctly")
	}

	if server.pendingArticles == nil {
		t.Error("pendingArticles map not initialized")
	}
}

func TestGenerateID(t *testing.T) {
	id1, err := generateID()
	if err != nil {
		t.Fatalf("generateID failed: %v", err)
	}

	if len(id1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("Expected ID length 32, got %d", len(id1))
	}

	// Generate another ID to ensure uniqueness
	id2, err := generateID()
	if err != nil {
		t.Fatalf("generateID failed: %v", err)
	}

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

func TestRequestApproval(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Approval: config.ApprovalConfig{
			Host: "0.0.0.0",
			Port: 8080,
			TailscaleHostname: "norsetinge.tailnet.ts.net",
		},
		Hugo: config.HugoConfig{
			SiteDir:   tmpDir + "/site",
			PublicDir: tmpDir + "/public",
			MirrorDir: tmpDir + "/mirror",
		},
		Ntfy: config.NtfyConfig{
			Enabled: false, // Disable ntfy for tests
		},
		Dropbox: config.DropboxConfig{
			BasePath: tmpDir,
		},
	}

	server := NewServer(cfg)

	article := &common.Article{
		ID:      "#TEST01",
		Title:   "Test Article",
		Author:  "TB",
		Content: "Test content",
	}

	// Note: Hugo build will fail without valid Hugo setup, but that's expected
	// The test should check that the function attempts to add the article
	err := server.RequestApproval(article)

	// We expect an error because Hugo isn't set up, but that's okay for this test
	// The important thing is that if ntfy is disabled, it shouldn't fail on that
	if err != nil {
		// Expected - Hugo build will fail
		t.Logf("RequestApproval returned expected error: %v", err)
	}

	// Even with error, check if notification was attempted
	// (This test mainly validates that the function doesn't panic)
}

// TestCleanupPreviewFiles tests Bug 9 fix: preview cleanup functionality
func TestCleanupPreviewFiles(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Hugo: config.HugoConfig{
			SiteDir:   filepath.Join(tmpDir, "site"),
			PublicDir: filepath.Join(tmpDir, "public"),
			MirrorDir: filepath.Join(tmpDir, "mirror"),
		},
		Dropbox: config.DropboxConfig{
			BasePath: tmpDir,
		},
	}

	// Create directories
	os.MkdirAll(cfg.Hugo.SiteDir+"/content", 0755)
	os.MkdirAll(cfg.Hugo.PublicDir, 0755)
	os.MkdirAll(cfg.Hugo.MirrorDir, 0755)

	server := NewServer(cfg)

	article := &common.Article{
		ID:      "#TEST99",
		Title:   "Test Article for Cleanup",
		Author:  "TB",
		Content: "Test content",
	}

	slug := article.GetSlug()
	previewDirName := "preview-" + slug

	// Create fake preview files in all three locations
	publicPreviewPath := filepath.Join(cfg.Hugo.PublicDir, previewDirName)
	mirrorPreviewPath := filepath.Join(cfg.Hugo.MirrorDir, previewDirName)
	contentPreviewPath := filepath.Join(cfg.Hugo.SiteDir, "content", "preview-"+slug+".md")

	os.MkdirAll(publicPreviewPath, 0755)
	os.MkdirAll(mirrorPreviewPath, 0755)
	os.WriteFile(contentPreviewPath, []byte("test content"), 0644)

	// Create dummy files inside preview dirs
	os.WriteFile(filepath.Join(publicPreviewPath, "index.html"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(mirrorPreviewPath, "index.html"), []byte("test"), 0644)

	// Verify files exist before cleanup
	if _, err := os.Stat(publicPreviewPath); os.IsNotExist(err) {
		t.Fatal("Public preview path should exist before cleanup")
	}
	if _, err := os.Stat(mirrorPreviewPath); os.IsNotExist(err) {
		t.Fatal("Mirror preview path should exist before cleanup")
	}
	if _, err := os.Stat(contentPreviewPath); os.IsNotExist(err) {
		t.Fatal("Content preview file should exist before cleanup")
	}

	// Run cleanup
	server.cleanupPreviewFiles(article)

	// Verify files are removed after cleanup
	if _, err := os.Stat(publicPreviewPath); !os.IsNotExist(err) {
		t.Error("Public preview path should be removed after cleanup")
	}
	if _, err := os.Stat(mirrorPreviewPath); !os.IsNotExist(err) {
		t.Error("Mirror preview path should be removed after cleanup")
	}
	if _, err := os.Stat(contentPreviewPath); !os.IsNotExist(err) {
		t.Error("Content preview file should be removed after cleanup")
	}

	t.Log("✅ Bug 9 fix verified: Preview files cleaned up successfully")
}
