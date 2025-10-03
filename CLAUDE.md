# CLAUDE.md: Norsetinge Project Initialization

## Project Status Analysis (2025-10-03)

**KISS Approach:** Simplified pipeline focusing on complete workflow before adding complexity.

## Current Project Structure

```
hugo-norsetinge/
├── doc/
│   └── project_plan.md       # Detailed Danish architecture documentation
├── site/                      # Hugo static site
│   ├── content/               # Article content (markdown)
│   ├── public/                # Hugo build output (generated)
│   ├── mirror/                # 1:1 copy of webhost (git tracked, rsync source)
│   ├── archetypes/
│   └── hugo.toml
├── src/                       # Go application source
│   └── go.mod
├── Dropbox -> /home/ubuntu/Dropbox  # Input directory symlink
├── GEMINI.md                  # English overview
└── CLAUDE.md                  # This file
```

## Simplified Pipeline Flow (KISS)

**Phase 1: Single Language (Danish/Original) Only**

1. **Draft** → Article placed in `Dropbox/Publisering/NorseTinge/kladde/`
2. **Ready to Publish** → Move article to `udgiv/` folder
3. **Watch & Preview** → Watcher detects file → Hugo builds preview
4. **Approval Request** → Ntfy push notification with Tailscale preview link
5. **Approve** → Via web interface over Tailscale
6. **Build Site** → Hugo builds full site to `site/public/`
7. **Mirror Sync** → Copy `site/public/` → `site/mirror/`
8. **Version Control** → Git commit + push `site/mirror/` to private repo
9. **Deploy** → Rsync `site/mirror/` → webhost with `--delete` flag
10. **Archive** → Move article to `udgivet/` folder

**Folder Structure (Simplified):**
```
Dropbox/Publisering/NorseTinge/
├── kladde/         # Drafts (work in progress)
├── udgiv/          # Ready to publish (triggers approval)
├── afvist/         # Rejected articles
├── udgivet/        # Published articles (archive)
└── skabeloner/     # Article templates
```

**Removed for simplicity:**
- ~~Complex status flag system~~ → Use folder location only
- ~~`afventer-rettelser/` folder~~ → Edit rejected articles in `afvist/`, move to `udgiv/` when ready
- ~~Email notifications~~ → Ntfy only
- ~~Multilingual translation~~ → Add later as Phase 2

## Deployment Architecture

**Mirror System:**
- `site/public/` → Hugo build output (temporary, not git tracked)
- `site/mirror/` → Exact copy of webhost (git tracked, private repo)
- Rsync syncs `site/mirror/` → webhost with `--delete` flag

**Benefits:**
- Full version control of published content
- Can test/validate in mirror before deployment
- Rsync automatically removes deleted content on webhost
- Git history provides rollback capability
- Complete backup of live site

## Implementation Status

> **Note**: This project follows a KISS (Keep It Simple, Stupid) approach. The original vision (doc/project_plan.md) included automatic translation to 22+ languages, but we've pivoted to complete the full pipeline for one language first. See CHANGELOG.md for the complete rationale.

### ✅ Phase 1: Core Infrastructure (COMPLETE)
- Dropbox directory structure ✓
- Go application with config loader, file watcher, markdown parser ✓
- Basic approval web server (port 8080, Tailscale) ✓
- Ntfy push notifications ✓

### ✅ Phase 2: Preview & Approval (COMPLETE)
- Hugo preview builder (single article) ✓
- Web-based approval interface ✓
- Ntfy notifications with preview links ✓
- Approve/Reject actions ✓
- File movement based on folder location ✓

### ✅ Phase 3: Publication Pipeline (COMPLETE)
- Hugo full-site build to `site/public/` ✓
- Mirror sync: `site/public/` → `site/mirror/` ✓
- Git automation: commit + push mirror to private repo ✓
- Rsync deployment: `site/mirror/` → webhost with `--delete` ✓
- Archive: move article to `udgivet/` after deployment ✓

### 📋 Phase 4: Future Enhancements (DEFERRED)
- Translation pipeline (OpenRouter API, 22 languages)
- Multilingual Hugo configuration
- Language switcher frontend
- Image processing automation
- Internal ad system

**For detailed change history and rationale, see CHANGELOG.md**

## Configuration Requirements

**Current config.yaml needs:**
```yaml
hugo:
  site_dir: "/path/to/site"
  public_dir: "/path/to/site/public"
  mirror_dir: "/path/to/site/mirror"

git:
  mirror_repo: "git@github.com:username/norsetinge-mirror.git"
  auto_commit: true

rsync:
  enabled: true
  host: "norsetinge.com"
  user: "deploy"
  target_path: "/var/www/norsetinge.com"
  ssh_key: "/path/to/ssh/key"
```

## Key Technical Decisions

- **Go version:** 1.25.1
- **Hugo:** Static site generator (single language initially)
- **Approval:** Tailscale-secured web app + Ntfy notifications
- **Deployment:** Git mirror + rsync with --delete
- **Status System:** Folder-based (no complex flags)
- **Translation:** Deferred to Phase 4

## Notes

- No complex status flags - folder location determines state
- Ntfy-only notifications (no email)
- Single language (Danish/original) initially
- Mirror provides git history and rsync source
- All components run in LXC container "norsetinge"
