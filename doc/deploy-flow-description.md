# Deploy Flow Description
**Version:** 1.0
**Dato:** 2025-10-03

---

## Formål

Dette dokument beskriver den komplette deployment flow fra godkendt artikel til live på webhost, inklusiv Hugo build, mirror sync, git version control, og rsync deployment.

---

## Oversigt

```
┌──────────────┐
│   Godkendt   │
│   Artikel    │
│  (udgivet/)  │
└──────┬───────┘
       │
       ↓
┌──────────────────────────────────────────────────────────┐
│              PERIODIC BUILD (Every 10 min)               │
│                  eller MANUAL (Deploy Nu)                │
└──────┬───────────────────────────────────────────────────┘
       │
       ↓
┌──────────────┐
│  Step 1:     │
│  Hugo Build  │
│  → public/   │
└──────┬───────┘
       │
       ↓
┌──────────────┐
│  Step 2:     │
│  Mirror Sync │
│  → mirror/   │
└──────┬───────┘
       │
       ↓
┌──────────────┐
│  Step 3:     │
│  Git Commit  │
│  + Push      │
└──────┬───────┘
       │
       ↓
┌──────────────┐
│  Step 4:     │
│  Rsync       │
│  → Webhost   │
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   LIVE ON    │
│  norsetinge  │
│    .com      │
└──────────────┘
```

---

## Step 1: Hugo Build (Full Site)

### Input
- Alle artikler i `/home/ubuntu/hugo-norsetinge/Dropbox/Publisering/NorseTinge/udgivet/`
- Hugo configuration: `/home/ubuntu/hugo-norsetinge/site/hugo.toml`
- Layouts/templates: `/home/ubuntu/hugo-norsetinge/site/layouts/`

### Process

**1.1. Clean Content Directory**
```go
contentDir := filepath.Join(cfg.Hugo.SiteDir, "content", "articles")
os.RemoveAll(contentDir)
os.MkdirAll(contentDir, 0755)
```

**Purpose:** Start fresh - remove old articles from previous builds

---

**1.2. Load Published Articles**
```go
publishedDir := filepath.Join(cfg.Dropbox.BasePath, "udgivet")
articles, err := loadPublishedArticles(publishedDir)
```

**What it does:**
- Scans `udgivet/` folder for `.md` files
- Parses each article (frontmatter + content)
- Validates required fields (title, author)
- Filters only articles with `published: 1` status

**Example:**
```
udgivet/
  ├── article-1.md  (published: 1) ✅ Included
  ├── article-2.md  (published: 1) ✅ Included
  └── old-draft.md  (published: 0) ❌ Skipped
```

---

**1.3. Copy Articles to Hugo Content**
```go
for _, article := range articles {
    targetPath := filepath.Join(contentDir, filename)
    copyFile(article.FilePath, targetPath)
}
```

**File structure created:**
```
site/content/articles/
  ├── article-1.md
  ├── article-2.md
  └── article-3.md
```

**Note:** Hugo vil bygge disse til multilang struktur automatisk (Phase 2)

---

**1.4. Run Hugo Build**
```bash
cd /home/ubuntu/hugo-norsetinge/site
hugo --gc --minify
```

**Command flags:**
- `--gc`: Garbage collect unused cache
- `--minify`: Minify HTML, CSS, JS (faster load times)

**Build process:**
1. Reads `hugo.toml` configuration
2. Processes all `.md` files in `content/`
3. Applies layouts from `layouts/`
4. Generates static HTML pages
5. Copies static assets
6. Creates RSS/sitemap
7. Outputs to `public/`

---

**1.5. Output Structure**
```
site/public/
├── index.html            # Homepage (English default)
├── sitemap.xml          # SEO sitemap
├── index.xml            # RSS feed
│
├── categories/          # Category archives
│   ├── teknologi/
│   └── ledelse/
│
├── tags/               # Tag archives
│   ├── devops/
│   └── kultur/
│
├── da/                 # Danish language (Phase 1: empty)
├── sv/                 # Swedish language (Phase 2)
├── no/                 # Norwegian language (Phase 2)
... (22 languages configured)
│
└── preview-[slug]/     # Preview builds (temporary)
```

**Current Phase 1 Status:**
- Only Danish/original language articles
- Multilang folders exist but are empty (Phase 2)
- No translation yet

---

### Output
- Complete static site in `/home/ubuntu/hugo-norsetinge/site/public/`
- All HTML, CSS, JS minified
- Sitemap and RSS generated
- Ready for deployment

### Logging
```
🔨 Building full site...
📚 Found 3 published articles
✅ Full site built successfully
   Public:  /home/ubuntu/hugo-norsetinge/site/public
   Mirror:  /home/ubuntu/hugo-norsetinge/site/mirror
```

---

## Step 2: Mirror Sync (public → mirror)

### Purpose
Create 1:1 copy of webhost locally for:
- Git version control
- Test/validate before deploy
- Rsync source (clean sync with --delete)

### Input
- Source: `/home/ubuntu/hugo-norsetinge/site/public/`
- Target: `/home/ubuntu/hugo-norsetinge/site/mirror/`

### Process

**2.1. Rsync Local Copy**
```go
func (d *Deployer) syncToMirror(publicDir, mirrorDir string) error {
    cmd := exec.Command("rsync",
        "-a",           // Archive mode (recursive, preserve permissions)
        "--delete",     // Delete files in mirror not in public
        "--exclude", ".git",  // Don't delete git metadata
        publicDir+"/",  // Source (trailing slash important!)
        mirrorDir+"/",  // Destination
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("rsync failed: %w\nOutput: %s", err, output)
    }

    return nil
}
```

**Rsync flags explained:**
- `-a` (archive): Preserves permissions, timestamps, symlinks
- `--delete`: Removes files in mirror that don't exist in public
- `--exclude .git`: Protect git repository in mirror

**Trailing slash importance:**
```bash
rsync public/ mirror/   # ✅ Correct - syncs CONTENTS of public into mirror
rsync public mirror/    # ❌ Wrong - creates mirror/public subfolder
```

---

**2.2. What Gets Synced**

**Added:**
```
New article added → Copied to mirror
New image added → Copied to mirror
Updated HTML → Overwritten in mirror
```

**Modified:**
```
Article edited → HTML regenerated → Mirror updated
CSS changed → Mirror updated
```

**Deleted:**
```
Article deleted from udgivet/ → Not in public/ → Removed from mirror/ (--delete)
Old preview/ folder → Not in public/ → Removed from mirror/
```

---

**2.3. Mirror Structure (After Sync)**
```
site/mirror/
├── .git/               # Git repository (preserved, not synced)
├── index.html         # Exact copy from public/
├── sitemap.xml
├── categories/
├── tags/
├── da/, sv/, no/, ... # Language folders
└── preview-[slug]/    # Only active previews
```

**1:1 Match:**
```bash
$ diff -r public/ mirror/ --exclude .git
# (no output = perfect match)
```

---

### Output
- Mirror directory updated to match public/ exactly
- Old files removed
- New files added
- Git metadata preserved

### Logging
```
📋 Syncing public → mirror...
✓ Synced to mirror: /home/ubuntu/hugo-norsetinge/site/mirror
```

---

## Step 3: Git Version Control (mirror → private repo)

### Purpose
- Version control of entire live website
- Rollback capability
- Backup of production site
- Audit trail of changes

### Input
- Mirror directory with changes: `/home/ubuntu/hugo-norsetinge/site/mirror/`

### Process

**3.1. Check Git Status**
```go
func (d *Deployer) gitCommitAndPush(mirrorDir string) error {
    // Check if git repo exists
    if !isGitRepo(mirrorDir) {
        return fmt.Errorf("mirror is not a git repository")
    }

    // Check if auto_commit enabled
    if !d.cfg.Git.AutoCommit {
        log.Printf("Git auto_commit disabled - skipping")
        return nil
    }

    // ... proceed with commit
}
```

---

**3.2. Stage All Changes**
```bash
cd /home/ubuntu/hugo-norsetinge/site/mirror
git add .
```

**What gets staged:**
- New files (articles, images, etc.)
- Modified files (updated HTML)
- Deleted files (removed articles)

---

**3.3. Generate Commit Message**
```go
timestamp := time.Now().Format("2006-01-02 15:04:05")
message := fmt.Sprintf("Deploy: %s\n\nAutomated deployment from Norsetinge", timestamp)
```

**Example commit message:**
```
Deploy: 2025-10-03 11:43:26

Automated deployment from Norsetinge
```

---

**3.4. Create Commit**
```bash
git commit -m "Deploy: 2025-10-03 11:43:26

Automated deployment from Norsetinge"
```

**Commit includes:**
- Added articles
- Modified content
- Deleted old content
- Updated sitemaps/RSS
- Changed assets

---

**3.5. Push to Remote**
```bash
git push origin main
```

**Remote repository:**
- Private GitHub/GitLab repo
- URL configured in `config.yaml`: `git.mirror_repo`
- SSH key authentication

**Example:**
```yaml
git:
  mirror_repo: "git@github.com:username/norsetinge-mirror.git"
  auto_commit: true
```

---

**3.6. Git History Example**

```bash
$ git log --oneline --graph
* a3f2e9a Deploy: 2025-10-03 11:43:26
* b4d1c8e Deploy: 2025-10-03 11:33:15
* c7e5f2a Deploy: 2025-10-03 11:23:02
```

**Each commit:**
- Full snapshot of website at that moment
- Can rollback with `git revert` or `git reset`
- Diff shows exactly what changed

**View changes:**
```bash
$ git diff HEAD~1 HEAD
# Shows what changed in last deployment
```

---

### Output
- Git commit created with timestamp
- Changes pushed to private remote repo
- Full version history maintained

### Logging
```
📦 Creating git commit...
✓ Git committed: Deploy: 2025-10-03 11:43:26
📤 Pushing to remote: git@github.com:username/norsetinge-mirror.git
✓ Git pushed to remote
```

---

## Step 4: Rsync to Webhost (mirror → production)

### Purpose
- Deploy to live webserver
- Sync only changes (efficient)
- Remove deleted content on webhost
- Match webhost to mirror exactly

### Input
- Source: `/home/ubuntu/hugo-norsetinge/site/mirror/`
- Target: `deploy@norsetinge.com:/var/www/norsetinge.com/`

### Process

**4.1. Check Rsync Enabled**
```go
if !d.cfg.Rsync.Enabled {
    log.Printf("Rsync disabled - skipping deployment")
    return nil
}
```

**Configuration:**
```yaml
rsync:
  enabled: true
  host: "norsetinge.com"
  user: "deploy"
  target_path: "/var/www/norsetinge.com"
  ssh_key: "/home/ubuntu/.ssh/norsetinge_deploy"  # Optional
```

---

**4.2. Build Rsync Command**
```go
target := fmt.Sprintf("%s@%s:%s",
    d.cfg.Rsync.User,
    d.cfg.Rsync.Host,
    d.cfg.Rsync.TargetPath,
)

args := []string{
    "-avz",          // Archive, verbose, compress
    "--delete",      // Delete files on remote not in source
    "--exclude", ".git",  // Don't upload git metadata
}

if d.cfg.Rsync.SSHKey != "" {
    args = append(args, "-e", fmt.Sprintf("ssh -i %s", d.cfg.Rsync.SSHKey))
}

args = append(args, mirrorDir+"/", target)

cmd := exec.Command("rsync", args...)
```

---

**4.3. Rsync Flags Explained**

**`-a` (archive mode):**
- `-r` recursive
- `-l` preserve symlinks
- `-p` preserve permissions
- `-t` preserve timestamps
- `-g` preserve group
- `-o` preserve owner

**`-v` (verbose):**
- Show files being transferred
- Useful for logging

**`-z` (compress):**
- Compress during transfer
- Faster over network
- Decompress on arrival

**`--delete`:**
- **Critical for hygiene!**
- Removes files on webhost not in mirror
- Ensures webhost matches mirror 100%

**Example scenario:**
```
# Article deleted from udgivet/
udgivet/old-article.md  ❌ Deleted

# Hugo build (article not included)
public/old-article/     ❌ Not generated

# Mirror sync (removed from mirror)
mirror/old-article/     ❌ Removed (rsync --delete)

# Rsync to webhost (removed from live site)
webhost/old-article/    ❌ Deleted (rsync --delete)
```

**Result:** Old article completely removed from live site

---

**4.4. Execute Rsync**
```bash
rsync -avz --delete \
  --exclude .git \
  -e "ssh -i /home/ubuntu/.ssh/norsetinge_deploy" \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Transfer process:**
1. SSH connection to webhost
2. Compare files (checksums)
3. Transfer only changed files
4. Delete files not in source
5. Update permissions/timestamps
6. Close connection

---

**4.5. Rsync Output Example**

```
sending incremental file list
./
index.html
sitemap.xml
categories/teknologi/index.html
da/index.html
preview-devops-paradigme/index.html
deleting old-article/index.html

sent 45,231 bytes  received 1,234 bytes  9,293.00 bytes/sec
total size is 2,145,678  speedup is 46.15
```

**Interpretation:**
- `sending incremental file list` - Only changes sent
- `deleting old-article/` - Old content removed
- `speedup is 46.15` - Highly efficient (only diffs transferred)

---

**4.6. SSH Key Authentication**

**Setup (one-time):**
```bash
# Generate deploy key
ssh-keygen -t ed25519 -f ~/.ssh/norsetinge_deploy -C "norsetinge-deploy"

# Copy to webhost
ssh-copy-id -i ~/.ssh/norsetinge_deploy.pub deploy@norsetinge.com

# Test connection
ssh -i ~/.ssh/norsetinge_deploy deploy@norsetinge.com "echo Connected"
```

**Security:**
- Deploy user has restricted permissions
- Only write access to `/var/www/norsetinge.com/`
- No sudo access
- Key-based auth (no password)

---

### Output
- Live website updated on `norsetinge.com`
- Only changed files transferred
- Old content removed
- Webhost matches mirror exactly

### Logging
```
🚀 Deploying to webhost: norsetinge.com
✓ Rsync completed successfully
   Target: deploy@norsetinge.com:/var/www/norsetinge.com
   Transferred: 45 KB (changed files only)
```

---

## Complete Flow Timing

### Periodic Build (Every 10 minutes)

**Timeline:**
```
00:00 - Trigger (cron ticker)
00:01 - Load 3 articles from udgivet/
00:02 - Hugo build (3 articles + multilang structure)
00:15 - Mirror sync (rsync local)
00:16 - Git commit + push
00:20 - Rsync to webhost
00:35 - ✅ Complete

Total: ~35 seconds
```

---

### Manual Deploy ("Godkend + Deploy Nu")

**Timeline:**
```
00:00 - User clicks "Deploy Nu"
00:01 - Article moved to udgivet/
00:02 - Trigger build+deploy
00:03 - Hugo build
00:18 - Mirror sync
00:19 - Git commit + push
00:23 - Rsync to webhost
00:38 - ✅ Complete
00:39 - Success page shown

Total: ~40 seconds
User sees: "Artikel er nu live!"
```

---

## Error Handling

### Hugo Build Fails

**Possible causes:**
- Invalid markdown syntax
- Missing required frontmatter
- Template errors

**Handling:**
```go
if err := h.buildSite(); err != nil {
    log.Printf("Error in periodic build: %v", err)
    return err  // Abort deployment
}
```

**Result:** No deployment occurs, error logged

---

### Mirror Sync Fails

**Possible causes:**
- Disk full
- Permission denied
- Rsync not installed

**Handling:**
```go
if err := d.syncToMirror(publicDir, mirrorDir); err != nil {
    log.Printf("Error syncing to mirror: %v", err)
    return err  // Abort deployment
}
```

**Result:** Git commit and rsync skipped

---

### Git Push Fails

**Possible causes:**
- No network connection
- SSH key issues
- Remote repo down

**Handling:**
```go
if err := d.gitCommitAndPush(mirrorDir); err != nil {
    log.Printf("Warning: Git push failed: %v", err)
    // Continue to rsync anyway
}
```

**Result:** Rsync still proceeds (git is backup, not critical path)

---

### Rsync Fails

**Possible causes:**
- Webhost down
- Network timeout
- Permission denied
- Disk full on webhost

**Handling:**
```go
if err := d.rsyncToWebhost(mirrorDir); err != nil {
    log.Printf("Error deploying to webhost: %v", err)
    return err  // Mark as failed
}
```

**Result:** Deployment marked as failed, retry on next periodic build

---

## Monitoring & Verification

### Success Indicators

**Logs:**
```
✅ Periodic build+deploy completed
   Duration: 35s
   Articles: 3
   Mirror: synced
   Git: committed + pushed
   Rsync: deployed
```

**Verification:**
```bash
# Check live site
curl https://norsetinge.com/sitemap.xml

# Check mirror matches webhost
ssh deploy@norsetinge.com "md5sum /var/www/norsetinge.com/index.html"
md5sum /home/ubuntu/hugo-norsetinge/site/mirror/index.html
# (checksums should match)
```

---

### Failure Indicators

**Logs:**
```
❌ Error in periodic build: ...
❌ Error syncing to mirror: ...
❌ Error deploying to webhost: ...
```

**Manual check:**
```bash
# Check Hugo build
cd /home/ubuntu/hugo-norsetinge/site
hugo --gc --minify

# Check mirror
ls -la /home/ubuntu/hugo-norsetinge/site/mirror/

# Test rsync connection
rsync -avz --dry-run \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

---

## Rollback Procedure

### Scenario: Bad deployment needs rollback

**Step 1: Identify good commit**
```bash
cd /home/ubuntu/hugo-norsetinge/site/mirror
git log --oneline
```

**Step 2: Revert to previous commit**
```bash
git revert HEAD
# Or
git reset --hard HEAD~1
```

**Step 3: Force push to remote**
```bash
git push --force origin main
```

**Step 4: Redeploy to webhost**
```bash
rsync -avz --delete \
  --exclude .git \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Result:** Live site reverted to previous state

---

## Configuration

### Minimal (Testing)

```yaml
hugo:
  site_dir: "/home/ubuntu/hugo-norsetinge/site"
  public_dir: "/home/ubuntu/hugo-norsetinge/site/public"
  mirror_dir: "/home/ubuntu/hugo-norsetinge/site/mirror"

git:
  mirror_repo: ""
  auto_commit: false  # Disabled

rsync:
  enabled: false      # Disabled
```

**Result:** Only Hugo build + mirror sync (local testing)

---

### Production

```yaml
hugo:
  site_dir: "/home/ubuntu/hugo-norsetinge/site"
  public_dir: "/home/ubuntu/hugo-norsetinge/site/public"
  mirror_dir: "/home/ubuntu/hugo-norsetinge/site/mirror"

git:
  mirror_repo: "git@github.com:username/norsetinge-mirror.git"
  auto_commit: true   # Enabled

rsync:
  enabled: true       # Enabled
  host: "norsetinge.com"
  user: "deploy"
  target_path: "/var/www/norsetinge.com"
  ssh_key: "/home/ubuntu/.ssh/norsetinge_deploy"
```

**Result:** Full deployment pipeline active

---

## Current Status (2025-10-03)

### Working
✅ Hugo builds to public/
✅ Mirror sync (public → mirror)
✅ Multilang structure created (22 languages)
✅ Periodic build every 10 minutes
✅ Manual "Deploy Nu" option

### Disabled (Testing)
⏸️ Git auto_commit: `false`
⏸️ Rsync enabled: `false`

### Not Yet Implemented
❌ Articles in multilang folders (Phase 2 - translation)
❌ Production rsync to real webhost
❌ Git remote repository configured

---

*Dette dokument beskriver den komplette deployment flow fra godkendt artikel til live produktion.*
