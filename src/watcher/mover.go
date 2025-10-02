package watcher

import (
	"fmt"
	"os"
	"path/filepath"

	"norsetinge/common"
	"norsetinge/config"
)

// FolderAliases maps status types to folder names based on language
type FolderAliases map[string]map[string]string

// LoadFolderAliases loads the folder-aliases.yaml file
func LoadFolderAliases(path string) (FolderAliases, error) {
	// TODO: Implement YAML loading from folder-aliases.yaml
	// For now, return hardcoded Danish aliases
	return FolderAliases{
		"da": {
			"draft":     "kladde",
			"revision":  "afventer-rettelser",
			"publish":   "udgiv",
			"published": "udgivet",
			"rejected":  "afvist",
			"update":    "opdater",
		},
		"en": {
			"draft":     "drafts",
			"revision":  "awaiting-revision",
			"publish":   "publish",
			"published": "published",
			"rejected":  "rejected",
			"update":    "update",
		},
	}, nil
}

// Mover handles moving files between folders based on status
type Mover struct {
	cfg     *config.Config
	aliases FolderAliases
}

// NewMover creates a new file mover
func NewMover(cfg *config.Config, aliasesPath string) (*Mover, error) {
	aliases, err := LoadFolderAliases(aliasesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load folder aliases: %w", err)
	}

	return &Mover{
		cfg:     cfg,
		aliases: aliases,
	}, nil
}

// GetFolderForStatus returns the folder path for a given status
func (m *Mover) GetFolderForStatus(status string) (string, error) {
	lang := m.cfg.Dropbox.FolderLanguage

	folderName, ok := m.aliases[lang][status]
	if !ok {
		return "", fmt.Errorf("no folder mapping for status '%s' in language '%s'", status, lang)
	}

	return filepath.Join(m.cfg.Dropbox.BasePath, folderName), nil
}

// MoveArticle moves an article to the appropriate folder based on its status
func (m *Mover) MoveArticle(article *common.Article) error {
	currentStatus := article.GetCurrentStatus()

	targetFolder, err := m.GetFolderForStatus(currentStatus)
	if err != nil {
		return fmt.Errorf("failed to get target folder: %w", err)
	}

	// Ensure target folder exists
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return fmt.Errorf("failed to create target folder: %w", err)
	}

	// Get filename
	filename := filepath.Base(article.FilePath)
	targetPath := filepath.Join(targetFolder, filename)

	// Check if source and target are the same
	if article.FilePath == targetPath {
		return nil // Already in correct location
	}

	// Move file
	if err := os.Rename(article.FilePath, targetPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Update article filepath
	article.FilePath = targetPath

	return nil
}

// ProcessArticleStatusChange handles status changes and moves files accordingly
func (m *Mover) ProcessArticleStatusChange(filePath string) error {
	// Parse article
	article, err := common.ParseArticle(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse article: %w", err)
	}

	// Move to appropriate folder
	if err := m.MoveArticle(article); err != nil {
		return fmt.Errorf("failed to move article: %w", err)
	}

	return nil
}

// GetAllMonitoredFolders returns all folders that should be monitored
func (m *Mover) GetAllMonitoredFolders() ([]string, error) {
	statuses := []string{"draft", "revision", "publish", "published", "rejected", "update"}
	folders := make([]string, 0, len(statuses))

	for _, status := range statuses {
		folder, err := m.GetFolderForStatus(status)
		if err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}
