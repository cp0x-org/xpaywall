package auth

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/cp0x-org/xpaywall/control-api/config"
)

// smtpSendTimeout bounds a single delivery so a slow/hung server can't stall the
// HTTP request that triggered it.
const smtpSendTimeout = 15 * time.Second

// Mailer delivers transactional emails. LogMailer is used when SMTP is not
// configured; SMTPMailer delivers over a real server.
type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail, link string) error
	SendWelcome(ctx context.Context, toEmail, username string) error
}

// LogMailer logs emails instead of sending them. Used when SMTP is unconfigured.
type LogMailer struct{}

func (LogMailer) SendPasswordReset(_ context.Context, toEmail, link string) error {
	log.Printf("[mailer] password reset for %s: %s", toEmail, link)
	return nil
}

func (LogMailer) SendWelcome(_ context.Context, toEmail, username string) error {
	log.Printf("[mailer] welcome email for %s (%s)", username, toEmail)
	return nil
}

// SMTPMailer sends email over SMTP using STARTTLS (port 587).
type SMTPMailer struct {
	addr       string // host:port
	auth       smtp.Auth
	from       string
	fromName   string
	appBaseURL string
}

// NewSMTPMailer builds an SMTPMailer from config. PLAIN auth is used when a
// username is set; From defaults to the SMTP username when SMTP_FROM is empty.
func NewSMTPMailer(cfg *config.ControlAPIConfig) *SMTPMailer {
	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUsername
	}
	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}
	return &SMTPMailer{
		addr:       fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort),
		auth:       auth,
		from:       from,
		fromName:   cfg.SMTPFromName,
		appBaseURL: cfg.AppBaseURL,
	}
}

func (m *SMTPMailer) SendPasswordReset(_ context.Context, toEmail, link string) error {
	body := `<p>We received a request to reset your password.</p>` +
		fmt.Sprintf(`<p><a href="%s">Choose a new password</a>. This link expires in 1 hour.</p>`, link) +
		`<p>If you didn't request this, you can safely ignore this email.</p>`
	return m.send(toEmail, "Reset your xpaywall password", body)
}

func (m *SMTPMailer) SendWelcome(_ context.Context, toEmail, username string) error {
	body := fmt.Sprintf(`<p>Hi %s,</p>`, html.EscapeString(username)) +
		`<p>Your xpaywall account has been created.</p>` +
		fmt.Sprintf(`<p><a href="%s/login">Sign in to the dashboard</a>.</p>`, m.appBaseURL)
	return m.send(toEmail, "Welcome to xpaywall", body)
}

// send delivers a single HTML message to one recipient, bounded by a timeout so
// a hung SMTP server cannot block the calling request indefinitely.
func (m *SMTPMailer) send(to, subject, htmlBody string) error {
	msg := m.message(to, subject, htmlBody)
	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(m.addr, m.auth, m.from, []string{to}, msg)
	}()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("smtp send: %w", err)
		}
		return nil
	case <-time.After(smtpSendTimeout):
		return fmt.Errorf("smtp send: timed out after %s", smtpSendTimeout)
	}
}

// message assembles RFC 5322 headers plus an HTML body.
func (m *SMTPMailer) message(to, subject, htmlBody string) []byte {
	var b strings.Builder
	if m.fromName != "" {
		fmt.Fprintf(&b, "From: %s <%s>\r\n", m.fromName, m.from)
	} else {
		fmt.Fprintf(&b, "From: %s\r\n", m.from)
	}
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return []byte(b.String())
}
