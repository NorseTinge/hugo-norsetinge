package approval

import (
	"fmt"
	"net/smtp"
	"strings"

	"norsetinge/config"
)

// EmailSender handles sending approval emails
type EmailSender struct {
	cfg *config.Config
}

// NewEmailSender creates a new email sender
func NewEmailSender(cfg *config.Config) *EmailSender {
	return &EmailSender{cfg: cfg}
}

// SendApprovalRequestHTML sends HTML email with link to Dropbox preview
func (e *EmailSender) SendApprovalRequestHTML(articleTitle, articleAuthor, dropboxHTMLPath, approvalURL, articleID string) error {
	subject := fmt.Sprintf("Artikel til godkendelse: %s", articleTitle)

	// Create simple email body with Dropbox link
	htmlBody := e.buildDropboxApprovalEmail(articleTitle, articleAuthor, dropboxHTMLPath, approvalURL, articleID)

	return e.sendHTML(subject, htmlBody, articleID)
}

// SendApprovalRequest sends plain text email (fallback)
func (e *EmailSender) SendApprovalRequest(articleTitle, articleAuthor, previewURL string) error {
	subject := fmt.Sprintf("Artikel til godkendelse: %s", articleTitle)
	body := fmt.Sprintf(`
Hej,

En ny artikel er klar til godkendelse:

Titel: %s
Forfatter: %s

Preview og godkendelse: %s

Venlig hilsen,
Norsetinge
`, articleTitle, articleAuthor, previewURL)

	return e.send(subject, body)
}

// buildDropboxApprovalEmail creates simple email with Dropbox preview link
func (e *EmailSender) buildDropboxApprovalEmail(title, author, dropboxPath, approvalURL, articleID string) string {
	// Extract ID without # for cleaner display
	cleanID := articleID
	if len(articleID) > 0 && articleID[0] == '#' {
		cleanID = articleID[1:]
	}

	approveCode := articleID + "-APPR"
	rejectCode := articleID + "-REJ"

	// Extract relative Dropbox path (from "Dropbox/" onwards)
	relativeDropboxPath := dropboxPath
	if idx := strings.Index(dropboxPath, "Dropbox/"); idx != -1 {
		relativeDropboxPath = dropboxPath[idx:]
	}

	// Create clickable Dropbox link
	dropboxLink := "file://" + dropboxPath

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Godkend: %s</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif; background: #f5f5f5;">
	<div style="max-width: 600px; margin: 40px auto; background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden;">
		<!-- Header -->
		<div style="background: #333; color: white; padding: 30px; text-align: center;">
			<h1 style="margin: 0; font-size: 24px;">📰 Ny artikel til godkendelse</h1>
			<p style="margin: 10px 0 0 0; font-size: 14px; opacity: 0.8;">Artikel ID: %s</p>
		</div>

		<!-- Content -->
		<div style="padding: 30px;">
			<h2 style="margin: 0 0 10px 0; color: #333;">%s</h2>
			<p style="margin: 0 0 20px 0; color: #666; font-size: 14px;">
				<strong>Forfatter:</strong> %s
			</p>

			<p style="color: #555; line-height: 1.6; margin: 0 0 25px 0;">
				En ny artikel er klar til godkendelse. Preview er gemt i din Dropbox.
			</p>

			<!-- Preview Location -->
			<div style="background: #e7f3ff; border-left: 4px solid #2196F3; padding: 15px; margin: 20px 0; border-radius: 4px;">
				<p style="margin: 0 0 10px 0; color: #333; font-weight: 600;">📁 Preview placering:</p>
				<a href="%s" style="background: white; padding: 10px 15px; border-radius: 4px; font-size: 14px; display: inline-block; color: #2196F3; text-decoration: none; border: 2px solid #2196F3; font-weight: 600;">
					🔗 Åbn i Dropbox: %s
				</a>
			</div>

			<!-- Approval Instructions -->
			<div style="border-top: 2px solid #eee; margin-top: 30px; padding-top: 25px;">
				<h3 style="color: #333; margin: 0 0 15px 0; text-align: center;">Godkend eller afvis</h3>
				<p style="color: #555; line-height: 1.6; text-align: center; margin: 0 0 20px 0;">
					Svar på denne email med en af disse koder:
				</p>
				<div style="background: #f8f9fa; padding: 20px; border-radius: 6px;">
					<div style="text-align: center; margin-bottom: 15px;">
						<strong style="color: #28a745; font-size: 18px;">✅ Godkend artiklen:</strong>
					</div>
					<div style="background: white; padding: 12px; border-radius: 4px; margin-bottom: 20px; text-align: center; border: 2px solid #28a745;">
						<code style="font-size: 20px; font-weight: bold; color: #28a745; font-family: monospace; letter-spacing: 2px;">%s</code>
					</div>

					<div style="border-top: 2px solid #ddd; margin: 20px 0;"></div>

					<div style="text-align: center; margin-bottom: 15px;">
						<strong style="color: #dc3545; font-size: 18px;">❌ Afvis artiklen:</strong>
					</div>
					<div style="background: white; padding: 12px; border-radius: 4px; text-align: center; border: 2px solid #dc3545;">
						<code style="font-size: 20px; font-weight: bold; color: #dc3545; font-family: monospace; letter-spacing: 2px;">%s</code>
					</div>
				</div>
				<p style="color: #888; font-size: 12px; text-align: center; margin: 20px 0 0 0;">
					💡 Tip: Kopiér koden og indsæt den i din svar-email<br>
					Du kan også bruge: GODKENDT, AFVIST, APPROVE, REJECT
				</p>
			</div>
		</div>

		<!-- Footer -->
		<div style="background: #f8f9fa; padding: 20px; text-align: center; border-top: 1px solid #e0e0e0;">
			<p style="margin: 0; color: #666; font-size: 13px;">
				<strong>NorseTinge</strong> · Automated Publishing System
			</p>
		</div>
	</div>
</body>
</html>`, title, cleanID, title, author, dropboxLink, relativeDropboxPath, approveCode, rejectCode)
}

// sendHTML sends HTML email via SMTP with article ID in custom header
func (e *EmailSender) sendHTML(subject, htmlBody, articleID string) error {
	auth := smtp.PlainAuth(
		"",
		e.cfg.Email.SMTPUser,
		e.cfg.Email.SMTPPassword,
		e.cfg.Email.SMTPHost,
	)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"X-NorseTinge-Article-ID: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s",
		e.cfg.Email.FromAddress,
		e.cfg.Email.ApprovalRecipient,
		subject,
		articleID,
		htmlBody,
	)

	addr := fmt.Sprintf("%s:%d", e.cfg.Email.SMTPHost, e.cfg.Email.SMTPPort)
	return smtp.SendMail(
		addr,
		auth,
		e.cfg.Email.FromAddress,
		[]string{e.cfg.Email.ApprovalRecipient},
		[]byte(message),
	)
}

// send sends plain text email via SMTP
func (e *EmailSender) send(subject, body string) error {
	auth := smtp.PlainAuth(
		"",
		e.cfg.Email.SMTPUser,
		e.cfg.Email.SMTPPassword,
		e.cfg.Email.SMTPHost,
	)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s",
		e.cfg.Email.FromAddress,
		e.cfg.Email.ApprovalRecipient,
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", e.cfg.Email.SMTPHost, e.cfg.Email.SMTPPort)
	return smtp.SendMail(
		addr,
		auth,
		e.cfg.Email.FromAddress,
		[]string{e.cfg.Email.ApprovalRecipient},
		[]byte(message),
	)
}
