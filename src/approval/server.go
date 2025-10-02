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

	"norsetinge/builder"
	"norsetinge/common"
	"norsetinge/config"
)

// Server handles approval web requests
type Server struct {
	cfg             *config.Config
	emailSender     *EmailSender
	hugoBuilder     *builder.HugoBuilder
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
	ID       string
	Article  *common.Article
	Approved bool
	Rejected bool
	Comments string
}

// NewServer creates a new approval server
func NewServer(cfg *config.Config) *Server {
	s := &Server{
		cfg:             cfg,
		emailSender:     NewEmailSender(cfg),
		hugoBuilder:     builder.NewHugoBuilder(cfg),
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
	http.HandleFunc("/action/reject/", s.handleReject)

	addr := fmt.Sprintf("%s:%d", s.cfg.Approval.Host, s.cfg.Approval.Port)
	log.Printf("Approval server starting on %s", addr)

	return http.ListenAndServe(addr, nil)
}

// RequestApproval creates approval request and sends email
func (s *Server) RequestApproval(article *common.Article) error {
	// Use article's unique ID (generated when first parsed)
	id := article.ID

	// Build Hugo preview
	log.Printf("Building Hugo preview for: %s", article.Title)
	htmlPath, err := s.hugoBuilder.BuildPreview(article)
	if err != nil {
		return fmt.Errorf("failed to build Hugo preview: %w", err)
	}

	// Store pending article with preview path
	s.mu.Lock()
	s.pendingArticles[id] = &PendingArticle{
		ID:      id,
		Article: article,
	}
	s.mu.Unlock()

	// Persist to disk
	if err := s.savePendingArticles(); err != nil {
		log.Printf("Warning: Failed to save pending articles: %v", err)
	}

	// Generate approval action URL
	approvalURL := fmt.Sprintf("http://%s:%d/approve/%s",
		s.cfg.Approval.TailscaleHostname,
		s.cfg.Approval.Port,
		id,
	)

	// Send HTML email with inline preview
	if err := s.emailSender.SendApprovalRequestHTML(
		article.Title,
		article.Author,
		htmlPath,
		approvalURL,
		id, // Pass article ID for email header
	); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
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

// handleApprove handles approval action
func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/action/approve/"):]

	s.mu.Lock()
	pending, exists := s.pendingArticles[id]
	if exists {
		pending.Approved = true
	}
	s.mu.Unlock()

	if !exists {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	log.Printf("Article approved: %s - continuing to translation", pending.Article.Title)

	// Remove from pending list
	if err := s.removePendingArticle(id); err != nil {
		log.Printf("Warning: Failed to remove pending article: %v", err)
	}

	// Status is already "publish" - article stays in udgiv/ for translation
	// No status change needed, translation will happen next

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html><head><meta charset="UTF-8"><title>Godkendt</title></head>
		<body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; text-align: center;">
			<h1>✅ Artikel Godkendt!</h1>
			<p>Artiklen vil nu blive oversat til 22 sprog og publiceret.</p>
		</body></html>
	`)
}

// handleReject handles rejection action
func (s *Server) handleReject(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/action/reject/"):]

	s.mu.Lock()
	pending, exists := s.pendingArticles[id]
	if exists {
		pending.Rejected = true
	}
	s.mu.Unlock()

	if !exists {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
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

	// Remove from pending list
	if err := s.removePendingArticle(id); err != nil {
		log.Printf("Warning: Failed to remove pending article: %v", err)
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

// HandleEmailReply processes an email reply and takes appropriate action
func (s *Server) HandleEmailReply(reply *EmailReply) error {
	// Find the pending article by article ID from email header
	s.mu.RLock()
	targetArticle, exists := s.pendingArticles[reply.ArticleID]
	s.mu.RUnlock()

	if !exists || targetArticle == nil {
		log.Printf("No pending article found for ArticleID: %s (Subject: %s)", reply.ArticleID, reply.Subject)
		return fmt.Errorf("no matching article found for ID: %s", reply.ArticleID)
	}

	// Process based on detected action
	switch reply.Action {
	case ActionApprove:
		log.Printf("Email approval received for: %s", targetArticle.Article.Title)
		// Article stays in udgiv/ for translation (no status change needed)
		targetArticle.Approved = true

		// Remove from pending list
		if err := s.removePendingArticle(reply.ArticleID); err != nil {
			log.Printf("Warning: Failed to remove pending article: %v", err)
		}

	case ActionReject:
		log.Printf("Email rejection received for: %s", targetArticle.Article.Title)
		targetArticle.Rejected = true
		targetArticle.Article.UpdateStatus("rejected")
		if err := targetArticle.Article.WriteFrontmatter(); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		if s.mover != nil {
			if err := s.mover.MoveArticle(targetArticle.Article); err != nil {
				return fmt.Errorf("failed to move article: %w", err)
			}
		}

		// Remove from pending list
		if err := s.removePendingArticle(reply.ArticleID); err != nil {
			log.Printf("Warning: Failed to remove pending article: %v", err)
		}

	default:
		log.Printf("Unknown action in email reply: %s", reply.Subject)
		return fmt.Errorf("unknown action: %v", reply.Action)
	}

	return nil
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

// removePendingArticle removes and persists the change
func (s *Server) removePendingArticle(id string) error {
	s.mu.Lock()
	delete(s.pendingArticles, id)
	s.mu.Unlock()

	return s.savePendingArticles()
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
        <h2>Preview</h2>
        <pre>{{.Article.Content}}</pre>
    </div>

    <div class="actions">
        <a href="/action/approve/{{.ID}}" class="button approve">✅ Godkend og publicer</a>
        <a href="/action/reject/{{.ID}}" class="button reject" onclick="return confirm('Afvis artikel?\n\nDu kan rette den i afvist/ mappen og sætte update: 1 for at sende den til godkendelse igen.')">❌ Afvis</a>
    </div>
</body>
</html>
`
