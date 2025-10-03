package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"norsetinge/src/common"
	"norsetinge/src/config"
)

// Watcher monitors Dropbox folders for file changes
type Watcher struct {
	cfg            *config.Config
	mover          *Mover
	watcher        *fsnotify.Watcher
	events         chan Event
	approvalServer ApprovalServer
}

// ApprovalServer interface for triggering approval
type ApprovalServer interface {
	RequestApproval(article *common.Article) error
}

// Event represents a file system event
type Event struct {
	Type     EventType
	FilePath string
}

// EventType represents the type of file event
type EventType int

const (
	EventCreated EventType = iota
	EventModified
	EventDeleted
)

// NewWatcher creates a new file watcher
func NewWatcher(cfg *config.Config) (*Watcher, error) {
	mover, err := NewMover(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create mover: %w", err)
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	return &Watcher{
		cfg:     cfg,
		mover:   mover,
		watcher: fsWatcher,
		events:  make(chan Event, 100),
	}, nil
}

// SetApprovalServer sets the approval server for handling approvals
func (w *Watcher) SetApprovalServer(server ApprovalServer) {
	w.approvalServer = server
}

// GetMover returns the mover for file operations
func (w *Watcher) GetMover() *Mover {
	return w.mover
}

// Start begins monitoring all configured folders
func (w *Watcher) Start() error {
	folders, err := w.mover.GetAllMonitoredFolders()
	if err != nil {
		return fmt.Errorf("failed to get monitored folders: %w", err)
	}

	// Add all folders to watcher
	for _, folder := range folders {
		if err := w.watcher.Add(folder); err != nil {
			return fmt.Errorf("failed to watch folder %s: %w", folder, err)
		}
		log.Printf("Watching folder: %s", folder)
	}

	// Start event processing goroutine
	go w.processEvents()

	// Start periodic folder scanner to enforce status-based file organization
	go w.periodicFolderScan()

	// Run initial scan immediately
	go w.scanAllFolders()

	return nil
}

// processEvents handles fsnotify events and converts them to our event type
func (w *Watcher) processEvents() {
	// Debounce timer to avoid processing rapid successive events
	debounce := make(map[string]*time.Timer)

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only process .md files
			if filepath.Ext(event.Name) != ".md" {
				continue
			}

			// Skip temp files
			if filepath.Base(event.Name)[0] == '.' {
				continue
			}

			// Debounce: wait 500ms before processing
			if timer, exists := debounce[event.Name]; exists {
				timer.Stop()
			}

			debounce[event.Name] = time.AfterFunc(500*time.Millisecond, func() {
				w.handleEvent(event)
				delete(debounce, event.Name)
			})

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// handleEvent processes a single file event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	var eventType EventType

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = EventCreated
		log.Printf("File created: %s", event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = EventModified
		log.Printf("File modified: %s", event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = EventDeleted
		log.Printf("File deleted: %s", event.Name)
	default:
		return // Ignore other events
	}

	// Send event to channel
	w.events <- Event{
		Type:     eventType,
		FilePath: event.Name,
	}

	// Process status changes (move files if needed)
	if eventType == EventCreated || eventType == EventModified {
		if err := w.mover.ProcessArticleStatusChange(event.Name); err != nil {
			log.Printf("Failed to process article status change: %v", err)
			return
		}

		// Check if article is now in publish folder (udgiv)
		article, err := common.ParseArticle(event.Name)
		if err != nil {
			log.Printf("Failed to parse article: %v", err)
			return
		}

		// Trigger approval for publish status OR update flag
		currentStatus := article.GetCurrentStatus()
		if (currentStatus == "publish" || currentStatus == "update") && w.approvalServer != nil {
			log.Printf("Triggering approval for: %s (status: %s)", article.Title, currentStatus)
			if err := w.approvalServer.RequestApproval(article); err != nil {
				log.Printf("Failed to request approval: %v", err)
			}
		} else if currentStatus == "rejected" && article.Status.Update == 1 && w.approvalServer != nil {
			// Article in rejected folder with update flag - send for re-approval
			log.Printf("Update requested for rejected article: %s", article.Title)

			// Change status to publish and move to publish folder
			article.UpdateStatus("publish")
			article.Status.Update = 0 // Clear update flag
			if err := article.WriteFrontmatter(); err != nil {
				log.Printf("Failed to update article status: %v", err)
				return
			}

			// Move to publish folder
			if err := w.mover.MoveArticle(article); err != nil {
				log.Printf("Failed to move article to publish folder: %v", err)
				return
			}

			log.Printf("Triggering re-approval for: %s", article.Title)
			if err := w.approvalServer.RequestApproval(article); err != nil {
				log.Printf("Failed to request approval: %v", err)
			}
		}
	}
}

// Events returns the event channel
func (w *Watcher) Events() <-chan Event {
	return w.events
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	close(w.events)
	return w.watcher.Close()
}

// periodicFolderScan runs a periodic scan of all folders to enforce status-based organization
// This ensures files are in the correct folder based on their status flags, regardless of file events
func (w *Watcher) periodicFolderScan() {
	ticker := time.NewTicker(2 * time.Minute) // Scan every 2 minutes
	defer ticker.Stop()

	for range ticker.C {
		w.scanAllFolders()
	}
}

// scanAllFolders scans all monitored folders and enforces status-based file organization
func (w *Watcher) scanAllFolders() {
	folders, err := w.mover.GetAllMonitoredFolders()
	if err != nil {
		log.Printf("Failed to get monitored folders for scan: %v", err)
		return
	}

	log.Printf("🔍 Starting folder scan to enforce status-based organization...")

	for _, folder := range folders {
		if err := w.scanFolder(folder); err != nil {
			log.Printf("Failed to scan folder %s: %v", folder, err)
		}
	}

	log.Printf("✅ Folder scan complete")
}

// scanFolder scans a single folder and processes all .md files
func (w *Watcher) scanFolder(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories and non-.md files
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		// Skip hidden files
		if entry.Name()[0] == '.' {
			continue
		}

		filePath := filepath.Join(folder, entry.Name())
		if err := w.processArticleFile(filePath); err != nil {
			log.Printf("Failed to process %s: %v", filePath, err)
		}
	}

	return nil
}

// processArticleFile processes a single article file and ensures it's in the correct folder
func (w *Watcher) processArticleFile(filePath string) error {
	// Parse the article
	article, err := common.ParseArticle(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse article: %w", err)
	}

	// Process status changes (will move file if needed)
	if err := w.mover.ProcessArticleStatusChange(filePath); err != nil {
		return fmt.Errorf("failed to process status change: %w", err)
	}

	// Re-parse to get updated file path after potential move
	article, err = common.ParseArticle(article.FilePath)
	if err != nil {
		// File may have been moved, try original path
		article, err = common.ParseArticle(filePath)
		if err != nil {
			return fmt.Errorf("failed to re-parse article: %w", err)
		}
	}

	// Check if article needs approval (publish or update status)
	currentStatus := article.GetCurrentStatus()
	if (currentStatus == "publish" || currentStatus == "update") && w.approvalServer != nil {
		// Check if already pending approval
		// We don't want to spam approval requests for articles already in the queue
		log.Printf("📋 Article ready for approval: %s (status: %s, ID: %s)", article.Title, currentStatus, article.ID)

		// Trigger approval
		if err := w.approvalServer.RequestApproval(article); err != nil {
			return fmt.Errorf("failed to request approval: %w", err)
		}
	}

	return nil
}
