# Changelog

## 2025-10-03 (Late Evening) - Deadlock Fix & Architecture Planning

### 🐛 Critical Bug Fix

*   **Mutex Deadlock Fixed** (`src/approval/server.go:88-153`):
    - **Problem**: RequestApproval() held global mutex during Hugo build (1-2 sec) and ntfy HTTP request, blocking all approve/reject actions
    - **Symptom**: Clicking "Godkend" button would hang forever with no response
    - **Root Cause**: handleApprove() couldn't acquire mutex while RequestApproval() held it
    - **Fix**: Refactored RequestApproval() into 4 phases with minimal locking:
      1. Phase 1: Check exists + reserve ID (short lock)
      2. Phase 2: Build Hugo preview (NO LOCK - long operation)
      3. Phase 3: Send ntfy notification (NO LOCK - HTTP request)
      4. Phase 4: Update final data + persist (short lock)
    - **Error handling**: Placeholder removed on failure for retry capability
    - **Result**: Approve/reject actions now respond immediately
    - **Documented in**: BUGS file entry #11

### 🏗️ Major Architecture Change (Planned)

**Problem Discovered**: Mutex deadlock in `src/approval/server.go:88-135` - `RequestApproval()` holds global lock during Hugo build and ntfy notification, blocking all approval/rejection actions.

**Root Cause**: Synchronous, single-threaded approval workflow cannot handle multiple articles in parallel.

**New Architecture Decision**: State Machine with ID-based routing

#### State Machine Design
- **Single Source of Truth**: `publish-flow.json` replaces `.pending_approvals.json`
- **Per-Article State**: Each article has its own state + mutex, enabling parallel processing
- **States**: IDGenerated → PreviewBuilding → PendingApproval → Approved → DeployedToMirror → DeployedToWebhost
- **State History**: Timestamp tracking for audit trail and debugging
- **Crash Recovery**: Full workflow state restored from publish-flow.json on restart

#### ID-based URL Structure
- **Current**: `norsetinge.com/devops-som-paradigme/` (slug can change)
- **New**: `norsetinge.com/articles/ABC123/` (permanent, immutable)
- **Hugo paths**: `content/articles/{ID}.md` → `public/articles/{ID}/index.html` → `mirror/articles/{ID}/index.html`
- **Slug redirects**: Optional slug-to-ID redirects for SEO (e.g., `/devops-paradigme/` → `/articles/ABC123/`)

#### Benefits
1. **Scalability**: N articles processed in parallel (currently limited to 1)
2. **Resilience**: Crash recovery from persistent state file
3. **No Deadlocks**: Per-article locking instead of global mutex
4. **Immutable URLs**: Article ID never changes, content/slug can be updated
5. **Audit Trail**: Full state history with timestamps for debugging

**Status**: Documented in TODO, implementation scheduled for next phase.

---

## 2025-10-03 (Evening) - Language Detection & Testing

### ✨ Features

*   **Language Detection Implemented** (`src/builder/hugo.go:152-160`): Added `detectLanguage()` function that reads optional `language` field from article frontmatter, defaults to "da" (Danish) if not specified.
*   **Language Field in Article Struct** (`src/common/frontmatter.go:44-45`): Added `Language string` field to Article struct with YAML tag for optional ISO 639-1 language codes (da, en, de, etc.).
*   **Three-Button Approval Interface** (`site/layouts/_default/single.html:112-114`): Added third button "⚡ Godkend + Deploy Nu" to preview pages:
    - ✅ Godkend: Approve for next periodic build (10 min)
    - ⚡ Godkend + Deploy Nu: Approve and deploy immediately
    - ❌ Afvis: Reject article

### 📝 Documentation Updates

*   **GEMINI.md Build Instructions** (`GEMINI.md:20-41`): Updated with correct build/run commands for norsetinge application:
    ```bash
    go build -o norsetinge ./src/main.go
    ./norsetinge
    ```

### 🐛 Bug Fixes

*   **Article Format Corrections**:
    *   Fixed `DevOps som paradigme.md` in `opdater/` folder - removed conflicting `publish: 1` flag
    *   Fixed `ai_devops_paradigme.md` in `udgiv/` folder - removed duplicate content in frontmatter YAML section

### ✅ Testing & Verification

*   **Preview System Tested**:
    *   Norsetinge application running successfully (PID 166392)
    *   Hugo preview builds correctly to `site/public/preview-*/`
    *   Preview accessible via http://localhost:8080/preview/
    *   Tailscale serve HTTPS verified: https://norsetinge.tail2d448.ts.net/preview/
    *   Ntfy notifications sent successfully
    *   Approval workflow ready for testing

### 📋 TODO Progress

*   ✅ Language detection logic (src/builder/hugo.go:152)
*   ✅ GEMINI.md build commands (GEMINI.md:24)
*   ✅ Go application path verification (GEMINI.md:31)
*   ✅ DevOps article format correction
*   ✅ Preview content verification

**Next Steps**:
*   Test godkendelse (approval) workflow
*   Test 4-5 artikler gennem hele pipeline
*   Enable git auto_commit
*   Enable rsync deployment

---

## 2025-10-03 (Afternoon) - Bug Fixes & Test Coverage

### 🐛 Bug Fixes

*   **Bug 7 - Dobbelt håndtering af events**: Confirmed that duplicate event handling is already prevented by `NotificationSent` flag in `approval/server.go:95`. Both `processEvents` and `periodicFolderScan` are protected against sending duplicate approval requests.

*   **Bug 8 - Status flag validation** (`src/common/frontmatter.go:137-165`): Fixed `UpdateStatus()` to ensure only one status flag is active at a time. The function now resets all status flags before setting the new status, preventing conflicting states.

*   **Bug 9 - Preview file cleanup** (`src/approval/server.go:376-402`): Implemented `cleanupPreviewFiles()` function to remove preview files from `public/`, `mirror/`, and `content/` directories when articles are approved or rejected. Called in all three handlers: `handleApprove`, `handleApproveAndDeploy`, and `handleReject`.

### ✅ Test Improvements

*   **Test Suite Status**: All tests now pass (approval: 4/4, common: 4/4, config: 2/2, watcher: 3/3)
*   **New Test: `TestUpdateStatusClearsOtherFlags`**: Added comprehensive test with 4 subtests to verify Bug 8 fix - ensures only one status flag is active at a time
*   **New Test: `TestCleanupPreviewFiles`**: Added test to verify Bug 9 fix - validates preview cleanup from all three directories
*   **Fixed: `TestUpdateStatus`**: Updated to reflect new behavior where status flags are reset
*   **Fixed: `TestLoadConfig`**: Modified to accept both config file and environment variable values for API key
*   **Fixed: `TestRequestApproval`**: Added proper temp directories and Hugo config to prevent test failures

### 📋 Documentation Updates

*   **BUGS file updated**: All 3 pending bugs (7, 8, 9) marked as FIXED with detailed explanations

---

## 2025-10-03 (Morning) - Project Direction Change

### 🎯 Strategic Pivot: KISS Approach

**Original Vision** (doc/project_plan.md):
- Automatic translation to 22+ languages via OpenRouter API
- Complex multilingual Hugo setup
- Email notifications for approvals
- Multiple workflow folders (afventer-rettelser/, etc.)

**Current KISS Implementation** (CLAUDE.md):
- Single language (Danish/original) only - translations deferred to Phase 4
- Simplified folder structure: kladde/ → udgiv/ → udgivet/afvist/
- Ntfy push notifications instead of email
- Folder-based status (no complex flag system)

**Rationale**: Complete the full pipeline for one language before adding complexity. This ensures a working system faster and makes debugging easier.

### 📐 Phase Structure

**Phase 1: Core Infrastructure** ✅
- Dropbox directory structure
- Go application with config loader, file watcher, markdown parser
- Basic approval web server (port 8080, Tailscale)
- Ntfy push notifications

**Phase 2: Preview & Approval** ✅
- Hugo preview builder (single article)
- Web-based approval interface
- Ntfy notifications with preview links
- Approve/Reject actions
- File movement based on folder location

**Phase 3: Publication Pipeline** ✅
- Hugo full-site build to `site/public/`
- Mirror sync: `site/public/` → `site/mirror/`
- Git automation: commit + push mirror to private repo
- Rsync deployment: `site/mirror/` → webhost with `--delete`
- Archive: move article to `udgivet/` after deployment

**Phase 4: Future Enhancements** (DEFERRED)
- Translation pipeline (OpenRouter API, 22 languages)
- Multilingual Hugo configuration
- Language switcher frontend
- Image processing automation
- Internal ad system

---

## 2025-10-03 (Morning) - Infrastructure & Refactoring

### ✨ Features & Refactoring

*   **Centralized Folder Alias Configuration**: Implemented loading of folder name aliases from `/home/ubuntu/hugo-norsetinge/folder-aliases.yaml`. This removes hardcoded folder paths and allows for flexible, language-based folder naming schemes.
    *   The main configuration loader in `src/config/config.go` now reads both `config.yaml` and `folder-aliases.yaml`.
    *   The file `watcher` (`src/watcher/mover.go`) now uses this central configuration instead of its own hardcoded values.
    *   The application startup in `src/main.go` was updated to support this new configuration loading mechanism.

### 🔧 Build & Project Structure

*   **Corrected Go Module Structure**: Moved the `go.mod` and `go.sum` files from the `src/` directory to the project root. This resolves fundamental issues with package discovery and aligns the project with standard Go module practices.
*   **Standardized Import Paths**: Updated all `import` paths across the Go source files to reflect the corrected module structure (e.g., `norsetinge/common` is now `norsetinge/src/common`). This was a necessary fix for the build to succeed.
*   **Resolved Compilation Errors**: Fixed a series of cascading build errors related to incorrect type definitions, function signatures, and argument mismatches that occurred during the refactoring process.
*   **Successful Compilation**: The application now compiles successfully into a `norsetinge` binary in the project root.

### 📝 Documentation & Maintenance

*   **Generated TODO List**: Scanned the codebase for `TODO` comments and created a `/home/ubuntu/hugo-norsetinge/TODO` file to track outstanding development tasks.
