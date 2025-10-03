# Mirror Site Design & Structure

**Opdateret:** 2025-10-03
**Status:** Phase 1 Implementation (Single Language)

---

## Oversigt

Mirror-systemet er en central del af Norsetinge's deployment-arkitektur. Det giver:

- 📋 **1:1 kopi** af webhost lokalt
- 🔄 **Git version control** af hele websitet
- 🧹 **Automatisk cleanup** via rsync --delete
- ⏪ **Rollback capability** via git history
- 🔒 **Backup** i privat repo

---

## Arkitektur Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Hugo Build Process                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Dropbox/udgivet/                                                │
│  └── artikler.md (source)                                        │
│           │                                                       │
│           ├──> Hugo Build                                        │
│           │                                                       │
│           v                                                       │
│  site/public/ (TEMPORARY - .gitignore)                          │
│  ├── index.html                                                  │
│  ├── articles/                                                   │
│  │   └── artikel-slug/                                          │
│  │       └── index.html                                         │
│  ├── categories/                                                 │
│  ├── tags/                                                       │
│  ├── css/                                                        │
│  ├── js/                                                         │
│  └── images/                                                     │
│           │                                                       │
│           ├──> rsync -a --delete                                │
│           │                                                       │
│           v                                                       │
│  site/mirror/ (GIT TRACKED)                                     │
│  ├── .git/                                                       │
│  ├── index.html                                                  │
│  ├── articles/                                                   │
│  │   └── artikel-slug/                                          │
│  │       └── index.html                                         │
│  ├── categories/                                                 │
│  ├── tags/                                                       │
│  ├── css/                                                        │
│  ├── js/                                                         │
│  └── images/                                                     │
│           │                                                       │
│           ├──> git commit + push                                │
│           │                                                       │
│           ├──> rsync -avz --delete                              │
│           │                                                       │
│           v                                                       │
│  Webhost: /var/www/norsetinge.com/ (PRODUCTION)                │
│  ├── index.html                                                  │
│  ├── articles/                                                   │
│  ├── categories/                                                 │
│  ├── tags/                                                       │
│  ├── css/                                                        │
│  ├── js/                                                         │
│  └── images/                                                     │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Folder Struktur

### Phase 1: Single Language (Current)

**site/public/ (temporary, not git tracked):**
```
public/
├── index.html              # Forside
├── articles/               # Alle artikler
│   ├── artikel-slug-1/
│   │   └── index.html
│   ├── artikel-slug-2/
│   │   └── index.html
│   └── ...
├── categories/             # Kategori-sider
│   ├── teknologi/
│   │   └── index.html
│   └── ...
├── tags/                   # Tag-sider
│   ├── ai/
│   │   └── index.html
│   └── ...
├── css/                    # Stylesheets
│   └── main.css
├── js/                     # JavaScript
│   └── main.js
├── images/                 # Billeder
│   └── ...
├── favicon.ico
├── robots.txt
└── sitemap.xml
```

**site/mirror/ (1:1 copy, git tracked):**
```
mirror/
├── .git/                   # Git repository
│   └── ...
├── index.html              # Identisk med public/
├── articles/               # Identisk med public/
│   ├── artikel-slug-1/
│   │   └── index.html
│   └── ...
├── categories/             # Identisk med public/
├── tags/                   # Identisk med public/
├── css/                    # Identisk med public/
├── js/                     # Identisk med public/
├── images/                 # Identisk med public/
├── favicon.ico
├── robots.txt
└── sitemap.xml
```

---

## Phase 2: Multilingual (Future)

### Planned Multilingual Structure

Hugo er konfigureret til 22 sprog, men i Phase 1 bruges kun dansk/original sprog.

**Når Phase 2 implementeres:**

```
public/ eller mirror/
├── index.html              # Forside (language selector)
├── da/                     # Dansk (original)
│   ├── index.html
│   ├── articles/
│   │   └── artikel-slug/
│   │       └── index.html
│   ├── categories/
│   └── tags/
├── en/                     # Engelsk
│   ├── index.html
│   ├── articles/
│   │   └── artikel-slug/  # Translated slug
│   │       └── index.html
│   ├── categories/
│   └── tags/
├── sv/                     # Svensk
│   ├── index.html
│   ├── articles/
│   └── ...
├── no/                     # Norsk
├── fi/                     # Finsk
├── de/                     # Tysk
├── fr/                     # Fransk
├── it/                     # Italiensk
├── es/                     # Spansk
├── el/                     # Græsk
├── kl/                     # Grønlandsk
├── is/                     # Islandsk
├── fo/                     # Færøsk
├── ru/                     # Russisk
├── tr/                     # Tyrkisk
├── uk/                     # Ukrainsk
├── et/                     # Estisk
├── lv/                     # Lettisk
├── lt/                     # Litauisk
├── zh/                     # Kinesisk
├── ko/                     # Koreansk
├── ja/                     # Japansk
├── css/                    # Delt mellem alle sprog
├── js/                     # Delt mellem alle sprog
└── images/                 # Delt mellem alle sprog
```

### Language Switcher (Phase 2)

**Frontend design:**
- Cookie-based language preference
- Flag/language selector i header
- Automatic redirect baseret på browser language
- Fallback til dansk hvis sprog ikke tilgængeligt

---

## Sync-mekanisme

### 1. Hugo Build → Public

```bash
cd /home/ubuntu/hugo-norsetinge/site
hugo --gc --minify
```

**Output:**
- Bygger alle artikler fra `udgivet/` til `public/`
- Genererer indexes, categories, tags
- Minifier CSS/JS
- Optimerer images (hvis configured)

**Tid:** ~2-5 sekunder (afhænger af antal artikler)

### 2. Public → Mirror Sync

```bash
rsync -a --delete \
  --exclude '.git' \
  /home/ubuntu/hugo-norsetinge/site/public/ \
  /home/ubuntu/hugo-norsetinge/site/mirror/
```

**Flags forklaret:**
- `-a` = Archive mode (preserve permissions, timestamps, etc.)
- `--delete` = Fjern filer i mirror/ der ikke findes i public/
- `--exclude '.git'` = Behold git repo i mirror/

**Hvad sker der:**
- Alle nye filer kopieres fra public/ → mirror/
- Ændrede filer opdateres
- Slettede filer fjernes fra mirror/
- Git repo (.git/) bevares intakt

**Tid:** ~1-2 sekunder (kun ændringer synces)

### 3. Git Commit & Push

```bash
cd /home/ubuntu/hugo-norsetinge/site/mirror
git add .
git commit -m "Update: $(date +%Y-%m-%d_%H:%M:%S)"
git push origin main
```

**Fordele:**
- Komplet version history af hele website
- Rollback capability (git reset --hard <commit>)
- Backup i privat GitHub repo
- Kan sammenligne ændringer mellem builds (git diff)

**Tid:** ~3-5 sekunder (afhænger af netværk)

### 4. Mirror → Webhost Deploy

```bash
rsync -avz --delete \
  --exclude '.git' \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Flags forklaret:**
- `-a` = Archive mode
- `-v` = Verbose (logging)
- `-z` = Compress under transfer
- `--delete` = Fjern filer på webhost der ikke findes i mirror/
- `--exclude '.git'` = Upload ikke .git folder til webhost

**Hvad sker der:**
- Kun ændrede filer uploades (delta transfer)
- Slettede artikler fjernes fra webhost
- Webhost matcher mirror/ 100%

**Tid:** ~5-15 sekunder (afhænger af antal ændringer + netværk)

---

## Hvorfor Mirror System?

### Problem: Hugo bygger til public/

Hugo bygger naturligt til `public/` folder, men:

❌ Vi kan ikke git-tracke `public/` (det er temporary build output)
❌ Rsync fra `public/` direkte betyder ingen version control
❌ Ingen rollback capability
❌ Ingen backup af live site

### Løsning: Mirror som Mellemled

✅ **public/** = Temporary Hugo output (.gitignored)
✅ **mirror/** = Git-tracked 1:1 kopi af webhost
✅ **mirror/** = Rsync source til webhost
✅ **mirror/.git/** = Version control + backup

---

## Deployment Scenarios

### Scenario 1: Ny Artikel Publiceres

1. Artikel godkendt → flyttet til `udgivet/`
2. Periodic build (hver 10. minut) trigger:
   - Hugo build → `public/`
   - Sync → `mirror/`
   - Git commit: "Update: 2025-10-03_14:20:00"
   - Git push → GitHub backup
   - Rsync → webhost
3. Ny artikel live på norsetinge.com

**Ændringer:**
- +1 ny artikel HTML fil
- +kategori/tag opdateringer
- +sitemap.xml update
- +index.html update (seneste artikler)

### Scenario 2: Artikel Rettes/Opdateres

1. Artikel i `udgivet/` rettes
2. Periodic build:
   - Hugo rebuild → `public/`
   - Sync → `mirror/` (kun ændret artikel synces)
   - Git commit: "Update: artikel-slug ændret"
   - Rsync → webhost (kun ændret fil uploades)

**Ændringer:**
- ~1 opdateret artikel HTML fil
- Evt. kategori/tag updates

### Scenario 3: Artikel Slettes

1. Artikel fjernes fra `udgivet/`
2. Periodic build:
   - Hugo rebuild → `public/` (artikel ikke inkluderet)
   - Sync → `mirror/` (artikel-folder slettes pga. --delete flag)
   - Git commit: "Update: removed artikel-slug"
   - Rsync → webhost (artikel-folder slettes på webhost)

**Ændringer:**
- -1 artikel HTML fil (slettet)
- -kategori/tag opdateringer
- -sitemap.xml update
- -index.html update

### Scenario 4: Rollback til Tidligere Version

**Problem:** Fejl opdaget på live site, skal rulle tilbage.

**Løsning:**

```bash
# 1. Find sidste fungerende commit
cd /home/ubuntu/hugo-norsetinge/site/mirror
git log --oneline
# 95f8ecf Update: 2025-10-03_14:00:00  (GOOD)
# a1b2c3d Update: 2025-10-03_14:10:00  (BAD - rollback from this)

# 2. Rollback til tidligere commit
git reset --hard 95f8ecf

# 3. Force push til backup repo
git push --force origin main

# 4. Rsync til webhost (restorer old version)
rsync -avz --delete \
  --exclude '.git' \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/

# 5. Webhost nu tilbage til fungerende version
```

**Tid:** ~30 sekunder total

---

## Git Strategi

### Repository: Private GitHub Repo

```
norsetinge-mirror (private repo)
├── main branch (production)
│   └── Matches webhost 100%
└── .github/
    └── workflows/ (future: auto-deploy on push)
```

### Commit Messages

**Format:**
```
Update: YYYY-MM-DD_HH:MM:SS

- Added: artikel-slug-1
- Modified: artikel-slug-2
- Removed: artikel-slug-3
```

**Automatisk genereret** af deployer.go baseret på git diff.

### Git Ignore

**mirror/.gitignore:**
```
# Nothing ignored in mirror - track everything except .git itself
```

**project root .gitignore:**
```
# Ignore Hugo temporary build output
site/public/

# Ignore Go build artifacts
norsetinge
*.exe

# Ignore config secrets
.env

# Track mirror fully (git controlled)
!site/mirror/
```

---

## Rsync Strategi

### From Mirror to Webhost

**Command:**
```bash
rsync -avz --delete \
  --exclude '.git' \
  --exclude '.gitignore' \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

### Rsync Fordele

✅ **Delta transfer:** Kun ændringer uploades
✅ **Preserve permissions:** File modes bevares
✅ **Preserve timestamps:** Correct Last-Modified headers
✅ **Automatic cleanup:** --delete fjerner gamle filer
✅ **Compression:** -z mindsker båndbredde
✅ **Resume capability:** Kan genoptage afbrudte transfers

### SSH Key Authentication

**Setup:**
```bash
# 1. Generer SSH key på norsetinge container
ssh-keygen -t ed25519 -f ~/.ssh/norsetinge_deploy

# 2. Copy public key til webhost
ssh-copy-id -i ~/.ssh/norsetinge_deploy.pub deploy@norsetinge.com

# 3. Test connection
ssh -i ~/.ssh/norsetinge_deploy deploy@norsetinge.com

# 4. Configure i config.yaml
rsync:
  ssh_key: "/home/ubuntu/.ssh/norsetinge_deploy"
```

---

## Monitoring & Validation

### Post-Deploy Validation

**Automatic checks (future enhancement):**

```bash
# 1. Verify webhost reachable
curl -I https://norsetinge.com

# 2. Check sitemap exists
curl -s https://norsetinge.com/sitemap.xml | grep -q "urlset"

# 3. Verify latest article accessible
curl -I https://norsetinge.com/articles/latest-slug/

# 4. Compare file count
LOCAL_COUNT=$(find site/mirror -type f -not -path '*/\.git/*' | wc -l)
REMOTE_COUNT=$(ssh deploy@norsetinge.com "find /var/www/norsetinge.com -type f | wc -l")

if [ "$LOCAL_COUNT" -eq "$REMOTE_COUNT" ]; then
  echo "✅ File count matches: $LOCAL_COUNT files"
else
  echo "⚠️ File count mismatch: Local=$LOCAL_COUNT Remote=$REMOTE_COUNT"
fi
```

### Log Files

**Deployment logs:**
```bash
# Check deployer logs
journalctl -u norsetinge -f | grep "📋 Syncing"

# Example log output:
📋 Syncing public → mirror...
✅ Mirror synced successfully
📋 Committing to git...
✅ Git committed: Update: 2025-10-03_14:20:00
📋 Pushing to remote repo...
✅ Git pushed successfully
📋 Rsyncing to webhost...
✅ Deployment complete!
```

---

## Performance Metrics

### Phase 1: Single Language (~50 artikler)

| Step                  | Tid       | Bandwidth |
|-----------------------|-----------|-----------|
| Hugo Build            | ~3 sek    | N/A       |
| Sync public→mirror    | ~1 sek    | N/A       |
| Git commit            | ~0.5 sek  | N/A       |
| Git push              | ~2 sek    | ~500 KB   |
| Rsync to webhost      | ~5 sek    | ~1 MB     |
| **Total**             | **~12 sek** | **~1.5 MB** |

### Phase 2: Multilingual (~50 artikler × 22 sprog = 1100 artikler)

| Step                  | Tid       | Bandwidth |
|-----------------------|-----------|-----------|
| Hugo Build            | ~15 sek   | N/A       |
| Sync public→mirror    | ~5 sek    | N/A       |
| Git commit            | ~2 sek    | N/A       |
| Git push              | ~10 sek   | ~10 MB    |
| Rsync to webhost      | ~30 sek   | ~20 MB    |
| **Total**             | **~62 sek** | **~30 MB** |

**Note:** Dette er estimater. Faktiske tal vil variere baseret på:
- Artikel størrelse (tekst + billeder)
- Netværkshastighed
- Antal ændringer per build

---

## Current Implementation Status

### ✅ Phase 1: COMPLETE

- [x] Hugo builds til `site/public/`
- [x] Rsync sync `public/` → `mirror/` med --delete
- [x] Git repo initialiseret i `site/mirror/`
- [x] Git commit automation (disabled for testing)
- [x] Rsync til webhost (disabled for testing)
- [x] Periodic build+deploy (every 10 min)
- [x] Manual fast deploy option

### ⚠️ Testing Phase

**Config status:**
```yaml
git:
  auto_commit: false  # Disabled for testing

rsync:
  enabled: false      # Disabled for testing
```

**Når testing complete:**
```yaml
git:
  auto_commit: true   # Enable for production

rsync:
  enabled: true       # Enable for production
```

### 📋 Phase 2: FUTURE

- [ ] Multilingual Hugo configuration
- [ ] Language folder structure (22 sprog)
- [ ] Translation automation (OpenRouter API)
- [ ] Language switcher frontend
- [ ] Cookie-based language preference
- [ ] Separate sitemaps per language
- [ ] hreflang tags for SEO

---

## Troubleshooting

### Problem 1: Mirror ikke synced korrekt

**Symptom:** Ændringer i `public/` vises ikke i `mirror/`

**Debug:**
```bash
# Check rsync output
rsync -avz --delete --dry-run \
  --exclude '.git' \
  site/public/ site/mirror/
```

**Fix:** Check at sync() kører uden fejl i deployer.go

### Problem 2: Git push fejler

**Symptom:** "Permission denied (publickey)"

**Debug:**
```bash
# Test SSH connection
ssh -T git@github.com

# Check SSH key
ls -la ~/.ssh/
```

**Fix:** Add SSH key til GitHub account

### Problem 3: Rsync til webhost fejler

**Symptom:** "Permission denied" eller "Connection refused"

**Debug:**
```bash
# Test SSH connection
ssh -i ~/.ssh/norsetinge_deploy deploy@norsetinge.com

# Test manual rsync
rsync -avz --dry-run \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Fix:**
- Verify SSH key permissions (chmod 600)
- Check webhost firewall rules
- Verify target path exists og er writable

### Problem 4: Webhost og mirror out of sync

**Symptom:** File count mismatch mellem mirror/ og webhost

**Debug:**
```bash
# Local count
find site/mirror -type f -not -path '*/\.git/*' | wc -l

# Remote count
ssh deploy@norsetinge.com "find /var/www/norsetinge.com -type f | wc -l"
```

**Fix:** Kør fuld rsync med --delete igen

---

## Security Considerations

### 1. Git Repository

✅ **Private repo:** Mirror repo er private på GitHub
✅ **SSH keys:** Read/write adgang via SSH keys (ikke HTTPS)
✅ **No secrets:** Ingen passwords eller API keys i mirror

### 2. Rsync Deployment

✅ **SSH key authentication:** Passwordless deploy via dedicated key
✅ **Restricted user:** Deploy user har kun write access til /var/www/norsetinge.com
✅ **No shell access:** Deploy user kan ikke login interaktivt

### 3. Webhost Access

✅ **Read-only webserver:** Nginx/Apache kan kun læse filer, ikke skrive
✅ **No PHP/exec:** Static HTML only, ingen dynamisk kode execution
✅ **HTTPS only:** Tvungen SSL via Let's Encrypt

---

## Future Enhancements

### 1. Automated Testing (Phase 3)

```yaml
# .github/workflows/test-before-deploy.yml
name: Test Before Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Validate HTML
        run: htmlproofer --check-html --check-links
      - name: Check broken links
        run: linkchecker index.html
      - name: Lighthouse CI
        run: lhci autorun
```

### 2. CDN Integration (Phase 4)

- Cloudflare CDN foran webhost
- Cache invalidation efter deploy
- Global edge caching for performance

### 3. Image Optimization (Phase 4)

- Automatic image compression før build
- WebP conversion
- Responsive image srcsets
- Lazy loading

### 4. Analytics Integration (Phase 5)

- Privacy-friendly analytics (Plausible/Fathom)
- Cookie-free tracking
- Real-time visitor monitoring

---

## Related Documentation

- **Complete workflow:** `doc/projekt-status-plan-current.md`
- **Deploy pipeline:** `doc/deploy-flow-description.md`
- **Article format:** `doc/article-stencil.md`
- **Preview system:** `doc/preview-html.md`
- **Notifications:** `doc/ntfy-notification.md`

---

*Dette dokument beskriver mirror-systemets design og implementation.*
*Opdateres løbende som systemet udvikles.*
