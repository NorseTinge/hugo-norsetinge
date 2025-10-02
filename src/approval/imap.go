package approval

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"norsetinge/config"
)

// IMAPReader reads and processes email replies
type IMAPReader struct {
	cfg    *config.Config
	client *client.Client
}

// EmailReply represents a parsed email reply
type EmailReply struct {
	MessageID string
	Subject   string
	Body      string
	Action    ApprovalAction
	ArticleID string // Article ID from X-NorseTinge-Article-ID header
}

// ApprovalAction represents the action to take
type ApprovalAction int

const (
	ActionUnknown ApprovalAction = iota
	ActionApprove
	ActionReject
)

// NewIMAPReader creates a new IMAP email reader
func NewIMAPReader(cfg *config.Config) *IMAPReader {
	return &IMAPReader{cfg: cfg}
}

// Connect connects to the IMAP server
func (r *IMAPReader) Connect() error {
	addr := fmt.Sprintf("%s:%d", r.cfg.Email.IMAPHost, r.cfg.Email.IMAPPort)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err := c.Login(r.cfg.Email.IMAPUser, r.cfg.Email.IMAPPassword); err != nil {
		c.Logout()
		return fmt.Errorf("login failed: %w", err)
	}

	r.client = c
	log.Printf("Connected to IMAP server: %s", addr)
	return nil
}

// CheckForReplies checks for new email replies
func (r *IMAPReader) CheckForReplies() ([]*EmailReply, error) {
	if r.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Select INBOX
	mbox, err := r.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select inbox: %w", err)
	}

	if mbox.Messages == 0 {
		return nil, nil
	}

	// Search for unseen messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := r.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	log.Printf("Found %d unread messages", len(ids))

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	// Fetch envelope, body, and RFC822 header to read custom X-NorseTinge-Article-ID
	section := &imap.BodySectionName{Peek: true}
	go func() {
		done <- r.client.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchBody,
			section.FetchItem(),
		}, messages)
	}()

	replies := []*EmailReply{}
	for msg := range messages {
		reply := r.parseMessage(msg)
		if reply != nil {
			replies = append(replies, reply)
		}

		// Mark as seen
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		seqSet := new(imap.SeqSet)
		seqSet.AddNum(msg.SeqNum)
		r.client.Store(seqSet, item, flags, nil)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	return replies, nil
}

// parseMessage parses an IMAP message into an EmailReply
func (r *IMAPReader) parseMessage(msg *imap.Message) *EmailReply {
	if msg.Envelope == nil {
		return nil
	}

	subject := msg.Envelope.Subject
	body := r.extractBody(msg)

	// Try to extract article ID from custom header first
	articleID := r.extractArticleID(msg)

	// If not found in header, try to extract from body/subject (#ABC123-APPR or #ABC123-REJ)
	if articleID == "" {
		searchText := subject + " " + body
		log.Printf("DEBUG: Searching for Article ID in text (length %d): %s", len(searchText), searchText[:min(200, len(searchText))])
		articleID = r.extractArticleIDFromText(searchText)
		if articleID != "" {
			log.Printf("DEBUG: Extracted Article ID from text: %s", articleID)
		} else {
			log.Printf("DEBUG: Failed to extract Article ID from text")
		}
	} else {
		log.Printf("DEBUG: Found Article ID in header: %s", articleID)
	}

	reply := &EmailReply{
		Subject:   subject,
		Body:      body,
		Action:    r.detectAction(subject, body),
		ArticleID: articleID,
	}

	return reply
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractArticleID extracts the Article ID from email headers
func (r *IMAPReader) extractArticleID(msg *imap.Message) string {
	// Look for X-NorseTinge-Article-ID header in the message body
	section := &imap.BodySectionName{Peek: true}
	bodyReader := msg.GetBody(section)
	if bodyReader == nil {
		return ""
	}

	// Parse email headers using textproto
	tp := textproto.NewReader(bufio.NewReader(bodyReader))
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Printf("Failed to read headers: %v", err)
		return ""
	}

	// Get custom header
	articleID := headers.Get("X-Norsetinge-Article-Id")
	if articleID == "" {
		// Try alternative casing
		articleID = headers.Get("X-NorseTinge-Article-ID")
	}

	return articleID
}

// extractArticleIDFromText extracts Article ID from text body (#ABC123-APPR or ABC123-APPR)
func (r *IMAPReader) extractArticleIDFromText(text string) string {
	// Look for patterns like #ABC123-APPR, ABC123-APPR, #ABC123-REJ, or ABC123-REJ
	// Convert to uppercase for matching
	upperText := strings.ToUpper(text)

	// Find XXXXXX-APPR or XXXXXX-REJ (with or without #)
	patterns := []string{"-APPR", "-REJ"}
	for _, pattern := range patterns {
		if idx := strings.Index(upperText, pattern); idx > 0 {
			// Look backwards to find start of ID
			// ID format: 6 uppercase hex characters (A-F0-9), optionally preceded by #
			start := idx - 1

			// Skip backwards while we have valid hex characters
			for start >= 0 && ((upperText[start] >= '0' && upperText[start] <= '9') ||
			                    (upperText[start] >= 'A' && upperText[start] <= 'F')) {
				start--
			}

			// Check if we have # before the ID
			hasHash := start >= 0 && upperText[start] == '#'
			if hasHash {
				start-- // Include the # in the result
			}

			// Extract the ID (should be 6 characters, optionally with # prefix)
			idStart := start + 1
			idEnd := idx

			// Verify we have a valid ID (6 hex chars)
			if hasHash {
				idLength := idEnd - idStart - 1 // -1 for the #
				if idLength == 6 {
					return upperText[idStart:idEnd]
				}
			} else {
				idLength := idEnd - idStart
				if idLength == 6 {
					// Add # prefix if missing
					return "#" + upperText[idStart:idEnd]
				}
			}
		}
	}

	return ""
}

// extractBody extracts body text from message using MIME parsing
func (r *IMAPReader) extractBody(msg *imap.Message) string {
	// Get the full message body (RFC822)
	section := &imap.BodySectionName{Peek: true}
	bodyReader := msg.GetBody(section)
	if bodyReader == nil {
		log.Printf("DEBUG: No body reader available")
		return ""
	}

	// Parse the email message
	mr, err := mail.CreateReader(bodyReader)
	if err != nil {
		log.Printf("DEBUG: Failed to create mail reader: %v", err)
		return ""
	}

	// Read all parts and concatenate text
	var bodyText strings.Builder
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("DEBUG: Error reading part: %v", err)
			break
		}

		// Read text/plain parts
		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := h.ContentType()
			if strings.HasPrefix(contentType, "text/plain") {
				body, err := io.ReadAll(part.Body)
				if err == nil {
					bodyText.WriteString(string(body))
					bodyText.WriteString("\n")
				}
			}
		}
	}

	result := bodyText.String()
	log.Printf("DEBUG: Extracted body text (length %d): %s", len(result), result[:min(500, len(result))])
	return result
}

// detectAction detects the approval action from subject/body
// Priority: 1) ID-based codes (#ABC123-APPR), 2) Fuzzy keywords
func (r *IMAPReader) detectAction(subject, body string) ApprovalAction {
	text := subject + " " + body
	textLower := strings.ToLower(text)

	// First check for ID-based approval/rejection codes
	// Pattern: #XXXXXX-APPR or #XXXXXX-REJ
	if strings.Contains(strings.ToUpper(text), "-APPR") {
		return ActionApprove
	}
	if strings.Contains(strings.ToUpper(text), "-REJ") {
		return ActionReject
	}

	// Fallback to fuzzy approval keywords (case-insensitive)
	approveKeywords := []string{
		"godkend", "godkendt", "godkender",
		"approve", "approved", "accept", "ok",
		"ja", "yes",
	}

	// Fuzzy rejection keywords (case-insensitive)
	rejectKeywords := []string{
		"afvis", "afvist", "afviser",
		"reject", "rejected", "decline",
		"nej", "no",
	}

	// Check for approval
	for _, keyword := range approveKeywords {
		if strings.Contains(textLower, keyword) {
			return ActionApprove
		}
	}

	// Check for rejection
	for _, keyword := range rejectKeywords {
		if strings.Contains(textLower, keyword) {
			return ActionReject
		}
	}

	return ActionUnknown
}

// StartMonitoring starts monitoring inbox for replies
func (r *IMAPReader) StartMonitoring(interval time.Duration, handler func(*EmailReply)) error {
	if err := r.Connect(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Started IMAP monitoring (checking every %v)", interval)

	for range ticker.C {
		replies, err := r.CheckForReplies()
		if err != nil {
			log.Printf("Error checking replies: %v", err)
			continue
		}

		for _, reply := range replies {
			log.Printf("Processing reply: Action=%v, Subject=%s", reply.Action, reply.Subject)
			handler(reply)
		}
	}

	return nil
}

// Close closes the IMAP connection
func (r *IMAPReader) Close() error {
	if r.client != nil {
		return r.client.Logout()
	}
	return nil
}
