package adapters

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/email"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Compile-time interface check.
var _ domain.EmailService = (*EmailService)(nil)

// EmailServiceConfig holds configuration for the EmailService adapter.
type EmailServiceConfig struct {
	SMTPSender *email.SMTPSender
	BaseURL    string // Base URL for magic link URLs (e.g. "http://localhost:3010")
}

// EmailService sends identity-related emails via SMTP.
type EmailService struct {
	smtp    *email.SMTPSender
	baseURL string
}

// NewEmailService creates a new EmailService.
func NewEmailService(cfg EmailServiceConfig) *EmailService {
	return &EmailService{
		smtp:    cfg.SMTPSender,
		baseURL: cfg.BaseURL,
	}
}

// SendMagicLink sends a magic link email for authentication.
func (s *EmailService) SendMagicLink(_ context.Context, to types.Email, token string, expiresIn int) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", s.baseURL, token)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2>Sign in to Nexus</h2>
  <p>Click the link below to sign in. This link expires in %d minutes.</p>
  <p><a href="%s" style="display: inline-block; padding: 12px 24px; background: #2563eb; color: #fff; text-decoration: none; border-radius: 6px;">Sign In</a></p>
  <p style="color: #6b7280; font-size: 14px;">If you didn't request this, you can safely ignore this email.</p>
</body>
</html>`, expiresIn/60, link)

	return s.smtp.Send(to.String(), "Sign in to Nexus", body)
}

// SendWelcome sends a welcome email to a new user.
func (s *EmailService) SendWelcome(_ context.Context, to types.Email, displayName string) error {
	greeting := "Welcome to Nexus"
	if displayName != "" {
		greeting = fmt.Sprintf("Welcome to Nexus, %s", displayName)
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2>%s</h2>
  <p>Your verified professional identity starts here. Connect your accounts, earn credentials, and own your reputation.</p>
  <p><a href="%s" style="display: inline-block; padding: 12px 24px; background: #2563eb; color: #fff; text-decoration: none; border-radius: 6px;">Get Started</a></p>
</body>
</html>`, greeting, s.baseURL)

	return s.smtp.Send(to.String(), greeting, body)
}
