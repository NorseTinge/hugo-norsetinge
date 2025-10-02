package approval

import (
	"testing"

	"norsetinge/config"
)

func TestNewEmailSender(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost:          "mail.norsetinge.com",
			SMTPPort:          587,
			SMTPUser:          "publisher@norsetinge.com",
			SMTPPassword:      "test-password",
			FromAddress:       "publisher@norsetinge.com",
			ApprovalRecipient: "editor@norsetinge.com",
		},
	}

	sender := NewEmailSender(cfg)
	if sender == nil {
		t.Fatal("NewEmailSender returned nil")
	}

	if sender.cfg != cfg {
		t.Error("Config not set correctly")
	}
}

// Note: Actual email sending test would require SMTP server
// Skipping integration test for now
func TestSendApprovalRequest_Mock(t *testing.T) {
	t.Skip("Skipping actual email send - requires SMTP credentials")

	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost:          "mail.norsetinge.com",
			SMTPPort:          587,
			SMTPUser:          "publisher@norsetinge.com",
			SMTPPassword:      "test-password",
			FromAddress:       "publisher@norsetinge.com",
			ApprovalRecipient: "editor@norsetinge.com",
		},
	}

	sender := NewEmailSender(cfg)
	err := sender.SendApprovalRequest(
		"Test Article",
		"TB (twisted brain)",
		"https://norsetinge.tailnet.ts.net/approve/123",
	)

	if err != nil {
		t.Errorf("SendApprovalRequest failed: %v", err)
	}
}
