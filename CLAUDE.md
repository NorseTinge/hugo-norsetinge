# CLAUDE.md: Norsetinge Project Initialization

## Project Status Analysis (2025-10-02)

Based on GEMINI.md and doc/project_plan.md, this is the current state and initialization plan for the Norsetinge automated multilingual news service.

## Current Project Structure

```
hugo-norsetinge/
├── doc/
│   └── project_plan.md       # Detailed Danish architecture documentation
├── site/                      # Hugo static site (basic setup exists)
│   ├── archetypes/
│   └── hugo.toml
├── src/                       # Go application source (module initialized)
│   └── go.mod
├── Dropbox -> /home/ubuntu/Dropbox  # Input directory symlink
├── GEMINI.md                  # English overview
└── CLAUDE.md                  # This file
```

## Project Architecture Summary

**Pipeline Flow:**
1. Markdown article placed in Dropbox `udgiv/` directory
2. Watch script detects new file (watcher monitors all 6 folders)
3. Hugo builds preview page for the article (single language - original)
4. Go app sends approval email with link to Hugo preview + web approval form
5. Approval options:
   - **Approve** → Article stays in `udgiv/`, continues to translation
   - **Request Revision** → Article moved to `afventer-rettelser/` with comments
   - **Reject** → Article moved to `afvist/` with rejected:1 status
6. Upon approval: OpenRouter API translates to 22 languages
7. Hugo generates complete static multilingual site
8. Site deployed via rsync to norsetinge.com
9. Original article moved to `udgivet/` with published:1 status

**Article Status System:**
Articles use a `status` field in frontmatter with ordered flags:
```yaml
status:
  draft: 1        # kladde/
  revision: 0     # afventer-rettelser/
  publish: 0      # udgiv/ (triggers pipeline)
  published: 0    # udgivet/
  rejected: 0     # afvist/
  update: 0       # opdater publiceret artikel (re-trigger pipeline)
```
**Rule:** Last `1` in the sequence determines current status. All previous flags can remain `1` to preserve history.
Example: `draft:1, revision:1, publish:1` → status is "publish"
Example: `draft:1, publish:1, published:1, update:1` → re-publish updated article

## Components to Build

### 1. Go Application (`src/`)
**Purpose:** Orchestrate entire pipeline
- File watcher for Dropbox `udgiv/` directory
- Approval web server (Tailscale accessible)
- OpenRouter API integration for translation
- **Image processor (`src/builder`)**: Automatic image scaling
  - Input: Original image (minimum size required)
  - Output: Multiple formats/sizes for responsive images
  - Formats: WebP, JPEG, PNG
  - Sizes: Thumbnail, medium, large, Open Graph (1200x630)
- Hugo content preparation and execution
- Email notification system
- File archival with metadata injection
- Deploy via rsync

### 2. Hugo Site Configuration (`site/`)
**Current:** Basic hugo.toml exists
**Needed:**
- Multilingual configuration (20+ languages)
- Custom theme with language switcher
- JavaScript for browser language detection
- Cookie-based language preference storage
- Footer with publication date, author, copyright
- GitHub Issues feedback link

### 3. Languages to Support
**Primary:** English (`en`)
**Core Nordic:** Danish (`da`), Swedish (`sv`), Norwegian (`no`), Finnish (`fi`), Icelandic (`is`), Faroese (`fo`), Greenlandic (`kl`)
**European:** German (`de`), French (`fr`), Italian (`it`), Spanish (`es`), Greek (`el`), Russian (`ru`), Turkish (`tr`), Ukrainian (`uk`), Estonian (`et`), Latvian (`lv`), Lithuanian (`lt`)
**Asian:** Chinese (`zh`), Korean (`ko`), Japanese (`ja`)

### 4. Directory Structure (Dropbox)
```
Dropbox/Publisering/NorseTinge/  (symlink to actual Dropbox)
├── kladde/                   # Drafts (status.draft: 1)
├── udgiv/                    # Ready to publish (status.publish: 1) - triggers pipeline
├── afventer-rettelser/       # Needs revision (status.revision: 1)
├── afvist/                   # Rejected (status.rejected: 1)
├── udgivet/                  # Published (status.published: 1) - with metadata
└── skabeloner/               # Article templates
```

### 5. Configuration Files Needed
- OpenRouter API credentials
- Email SMTP configuration
- Tailscale network setup
- Deployment rsync target
- Hugo multilingual config
- Language mapping (code -> full name)
- **Folder aliases** (`folder-aliases.yaml`) - Multilingual folder names
  - Currently supports: English (en), Danish (da)
  - New languages can be added manually or via OpenRouter LLM on demand

## Development Order

1. **Phase 1: Core Infrastructure** ✅ COMPLETE
   - Set up Dropbox directory structure ✓
   - Create Go application skeleton with modules:
     - Config loader (config.yaml) ✓
     - File watcher (monitor all 6 Dropbox folders) ✓
     - Markdown parser (frontmatter: title, author, status) ✓
     - Status evaluator (find last `1` in status sequence) ✓
     - File mover (move files based on status changes) ✓
     - Approval web server (port 8080) ✓
     - Email sender (SMTP via mail.norsetinge.com) ✓
   - Basic Hugo multilingual configuration ✓
   - Folder aliases system (en/da) ✓

2. **Phase 2: Approval System** ✅ COMPLETE
   - Web-based approval interface ✓
   - Email approval notifications ✓
   - IMAP email reader for reply monitoring ✓
   - Three approval actions (approve/revision/reject) ✓
   - File movement based on approval decision ✓

3. **Phase 3: Hugo Preview Builder** ✅ COMPLETE
   - Hugo single-article preview builder ✓
   - Pre-approval article rendering (original language only) ✓
   - Email links to Hugo-built preview page ✓
   - Serve preview via approval web server (/preview/) ✓
   - Simple Hugo layout for article display ✓

4. **Phase 4: Translation Pipeline**
   - OpenRouter API integration
   - Language contract system (`hugin.json`)
   - Translate article to 22 languages
   - Hugo multilingual content structure generation

5. **Phase 5: Frontend**
   - Hugo theme with language switcher
   - JavaScript language detection
   - Cookie preference management
   - Footer with metadata display

6. **Phase 6: Deployment**
   - Hugo build automation (full multilingual site)
   - rsync deployment script
   - Article archival with metadata (move to `udgivet/`)
   - Error handling and logging

7. **Phase 7: Enhancement**
   - Pipeline status dashboard
   - Internal ad system (data/ads.yaml)
   - CLI tool expansion (`hugin`)
   - Image processing automation

## Key Technical Decisions

- **Go version:** 1.25.1 (per src/go.mod)
- **Hugo:** Static site generator
- **Translation:** OpenRouter API
- **Approval:** Tailscale-secured web app
- **Deployment:** rsync to standard webhost
- **URL structure:** `norsetinge.com/artikel/{lang}/{slug}/` (English default without lang code)

## Next Steps for Implementation

1. Create Dropbox directory structure
2. Initialize Go application structure with proper modules
3. Set up Hugo multilingual configuration
4. Implement file watcher and basic approval flow
5. Add OpenRouter translation integration
6. Build frontend language handling
7. Implement deployment pipeline

## Notes

- Approval email contains Tailscale link to preview
- Author format: "TB (twisted brain)" in frontmatter
- Cookie consent message in user's browser language
- Publication metadata injected post-deployment
- All components run in LXC container "norsetinge"
