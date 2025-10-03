package deployer

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"norsetinge/config"
)

// Deployer handles deployment pipeline
type Deployer struct {
	cfg *config.Config
}

// NewDeployer creates a new deployer
func NewDeployer(cfg *config.Config) *Deployer {
	return &Deployer{cfg: cfg}
}

// Deploy runs the complete deployment pipeline
func (d *Deployer) Deploy(publicDir, mirrorDir string) error {
	log.Printf("🚀 Starting deployment pipeline...")

	// 1. Sync public to mirror
	if err := d.syncToMirror(publicDir, mirrorDir); err != nil {
		return fmt.Errorf("failed to sync to mirror: %w", err)
	}

	// 2. Git commit and push mirror
	if d.cfg.Git.AutoCommit {
		if err := d.gitCommitAndPush(mirrorDir); err != nil {
			return fmt.Errorf("failed to git commit/push: %w", err)
		}
	}

	// 3. Rsync to webhost
	if d.cfg.Rsync.Enabled {
		if err := d.rsyncToWebhost(mirrorDir); err != nil {
			return fmt.Errorf("failed to rsync to webhost: %w", err)
		}
	}

	log.Printf("✅ Deployment complete!")
	return nil
}

// syncToMirror copies public directory to mirror
func (d *Deployer) syncToMirror(publicDir, mirrorDir string) error {
	log.Printf("📋 Syncing public → mirror...")

	// Remove existing mirror content (but keep .git)
	entries, err := os.ReadDir(mirrorDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read mirror directory: %w", err)
	}

	for _, entry := range entries {
		// Keep .git directory
		if entry.Name() == ".git" {
			continue
		}

		path := filepath.Join(mirrorDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	// Copy all files from public to mirror
	if err := copyDir(publicDir, mirrorDir); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	log.Printf("✓ Synced to mirror: %s", mirrorDir)
	return nil
}

// gitCommitAndPush commits and pushes mirror to private repo
func (d *Deployer) gitCommitAndPush(mirrorDir string) error {
	log.Printf("📦 Committing and pushing to git...")

	// Initialize git repo if not exists
	gitDir := filepath.Join(mirrorDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := d.gitInit(mirrorDir); err != nil {
			return err
		}
	}

	// Git add all
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s", string(output))
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = mirrorDir
	if err := cmd.Run(); err == nil {
		log.Printf("ℹ️  No changes to commit")
		return nil
	}

	// Git commit with timestamp
	commitMsg := fmt.Sprintf("Deploy: %s", time.Now().Format("2006-01-02 15:04:05"))
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}

	// Git push
	cmd = exec.Command("git", "push")
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s", string(output))
	}

	log.Printf("✓ Pushed to git: %s", d.cfg.Git.MirrorRepo)
	return nil
}

// gitInit initializes git repo and sets remote
func (d *Deployer) gitInit(mirrorDir string) error {
	log.Printf("Initializing git repository...")

	// Git init
	cmd := exec.Command("git", "init")
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %s", string(output))
	}

	// Set remote
	cmd = exec.Command("git", "remote", "add", "origin", d.cfg.Git.MirrorRepo)
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add failed: %s", string(output))
	}

	// Set default branch to main
	cmd = exec.Command("git", "branch", "-M", "main")
	cmd.Dir = mirrorDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git branch failed: %s", string(output))
	}

	log.Printf("✓ Git repository initialized")
	return nil
}

// rsyncToWebhost syncs mirror to webhost using rsync
func (d *Deployer) rsyncToWebhost(mirrorDir string) error {
	log.Printf("🌐 Deploying to webhost via rsync...")

	// Build rsync command with --delete flag
	target := fmt.Sprintf("%s@%s:%s", d.cfg.Rsync.User, d.cfg.Rsync.Host, d.cfg.Rsync.TargetPath)

	args := []string{
		"-avz",
		"--delete", // Remove files on remote that don't exist in source
		"--exclude", ".git",
	}

	// Add SSH key if specified
	if d.cfg.Rsync.SSHKey != "" {
		args = append(args, "-e", fmt.Sprintf("ssh -i %s", d.cfg.Rsync.SSHKey))
	}

	// Add source and target
	args = append(args, mirrorDir+"/", target)

	cmd := exec.Command("rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %s", string(output))
	}

	log.Printf("✓ Deployed to: %s", target)
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy directory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}
