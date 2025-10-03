package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Article represents a parsed markdown article with frontmatter
type Article struct {
	FilePath string `yaml:"-"` // Internal use only - do not read from YAML

	// Unique identifier (auto-generated, persistent)
	ID string `yaml:"id"`

	// Required fields
	Title  string `yaml:"title"`
	Author string `yaml:"author"`
	Status Status `yaml:"status"`

	// Optional SEO fields
	Description string   `yaml:"description,omitempty"`
	Images      []string `yaml:"images,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Videos      []string `yaml:"videos,omitempty"`
	Audio       []string `yaml:"audio,omitempty"`

	// Optional organization fields
	Slug       string   `yaml:"slug,omitempty"`
	Categories []string `yaml:"categories,omitempty"`
	Series     string   `yaml:"series,omitempty"`

	// Optional branding fields
	Favicon string `yaml:"favicon,omitempty"`
	AppIcon string `yaml:"app_icon,omitempty"`

	// Raw content (after frontmatter)
	Content string
}

// Status represents the article workflow status
type Status struct {
	Draft     int `yaml:"draft"`
	Revision  int `yaml:"revision"`
	Publish   int `yaml:"publish"`
	Published int `yaml:"published"`
	Rejected  int `yaml:"rejected"`
	Update    int `yaml:"update"`
}

// ParseArticle reads a markdown file and parses the YAML frontmatter
func ParseArticle(filePath string) (*Article, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split frontmatter and content
	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid frontmatter: missing --- delimiters")
	}

	// Parse YAML frontmatter
	article := &Article{
		FilePath: filePath,
		Content:  string(bytes.TrimSpace(parts[2])),
	}

	if err := yaml.Unmarshal(parts[1], article); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Validate required fields
	if article.Title == "" {
		return nil, fmt.Errorf("missing required field: title")
	}
	if article.Author == "" {
		return nil, fmt.Errorf("missing required field: author")
	}

	// Generate ID if missing
	if article.ID == "" {
		article.ID = article.generateID()
		// Save ID to frontmatter
		if err := article.WriteFrontmatter(); err != nil {
			return nil, fmt.Errorf("failed to save generated ID: %w", err)
		}
	}

	return article, nil
}

// GetCurrentStatus returns the current status based on "last 1 wins" rule
func (a *Article) GetCurrentStatus() string {
	statuses := []struct {
		name  string
		value int
	}{
		{"draft", a.Status.Draft},
		{"revision", a.Status.Revision},
		{"publish", a.Status.Publish},
		{"published", a.Status.Published},
		{"rejected", a.Status.Rejected},
		{"update", a.Status.Update},
	}

	currentStatus := "unknown"
	for _, s := range statuses {
		if s.value == 1 {
			currentStatus = s.name
		}
	}

	return currentStatus
}

// GetSlug returns a URL-friendly slug from the title
func (a *Article) GetSlug() string {
	slug := strings.ToLower(a.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "æ", "ae")
	slug = strings.ReplaceAll(slug, "ø", "oe")
	slug = strings.ReplaceAll(slug, "å", "aa")
	// Remove non-alphanumeric characters except hyphens
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "")
	return slug
}

// UpdateStatus sets a new status and preserves history
func (a *Article) UpdateStatus(newStatus string) error {
	switch newStatus {
	case "draft":
		a.Status.Draft = 1
	case "revision":
		a.Status.Revision = 1
	case "publish":
		a.Status.Publish = 1
	case "published":
		a.Status.Published = 1
	case "rejected":
		a.Status.Rejected = 1
	case "update":
		a.Status.Update = 1
	default:
		return fmt.Errorf("invalid status: %s", newStatus)
	}
	return nil
}

// WriteFrontmatter updates the file with modified frontmatter
func (a *Article) WriteFrontmatter() error {
	frontmatter, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	newContent := fmt.Sprintf("---\n%s---\n\n%s", string(frontmatter), a.Content)

	if err := os.WriteFile(a.FilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// generateID creates a unique ID based on title, author, and timestamp
// Format: #ABC123 (6 characters: 3 letters + 3 numbers/letters, always uppercase)
func (a *Article) generateID() string {
	// Combine title, author, and current timestamp for uniqueness
	data := fmt.Sprintf("%s|%s|%d", a.Title, a.Author, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))

	// Convert to uppercase hex and take first 6 characters
	hexStr := strings.ToUpper(hex.EncodeToString(hash[:]))
	return "#" + hexStr[:6]
}
