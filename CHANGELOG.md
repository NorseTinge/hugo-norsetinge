# Changelog

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
