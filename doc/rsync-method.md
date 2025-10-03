# Rsync Method & Best Practices

**Opdateret:** 2025-10-03
**Status:** Production Ready

---

## Oversigt

Rsync er det primære værktøj til deployment i Norsetinge. Det bruges til:

1. **Local sync:** `site/public/` → `site/mirror/` (1:1 kopi)
2. **Remote deploy:** `site/mirror/` → webhost (production deployment)

---

## Rsync Basics

### Hvad er Rsync?

Rsync er et værktøj til effektiv fil-synkronisering:

✅ **Delta transfer:** Kun ændringer overføres (ikke hele filer)
✅ **Preserve attributes:** Permissions, timestamps, ownership bevares
✅ **Compression:** Data komprimeres under transfer
✅ **Resume capability:** Kan genoptage afbrudte transfers
✅ **Delete handling:** Kan fjerne filer på destination der ikke findes i source
✅ **Network efficient:** Minimerer båndbredde-forbrug

### Grundlæggende Syntax

```bash
rsync [OPTIONS] SOURCE DESTINATION
```

**Vigtigt:** Trailing slash `/` betyder noget!

```bash
# MED trailing slash - kopier INDHOLDET af source/ til dest/
rsync -a source/ dest/

# UDEN trailing slash - kopier source/ FOLDER til dest/
rsync -a source dest/
```

---

## Norsetinge Rsync Commands

### 1. Local Sync: public → mirror

**Formål:** Sync Hugo build output til git-tracked mirror.

**Command:**
```bash
rsync -a --delete \
  --exclude '.git' \
  /home/ubuntu/hugo-norsetinge/site/public/ \
  /home/ubuntu/hugo-norsetinge/site/mirror/
```

**Flags forklaret:**

| Flag | Betydning |
|------|-----------|
| `-a` | Archive mode = -rlptgoD (recursive, links, perms, times, group, owner, devices) |
| `--delete` | Slet filer i mirror/ der ikke findes i public/ |
| `--exclude '.git'` | Bevar .git/ folder i mirror/ (skip ikke) |

**Hvad sker der:**

1. Rsync sammenligner `public/` med `mirror/`
2. Nye filer kopieres
3. Ændrede filer opdateres
4. Slettede filer fjernes fra mirror/ (pga. --delete)
5. `.git/` folder bevares (pga. --exclude)

**Output eksempel:**
```
sending incremental file list
articles/ny-artikel/index.html
articles/opdateret-artikel/index.html
deleting articles/gammel-artikel/

sent 52,847 bytes  received 156 bytes  106,006.00 bytes/sec
total size is 4,582,391  speedup is 86.36
```

**Tid:** ~1-2 sekunder (afhænger af antal ændringer)

---

### 2. Remote Deploy: mirror → webhost

**Formål:** Deploy til production webhost over SSH.

**Command:**
```bash
rsync -avz --delete \
  --exclude '.git' \
  --exclude '.gitignore' \
  -e "ssh -i /home/ubuntu/.ssh/norsetinge_deploy" \
  /home/ubuntu/hugo-norsetinge/site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Flags forklaret:**

| Flag | Betydning |
|------|-----------|
| `-a` | Archive mode (preserve everything) |
| `-v` | Verbose (vis hvad der sker) |
| `-z` | Compress data under transfer |
| `--delete` | Slet filer på webhost der ikke findes i mirror/ |
| `--exclude '.git'` | Upload ikke .git/ folder til webhost |
| `--exclude '.gitignore'` | Upload ikke .gitignore til webhost |
| `-e "ssh -i ..."` | Specify SSH key for authentication |

**Hvad sker der:**

1. Rsync connecter til webhost via SSH
2. Sammenligner mirror/ med /var/www/norsetinge.com/
3. Komprimerer og uploader kun ændrede filer
4. Sletter filer på webhost der ikke findes lokalt
5. Bevarer file permissions og timestamps

**Output eksempel:**
```
sending incremental file list
articles/ny-artikel/
articles/ny-artikel/index.html
articles/opdateret-artikel/index.html
deleting articles/gammel-artikel/index.html
deleting articles/gammel-artikel/

sent 15,428 bytes  received 89 bytes  3,103.40 bytes/sec
total size is 4,582,391  speedup is 295.24
```

**Tid:** ~5-15 sekunder (afhænger af antal ændringer + netværk)

---

## Flag Reference

### Mest brugte flags i Norsetinge

| Flag | Betydning | Hvorfor vi bruger det |
|------|-----------|----------------------|
| `-a` | Archive mode | Bevar alle file attributes (permissions, timestamps, etc.) |
| `-v` | Verbose | Se hvad der synces (debugging + logging) |
| `-z` | Compress | Reducer båndbredde ved remote sync |
| `--delete` | Delete extraneous | Hold destination clean (fjern gamle artikler) |
| `--exclude` | Exclude pattern | Skip .git/, .gitignore, etc. |
| `-e` | Remote shell | Specify SSH key for authentication |
| `--dry-run` | Test run | Preview hvad der ville ske (UDEN at ændre noget) |

### Archive Mode (-a) breakdown

`-a` er egentlig en kombination af:

| Flag | Betydning |
|------|-----------|
| `-r` | Recursive (inkluder subfolders) |
| `-l` | Copy symlinks as symlinks |
| `-p` | Preserve permissions (chmod) |
| `-t` | Preserve modification times |
| `-g` | Preserve group |
| `-o` | Preserve owner |
| `-D` | Preserve device/special files |

**Hvorfor `-a` er perfekt til static sites:**
- Bevarer file permissions → webserver kan læse filer korrekt
- Bevarer timestamps → correct Last-Modified HTTP headers
- Recursive → alle subfolders inkluderes automatisk

---

## Advanced Rsync Patterns

### 1. Dry Run (Test Sync)

**Test sync UDEN at ændre noget:**

```bash
rsync -avz --delete --dry-run \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Output viser:**
- Hvilke filer ville blive kopieret
- Hvilke filer ville blive slettet
- Total størrelse af transfer

**Brug case:** Test før deploy til produktion

---

### 2. Bandwidth Limit

**Begræns båndbredde til 1000 KB/s:**

```bash
rsync -avz --delete \
  --bwlimit=1000 \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Brug case:** Deploy i arbejdstiden uden at fylde netværket

---

### 3. Progress Display

**Vis progress for hver fil:**

```bash
rsync -avz --delete --progress \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Output:**
```
articles/lang-artikel/index.html
    12,847 100%  2.45MB/s    0:00:00 (xfr#1, to-chk=425/1247)
articles/ny-artikel/index.html
     5,234 100%  1.23MB/s    0:00:00 (xfr#2, to-chk=424/1247)
```

**Brug case:** Manual deploy hvor du vil se fremgang

---

### 4. Exclude Multiple Patterns

**Skip flere file patterns:**

```bash
rsync -avz --delete \
  --exclude '.git' \
  --exclude '.gitignore' \
  --exclude '*.tmp' \
  --exclude '*.log' \
  --exclude '.DS_Store' \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Eller via exclude-file:**

```bash
# Create exclude list
cat > rsync-excludes.txt <<EOF
.git
.gitignore
*.tmp
*.log
.DS_Store
EOF

# Use exclude file
rsync -avz --delete \
  --exclude-from=rsync-excludes.txt \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

---

### 5. Partial Transfer Resume

**Genoptag afbrudte transfers:**

```bash
rsync -avz --delete \
  --partial \
  --partial-dir=.rsync-partial \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Hvad sker der:**
- Hvis transfer afbrydes, gemmes del-overførte filer i `.rsync-partial/`
- Næste rsync kører genoptager fra hvor den stoppede

**Brug case:** Ustabilt netværk, store files

---

### 6. Stats & Summary

**Vis detaljeret statistik:**

```bash
rsync -avz --delete --stats \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Output:**
```
Number of files: 1,247 (reg: 1,235, dir: 12)
Number of created files: 5
Number of deleted files: 2
Number of regular files transferred: 8
Total file size: 4,582,391 bytes
Total transferred file size: 52,847 bytes
Literal data: 52,847 bytes
Matched data: 0 bytes
File list size: 24,567
File list generation time: 0.001 seconds
File list transfer time: 0.000 seconds
Total bytes sent: 78,392
Total bytes received: 256

sent 78,392 bytes  received 256 bytes  15,729.60 bytes/sec
total size is 4,582,391  speedup is 58.23
```

---

## SSH Key Setup

### 1. Generate SSH Key

```bash
# Generate dedicated deploy key
ssh-keygen -t ed25519 -f ~/.ssh/norsetinge_deploy -C "norsetinge-deploy"

# Output:
# ~/.ssh/norsetinge_deploy (private key)
# ~/.ssh/norsetinge_deploy.pub (public key)
```

### 2. Copy Public Key to Webhost

```bash
# Method 1: Using ssh-copy-id
ssh-copy-id -i ~/.ssh/norsetinge_deploy.pub deploy@norsetinge.com

# Method 2: Manual
cat ~/.ssh/norsetinge_deploy.pub | \
  ssh deploy@norsetinge.com "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

### 3. Test SSH Connection

```bash
# Test connection with key
ssh -i ~/.ssh/norsetinge_deploy deploy@norsetinge.com

# Should login without password
```

### 4. Configure SSH Config (Optional)

**Simplify commands med SSH config:**

```bash
# Edit ~/.ssh/config
cat >> ~/.ssh/config <<EOF

Host norsetinge-deploy
  HostName norsetinge.com
  User deploy
  IdentityFile ~/.ssh/norsetinge_deploy
  IdentitiesOnly yes
EOF

# Now you can use:
rsync -avz --delete site/mirror/ norsetinge-deploy:/var/www/norsetinge.com/
```

### 5. Set Correct Permissions

```bash
# Private key must be readable only by owner
chmod 600 ~/.ssh/norsetinge_deploy

# Public key can be world-readable
chmod 644 ~/.ssh/norsetinge_deploy.pub

# SSH config should be private
chmod 600 ~/.ssh/config
```

---

## Rsync in Go Code

### Norsetinge Implementation

**File:** `src/deployer/deployer.go`

#### 1. Local Sync: syncToMirror()

```go
func (d *Deployer) syncToMirror(publicDir, mirrorDir string) error {
    log.Printf("📋 Syncing public → mirror...")

    // rsync -a --delete --exclude '.git' publicDir/ mirrorDir/
    args := []string{
        "-a",              // Archive mode
        "--delete",        // Remove files not in source
        "--exclude", ".git", // Preserve git repo
        publicDir + "/",   // Source (trailing slash important!)
        mirrorDir + "/",   // Destination
    }

    cmd := exec.Command("rsync", args...)
    output, err := cmd.CombinedOutput()

    if err != nil {
        return fmt.Errorf("rsync failed: %w\nOutput: %s", err, output)
    }

    log.Printf("✅ Mirror synced successfully")
    return nil
}
```

#### 2. Remote Deploy: rsyncToWebhost()

```go
func (d *Deployer) rsyncToWebhost(mirrorDir string) error {
    if !d.cfg.Rsync.Enabled {
        log.Printf("⏭️ Rsync disabled in config")
        return nil
    }

    log.Printf("📋 Rsyncing to webhost...")

    // Build target: user@host:/path/
    target := fmt.Sprintf("%s@%s:%s",
        d.cfg.Rsync.User,
        d.cfg.Rsync.Host,
        d.cfg.Rsync.TargetPath,
    )

    // Build rsync args
    args := []string{
        "-avz",            // Archive + verbose + compress
        "--delete",        // Remove files not in source
        "--exclude", ".git",
        "--exclude", ".gitignore",
    }

    // Add SSH key if specified
    if d.cfg.Rsync.SSHKey != "" {
        sshCmd := fmt.Sprintf("ssh -i %s", d.cfg.Rsync.SSHKey)
        args = append(args, "-e", sshCmd)
    }

    // Add source and target
    args = append(args, mirrorDir+"/", target)

    // Execute rsync
    cmd := exec.Command("rsync", args...)
    output, err := cmd.CombinedOutput()

    if err != nil {
        return fmt.Errorf("rsync to webhost failed: %w\nOutput: %s", err, output)
    }

    log.Printf("✅ Deployment complete!")
    log.Printf("📊 Rsync output:\n%s", output)

    return nil
}
```

---

## Troubleshooting

### Problem 1: Permission Denied

**Symptom:**
```
rsync: send_files failed to open "file.html": Permission denied (13)
```

**Cause:** Rsync kan ikke læse source file eller skrive til destination

**Fix:**
```bash
# Check source permissions
ls -la site/mirror/

# Check destination permissions (remote)
ssh deploy@norsetinge.com "ls -la /var/www/norsetinge.com/"

# Fix permissions if needed
chmod 755 site/mirror/
chmod 644 site/mirror/**/*.html
```

---

### Problem 2: Connection Refused

**Symptom:**
```
ssh: connect to host norsetinge.com port 22: Connection refused
```

**Cause:** SSH ikke accessible på webhost

**Fix:**
```bash
# Check SSH running on webhost
ssh deploy@norsetinge.com "systemctl status ssh"

# Check firewall allows SSH
ssh deploy@norsetinge.com "sudo ufw status"

# Test connection
telnet norsetinge.com 22
```

---

### Problem 3: Files Not Deleted

**Symptom:** Gamle filer forbliver på destination trods --delete flag

**Cause:** Trailing slash fejl eller rsync version mismatch

**Fix:**
```bash
# Verify trailing slash PRESENT on source:
rsync -avz --delete site/mirror/ user@host:/path/

# NOT this (missing trailing slash):
rsync -avz --delete site/mirror user@host:/path/

# Check rsync version
rsync --version
# Should be 3.1.0 or newer for best --delete behavior
```

---

### Problem 4: Slow Transfer

**Symptom:** Rsync tager meget lang tid

**Debug:**
```bash
# Run with --progress to see what's slow
rsync -avz --delete --progress \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/

# Check bandwidth
rsync -avz --delete --stats \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

**Possible causes:**
- Mange små filer (overhead per fil)
- Slow network connection
- Webhost I/O performance

**Optimizations:**
```bash
# Skip compression for already-compressed files
rsync -av --delete \
  --skip-compress=jpg,jpeg,png,gif,webp,mp4,mp3 \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/

# Use whole-file transfer (skip delta algorithm for faster local networks)
rsync -avz --delete --whole-file \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

---

### Problem 5: SSH Key Not Working

**Symptom:**
```
Permission denied (publickey)
```

**Debug:**
```bash
# Test SSH connection with verbose
ssh -v -i ~/.ssh/norsetinge_deploy deploy@norsetinge.com

# Check key permissions
ls -la ~/.ssh/norsetinge_deploy
# Should be: -rw------- (600)

# Check key is correct format
file ~/.ssh/norsetinge_deploy
# Should say: OpenSSH private key

# Check public key on remote
ssh deploy@norsetinge.com "cat ~/.ssh/authorized_keys"
```

**Fix:**
```bash
# Fix permissions
chmod 600 ~/.ssh/norsetinge_deploy
chmod 644 ~/.ssh/norsetinge_deploy.pub

# Re-copy public key
ssh-copy-id -i ~/.ssh/norsetinge_deploy.pub deploy@norsetinge.com
```

---

## Best Practices

### 1. Always Use --dry-run First

Before deploying to production:

```bash
# Test sync locally
rsync -a --delete --dry-run site/public/ site/mirror/

# Test deploy to webhost
rsync -avz --delete --dry-run \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

### 2. Use Trailing Slash Consistently

**Always include trailing slash on source:**

```bash
# CORRECT
rsync -a source/ dest/

# WRONG (creates dest/source/ instead of syncing contents)
rsync -a source dest/
```

### 3. Exclude Git Folders

**Never upload .git/ to webhost:**

```bash
rsync -avz --delete \
  --exclude '.git' \
  --exclude '.gitignore' \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

### 4. Log Rsync Output

**Capture rsync output for debugging:**

```go
cmd := exec.Command("rsync", args...)
output, err := cmd.CombinedOutput()

log.Printf("📊 Rsync output:\n%s", output)
```

### 5. Use Dedicated SSH Key

**Don't use personal SSH key for automated deploys:**

```bash
# Generate dedicated key
ssh-keygen -t ed25519 -f ~/.ssh/norsetinge_deploy

# Use in rsync
rsync -avz -e "ssh -i ~/.ssh/norsetinge_deploy" \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

### 6. Set Bandwidth Limits (If Needed)

**Avoid saturating network:**

```bash
# Limit to 5000 KB/s
rsync -avz --delete --bwlimit=5000 \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/
```

### 7. Monitor Deployment Size

**Check what's being transferred:**

```bash
rsync -avz --delete --stats \
  site/mirror/ \
  deploy@norsetinge.com:/var/www/norsetinge.com/

# Look for:
# - Total bytes sent
# - Number of files transferred
# - Speedup ratio
```

---

## Performance Optimization

### Local Sync (public → mirror)

**Current:**
```bash
rsync -a --delete --exclude '.git' public/ mirror/
```

**Optimized for speed:**
```bash
# Use --whole-file for local sync (skip delta algorithm)
rsync -a --delete --whole-file --exclude '.git' public/ mirror/
```

**Why:** Delta algorithm overhead not worth it for local sync.

---

### Remote Deploy (mirror → webhost)

**Current:**
```bash
rsync -avz --delete -e "ssh -i key" mirror/ user@host:/path/
```

**Optimized for bandwidth:**
```bash
# Skip compression for already-compressed files
rsync -av --delete \
  --compress-level=6 \
  --skip-compress=jpg,jpeg,png,gif,webp,mp4,mp3,woff,woff2 \
  -e "ssh -i key -C" \
  mirror/ user@host:/path/
```

**Why:** Images/fonts already compressed - skip CPU overhead.

---

### Parallel Rsync (Future Enhancement)

For meget store sites kan rsync paralleliseres:

```bash
# Split sync into chunks (requires GNU parallel)
find site/mirror -type d -maxdepth 1 | \
  parallel rsync -avz --delete \
    {}/ \
    deploy@norsetinge.com:/var/www/norsetinge.com/{/}/
```

**Note:** Ikke nødvendigt for Phase 1 (~50 artikler)

---

## Config.yaml Integration

### Current Configuration

```yaml
rsync:
  enabled: false           # Toggle remote deploy
  host: "norsetinge.com"   # Webhost hostname
  user: "deploy"           # SSH user
  target_path: "/var/www/norsetinge.com"  # Webhost path
  ssh_key: "/home/ubuntu/.ssh/norsetinge_deploy"  # SSH key path
```

### Production Configuration

```yaml
rsync:
  enabled: true            # Enable for production
  host: "norsetinge.com"
  user: "deploy"
  target_path: "/var/www/norsetinge.com"
  ssh_key: "/home/ubuntu/.ssh/norsetinge_deploy"
  bandwidth_limit: 0       # 0 = unlimited (KB/s)
  timeout: 300             # 5 min timeout
```

---

## Related Documentation

- **Mirror system:** `doc/mirror-site-design.md`
- **Deploy flow:** `doc/deploy-flow-description.md`
- **Project status:** `doc/projekt-status-plan-current.md`

---

*Dette dokument beskriver rsync usage i Norsetinge projektet.*
*Opdateres løbende som systemet udvikles.*
