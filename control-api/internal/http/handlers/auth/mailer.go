package auth

import (
	"context"
	"log"
)

// Mailer delivers transactional emails. The project has no SMTP infrastructure
// yet, so LogMailer is the default implementation; a real SMTP mailer can be
// added behind this interface without touching the handlers.
type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail, link string) error
}

// LogMailer logs the reset link instead of sending it. Temporary until SMTP.
type LogMailer struct{}

func (LogMailer) SendPasswordReset(_ context.Context, toEmail, link string) error {
	log.Printf("[mailer] password reset for %s: %s", toEmail, link)
	return nil
}
