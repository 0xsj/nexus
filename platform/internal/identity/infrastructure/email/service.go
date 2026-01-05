package email

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Configuration
// ============================================================================

// Config contains email service configuration.
type Config struct {
	// BaseURL is the application base URL for links.
	BaseURL string

	// FromAddress is the sender email address.
	FromAddress string

	// FromName is the sender display name.
	FromName string
}

// DefaultConfig returns default configuration.
func DefaultConfig() Config {
	return Config{
		BaseURL:     "https://nexus.io",
		FromAddress: "noreply@nexus.io",
		FromName:    "Nexus",
	}
}

// ============================================================================
// Service Interface
// ============================================================================

// Sender is the low-level email sending interface.
// Implementations: SMTP, SendGrid, AWS SES, Resend, etc.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Message represents an email message.
type Message struct {
	To       string
	From     string
	FromName string
	Subject  string
	TextBody string
	HTMLBody string
}

// ============================================================================
// Service
// ============================================================================

// Service implements domain.EmailService.
type Service struct {
	sender Sender
	config Config
	logger log.Logger
}

// Ensure Service implements domain.EmailService.
var _ domain.EmailService = (*Service)(nil)

// NewService creates a new email service.
func NewService(sender Sender, config Config, logger log.Logger) *Service {
	return &Service{
		sender: sender,
		config: config,
		logger: logger,
	}
}

// ============================================================================
// Email Operations
// ============================================================================

// SendMagicLink sends a magic link email.
func (s *Service) SendMagicLink(ctx context.Context, params domain.SendMagicLinkParams) error {
	const op = "email.Service.SendMagicLink"

	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", s.config.BaseURL, params.Token)

	subject := s.getMagicLinkSubject(params.Purpose)
	textBody := s.getMagicLinkTextBody(params.Purpose, verifyURL, params.ExpiresAt.Format("15:04 MST"))
	htmlBody := s.getMagicLinkHTMLBody(params.Purpose, verifyURL, params.ExpiresAt.Format("15:04 MST"))

	msg := Message{
		To:       params.To,
		From:     s.config.FromAddress,
		FromName: s.config.FromName,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	if err := s.sender.Send(ctx, msg); err != nil {
		s.logger.Error("failed to send magic link email",
			log.String("op", op),
			log.String("to", params.To),
			log.String("purpose", params.Purpose.String()),
			log.Err(err),
		)
		return domain.ErrEmailSendFailed(op, err)
	}

	s.logger.Info("magic link email sent",
		log.String("to", params.To),
		log.String("purpose", params.Purpose.String()),
	)

	return nil
}

// SendWelcome sends a welcome email to new users.
func (s *Service) SendWelcome(ctx context.Context, params domain.SendWelcomeParams) error {
	const op = "email.Service.SendWelcome"

	subject := "Welcome to Nexus!"
	textBody := fmt.Sprintf("Hi %s,\n\nWelcome to Nexus - your verified identity platform.\n\nGet started by connecting your professional profiles and building your verified credentials.\n\nBest,\nThe Nexus Team", params.Username)
	htmlBody := fmt.Sprintf(`
		<h1>Welcome to Nexus!</h1>
		<p>Hi %s,</p>
		<p>Welcome to Nexus - your verified identity platform.</p>
		<p>Get started by connecting your professional profiles and building your verified credentials.</p>
		<p>Best,<br>The Nexus Team</p>
	`, params.Username)

	msg := Message{
		To:       params.To,
		From:     s.config.FromAddress,
		FromName: s.config.FromName,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	if err := s.sender.Send(ctx, msg); err != nil {
		s.logger.Error("failed to send welcome email",
			log.String("op", op),
			log.String("to", params.To),
			log.Err(err),
		)
		return domain.ErrEmailSendFailed(op, err)
	}

	s.logger.Info("welcome email sent", log.String("to", params.To))

	return nil
}

// SendSecurityAlert sends a security alert email.
func (s *Service) SendSecurityAlert(ctx context.Context, params domain.SendSecurityAlertParams) error {
	const op = "email.Service.SendSecurityAlert"

	subject := fmt.Sprintf("Security Alert: %s", params.AlertType)
	textBody := fmt.Sprintf("Security Alert\n\nType: %s\nMessage: %s\nIP Address: %s\nTime: %s\n\nIf this wasn't you, please secure your account immediately.",
		params.AlertType, params.Message, params.IPAddress, params.Timestamp.Format("Jan 2, 2006 at 15:04 MST"))
	htmlBody := fmt.Sprintf(`
		<h1>Security Alert</h1>
		<p><strong>Type:</strong> %s</p>
		<p><strong>Message:</strong> %s</p>
		<p><strong>IP Address:</strong> %s</p>
		<p><strong>Time:</strong> %s</p>
		<p>If this wasn't you, please secure your account immediately.</p>
	`, params.AlertType, params.Message, params.IPAddress, params.Timestamp.Format("Jan 2, 2006 at 15:04 MST"))

	msg := Message{
		To:       params.To,
		From:     s.config.FromAddress,
		FromName: s.config.FromName,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}

	if err := s.sender.Send(ctx, msg); err != nil {
		s.logger.Error("failed to send security alert email",
			log.String("op", op),
			log.String("to", params.To),
			log.String("alert_type", params.AlertType),
			log.Err(err),
		)
		return domain.ErrEmailSendFailed(op, err)
	}

	s.logger.Info("security alert email sent",
		log.String("to", params.To),
		log.String("alert_type", params.AlertType),
	)

	return nil
}

// ============================================================================
// Email Templates
// ============================================================================

func (s *Service) getMagicLinkSubject(purpose domain.MagicLinkPurpose) string {
	switch purpose {
	case domain.MagicLinkPurposeLogin:
		return "Sign in to Nexus"
	case domain.MagicLinkPurposeRegister:
		return "Complete your Nexus registration"
	case domain.MagicLinkPurposeVerify:
		return "Verify your email address"
	case domain.MagicLinkPurposeLink:
		return "Link your email to Nexus"
	default:
		return "Your Nexus magic link"
	}
}

func (s *Service) getMagicLinkTextBody(purpose domain.MagicLinkPurpose, url, expiresAt string) string {
	action := s.getMagicLinkAction(purpose)
	return fmt.Sprintf("%s\n\nClick the link below to %s:\n\n%s\n\nThis link expires at %s.\n\nIf you didn't request this, you can safely ignore this email.\n\nBest,\nThe Nexus Team",
		s.getMagicLinkGreeting(purpose), action, url, expiresAt)
}

func (s *Service) getMagicLinkHTMLBody(purpose domain.MagicLinkPurpose, url, expiresAt string) string {
	action := s.getMagicLinkAction(purpose)
	buttonText := s.getMagicLinkButtonText(purpose)

	return fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
			<h1 style="color: #333;">%s</h1>
			<p>Click the button below to %s:</p>
			<p style="margin: 30px 0;">
				<a href="%s" style="background-color: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block;">
					%s
				</a>
			</p>
			<p style="color: #666; font-size: 14px;">This link expires at %s.</p>
			<p style="color: #666; font-size: 14px;">If you didn't request this, you can safely ignore this email.</p>
			<p style="margin-top: 40px; color: #666;">Best,<br>The Nexus Team</p>
		</div>
	`, s.getMagicLinkGreeting(purpose), action, url, buttonText, expiresAt)
}

func (s *Service) getMagicLinkGreeting(purpose domain.MagicLinkPurpose) string {
	switch purpose {
	case domain.MagicLinkPurposeRegister:
		return "Welcome to Nexus!"
	default:
		return "Hello!"
	}
}

func (s *Service) getMagicLinkAction(purpose domain.MagicLinkPurpose) string {
	switch purpose {
	case domain.MagicLinkPurposeLogin:
		return "sign in to your account"
	case domain.MagicLinkPurposeRegister:
		return "complete your registration"
	case domain.MagicLinkPurposeVerify:
		return "verify your email address"
	case domain.MagicLinkPurposeLink:
		return "link this email to your account"
	default:
		return "continue"
	}
}

func (s *Service) getMagicLinkButtonText(purpose domain.MagicLinkPurpose) string {
	switch purpose {
	case domain.MagicLinkPurposeLogin:
		return "Sign In"
	case domain.MagicLinkPurposeRegister:
		return "Complete Registration"
	case domain.MagicLinkPurposeVerify:
		return "Verify Email"
	case domain.MagicLinkPurposeLink:
		return "Link Email"
	default:
		return "Continue"
	}
}

// ============================================================================
// Console Sender (Development)
// ============================================================================

// ConsoleSender logs emails to console instead of sending.
// Use for local development.
type ConsoleSender struct {
	logger log.Logger
}

// NewConsoleSender creates a new console sender.
func NewConsoleSender(logger log.Logger) *ConsoleSender {
	return &ConsoleSender{logger: logger}
}

// Send logs the email to console.
func (s *ConsoleSender) Send(ctx context.Context, msg Message) error {
	s.logger.Info("========== EMAIL ==========",
		log.String("to", msg.To),
		log.String("from", fmt.Sprintf("%s <%s>", msg.FromName, msg.From)),
		log.String("subject", msg.Subject),
	)
	s.logger.Info("--- TEXT BODY ---")
	s.logger.Info(msg.TextBody)
	s.logger.Info("========== END EMAIL ==========")

	return nil
}

// Ensure ConsoleSender implements Sender.
var _ Sender = (*ConsoleSender)(nil)
