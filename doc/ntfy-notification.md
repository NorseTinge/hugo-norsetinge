# Ntfy Notification System
**Version:** 1.0
**Dato:** 2025-10-03

---

## Formål

Ntfy.sh bruges til at sende push-notifikationer til editor's mobil når en artikel er klar til godkendelse. Notifikationer skal være klare, actionable, og automatisk ryddes når de ikke længere er relevante.

---

## Arkitektur

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Watcher    │ --> │ Approval     │ --> │   Ntfy.sh    │
│  (Detect)    │     │  Server      │     │   (Push)     │
└──────────────┘     └──────────────┘     └──────────────┘
                            │                     │
                            ↓                     ↓
                     ┌──────────────┐     ┌──────────────┐
                     │  Pending     │     │   Mobile     │
                     │  Approvals   │     │   Device     │
                     │   (State)    │     │  (Receive)   │
                     └──────────────┘     └──────────────┘
                            │
                            ↓
                     ┌──────────────┐
                     │   Action     │
                     │  (Approve/   │
                     │   Reject)    │
                     └──────────────┘
                            │
                            ↓
                     ┌──────────────┐
                     │  Clear/Delete│
                     │ Notification │
                     └──────────────┘
```

---

## Ntfy Message Format

### Standard Notification

```json
{
  "topic": "norsetinge-approvals",
  "title": "📰 [Article Title]",
  "message": "Af: [Author]\n\n[Preview URL]",
  "priority": 4,
  "tags": ["newspaper"],
  "actions": [
    {
      "action": "view",
      "label": "Åbn",
      "url": "https://norsetinge.tail2d448.ts.net/preview/...",
      "clear": false
    }
  ],
  "click": "https://norsetinge.tail2d448.ts.net/approve/[article_id]"
}
```

### Field Beskrivelser

**topic:** `norsetinge-approvals`
- Subscriber topic (private, secret)
- Samme topic for alle notifikationer
- Mobil app subscriber til denne topic

**title:** `📰 [Article Title]`
- Emoji prefix for nem genkendelse
- Artikel titel (max 100 chars)
- Format: `📰 DevOps som paradigme: Samarbejde, kultur og inklusion`

**message:**
```
Af: [Author]

[Preview URL]
```
- Linje 1: Forfatter navn
- Linje 2: Blank
- Linje 3: Direkte link til preview

**priority:** `4` (high)
- Ntfy priority levels: 1 (min) til 5 (max urgent)
- 4 = high priority (vibration + sound)
- Sikrer editor ser notifikationen

**tags:** `["newspaper"]`
- Emoji/icon i notification
- Ntfy bruger tags til at vise relevante emojis

**actions:**
```json
[
  {
    "action": "view",
    "label": "Åbn",
    "url": "[preview_url]",
    "clear": false
  }
]
```
- `action: "view"` - Åbn URL i browser
- `label: "Åbn"` - Knap tekst (dansk)
- `url` - Direkte link til preview
- `clear: false` - **VIGTIGT:** Behold notifikation efter klik (ryddes kun ved approve/reject)

**click:** `[approval_url]`
- URL der åbnes ved klik på selve notifikationen
- Går direkte til approval side

---

## HTTP Request Format

Ntfy.sh bruger **HTTP headers** for metadata, ikke JSON body:

```http
POST https://ntfy.sh/norsetinge-approvals
Content-Type: text/plain
Title: 📰 DevOps som paradigme
Priority: 4
Tags: newspaper
Actions: [{"action":"view","label":"Åbn","url":"https://...","clear":false}]
Click: https://norsetinge.tail2d448.ts.net/approve/ECB1F8

Af: TB (twisted brain)

https://norsetinge.tail2d448.ts.net/preview/preview-devops-som-paradigme/
```

**Body:** Plain text message (forfatter + URL)

---

## Go Implementation

### Ntfy Sender

```go
package approval

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type NtfySender struct {
    cfg *config.Config
}

type NtfyMessage struct {
    Topic    string       `json:"topic"`
    Title    string       `json:"title"`
    Message  string       `json:"message"`
    Priority int          `json:"priority"`
    Tags     []string     `json:"tags,omitempty"`
    Actions  []NtfyAction `json:"actions,omitempty"`
    Click    string       `json:"click,omitempty"`
}

type NtfyAction struct {
    Action string `json:"action"`
    Label  string `json:"label"`
    URL    string `json:"url"`
    Clear  bool   `json:"clear,omitempty"`
}

func (n *NtfySender) SendApprovalNotification(
    title, author, previewURL, approvalURL, articleID string,
) error {
    if !n.cfg.Ntfy.Enabled {
        return nil
    }

    msg := NtfyMessage{
        Topic:    n.cfg.Ntfy.Topic,
        Title:    fmt.Sprintf("📰 %s", title),
        Message:  fmt.Sprintf("Af: %s\n\n%s", author, previewURL),
        Priority: 4, // High priority
        Tags:     []string{"newspaper"},
        Click:    approvalURL,
        Actions: []NtfyAction{
            {
                Action: "view",
                Label:  "Åbn",
                URL:    previewURL,
                Clear:  false, // Keep notification until explicitly cleared
            },
        },
    }

    return n.send(msg)
}

// Send via HTTP headers (ntfy.sh preferred method)
func (n *NtfySender) send(msg NtfyMessage) error {
    url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, msg.Topic)

    // Create request with message as body
    req, err := http.NewRequest("POST", url, bytes.NewBufferString(msg.Message))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    // Set metadata via headers
    req.Header.Set("Title", msg.Title)
    req.Header.Set("Priority", fmt.Sprintf("%d", msg.Priority))
    req.Header.Set("Tags", "newspaper")

    if msg.Click != "" {
        req.Header.Set("Click", msg.Click)
    }

    if len(msg.Actions) > 0 {
        actionsJSON, _ := json.Marshal(msg.Actions)
        req.Header.Set("Actions", string(actionsJSON))
    }

    // Send request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send notification: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
    }

    log.Printf("📱 Ntfy notification sent: %s", msg.Title)
    return nil
}
```

---

## Notification Lifecycle

### 1. Create (Send)

**Trigger:** Artikel detekteret med `publish: 1` eller `update: 1`

**Process:**
1. Watcher detekterer artikel i `udgiv/` folder
2. Approval server bygger Hugo preview
3. Ntfy sender genererer notification
4. HTTP POST til ntfy.sh
5. Ntfy.sh sender push til subscriber (mobil)

**Logging:**
```
📋 Article ready for approval: [title] (status: publish, ID: #ECB1F8)
Building Hugo preview for: [title]
Hugo build successful
Preview built and ready at: /preview/...
📤 Sending ntfy notification: 📰 [title]
📱 ntfy notification sent: 📰 [title]
```

---

### 2. Active (Pending Approval)

**State:** Notifikation vises på mobil device

**User Actions:**
- **Klik på notifikation** → Åbner approval URL i browser
- **Klik "Åbn" knap** → Åbner preview URL i browser
- **Ignorer** → Notifikation forbliver synlig

**Backend State:**
```json
// .pending_approvals.json
{
  "articles": [
    {
      "id": "#ECB1F8",
      "title": "DevOps som paradigme",
      "author": "TB",
      "preview_path": "preview-devops-som-paradigme/index.html",
      "notification_sent": true,
      "notification_time": "2025-10-03T11:23:46Z"
    }
  ]
}
```

---

### 3. Clear (After Action)

**Trigger:** Editor godkender eller afviser artikel

**Process:**
1. Editor klikker knap i approval UI:
   - ✅ Godkend
   - ⚡ Godkend + Deploy Nu
   - ❌ Afvis

2. Backend handler action:
   - Flytter artikel til korrekt folder
   - Opdaterer status flags
   - **Fjerner artikel fra pending list**

3. **Send clear/delete notification til ntfy.sh:**

```go
func (n *NtfySender) ClearNotification(articleID string) error {
    // Send empty message with tag to delete
    url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, n.cfg.Ntfy.Topic)

    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }

    // Use article ID as tag for deletion
    req.Header.Set("X-Tags", articleID)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    log.Printf("🗑️ Cleared ntfy notification for: %s", articleID)
    return nil
}
```

**Note:** Ntfy.sh doesn't directly support deleting specific notifications via API. Instead:
- Send new notification med `clear: true` action
- Eller rely on user at dismiss efter action
- Backend tracker kun active pending approvals

---

## Pending Approvals Management

### State File: `.pending_approvals.json`

**Location:** `/home/ubuntu/hugo-norsetinge/Dropbox/Publisering/NorseTinge/.pending_approvals.json`

**Format:**
```json
{
  "articles": [
    {
      "id": "#ECB1F8",
      "title": "DevOps som paradigme: Samarbejde, kultur og inklusion",
      "author": "TB (twisted brain)",
      "file_path": "/path/to/article.md",
      "preview_path": "preview-devops-som-paradigme/index.html",
      "notification_sent": true,
      "notification_time": "2025-10-03T11:23:46Z",
      "created_at": "2025-10-03T11:23:46Z"
    }
  ]
}
```

### Operations

#### Add Pending Article

```go
func (s *Server) addPendingArticle(article *common.Article, previewPath string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    pending := &PendingArticle{
        ID:                article.ID,
        Article:           article,
        PreviewPath:       previewPath,
        NotificationSent:  true,
        NotificationTime:  time.Now(),
    }

    s.pendingArticles[article.ID] = pending
    s.savePendingArticles()
}
```

#### Remove Pending Article (After Action)

```go
func (s *Server) handleApprove(w http.ResponseWriter, r *http.Request) {
    articleID := extractArticleID(r.URL.Path)

    s.mu.Lock()
    pending, exists := s.pendingArticles[articleID]
    if !exists {
        s.mu.Unlock()
        http.Error(w, "Article not found", http.StatusNotFound)
        return
    }

    // Process approval
    pending.Article.UpdateStatus("published")
    s.mover.MoveArticle(pending.Article)

    // REMOVE from pending list
    delete(s.pendingArticles, articleID)
    s.mu.Unlock()

    // Save state
    s.savePendingArticles()

    // Optional: Clear notification (best-effort)
    s.ntfySender.ClearNotification(articleID)

    // Show success page
    http.Redirect(w, r, "/success?action=approved", http.StatusSeeOther)
}
```

---

## Cleanup Strategies

### Strategy 1: Delete from Pending List (Current)

**Pro:**
- Simple implementation
- Reliable state management
- Works offline

**Con:**
- Notification remains on device until user dismisses
- No automatic cleanup on device

**Implementation:**
```go
// After approve/reject action
delete(s.pendingArticles, articleID)
s.savePendingArticles()
```

---

### Strategy 2: Send Clear Message (Future Enhancement)

**Ntfy.sh supports "clear on action":**

```json
{
  "actions": [
    {
      "action": "http",
      "label": "Godkend",
      "url": "https://norsetinge.tail2d448.ts.net/action/approve/ECB1F8",
      "method": "POST",
      "clear": true  // ← Clear notification after click
    }
  ]
}
```

**Pro:**
- Automatic cleanup when action clicked
- Better UX

**Con:**
- Only clears if action clicked via notification
- Doesn't clear if approved via web UI directly

**Implementation:**
```go
Actions: []NtfyAction{
    {
        Action: "http",
        Label:  "Godkend",
        URL:    approveURL,
        Method: "POST",
        Clear:  true,  // Auto-clear on click
    },
    {
        Action: "http",
        Label:  "Afvis",
        URL:    rejectURL,
        Method: "POST",
        Clear:  true,
    },
}
```

---

### Strategy 3: Periodic Cleanup (Recommended)

**Automatically clean old notifications:**

```go
func (s *Server) periodicCleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        s.cleanupOldPending()
    }
}

func (s *Server) cleanupOldPending() {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now()
    maxAge := 24 * time.Hour  // Remove after 24 hours

    for id, pending := range s.pendingArticles {
        if now.Sub(pending.NotificationTime) > maxAge {
            log.Printf("🗑️ Cleaning up old pending approval: %s", pending.Article.Title)
            delete(s.pendingArticles, id)
        }
    }

    s.savePendingArticles()
}
```

---

## Anti-Spam Protection

### Problem: Periodic Folder Scan Resends Notifications

**Issue:** Folder scanner kører hver 2. minut og gensender notifikationer for artikler der allerede er pending.

**Solution: Track Notification State**

```go
func (s *Server) RequestApproval(article *common.Article) error {
    s.mu.Lock()

    // Check if already pending
    if pending, exists := s.pendingArticles[article.ID]; exists {
        if pending.NotificationSent {
            s.mu.Unlock()
            log.Printf("⏭️ Skipping notification - already sent for: %s", article.Title)
            return nil
        }
    }

    s.mu.Unlock()

    // Build preview
    previewPath, err := s.hugoBuilder.BuildPreview(article)
    if err != nil {
        return err
    }

    // Generate URLs
    previewURL := fmt.Sprintf("https://%s/preview/%s",
        s.cfg.Approval.TailscaleHostname, previewPath)

    approvalURL := fmt.Sprintf("https://%s/approve/%s",
        s.cfg.Approval.TailscaleHostname, article.ID)

    // Send notification
    err = s.ntfySender.SendApprovalNotification(
        article.Title,
        article.Author,
        previewURL,
        approvalURL,
        article.ID,
    )
    if err != nil {
        return err
    }

    // Mark as sent
    s.addPendingArticle(article, previewPath)

    return nil
}
```

**Key Points:**
- Check `s.pendingArticles[article.ID]` before sending
- Skip if already exists AND notification sent
- Only send once per article lifecycle

---

## Testing

### Manual Test: Send Notification

```bash
# Send test notification via curl
curl -X POST https://ntfy.sh/norsetinge-approvals \
  -H "Title: 📰 Test Artikel" \
  -H "Priority: 4" \
  -H "Tags: newspaper" \
  -H "Actions: [{\"action\":\"view\",\"label\":\"Åbn\",\"url\":\"https://norsetinge.tail2d448.ts.net/preview/test\"}]" \
  -d "Af: Test Forfatter

https://norsetinge.tail2d448.ts.net/preview/test"
```

### Subscribe on Mobile

**iOS:**
1. Download Ntfy app from App Store
2. Add subscription: `ntfy.sh/norsetinge-approvals`

**Android:**
1. Download Ntfy app from Google Play
2. Add subscription: `ntfy.sh/norsetinge-approvals`

**Web:**
- Visit: `https://ntfy.sh/norsetinge-approvals`

---

## Security

### Topic Privacy

**Topic naming:** `norsetinge-approvals` (not guessable)
- Random/unique topic name
- Secret stored in `.env`: `NTFY_TOPIC=norsetinge-approvals`
- NOT in git repository

**Access control:**
- Anyone with topic name can subscribe
- Use complex topic name as security
- Future: Ntfy.sh supports auth tokens for paid plans

---

## Configuration

### config.yaml

```yaml
ntfy:
  enabled: true
  server: "https://ntfy.sh"
  topic: ""  # Set via env: NTFY_TOPIC
```

### .env

```bash
NTFY_TOPIC=norsetinge-approvals
```

### Environment Variable

```go
func (c *Config) Load() error {
    // Load from .env if exists
    godotenv.Load()

    // Override from environment
    if topic := os.Getenv("NTFY_TOPIC"); topic != "" {
        c.Ntfy.Topic = topic
    }

    return nil
}
```

---

## Monitoring

### Logs

```bash
# Successful notification
📋 Article ready for approval: DevOps som paradigme (status: publish, ID: #ECB1F8)
📤 Sending ntfy notification: 📰 DevOps som paradigme: Samarbejde, kultur og inklusion
📱 ntfy notification sent: 📰 DevOps som paradigme: Samarbejde, kultur og inklusion

# Skipped (already sent)
⏭️ Skipping notification - already sent for: DevOps som paradigme

# Cleared after action
🗑️ Cleared ntfy notification for: #ECB1F8

# Periodic cleanup
🗑️ Cleaning up old pending approval: Old Article (24h+ old)
```

### Metrics

Track:
- Total notifications sent
- Active pending approvals
- Average time to approval
- Notification → Action conversion rate

---

## Future Enhancements

### 1. Rich Actions in Notification

```json
{
  "actions": [
    {
      "action": "http",
      "label": "✅ Godkend",
      "url": "https://.../action/approve/...",
      "method": "POST",
      "clear": true
    },
    {
      "action": "http",
      "label": "❌ Afvis",
      "url": "https://.../action/reject/...",
      "method": "POST",
      "clear": true
    }
  ]
}
```

**Benefit:** Approve/reject directly from notification without opening browser

---

### 2. Attachment with Preview Image

```http
POST https://ntfy.sh/topic
Attach: https://norsetinge.tail2d448.ts.net/preview/image.png
```

**Benefit:** Visual preview in notification

---

### 3. Custom Icons per Category

```http
Tags: tech,blue
```

**Benefit:** Different colors/icons for different article types

---

## Troubleshooting

### Issue: Notification not received

**Check:**
1. Ntfy enabled in config: `ntfy.enabled: true`
2. Topic set correctly: `NTFY_TOPIC` in .env
3. Mobile subscribed to correct topic
4. Internet connection on mobile
5. Ntfy app notifications enabled in OS settings

**Debug:**
```bash
# Test direct curl
curl -d "Test" https://ntfy.sh/norsetinge-approvals
```

---

### Issue: Multiple notifications for same article

**Cause:** Periodic folder scan resends

**Fix:** Anti-spam protection (check pending list before send)

**Verify:**
```bash
# Check pending approvals file
cat /home/ubuntu/hugo-norsetinge/Dropbox/Publisering/NorseTinge/.pending_approvals.json
```

---

### Issue: Notification not cleared after approval

**Expected:** Ntfy.sh doesn't support server-side deletion of notifications

**Solution:**
- Use `clear: true` in actions for auto-dismiss on click
- Or user manually dismisses
- Backend cleans up pending list regardless

---

## Best Practices

✅ **DO:**
- Send notification only once per article
- Track pending state in `.pending_approvals.json`
- Clean up old pending approvals (24h+)
- Use descriptive titles with emoji
- Set priority to 4 (high) for important approvals
- Include direct preview link

❌ **DON'T:**
- Send duplicate notifications
- Include sensitive data in notification body
- Use public/guessable topic names
- Forget to clean up pending list after action
- Send low priority for time-sensitive approvals

---

*Dette dokument definerer den komplette Ntfy notification strategi for Norsetinge approval systemet.*
