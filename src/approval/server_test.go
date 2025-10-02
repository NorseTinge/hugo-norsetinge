package approval

import (
	"testing"

	"norsetinge/common"
	"norsetinge/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Approval: config.ApprovalConfig{
			Host: "0.0.0.0",
			Port: 8080,
			TailscaleHostname: "norsetinge.tailnet.ts.net",
		},
	}

	server := NewServer(cfg)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.cfg != cfg {
		t.Error("Config not set correctly")
	}

	if server.pendingArticles == nil {
		t.Error("pendingArticles map not initialized")
	}
}

func TestGenerateID(t *testing.T) {
	id1, err := generateID()
	if err != nil {
		t.Fatalf("generateID failed: %v", err)
	}

	if len(id1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("Expected ID length 32, got %d", len(id1))
	}

	// Generate another ID to ensure uniqueness
	id2, err := generateID()
	if err != nil {
		t.Fatalf("generateID failed: %v", err)
	}

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}
}

func TestRequestApproval(t *testing.T) {
	cfg := &config.Config{
		Approval: config.ApprovalConfig{
			Host: "0.0.0.0",
			Port: 8080,
			TailscaleHostname: "norsetinge.tailnet.ts.net",
		},
		Email: config.EmailConfig{
			SMTPHost:          "mail.norsetinge.com",
			SMTPPort:          587,
			SMTPUser:          "publisher@norsetinge.com",
			SMTPPassword:      "test",
			FromAddress:       "publisher@norsetinge.com",
			ApprovalRecipient: "lpm@lpmathiasen.com",
		},
	}

	server := NewServer(cfg)

	article := &common.Article{
		Title:  "Test Article",
		Author: "TB",
		Content: "Test content",
	}

	// Note: This will fail to send email without valid SMTP
	// but should still create pending article
	_ = server.RequestApproval(article)

	// Check if article was added to pending
	server.mu.RLock()
	count := len(server.pendingArticles)
	server.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 pending article, got %d", count)
	}
}
