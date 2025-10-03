# Changelog

## 2025-10-03

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
