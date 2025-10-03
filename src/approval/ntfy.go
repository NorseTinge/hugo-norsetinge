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

// SendApprovalNotification sends a simple notification with preview URL
func (n *NtfySender) SendApprovalNotification(title, author, previewURLPath, articleID string) error {
	if !n.cfg.Ntfy.Enabled {
		log.Printf("ntfy notifications disabled")
		return nil
	}

	// Generate Tailscale preview URL
	previewURL := fmt.Sprintf("http://%s:%d/preview/%s",
		n.cfg.Approval.TailscaleHostname,
		n.cfg.Approval.Port,
		previewURLPath)

	msg := NtfyMessage{
		Topic:    n.cfg.Ntfy.Topic,
		Title:    fmt.Sprintf("📰 %s", title),
		Message:  fmt.Sprintf("Af: %s\n\n%s", author, previewURL),
		Priority: 4, // High priority
		Tags:     []string{"newspaper"},
		Actions: []NtfyAction{
			{
				Action: "view",
				Label:  "Åbn",
				URL:    previewURL,
			},
		},
	}

	return n.send(msg)
}

// send sends a ntfy notification
func (n *NtfySender) send(msg NtfyMessage) error {
	url := fmt.Sprintf("%s/%s", n.cfg.Ntfy.Server, n.cfg.Ntfy.Topic)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal ntfy message: %w", err)
	}

	// Debug: Log the JSON being sent
	log.Printf("📤 Sending ntfy JSON: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
