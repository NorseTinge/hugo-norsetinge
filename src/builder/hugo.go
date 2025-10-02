package builder

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"norsetinge/common"
	"norsetinge/config"
)

// HugoBuilder handles Hugo site building
type HugoBuilder struct {
	cfg *config.Config
}

// NewHugoBuilder creates a new Hugo builder
func NewHugoBuilder(cfg *config.Config) *HugoBuilder {
	return &HugoBuilder{cfg: cfg}
}

// BuildPreview builds a single-article preview for approval
// Copies complete Hugo-built site to Dropbox and returns the Dropbox path
func (h *HugoBuilder) BuildPreview(article *common.Article) (string, error) {
	// Detect article language from content or default to Danish
	lang := h.detectLanguage(article)

	// Create content file in Hugo structure
	contentPath := filepath.Join(h.cfg.Hugo.SiteDir, "content", fmt.Sprintf("preview-%s.md", article.GetSlug()))

	// Write article as Hugo content
	if err := h.writeHugoContent(contentPath, article, lang); err != nil {
		return "", fmt.Errorf("failed to write Hugo content: %w", err)
	}
	defer os.Remove(contentPath) // Clean up after build

	// Build Hugo site
	if err := h.buildSite(); err != nil {
		return "", fmt.Errorf("failed to build Hugo site: %w", err)
	}

	// Copy complete preview to Dropbox
	slug := article.GetSlug()
	hugoPreviewDir := filepath.Join(h.cfg.Hugo.PublicDir, fmt.Sprintf("preview-%s", slug))
	dropboxPreviewDir := filepath.Join(h.cfg.Dropbox.BasePath, "godkendelse", slug)

	if err := h.copyPreviewToDropbox(hugoPreviewDir, dropboxPreviewDir); err != nil {
		return "", fmt.Errorf("failed to copy preview to Dropbox: %w", err)
	}

	// Return Dropbox path to index.html
	dropboxHTMLPath := filepath.Join(dropboxPreviewDir, "index.html")
	log.Printf("Preview copied to Dropbox: %s", dropboxHTMLPath)
	return dropboxHTMLPath, nil
}

// copyPreviewToDropbox copies complete Hugo preview directory to Dropbox
func (h *HugoBuilder) copyPreviewToDropbox(src, dst string) error {
	// Remove existing Dropbox preview if it exists
	if err := os.RemoveAll(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old preview: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create Dropbox directory: %w", err)
	}

	// Copy all files recursively
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// writeHugoContent writes article in Hugo content format
func (h *HugoBuilder) writeHugoContent(path string, article *common.Article, lang string) error {
	content := fmt.Sprintf(`---
title: "%s"
author: "%s"
draft: false
---

%s
`, article.Title, article.Author, article.Content)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create content directory: %w", err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// buildSite runs hugo build command
func (h *HugoBuilder) buildSite() error {
	// Get absolute paths
	siteDir, err := filepath.Abs(h.cfg.Hugo.SiteDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute site path: %w", err)
	}

	publicDir, err := filepath.Abs(h.cfg.Hugo.PublicDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute public path: %w", err)
	}

	cmd := exec.Command("hugo", "--source", siteDir, "--destination", publicDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Hugo build error: %s", string(output))
		return fmt.Errorf("hugo build failed: %w", err)
	}
	log.Printf("Hugo build successful")
	return nil
}

// detectLanguage detects article language (for now, assumes Danish)
// TODO: Add language detection logic
func (h *HugoBuilder) detectLanguage(article *common.Article) string {
	// For now, default to Danish
	// Later: detect from content or add language field to frontmatter
	return "da"
}
