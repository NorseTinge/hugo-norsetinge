# Preview HTML Specification
**Version:** 1.0
**Dato:** 2025-10-03

---

## Formål

Preview HTML skal vise artiklen præcis som den vil se ud når den er publiceret online, med godkendelses-knapper til editor.

---

## Arkitektur

```
┌─────────────────────────────────────────────────┐
│           APPROVAL INTERFACE (sticky)           │
│  [✅ Godkend] [⚡ Deploy Nu] [❌ Afvis]         │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│              ARTICLE PREVIEW                    │
│                                                 │
│  Præcis som artiklen vil se ud online:         │
│  - Header med titel + metadata                  │
│  - Artikel indhold (rendered Markdown)          │
│  - Footer med copyright + dato                  │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## Layout Komponenter

### 1. Approval Bar (Sticky Top)

**Position:** Fast øverst på siden (sticky)
**Formål:** Altid tilgængelige godkendelses-knapper

```html
<div class="approval-bar">
    <div class="approval-container">
        <div class="approval-info">
            <h2>📰 Artikel til Godkendelse</h2>
            <p><strong>Titel:</strong> {article.title}</p>
            <p><strong>Forfatter:</strong> {article.author}</p>
            <p><strong>ID:</strong> {article.id}</p>
        </div>

        <div class="approval-actions">
            <button class="btn-approve" onclick="approve()">
                ✅ Godkend
            </button>
            <button class="btn-deploy" onclick="approveDeploy()">
                ⚡ Godkend + Deploy Nu
            </button>
            <button class="btn-reject" onclick="reject()">
                ❌ Afvis
            </button>
        </div>
    </div>
</div>
```

**Styling:**
- `position: sticky; top: 0;`
- `background: rgba(255, 255, 255, 0.98);` (næsten hvid, let transparent)
- `backdrop-filter: blur(10px);` (moderne glasmorfisme effekt)
- `border-bottom: 2px solid #333;`
- `z-index: 1000;` (altid øverst)
- `box-shadow: 0 2px 10px rgba(0,0,0,0.1);`

---

### 2. Article Preview (Main Content)

**Position:** Under approval bar
**Formål:** Vise artiklen præcis som den vil se ud online

```html
<article class="article-preview">
    <header class="article-header">
        <h1>{article.title}</h1>
        <div class="article-meta">
            <span class="author">Af {article.author}</span>
            <span class="date">{current_date}</span>
            {if article.categories}
                <span class="categories">{categories}</span>
            {endif}
        </div>
        {if article.description}
            <p class="lead">{article.description}</p>
        {endif}
    </header>

    <div class="article-content">
        {rendered_markdown_content}
    </div>

    <footer class="article-footer">
        <div class="article-tags">
            {foreach article.tags as tag}
                <span class="tag">{tag}</span>
            {endforeach}
        </div>
        <div class="article-info">
            <p>&copy; {current_year} Norsetinge</p>
            <p>Forfatter: {article.author}</p>
        </div>
    </footer>
</article>
```

---

## Knap Specifikationer

### ✅ Godkend (Normal Approval)

**Farve:** Grøn (`#28a745`)
**Funktion:**
1. Flyt artikel til `udgivet/` folder
2. Sæt `published: 1` i frontmatter
3. Vent på næste periodiske build (maks 10 minutter)
4. Vis success besked

**JavaScript:**
```javascript
function approve() {
    if (confirm('Godkend artikel?\n\nArtiklen vil blive publiceret ved næste periodiske build (ca. 10 minutter).')) {
        window.location.href = `/action/approve/${articleId}`;
    }
}
```

**Backend action:** `POST /action/approve/{article_id}`
- Flytter fil til `udgivet/`
- Opdaterer status flags
- Returnerer success page

---

### ⚡ Godkend + Deploy Nu (Fast Approval)

**Farve:** Orange (`#ff9800`)
**Funktion:**
1. Flyt artikel til `udgivet/` folder
2. Sæt `published: 1` i frontmatter
3. Byg HELE sitet øjeblikkeligt (Hugo)
4. Sync til mirror
5. Git commit + push
6. Rsync til webhost
7. Vis success besked med deploy status

**JavaScript:**
```javascript
function approveDeploy() {
    if (confirm('Godkend + Deploy Nu?\n\n⚠️ Artiklen vil blive bygget og deployeret med det samme.\n\nDette kan tage 30-60 sekunder.')) {
        // Vis loading spinner
        showLoading('Deployer artikel...');

        window.location.href = `/action/approve-deploy/${articleId}`;
    }
}
```

**Backend action:** `POST /action/approve-deploy/{article_id}`
- Flytter fil til `udgivet/`
- Kører komplet build+deploy pipeline
- Returnerer deploy status page

---

### ❌ Afvis (Reject)

**Farve:** Rød (`#dc3545`)
**Funktion:**
1. Flyt artikel til `afvist/` folder
2. Sæt `rejected: 1` i frontmatter
3. Vis besked om hvordan man retter og genindsender

**JavaScript:**
```javascript
function reject() {
    if (confirm('Afvis artikel?\n\nDu kan rette den i afvist/ mappen og sende til godkendelse igen med update: 1')) {
        window.location.href = `/action/reject/${articleId}`;
    }
}
```

**Backend action:** `POST /action/reject/{article_id}`
- Flytter fil til `afvist/`
- Opdaterer status flags
- Returnerer info page om næste skridt

---

## Styling (CSS)

### Approval Bar

```css
.approval-bar {
    position: sticky;
    top: 0;
    background: rgba(255, 255, 255, 0.98);
    backdrop-filter: blur(10px);
    border-bottom: 2px solid #333;
    padding: 15px 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}

.approval-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 20px;
}

.approval-info h2 {
    margin: 0 0 10px 0;
    font-size: 1.2em;
}

.approval-info p {
    margin: 5px 0;
    font-size: 0.9em;
    color: #666;
}

.approval-actions {
    display: flex;
    gap: 15px;
}

.approval-actions button {
    padding: 12px 30px;
    font-size: 16px;
    font-weight: 600;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.3s ease;
    color: white;
}

.btn-approve {
    background: #28a745;
}

.btn-approve:hover {
    background: #218838;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(40, 167, 69, 0.3);
}

.btn-deploy {
    background: #ff9800;
}

.btn-deploy:hover {
    background: #e68900;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(255, 152, 0, 0.3);
}

.btn-reject {
    background: #dc3545;
}

.btn-reject:hover {
    background: #c82333;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(220, 53, 69, 0.3);
}
```

### Article Preview

```css
.article-preview {
    max-width: 800px;
    margin: 40px auto;
    padding: 0 20px;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    line-height: 1.6;
    color: #333;
}

.article-header {
    border-bottom: 2px solid #333;
    padding-bottom: 20px;
    margin-bottom: 30px;
}

.article-header h1 {
    font-size: 2.5em;
    margin: 0 0 15px 0;
    line-height: 1.2;
}

.article-meta {
    display: flex;
    gap: 15px;
    flex-wrap: wrap;
    font-size: 0.9em;
    color: #666;
}

.article-meta .author {
    font-weight: 600;
}

.article-meta .categories {
    background: #f0f0f0;
    padding: 3px 10px;
    border-radius: 3px;
}

.lead {
    font-size: 1.2em;
    color: #666;
    font-style: italic;
    margin-top: 20px;
}

.article-content {
    font-size: 1.1em;
    line-height: 1.8;
}

.article-content h2 {
    font-size: 1.8em;
    margin-top: 40px;
    margin-bottom: 15px;
}

.article-content h3 {
    font-size: 1.4em;
    margin-top: 30px;
    margin-bottom: 10px;
}

.article-content p {
    margin-bottom: 15px;
}

.article-content ul,
.article-content ol {
    margin-bottom: 15px;
    padding-left: 30px;
}

.article-content li {
    margin-bottom: 8px;
}

.article-content code {
    background: #f5f5f5;
    padding: 2px 6px;
    border-radius: 3px;
    font-family: "Courier New", monospace;
    font-size: 0.9em;
}

.article-content pre {
    background: #f5f5f5;
    padding: 15px;
    border-radius: 5px;
    overflow-x: auto;
}

.article-content blockquote {
    border-left: 4px solid #333;
    padding-left: 20px;
    margin-left: 0;
    color: #666;
    font-style: italic;
}

.article-footer {
    margin-top: 50px;
    padding-top: 30px;
    border-top: 1px solid #ddd;
}

.article-tags {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    margin-bottom: 20px;
}

.tag {
    background: #e0e0e0;
    padding: 5px 12px;
    border-radius: 15px;
    font-size: 0.9em;
    color: #333;
}

.article-info {
    font-size: 0.9em;
    color: #666;
}

.article-info p {
    margin: 5px 0;
}
```

---

## Responsivt Design

### Mobile (< 768px)

```css
@media (max-width: 768px) {
    .approval-container {
        flex-direction: column;
        align-items: stretch;
    }

    .approval-actions {
        flex-direction: column;
        width: 100%;
    }

    .approval-actions button {
        width: 100%;
    }

    .article-header h1 {
        font-size: 1.8em;
    }

    .article-content {
        font-size: 1em;
    }
}
```

---

## Backend Template (Go html/template)

```html
<!DOCTYPE html>
<html lang="da">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Preview: {{.Article.Title}}</title>
    <link rel="stylesheet" href="/assets/preview.css">
</head>
<body>
    <!-- Approval Bar (Sticky) -->
    <div class="approval-bar">
        <div class="approval-container">
            <div class="approval-info">
                <h2>📰 Artikel til Godkendelse</h2>
                <p><strong>Titel:</strong> {{.Article.Title}}</p>
                <p><strong>Forfatter:</strong> {{.Article.Author}}</p>
                <p><strong>ID:</strong> {{.Article.ID}}</p>
            </div>

            <div class="approval-actions">
                <button class="btn-approve" onclick="approve('{{.Article.ID}}')">
                    ✅ Godkend
                </button>
                <button class="btn-deploy" onclick="approveDeploy('{{.Article.ID}}')">
                    ⚡ Godkend + Deploy Nu
                </button>
                <button class="btn-reject" onclick="reject('{{.Article.ID}}')">
                    ❌ Afvis
                </button>
            </div>
        </div>
    </div>

    <!-- Article Preview -->
    <article class="article-preview">
        <header class="article-header">
            <h1>{{.Article.Title}}</h1>
            <div class="article-meta">
                <span class="author">Af {{.Article.Author}}</span>
                <span class="date">{{.Date}}</span>
                {{if .Article.Categories}}
                    {{range .Article.Categories}}
                        <span class="categories">{{.}}</span>
                    {{end}}
                {{end}}
            </div>
            {{if .Article.Description}}
                <p class="lead">{{.Article.Description}}</p>
            {{end}}
        </header>

        <div class="article-content">
            {{.RenderedContent}}
        </div>

        <footer class="article-footer">
            {{if .Article.Tags}}
                <div class="article-tags">
                    {{range .Article.Tags}}
                        <span class="tag">{{.}}</span>
                    {{end}}
                </div>
            {{end}}
            <div class="article-info">
                <p>&copy; {{.Year}} Norsetinge</p>
                <p>Forfatter: {{.Article.Author}}</p>
            </div>
        </footer>
    </article>

    <script src="/assets/preview.js"></script>
</body>
</html>
```

---

## JavaScript Interaktivitet

```javascript
// preview.js

function approve(articleId) {
    if (confirm('Godkend artikel?\n\nArtiklen vil blive publiceret ved næste periodiske build (ca. 10 minutter).')) {
        window.location.href = `/action/approve/${articleId}`;
    }
}

function approveDeploy(articleId) {
    if (confirm('Godkend + Deploy Nu?\n\n⚠️ Artiklen vil blive bygget og deployeret med det samme.\n\nDette kan tage 30-60 sekunder.')) {
        // Vis loading overlay
        showLoading('Deployer artikel...');

        window.location.href = `/action/approve-deploy/${articleId}`;
    }
}

function reject(articleId) {
    if (confirm('Afvis artikel?\n\nDu kan rette den i afvist/ mappen og sende til godkendelse igen med update: 1')) {
        window.location.href = `/action/reject/${articleId}`;
    }
}

function showLoading(message) {
    const overlay = document.createElement('div');
    overlay.id = 'loading-overlay';
    overlay.innerHTML = `
        <div class="loading-spinner">
            <div class="spinner"></div>
            <p>${message}</p>
        </div>
    `;
    document.body.appendChild(overlay);
}

// Loading overlay styling (add to CSS)
/*
#loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
}

.loading-spinner {
    text-align: center;
    color: white;
}

.spinner {
    border: 4px solid rgba(255, 255, 255, 0.3);
    border-top: 4px solid white;
    border-radius: 50%;
    width: 50px;
    height: 50px;
    animation: spin 1s linear infinite;
    margin: 0 auto 20px;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}
*/
```

---

## Success/Error Pages

### Success Page (efter godkendelse)

```html
<!DOCTYPE html>
<html lang="da">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Artikel Godkendt</title>
</head>
<body>
    <div class="success-container">
        <div class="success-icon">✅</div>
        <h1>Artikel Godkendt!</h1>
        <p><strong>{{.Article.Title}}</strong></p>

        {{if .FastDeploy}}
            <div class="deploy-status">
                <h2>Deployment Status</h2>
                <ul>
                    <li>✓ Flyttet til udgivet/</li>
                    <li>✓ Hugo site bygget</li>
                    <li>✓ Sync til mirror</li>
                    <li>✓ Git commit + push</li>
                    <li>✓ Rsync til webhost</li>
                </ul>
                <p class="success-message">
                    Artiklen er nu live på norsetinge.com
                </p>
            </div>
        {{else}}
            <p class="info-message">
                Artiklen vil blive publiceret ved næste periodiske build (maks 10 minutter).
            </p>
        {{end}}

        <a href="/dashboard" class="btn-primary">Tilbage til Dashboard</a>
    </div>
</body>
</html>
```

### Rejection Page

```html
<!DOCTYPE html>
<html lang="da">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Artikel Afvist</title>
</head>
<body>
    <div class="reject-container">
        <div class="reject-icon">❌</div>
        <h1>Artikel Afvist</h1>
        <p><strong>{{.Article.Title}}</strong></p>

        <div class="info-box">
            <h2>Næste skridt:</h2>
            <ol>
                <li>Find artiklen i <code>afvist/</code> mappen</li>
                <li>Foretag de nødvendige rettelser</li>
                <li>Sæt <code>update: 1</code> i frontmatter</li>
                <li>Systemet sender ny godkendelse automatisk</li>
            </ol>
        </div>

        <a href="/dashboard" class="btn-primary">Tilbage til Dashboard</a>
    </div>
</body>
</html>
```

---

## Markdown Rendering

**Backend skal konvertere Markdown til HTML:**

```go
import (
    "github.com/gomarkdown/markdown"
    "github.com/gomarkdown/markdown/html"
    "github.com/gomarkdown/markdown/parser"
)

func renderMarkdown(content string) string {
    // Create parser with extensions
    extensions := parser.CommonExtensions | parser.AutoHeadingIDs
    p := parser.NewWithExtensions(extensions)

    // Create HTML renderer
    htmlFlags := html.CommonFlags | html.HrefTargetBlank
    opts := html.RendererOptions{Flags: htmlFlags}
    renderer := html.NewRenderer(opts)

    // Parse and render
    doc := p.Parse([]byte(content))
    return string(markdown.Render(doc, renderer))
}
```

---

## Accessibility

- `<button>` elementer har tydelige labels
- ARIA labels hvor nødvendigt
- Keyboard navigation understøttet
- Kontrast ratio > 4.5:1 (WCAG AA)
- Focus states på alle interaktive elementer

---

## Performance

- CSS inlined i `<head>` for faster first paint
- Minimal JavaScript (vanilla JS, no frameworks)
- Lazy loading ikke nødvendig (single page)
- Caching headers for static assets

---

## Security

- CSRF tokens på alle POST requests
- Input sanitization på markdown rendering
- XSS protection via proper HTML escaping
- Tailscale network isolation (ikke public internet)

---

## Testing Checklist

- [ ] Preview viser korrekt artikel content
- [ ] Alle 3 knapper virker
- [ ] Sticky header forbliver øverst ved scroll
- [ ] Responsivt på mobil (320px+)
- [ ] Confirmation dialogs før actions
- [ ] Success/error pages vises korrekt
- [ ] Markdown rendering (headings, lists, code, etc.)
- [ ] Tags og categories vises korrekt

---

*Dette dokument definerer den komplette preview HTML specification for Norsetinge approval systemet.*
