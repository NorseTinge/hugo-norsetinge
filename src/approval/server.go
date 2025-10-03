package approval

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"norsetinge/src/builder"
	"norsetinge/src/common"
	"norsetinge/src/config"
	"norsetinge/src/deployer"
)

// Server handles approval web requests
type Server struct {
	cfg             *config.Config
	ntfySender      *NtfySender
	hugoBuilder     *builder.HugoBuilder
	deployer        *deployer.Deployer
	pendingArticles map[string]*PendingArticle
	mu              sync.RWMutex
	mover           FileMover
}

// FileMover interface for moving files based on status
type FileMover interface {
	MoveArticle(article *common.Article) error
}

// PendingArticle represents an article awaiting approval
type PendingArticle struct {
	ID               string
	Article          *common.Article
	PreviewPath      string // Hugo preview HTML path (for iframe)
	Approved         bool
	Rejected         bool
	Comments         string
	NotificationSent bool   // To prevent re-sending notifications
}

// NewServer creates a new approval server
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg:             cfg,
		ntfySender:      NewNtfySender(cfg),
		hugoBuilder:     builder.NewHugoBuilder(cfg),
		deployer:        deployer.NewDeployer(cfg),
		pendingArticles: make(map[string]*PendingArticle),
	}

	// Load pending articles from disk
	if err := s.loadPendingArticles(); err != nil {
		log.Printf("Warning: Failed to load pending articles: %v", err)
	}

	return s
}

// SetMover sets the file mover for handling file movements
func (s *Server) SetMover(mover FileMover) {
	s.mover = mover
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Serve Hugo public directory for previews
	http.Handle("/preview/", http.StripPrefix("/preview/", http.FileServer(http.Dir(s.cfg.Hugo.PublicDir))))

	http.HandleFunc("/approve/", s.handleApproval)
	http.HandleFunc("/action/approve/", s.handleApprove)
	http.HandleFunc("/action/approve-deploy/", s.handleApproveAndDeploy)
	http.HandleFunc("/action/reject/", s.handleReject)

	addr := fmt.Sprintf("%s:%d", s.cfg.Approval.Host, s.cfg.Approval.Port)
	log.Printf("Approval server starting on %s", addr)

	return http.ListenAndServe(addr, nil)
}

// RequestApproval creates approval request and sends notification if not already sent.
func (s *Server) RequestApproval(article *common.Article) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := article.ID

	// Check if a notification has already been sent for this article.
	if pending, exists := s.pendingArticles[id]; exists && pending.NotificationSent {
		log.Printf("⏭️ Skipping notification, already pending for: %s", article.Title)
		return nil
	}

	// --- The rest of the operations happen inside the lock to ensure consistency ---

	// Build Hugo preview
	log.Printf("Building Hugo preview for: %s", article.Title)
	htmlPath, err := s.hugoBuilder.BuildPreview(article)
	if err != nil {
		return fmt.Errorf("failed to build Hugo preview: %w", err)
	}

	// Send ntfy push notification
	if s.cfg.Ntfy.Enabled {
		if err := s.ntfySender.SendApprovalNotification(
			article.Title,
			article.Author,
			htmlPath,
			id,
		); err != nil {
			log.Printf("Warning: Failed to send ntfy notification: %v", err)
			// Do not mark as sent if it fails, so it can be retried.
			return err
		}
	}

	// Store pending article with preview path and mark notification as sent.
	s.pendingArticles[id] = &PendingArticle{
		ID:               id,
		Article:          article,
		PreviewPath:      htmlPath,
		NotificationSent: true, // Mark as sent
	}

	// Persist to disk
	if err := s.savePendingArticles(); err != nil {
		log.Printf("Warning: Failed to save pending articles: %v", err)
	}

	log.Printf("Approval request sent for: %s (ID: %s, Preview: %s)", article.Title, id, htmlPath)
	return nil
}

// handleApproval shows the approval page
func (s *Server) handleApproval(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/approve/"):]

	s.mu.RLock()
	pending, exists := s.pendingArticles[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.New("approval").Parse(approvalTemplate))
	tmpl.Execute(w, pending)
}

// handleApprove handles normal approval (no immediate deploy)
func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/action/approve/"):]

	s.mu.Lock()
	pending, exists := s.pendingArticles[id]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}
	delete(s.pendingArticles, id)
	s.mu.Unlock()

	// Persist the change to the pending articles list on disk
	if err := s.savePendingArticles(); err != nil {
		log.Printf("Warning: Failed to save pending articles list: %v", err)
	}

	log.Printf("Article approved: %s - moving to udgivet/", pending.Article.Title)

	// Move article to udgivet/
	pending.Article.UpdateStatus("published")
	if err := pending.Article.WriteFrontmatter(); err != nil {
		log.Printf("Error updating article status: %v", err)
		http.Error(w, "Failed to update article", http.StatusInternalServerError)
		return
	}

	if s.mover != nil {
		if err := s.mover.MoveArticle(pending.Article); err != nil {
			log.Printf("Error moving article: %v", err)
		} else {
			log.Printf("✓ Article moved to udgivet/: %s", pending.Article.Title)
		}
	}

	// Clean up preview files
	s.cleanupPreviewFiles(pending.Article)

	// Clear ntfy notification for this approval
	if s.cfg.Ntfy.Enabled {
		if err := s.ntfySender.ClearAllNotifications(); err != nil {
			log.Printf("Warning: Failed to clear ntfy notifications: %v", err)
		}
	}

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html><head><meta charset="UTF-8"><title>Godkendt</title></head>
		<body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; text-align: center;">
			<h1>✅ Artikel Godkendt!</h1>
			<p>Artiklen er flyttet til udgivet/ og vil blive deployeret ved næste automatiske build (hver 10-15 min).</p>
		</body></html>
	`)
}

// handleApproveAndDeploy handles immediate approval + deploy
func (s *Server) handleApproveAndDeploy(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/action/approve-deploy/"):]

	s.mu.Lock()
	pending, exists := s.pendingArticles[id]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}
	delete(s.pendingArticles, id)
	s.mu.Unlock()

	// Persist the change to the pending articles list on disk
	if err := s.savePendingArticles(); err != nil {
		log.Printf("Warning: Failed to save pending articles list: %v", err)
	}

	log.Printf("Article approved with immediate deploy: %s", pending.Article.Title)

	// 1. Move article to udgivet/
	pending.Article.UpdateStatus("published")
	if err := pending.Article.WriteFrontmatter(); err != nil {
		log.Printf("Error updating article status: %v", err)
		http.Error(w, "Failed to update article", http.StatusInternalServerError)
		return
	}

	if s.mover != nil {
		if err := s.mover.MoveArticle(pending.Article); err != nil {
			log.Printf("Error moving article: %v", err)
		} else {
			log.Printf("✓ Article moved to udgivet/: %s", pending.Article.Title)
		}
	}

	// Clean up preview files
	s.cleanupPreviewFiles(pending.Article)

	// Clear ntfy notification for this approval
	if s.cfg.Ntfy.Enabled {
		if err := s.ntfySender.ClearAllNotifications(); err != nil {
			log.Printf("Warning: Failed to clear ntfy notifications: %v", err)
		}
	}

	// 2. Build full Hugo site
	publicDir, mirrorDir, err := s.hugoBuilder.BuildFullSite()
	if err != nil {
		log.Printf("Error building site: %v", err)
		http.Error(w, "Failed to build site", http.StatusInternalServerError)
		return
	}

	// 3. Deploy (mirror-sync + git + rsync)
	if err := s.deployer.Deploy(publicDir, mirrorDir); err != nil {
		log.Printf("Error deploying: %v", err)
		http.Error(w, "Failed to deploy site", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html><head><meta charset="UTF-8"><title>Deployeret</title></head>
		<body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; text-align: center;">
			<h1>⚡ Artikel Godkendt & Deployeret!</h1>
			<p>Artiklen er nu live på norsetinge.com</p>
			<p style="color: #666; font-size: 14px;">Bygget, deployeret og arkiveret i udgivet/</p>
		</body></html>
	`)
}

// handleReject handles rejection action
func (s *Server) handleReject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/action/reject/"):]

	s.mu.Lock()
	pending, exists := s.pendingArticles[id]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}
	delete(s.pendingArticles, id)
	s.mu.Unlock()

	// Persist the change to the pending articles list on disk
	if err := s.savePendingArticles(); err != nil {
		log.Printf("Warning: Failed to save pending articles list: %v", err)
	}

	// Update article status to rejected and move file
	pending.Article.UpdateStatus("rejected")
	if err := pending.Article.WriteFrontmatter(); err != nil {
		log.Printf("Error updating article status: %v", err)
		http.Error(w, "Failed to update article", http.StatusInternalServerError)
		return
	}

	// Move file to afvist/
	if s.mover != nil {
		if err := s.mover.MoveArticle(pending.Article); err != nil {
			log.Printf("Error moving article: %v", err)
		}
	}

	log.Printf("Article rejected: %s", pending.Article.Title)

	// Clean up preview files
	s.cleanupPreviewFiles(pending.Article)

	// Clear ntfy notification for this rejection
	if s.cfg.Ntfy.Enabled {
		if err := s.ntfySender.ClearAllNotifications(); err != nil {
			log.Printf("Warning: Failed to clear ntfy notifications: %v", err)
		}
	}

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html><head><meta charset="UTF-8"><title>Afvist</title></head>
		<body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; text-align: center;">
			<h1>❌ Artikel Afvist</h1>
			<p>Artiklen er flyttet til afvist/ mappen.</p>
		</body></html>
	`)
}

// generateID generates a random ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getPendingArticlesPath returns the path to pending articles storage file
func (s *Server) getPendingArticlesPath() string {
	return filepath.Join(s.cfg.Dropbox.BasePath, ".pending_approvals.json")
}

// savePendingArticles persists pending articles to disk
func (s *Server) savePendingArticles() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s.pendingArticles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal pending articles: %w", err)
	}

	path := s.getPendingArticlesPath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write pending articles file: %w", err)
	}

	log.Printf("Saved %d pending articles to %s", len(s.pendingArticles), path)
	return nil
}

// cleanupPreviewFiles removes preview files from public and mirror directories
func (s *Server) cleanupPreviewFiles(article *common.Article) {
	slug := article.GetSlug()
	previewDirName := fmt.Sprintf("preview-%s", slug)

	// Clean up from public directory
	publicPreviewPath := filepath.Join(s.cfg.Hugo.PublicDir, previewDirName)
	if err := os.RemoveAll(publicPreviewPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove preview from public: %v", err)
	} else if err == nil {
		log.Printf("🧹 Cleaned up preview from public/: %s", previewDirName)
	}

	// Clean up from mirror directory
	mirrorPreviewPath := filepath.Join(s.cfg.Hugo.MirrorDir, previewDirName)
	if err := os.RemoveAll(mirrorPreviewPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove preview from mirror: %v", err)
	} else if err == nil {
		log.Printf("🧹 Cleaned up preview from mirror/: %s", previewDirName)
	}

	// Also clean up the temporary content file if it still exists
	contentPath := filepath.Join(s.cfg.Hugo.SiteDir, "content", fmt.Sprintf("preview-%s.md", slug))
	if err := os.Remove(contentPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove preview content: %v", err)
	}
}

// loadPendingArticles loads pending articles from disk
func (s *Server) loadPendingArticles() error {
	path := s.getPendingArticlesPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, not an error
			return nil
		}
		return fmt.Errorf("failed to read pending articles file: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, &s.pendingArticles); err != nil {
		return fmt.Errorf("failed to unmarshal pending articles: %w", err)
	}

	log.Printf("Loaded %d pending articles from %s", len(s.pendingArticles), path)
	return nil
}


const approvalTemplate = `
<!DOCTYPE html>
<html lang="da">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Godkend Artikel - {{.Article.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            line-height: 1.6;
        }
        .header {
            border-bottom: 2px solid #333;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .article-preview {
            background: #f5f5f5;
            padding: 30px;
            margin: 30px 0;
            border-radius: 8px;
        }
        .actions {
            display: flex;
            gap: 15px;
            margin-top: 30px;
            justify-content: center;
        }
        .button {
            padding: 15px 40px;
            font-size: 18px;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            font-weight: 600;
        }
        .approve { background: #28a745; color: white; }
        .approve-deploy { background: #ff9800; color: white; }
        .reject { background: #dc3545; color: white; }
        .info-box {
            background: #e7f3ff;
            border-left: 4px solid #2196F3;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📰 Artikel til Godkendelse</h1>
        <p><strong>Titel:</strong> {{.Article.Title}}</p>
        <p><strong>Forfatter:</strong> {{.Article.Author}}</p>
    </div>

    <div class="info-box">
        <strong>💡 Tip:</strong> Hvis artiklen skal rettes, afvis den. Du kan derefter rette den i <code>afvist/</code> mappen og sætte <code>update: 1</code> for at sende den til godkendelse igen.
    </div>

    <div class="article-preview">
        <h2>Artikel Preview</h2>
        <iframe src="/preview/{{.PreviewPath}}" style="width: 100%; height: 600px; border: 1px solid #ddd; border-radius: 4px;"></iframe>
    </div>

    <div class="actions">
        <a href="/action/approve/{{.ID}}" class="button approve">✅ Godkend</a>
        <a href="/action/approve-deploy/{{.ID}}" class="button approve-deploy" onclick="return confirm('Deploy øjeblikkeligt?\n\nArtiklen vil blive bygget og deployeret med det samme.')">⚡ Godkend + Deploy Nu</a>
        <a href="/action/reject/{{.ID}}" class="button reject" onclick="return confirm('Afvis artikel?\n\nDu kan rette den i afvist/ mappen.')">❌ Afvis</a>
    </div>
</body>
</html>
`
