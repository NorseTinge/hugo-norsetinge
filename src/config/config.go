package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Dropbox    DropboxConfig    `yaml:"dropbox"`
	OpenRouter OpenRouterConfig `yaml:"openrouter"`
	Email      EmailConfig      `yaml:"email"`
	Approval   ApprovalConfig   `yaml:"approval"`
	Ntfy       NtfyConfig       `yaml:"ntfy"`
	Hugo       HugoConfig       `yaml:"hugo"`
	Git        GitConfig        `yaml:"git"`
	Rsync      RsyncConfig      `yaml:"rsync"`
	Images     ImagesConfig     `yaml:"images"`
	Deploy     DeployConfig     `yaml:"deploy"`
	Languages  []string         `yaml:"languages"`
}

type DropboxConfig struct {
	BasePath       string `yaml:"base_path"`
	FolderLanguage string `yaml:"folder_language"`
}

type OpenRouterConfig struct {
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	Endpoint string `yaml:"endpoint"`
}

type EmailConfig struct {
	SMTPHost          string `yaml:"smtp_host"`
	SMTPPort          int    `yaml:"smtp_port"`
	SMTPUser          string `yaml:"smtp_user"`
	SMTPPassword      string `yaml:"smtp_password"`
	FromAddress       string `yaml:"from_address"`
	ApprovalRecipient string `yaml:"approval_recipient"`

	// IMAP for reading replies
	IMAPHost string `yaml:"imap_host"`
	IMAPPort int    `yaml:"imap_port"`
	IMAPUser string `yaml:"imap_user"`
	IMAPPassword string `yaml:"imap_password"`
}

type ApprovalConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	TailscaleHostname string `yaml:"tailscale_hostname"`
}

type NtfyConfig struct {
	Enabled bool   `yaml:"enabled"`
	Server  string `yaml:"server"`
	Topic   string `yaml:"topic"`
}

type HugoConfig struct {
	SiteDir   string `yaml:"site_dir"`
	PublicDir string `yaml:"public_dir"`
	MirrorDir string `yaml:"mirror_dir"`
}

type ImagesConfig struct {
	MinWidth int               `yaml:"min_width"`
	MinHeight int              `yaml:"min_height"`
	Formats  []string          `yaml:"formats"`
	Sizes    map[string][2]int `yaml:"sizes"`
	Quality  map[string]int    `yaml:"quality"`
	Icons    IconsConfig       `yaml:"icons"`
}

type IconsConfig struct {
	FaviconSizes         []int `yaml:"favicon_sizes"`
	AppleTouchIconSizes  []int `yaml:"apple_touch_icon_sizes"`
	AndroidIconSizes     []int `yaml:"android_icon_sizes"`
}

type DeployConfig struct {
	Method      string `yaml:"method"`
	RsyncTarget string `yaml:"rsync_target"`
	RsyncOpts   string `yaml:"rsync_opts"`
}

type GitConfig struct {
	MirrorRepo string `yaml:"mirror_repo"`
	AutoCommit bool   `yaml:"auto_commit"`
}

type RsyncConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Host       string `yaml:"host"`
	User       string `yaml:"user"`
	TargetPath string `yaml:"target_path"`
	SSHKey     string `yaml:"ssh_key"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Load .env file from project root
	envPath := "/home/ubuntu/hugo-norsetinge/.env"
	if err := godotenv.Load(envPath); err != nil {
		// .env is optional, don't fail if it doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override secrets from environment variables
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		cfg.OpenRouter.APIKey = apiKey
	}
	if smtpPass := os.Getenv("SMTP_PASSWORD"); smtpPass != "" {
		cfg.Email.SMTPPassword = smtpPass
	}
	if imapPass := os.Getenv("IMAP_PASSWORD"); imapPass != "" {
		cfg.Email.IMAPPassword = imapPass
	}
	if ntfyTopic := os.Getenv("NTFY_TOPIC"); ntfyTopic != "" {
		cfg.Ntfy.Topic = ntfyTopic
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks if required configuration fields are set
func (c *Config) Validate() error {
	if c.Dropbox.BasePath == "" {
		return fmt.Errorf("dropbox.base_path is required")
	}
	if c.Dropbox.FolderLanguage == "" {
		return fmt.Errorf("dropbox.folder_language is required")
	}
	if c.Hugo.SiteDir == "" {
		return fmt.Errorf("hugo.site_dir is required")
	}
	// Add more validation as needed
	return nil
}

// GetFolderPath returns the full path to a specific folder based on folder_language
func (c *Config) GetFolderPath(folderType string) (string, error) {
	// TODO: Load folder aliases from folder-aliases.yaml
	// For now, return a placeholder
	return fmt.Sprintf("%s/%s", c.Dropbox.BasePath, folderType), nil
}
