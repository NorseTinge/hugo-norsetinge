package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
dropbox:
  base_path: "Dropbox/Publisering/NorseTinge"
  folder_language: "da"

openrouter:
  api_key: "test-key"
  model: "anthropic/claude-3.5-sonnet"
  endpoint: "https://openrouter.ai/api/v1"

email:
  smtp_host: "smtp.example.com"
  smtp_port: 587
  smtp_user: "user@example.com"
  smtp_password: "password"
  from_address: "norsetinge@example.com"
  approval_recipient: "editor@example.com"

approval:
  host: "0.0.0.0"
  port: 8080
  tailscale_hostname: "norsetinge.tailnet.ts.net"

hugo:
  site_dir: "site"
  public_dir: "site/public"

images:
  min_width: 1200
  min_height: 630
  formats: ["webp", "jpeg", "png"]
  quality:
    webp: 85
    jpeg: 90
    png: 95

deploy:
  method: "rsync"
  rsync_target: "user@host.com:/var/www/"
  rsync_opts: "-avz --delete"

languages:
  - en
  - da
  - sv
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Load config
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Validate fields
	if cfg.Dropbox.BasePath != "Dropbox/Publisering/NorseTinge" {
		t.Errorf("Expected base_path 'Dropbox/Publisering/NorseTinge', got '%s'", cfg.Dropbox.BasePath)
	}

	if cfg.Dropbox.FolderLanguage != "da" {
		t.Errorf("Expected folder_language 'da', got '%s'", cfg.Dropbox.FolderLanguage)
	}

	if cfg.OpenRouter.APIKey != "test-key" {
		t.Errorf("Expected api_key 'test-key', got '%s'", cfg.OpenRouter.APIKey)
	}

	if cfg.Approval.Port != 8080 {
		t.Errorf("Expected approval port 8080, got %d", cfg.Approval.Port)
	}

	if len(cfg.Languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(cfg.Languages))
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Dropbox: DropboxConfig{
					BasePath:       "test/path",
					FolderLanguage: "en",
				},
				Hugo: HugoConfig{
					SiteDir: "site",
				},
			},
			wantErr: false,
		},
		{
			name: "missing base_path",
			config: Config{
				Dropbox: DropboxConfig{
					FolderLanguage: "en",
				},
				Hugo: HugoConfig{
					SiteDir: "site",
				},
			},
			wantErr: true,
		},
		{
			name: "missing folder_language",
			config: Config{
				Dropbox: DropboxConfig{
					BasePath: "test/path",
				},
				Hugo: HugoConfig{
					SiteDir: "site",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
