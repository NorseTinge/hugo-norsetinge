package builder

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"norsetinge/src/common"
	"norsetinge/src/config"
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
// Returns the URL path relative to /preview/ endpoint (served via Tailscale)
func (h *HugoBuilder) BuildPreview(article *common.Article) (string, error) {
	// Detect article language from content or default to Danish
	lang := h.detectLanguage(article)

	// Create content file in Hugo structure
	slug := article.GetSlug()
	contentPath := filepath.Join(h.cfg.Hugo.SiteDir, "content", fmt.Sprintf("preview-%s.md", slug))

	// Write article as Hugo content
	if err := h.writeHugoContent(contentPath, article, lang); err != nil {
		return "", fmt.Errorf("failed to write Hugo content: %w", err)
	}
	defer os.Remove(contentPath) // Clean up after build

	// Build Hugo site
	if err := h.buildSite(); err != nil {
		return "", fmt.Errorf("failed to build Hugo site: %w", err)
	}

	// Return URL path for preview (served via /preview/ endpoint)
	previewURLPath := fmt.Sprintf("preview-%s/index.html", slug)
	log.Printf("Preview built and ready at: /preview/%s", previewURLPath)
	return previewURLPath, nil
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
preview: true
articleID: "%s"
---

%s
`, article.Title, article.Author, article.ID, article.Content)

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

// BuildFullSite builds complete Hugo site with all published articles
// Returns paths to public and mirror directories
func (h *HugoBuilder) BuildFullSite() (publicDir, mirrorDir string, err error) {
	log.Printf("🔨 Building full site...")

	// 1. Clean and prepare content directory
	contentDir := filepath.Join(h.cfg.Hugo.SiteDir, "content", "articles")
	if err := os.RemoveAll(contentDir); err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to clean content directory: %w", err)
	}
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create content directory: %w", err)
	}

	// 2. Copy all published articles to Hugo content
	publishedDir := filepath.Join(h.cfg.Dropbox.BasePath, "udgivet")
	articles, err := h.loadPublishedArticles(publishedDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to load published articles: %w", err)
	}

	log.Printf("📚 Found %d published articles", len(articles))

	for _, article := range articles {
		slug := article.GetSlug()
		articlePath := filepath.Join(contentDir, fmt.Sprintf("%s.md", slug))

		if err := h.writeHugoContent(articlePath, article, h.detectLanguage(article)); err != nil {
			return "", "", fmt.Errorf("failed to write article %s: %w", slug, err)
		}
		log.Printf("  ✓ Added: %s", article.Title)
	}

	// 3. Build Hugo site
	if err := h.buildSite(); err != nil {
		return "", "", fmt.Errorf("failed to build Hugo site: %w", err)
	}

	// 4. Return paths
	publicDir = h.cfg.Hugo.PublicDir
	mirrorDir = h.cfg.Hugo.MirrorDir

	log.Printf("✅ Full site built successfully")
	log.Printf("   Public:  %s", publicDir)
	log.Printf("   Mirror:  %s", mirrorDir)

	return publicDir, mirrorDir, nil
}

// loadPublishedArticles loads all articles from the published directory
func (h *HugoBuilder) loadPublishedArticles(publishedDir string) ([]*common.Article, error) {
	var articles []*common.Article

	entries, err := os.ReadDir(publishedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return articles, nil // No published articles yet
		}
		return nil, fmt.Errorf("failed to read published directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		articlePath := filepath.Join(publishedDir, entry.Name())
		article, err := common.ParseArticle(articlePath)
		if err != nil {
			log.Printf("Warning: Failed to parse %s: %v", entry.Name(), err)
			continue
		}

		articles = append(articles, article)
	}

	return articles, nil
}
