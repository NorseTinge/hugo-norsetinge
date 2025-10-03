# Norsetinge - Project Status & Current Plan
**Opdateret:** 2025-10-03
**Status:** Phase 1 Implementation Complete, Testing in Progress

---

## Projekt Oversigt

**Norsetinge** er et automatiseret nyhedspubliceringssystem der håndterer hele workflow'et fra artikel-skrivning til publicering på webhost, med godkendelse via mobil-notifikationer.

### Vision
At skabe en højt automatiseret nyhedsplatform der kan:
1. **Phase 1 (Current):** Publicere artikler i original sprog (dansk) med minimal manuel indgriben
2. **Phase 2 (Future):** Oversætte artikler til 22 sprog og publicere globalt

---

## Arkitektur (KISS-princippet)

### Simplified Pipeline - Phase 1

```
┌──────────┐     ┌─────────────┐     ┌──────┐     ┌────────┐     ┌────────┐     ┌─────────┐
│ Dropbox  │ --> │ Go Watcher  │ --> │ Hugo │ --> │ Mirror │ --> │  Git   │ --> │ Webhost │
│ (Input)  │     │ + Approval  │     │Build │     │ (1:1)  │     │ Backup │     │ (Rsync) │
└──────────┘     └─────────────┘     └──────┘     └────────┘     └────────┘     └─────────┘
                        ↓
                   ┌─────────┐
                   │  Ntfy   │
                   │  Push   │
                   └─────────┘
```

### Komponenter

1. **Input:** Dropbox folders (folder-baseret workflow)
2. **Processing:** Go application med:
   - File watcher (fsnotify)
   - Markdown parser (frontmatter + content)
   - Hugo builder (preview + full site)
   - Approval server (HTTP over Tailscale)
   - File mover (status-based organization)
   - Deployer (mirror + git + rsync)
3. **Notifications:** Ntfy.sh push til mobil
4. **Access:** Tailscale privat netværk
5. **Build:** Hugo static site generator
6. **Deploy:** Git mirror + Rsync til webhost

---

## Workflow: Fra Artikel til Webhost

### 1. Skriv Artikel (kladde/)

```markdown
---
title: "Din artikel titel"
author: "TB (twisted brain)"
description: "Kort resume"
tags: ["tag1", "tag2"]
categories: ["Kategori"]
status:
  draft: 0
  revision: 0
  publish: 0    # Sæt til 1 når klar
  published: 0
  rejected: 0
  update: 0
---

## Din artikel indhold her

Skriv i Markdown format...
```

**Status:** Alle flags = `0` → Systemet ignorerer artiklen

### 2. Klar til Publicering (udgiv/)

- Sæt `status: publish: 1` i frontmatter
- Systemet opdager automatisk (folder scan hver 2. minut)
- Artiklen flyttes automatisk til `udgiv/` folder

### 3. Preview & Godkendelse

**Automatisk proces:**
1. Hugo bygger preview HTML af artiklen
2. Ntfy.sh sender push-notifikation til mobil
3. Notifikation indeholder:
   - Artikel titel + forfatter
   - Link til preview: `https://norsetinge.tail2d448.ts.net/preview/...`
4. Åbn preview i browser (via Tailscale - sikker)

### 4. Godkendelse (3 valgmuligheder)

**✅ Godkend:**
- Flytter artikel til `udgivet/`
- Venter på periodisk build (hver 10. minut)

**⚡ Godkend + Deploy Nu:**
- Flytter artikel til `udgivet/`
- Bygger + deployer øjeblikkeligt
- Bruger samme pipeline som periodisk build

**❌ Afvis:**
- Flytter artikel til `afvist/`
- Kan rettes og genindsendes med `update: 1`

### 5. Build & Deploy

**A. Periodisk Build (hver 10. minut):**

```
1. Hugo Build:
   └─> Læs ALLE artikler fra udgivet/
   └─> Byg komplet site til site/public/

2. Mirror Sync:
   └─> Copy site/public/ → site/mirror/ (1:1 kopi)

3. Git Version Control:
   └─> git add site/mirror/
   └─> git commit -m "Update: [timestamp]"
   └─> git push origin main

4. Rsync Deploy:
   └─> rsync -avz --delete site/mirror/ user@webhost:/path/
```

**B. Manuel Deploy (via "Godkend + Deploy Nu"):**
- Samme proces som ovenfor
- Køres øjeblikkeligt i stedet for at vente

---

## Folder Struktur

```
Dropbox/Publisering/NorseTinge/
├── kladde/                 # Arbejde under udvikling
│   └── artikel.md          # status flags alle = 0
│
├── udgiv/                  # Klar til godkendelse
│   └── artikel.md          # publish: 1 eller update: 1
│
├── afvist/                 # Afviste artikler
│   └── artikel.md          # rejected: 1 (kan rettes)
│
├── udgivet/                # Publicerede artikler
│   └── artikel.md          # published: 1 (Hugo source)
│
└── skabeloner/             # Artikel templates
    └── artikel-skabelon.md
```

---

## Status Flag System (KISS)

### Princip
**Systemet reagerer KUN hvis mindst ét flag er sat til `1`**

### Status Types

| Flag       | Betydning                    | Handling                        |
|------------|------------------------------|---------------------------------|
| `draft: 1` | Artikel under udvikling      | Ingen - forbliver i kladde/     |
| `publish: 1` | Klar til første godkendelse | → Preview + Ntfy notification  |
| `update: 1` | Opdatering/rettelse          | → Preview + Ntfy notification  |
| `rejected: 1` | Afvist af editor          | Flyttet til afvist/             |
| `published: 1` | Godkendt og publiceret   | Flyttet til udgivet/            |

### Regler
- **Alle flags = 0:** Systemet ignorerer filen
- **"Last 1 wins":** Hvis flere flags er `1`, tæller den sidste i rækkefølgen
- **Auto-move:** Filer flyttes automatisk baseret på status

---

## Mirror System (Smart Deploy)

### Struktur

```
hugo-norsetinge/
├── site/
│   ├── public/     # Hugo build output (temporary, .gitignore)
│   │   ├── index.html
│   │   ├── articles/
│   │   └── ...
│   │
│   └── mirror/     # 1:1 kopi af webhost (git tracked)
│       ├── .git/   # Version control
│       ├── index.html
│       ├── articles/
│       └── ...
```

### Hvorfor Mirror?

**Problem:** Hugo bygger til `public/`, men vi vil have:
- Git version control af live site
- Automatisk sletning af fjernede filer på webhost

**Løsning:** Mirror system

1. **Hugo bygger til public/**
   - Temporary output
   - Ikke git tracked (i .gitignore)

2. **Sync til mirror/**
   - `rsync -a --delete public/ mirror/`
   - 1:1 kopi af hvad der skal være på webhost
   - Git tracked (privat repo)

3. **Git commit + push**
   - Komplet version history
   - Rollback capability
   - Backup af hele site

4. **Rsync til webhost**
   - `rsync -avz --delete mirror/ user@webhost:/path/`
   - `--delete` fjerner filer på webhost der ikke findes i mirror
   - Hygiejnisk: webhost matcher mirror 100%

### Fordele

✅ **Version Control:** Hele websitet er git tracked
✅ **Rollback:** Git history giver mulighed for at rulle tilbage
✅ **Backup:** Privat repo er komplet backup
✅ **Auto-cleanup:** `--delete` flag fjerner gamle/slettede artikler
✅ **Test/Validate:** Mirror kan testes lokalt før deploy

---

## Teknisk Stack

### Backend
- **Go 1.25.1:** Main application language
- **fsnotify:** File system watcher
- **gopkg.in/yaml.v3:** YAML frontmatter parsing

### Frontend/Build
- **Hugo:** Static site generator
- **Markdown:** Article format

### Infrastructure
- **Tailscale:** Private network (HTTPS proxy)
- **Ntfy.sh:** Push notifications
- **Git:** Version control (private repo)
- **Rsync:** Deploy tool
- **Dropbox:** Input directory sync

### Services
- **IMAP:** Email reply monitoring (future)
- **OpenRouter API:** Translation service (Phase 2)

---

## Aktuel Implementation Status

### ✅ Phase 1: COMPLETE

**Core Infrastructure:**
- [x] Dropbox directory structure
- [x] Go application skeleton
- [x] Config loader (config.yaml)
- [x] File watcher (monitors 6 folders)
- [x] Markdown parser (frontmatter + content)
- [x] Status flag system (KISS principle)
- [x] File mover (status-based organization)

**Preview & Approval:**
- [x] Hugo preview builder (single article)
- [x] Approval web server (HTTP over Tailscale)
- [x] Tailscale HTTPS proxy setup
- [x] Ntfy.sh push notifications
- [x] 3-button approval UI:
  - ✅ Godkend (normal)
  - ⚡ Godkend + Deploy Nu (fast)
  - ❌ Afvis

**Publication Pipeline:**
- [x] Hugo full-site builder (all articles)
- [x] Mirror sync (public → mirror)
- [x] Git automation (commit + push)
- [x] Rsync deployment (mirror → webhost)
- [x] Periodic build+deploy (every 10 minutes)
- [x] Manual fast deploy (on demand)
- [x] Auto-archive (move to udgivet/)

**Bug Fixes (2025-10-03):**
- [x] Fixed: Status flag system ignorerede artikler uden flags
- [x] Fixed: FilePath blev læst fra YAML (skulle være internt)
- [x] Fixed: Gamle test-artikler sendte spam-notifikationer

### ⚠️ Current Issues

1. **Preview Content Missing:**
   - Problem: Artikel har `content:` som YAML felt i stedet for efter `---`
   - Impact: Approval UI viser tom preview
   - Fix: Ret artikel format

### 📋 Phase 2: FUTURE (Deferred)

**Translation Pipeline:**
- [ ] OpenRouter API integration
- [ ] Automatic translation to 22 languages
- [ ] Multilingual Hugo configuration
- [ ] Language switcher frontend
- [ ] Cookie-based language preference

**Advanced Features:**
- [ ] Image processing automation
- [ ] Favicon/icon generation
- [ ] Internal ad system
- [ ] Analytics integration

---

## Configuration

### config.yaml

```yaml
# Dropbox paths
dropbox:
  base_path: "/home/ubuntu/hugo-norsetinge/Dropbox/Publisering/NorseTinge"
  folder_language: "da"

# Approval server
approval:
  host: "0.0.0.0"
  port: 8080
  tailscale_hostname: "norsetinge.tail2d448.ts.net"

# Ntfy push notifications
ntfy:
  enabled: true
  server: "https://ntfy.sh"
  topic: "[SECRET]"  # Set in .env as NTFY_TOPIC

# Hugo paths
hugo:
  site_dir: "/home/ubuntu/hugo-norsetinge/site"
  public_dir: "/home/ubuntu/hugo-norsetinge/site/public"
  mirror_dir: "/home/ubuntu/hugo-norsetinge/site/mirror"

# Git mirror
git:
  mirror_repo: "git@github.com:username/norsetinge-mirror.git"
  auto_commit: false  # Disabled for testing

# Rsync deployment
rsync:
  enabled: false  # Disabled for testing
  host: "norsetinge.com"
  user: "deploy"
  target_path: "/var/www/norsetinge.com"
  ssh_key: ""

# Supported languages (Phase 2)
languages:
  - en  # English (primary)
  - da  # Danish
  - sv  # Swedish
  - no  # Norwegian
  - fi  # Finnish
  - de  # German
  - fr  # French
  - it  # Italian
  - es  # Spanish
  - el  # Greek
  - kl  # Greenlandic
  - is  # Icelandic
  - fo  # Faroese
  - ru  # Russian
  - tr  # Turkish
  - uk  # Ukrainian
  - et  # Estonian
  - lv  # Latvian
  - lt  # Lithuanian
  - zh  # Chinese
  - ko  # Korean
  - ja  # Japanese
```

---

## Testing Plan

### ✅ Unit Testing (Done)
- Status flag parsing
- File movement logic
- YAML frontmatter parsing

### 🔄 Integration Testing (In Progress)
- [ ] Fix artikel format (content efter frontmatter)
- [ ] Test complete pipeline med 4-5 artikler:
  1. Kladde → udgiv (med publish: 1)
  2. Preview + Ntfy notification
  3. Godkend (normal) → Vent på periodic build
  4. Godkend + Deploy Nu → Øjeblikkelig deploy
  5. Afvis → Ret → Update (med update: 1)

### 📋 Production Testing (Pending)
- [ ] Enable git auto_commit
- [ ] Enable rsync deployment
- [ ] Test rollback fra git history
- [ ] Verify --delete flag på webhost
- [ ] Monitor periodic build (10 min interval)

---

## Deployment

### Development Environment
- **Host:** LXC container "norsetinge"
- **OS:** Ubuntu Linux
- **Network:** Tailscale private network
- **Access:** SSH + Tailscale web UI

### Production Deployment (når klar)

```bash
# 1. Enable git + rsync i config.yaml
git:
  auto_commit: true

rsync:
  enabled: true

# 2. Setup SSH keys for rsync
ssh-keygen -t ed25519 -f ~/.ssh/norsetinge_deploy
ssh-copy-id -i ~/.ssh/norsetinge_deploy.pub deploy@norsetinge.com

# 3. Build production binary
cd /home/ubuntu/hugo-norsetinge/src
go build -o ../norsetinge

# 4. Setup systemd service (auto-start)
sudo systemctl enable norsetinge.service
sudo systemctl start norsetinge.service

# 5. Setup Tailscale serve (permanent)
sudo tailscale serve --bg --https=443 localhost:8080
```

---

## Monitoring & Logs

### Real-time Logs

```bash
# Follow application logs
journalctl -u norsetinge -f

# Or hvis kører i background:
tail -f /var/log/norsetinge.log
```

### Key Log Events

```
📄 Event: 0 - /path/to/article.md           # File created
📄 Event: 1 - /path/to/article.md           # File modified
📋 Article ready for approval: [title]      # Detected publish: 1
🔨 Building full site...                    # Periodic build started
✅ Full site built successfully             # Build complete
📋 Syncing public → mirror...               # Mirror sync
✅ Deployment complete!                     # Rsync finished
📤 Sending ntfy notification: [title]       # Push notification
📱 ntfy notification sent: [title]          # Notification delivered
```

---

## Known Issues & Workarounds

### Issue 1: Artikel Format
**Problem:** Cursor/Windsurf AI genererede artikel med `content:` som YAML felt
**Impact:** Preview viser tom content
**Fix:** Ret artikel så content er EFTER `---` frontmatter

### Issue 2: Periodic Scan Spam
**Problem:** Gamle test-artikler sendte notifikationer hver 2. minut
**Status:** FIXED - ryddet `.pending_approvals.json` og gamle filer

### Issue 3: FilePath YAML Pollution
**Problem:** `filepath:` felt i YAML blev læst af parser
**Status:** FIXED - tilføjet `yaml:"-"` tag til FilePath field

---

## Next Steps (Prioriteret)

### 1. Fix Current Article Format ⚠️
- [ ] Ret "DevOps som paradigme.md" til korrekt format
- [ ] Test preview viser content korrekt
- [ ] Test godkendelse workflow

### 2. Complete Pipeline Test
- [ ] Test 4-5 artikler gennem hele pipeline
- [ ] Verify periodic build (10 min)
- [ ] Verify manual deploy
- [ ] Check mirror sync

### 3. Enable Production Features
- [ ] Enable git auto_commit
- [ ] Enable rsync deployment
- [ ] Test complete deploy to webhost

### 4. Documentation
- [ ] User guide (dansk)
- [ ] Troubleshooting guide
- [ ] API documentation (for fremtidig integration)

### 5. Phase 2 Planning
- [ ] OpenRouter API research
- [ ] Translation workflow design
- [ ] Multilingual Hugo setup

---

## Support & Troubleshooting

### Common Problems

**Q: Ntfy notification ikke modtaget?**
A: Check at `ntfy.enabled: true` og topic er korrekt i config.yaml

**Q: Preview viser ikke på Tailscale URL?**
A: Verify `tailscale serve status` - skal vise proxy til localhost:8080

**Q: Artikel flytter ikke automatisk?**
A: Check status flags - mindst ét skal være `1`

**Q: Periodic build kører ikke?**
A: Check logs for "⏰ Running periodic build+deploy..."

### Debug Mode

```bash
# Run med verbose logging
./norsetinge --debug

# Check watcher events
grep "📄 Event:" /var/log/norsetinge.log

# Check status changes
grep "status:" /var/log/norsetinge.log
```

---

## Team & Contacts

**Developer:** Claude + TB (twisted brain)
**Project Start:** 2025-10-02
**Current Status:** Phase 1 Implementation Complete
**Last Updated:** 2025-10-03

---

## License & Repository

**Private Project** - Not open source
**Mirror Repo:** git@github.com:username/norsetinge-mirror.git (private)
**Main Repo:** TBD

---

*Dette dokument opdateres løbende som projektet udvikler sig.*
