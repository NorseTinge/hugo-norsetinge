package approval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"norsetinge/config"
)

// NtfySender sends push notifications via ntfy.sh
type NtfySender struct {
	cfg *config.Config
}

// NtfyMessage represents a ntfy notification
type NtfyMessage struct {
	Topic   string        `json:"topic"`
	Title   string        `json:"title"`
	Message string        `json:"message"`
	Actions []NtfyAction  `json:"actions,omitempty"`
	Tags    []string      `json:"tags,omitempty"`
	Priority int          `json:"priority,omitempty"`
}

// NtfyAction represents a clickable action button
type NtfyAction struct {
	Action string `json:"action"` // "view" or "http"
	Label  string `json:"label"`
	URL    string `json:"url"`
	Clear  bool   `json:"clear,omitempty"` // Clear notification after click
}

// NewNtfySender creates a new ntfy sender
func NewNtfySender(cfg *config.Config) *NtfySender {
	return &NtfySender{cfg: cfg}
}

// SendApprovalNotification sends a simple notification with approval URL
func (n *NtfySender) SendApprovalNotification(title, author, previewURLPath, articleID string) error {
	if !n.cfg.Ntfy.Enabled {
		log.Printf("ntfy notifications disabled")
		return nil
	}

	// Generate Tailscale approval URL (https via tailscale serve)
	// This shows the approval page with 3 buttons + article preview
	approvalURL := fmt.Sprintf("https://%s/approve/%s",
		n.cfg.Approval.TailscaleHostname,
		articleID)

	msg := NtfyMessage{
		Topic:    n.cfg.Ntfy.Topic,
		Title:    fmt.Sprintf("📰 %s", title),
		Message:  fmt.Sprintf("Af: %s\n\nKlik for at godkende/afvise", author),
		Priority: 4, // High priority
		Tags:     []string{"newspaper"},
		Actions: []NtfyAction{
			{
				Action: "view",
				Label:  "Godkend artikel",
				URL:    approvalURL,
			},
		},
	}

	return n.send(msg)
}

// send sends a ntfy notification using headers (not JSON body)
func (n *NtfySender) send(msg NtfyMessage) error {
	url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, n.cfg.Ntfy.Topic)

	// Send message as body, metadata as headers
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(msg.Message))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers according to ntfy documentation
	req.Header.Set("Title", msg.Title)
	req.Header.Set("Priority", fmt.Sprintf("%d", msg.Priority))
	req.Header.Set("Tags", "newspaper")

	// Add action button as JSON in header
	if len(msg.Actions) > 0 {
		actionsJSON, _ := json.Marshal(msg.Actions)
		req.Header.Set("Actions", string(actionsJSON))
	}

	log.Printf("📤 Sending ntfy notification: %s", msg.Title)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ntfy notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	log.Printf("📱 ntfy notification sent: %s", msg.Title)
	return nil
}

// ClearAllNotifications deletes all notifications from the topic
func (n *NtfySender) ClearAllNotifications() error {
	if !n.cfg.Ntfy.Enabled {
		return nil
	}

	url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, n.cfg.Ntfy.Topic)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create DELETE request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to clear ntfy notifications: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy DELETE returned status %d", resp.StatusCode)
	}

	log.Printf("🧹 Cleared all ntfy notifications from topic")
	return nil
}
