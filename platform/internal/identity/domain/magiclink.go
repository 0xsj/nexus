package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Magic Link Token
// ============================================================================

// MagicLinkToken represents a one-time use token for passwordless authentication.
type MagicLinkToken struct {
	id        string
	email     string
	token     string
	tokenHash string
	purpose   MagicLinkPurpose
	status    MagicLinkStatus
	createdAt time.Time
	expiresAt time.Time
	usedAt    *time.Time
	ipAddress string
	userAgent string
}

// MagicLinkPurpose indicates why the magic link was created.
type MagicLinkPurpose string

const (
	// MagicLinkPurposeLogin is for logging in existing users.
	MagicLinkPurposeLogin MagicLinkPurpose = "login"

	// MagicLinkPurposeRegister is for registering new users.
	MagicLinkPurposeRegister MagicLinkPurpose = "register"

	// MagicLinkPurposeVerify is for verifying email ownership.
	MagicLinkPurposeVerify MagicLinkPurpose = "verify"

	// MagicLinkPurposeLink is for linking email to existing account.
	MagicLinkPurposeLink MagicLinkPurpose = "link"
)

// String returns the string representation.
func (p MagicLinkPurpose) String() string {
	return string(p)
}

// IsValid checks if the purpose is valid.
func (p MagicLinkPurpose) IsValid() bool {
	switch p {
	case MagicLinkPurposeLogin, MagicLinkPurposeRegister, MagicLinkPurposeVerify, MagicLinkPurposeLink:
		return true
	default:
		return false
	}
}

// MagicLinkStatus represents the status of a magic link token.
type MagicLinkStatus string

const (
	// MagicLinkStatusPending means the token is valid and unused.
	MagicLinkStatusPending MagicLinkStatus = "pending"

	// MagicLinkStatusUsed means the token has been used.
	MagicLinkStatusUsed MagicLinkStatus = "used"

	// MagicLinkStatusExpired means the token has expired.
	MagicLinkStatusExpired MagicLinkStatus = "expired"

	// MagicLinkStatusRevoked means the token was manually revoked.
	MagicLinkStatusRevoked MagicLinkStatus = "revoked"
)

// String returns the string representation.
func (s MagicLinkStatus) String() string {
	return string(s)
}

// ============================================================================
// Constructor
// ============================================================================

// NewMagicLinkToken creates a new magic link token.
func NewMagicLinkToken(
	id string,
	email string,
	token string,
	tokenHash string,
	purpose MagicLinkPurpose,
	ttl time.Duration,
	ipAddress string,
	userAgent string,
) *MagicLinkToken {
	now := time.Now()
	return &MagicLinkToken{
		id:        id,
		email:     email,
		token:     token,
		tokenHash: tokenHash,
		purpose:   purpose,
		status:    MagicLinkStatusPending,
		createdAt: now,
		expiresAt: now.Add(ttl),
		ipAddress: ipAddress,
		userAgent: userAgent,
	}
}

// ReconstituteMagicLinkToken recreates a token from persistence.
func ReconstituteMagicLinkToken(
	id string,
	email string,
	tokenHash string,
	purpose MagicLinkPurpose,
	status MagicLinkStatus,
	createdAt time.Time,
	expiresAt time.Time,
	usedAt *time.Time,
	ipAddress string,
	userAgent string,
) *MagicLinkToken {
	return &MagicLinkToken{
		id:        id,
		email:     email,
		tokenHash: tokenHash,
		purpose:   purpose,
		status:    status,
		createdAt: createdAt,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		ipAddress: ipAddress,
		userAgent: userAgent,
	}
}

// ============================================================================
// Accessors
// ============================================================================

func (t *MagicLinkToken) ID() string                { return t.id }
func (t *MagicLinkToken) Email() string             { return t.email }
func (t *MagicLinkToken) Token() string             { return t.token }
func (t *MagicLinkToken) TokenHash() string         { return t.tokenHash }
func (t *MagicLinkToken) Purpose() MagicLinkPurpose { return t.purpose }
func (t *MagicLinkToken) Status() MagicLinkStatus   { return t.status }
func (t *MagicLinkToken) CreatedAt() time.Time      { return t.createdAt }
func (t *MagicLinkToken) ExpiresAt() time.Time      { return t.expiresAt }
func (t *MagicLinkToken) UsedAt() *time.Time        { return t.usedAt }
func (t *MagicLinkToken) IPAddress() string         { return t.ipAddress }
func (t *MagicLinkToken) UserAgent() string         { return t.userAgent }

// ============================================================================
// State Checks
// ============================================================================

// IsValid checks if the token is valid (pending and not expired).
func (t *MagicLinkToken) IsValid() bool {
	return t.status == MagicLinkStatusPending && !t.IsExpired()
}

// IsExpired checks if the token has expired.
func (t *MagicLinkToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// IsUsed checks if the token has been used.
func (t *MagicLinkToken) IsUsed() bool {
	return t.status == MagicLinkStatusUsed
}

// ============================================================================
// Operations
// ============================================================================

// MarkUsed marks the token as used.
func (t *MagicLinkToken) MarkUsed() error {
	if t.status != MagicLinkStatusPending {
		return ErrMagicLinkAlreadyUsed("MagicLinkToken.MarkUsed", t.id)
	}

	if t.IsExpired() {
		t.status = MagicLinkStatusExpired
		return ErrMagicLinkExpired("MagicLinkToken.MarkUsed", t.id)
	}

	now := time.Now()
	t.status = MagicLinkStatusUsed
	t.usedAt = &now
	return nil
}

// Revoke revokes the token.
func (t *MagicLinkToken) Revoke() {
	t.status = MagicLinkStatusRevoked
}

// ============================================================================
// Magic Link Error Codes
// ============================================================================

const (
	CodeMagicLinkNotFound    errors.Code = "MAGIC_LINK_NOT_FOUND"
	CodeMagicLinkExpired     errors.Code = "MAGIC_LINK_EXPIRED"
	CodeMagicLinkAlreadyUsed errors.Code = "MAGIC_LINK_ALREADY_USED"
	CodeMagicLinkInvalid     errors.Code = "MAGIC_LINK_INVALID"
	CodeEmailSendFailed      errors.Code = "EMAIL_SEND_FAILED"
	CodeEmailRateLimited     errors.Code = "EMAIL_RATE_LIMITED"
)

// ============================================================================
// Magic Link Errors
// ============================================================================

// ErrMagicLinkNotFound creates a magic link not found error.
func ErrMagicLinkNotFound(operation string, identifier string) *errors.Error {
	return errors.NotFound(operation, "magic link: "+identifier).
		WithCode(CodeMagicLinkNotFound).
		WithMeta("identifier", identifier)
}

// ErrMagicLinkExpired creates a magic link expired error.
func ErrMagicLinkExpired(operation string, tokenID string) *errors.Error {
	return errors.Unauthorized(operation, "magic link has expired").
		WithCode(CodeMagicLinkExpired).
		WithMeta("token_id", tokenID)
}

// ErrMagicLinkAlreadyUsed creates a magic link already used error.
func ErrMagicLinkAlreadyUsed(operation string, tokenID string) *errors.Error {
	return errors.Unauthorized(operation, "magic link has already been used").
		WithCode(CodeMagicLinkAlreadyUsed).
		WithMeta("token_id", tokenID)
}

// ErrMagicLinkInvalid creates a magic link invalid error.
func ErrMagicLinkInvalid(operation string, reason string) *errors.Error {
	return errors.Unauthorized(operation, "invalid magic link: "+reason).
		WithCode(CodeMagicLinkInvalid)
}

// ErrEmailSendFailed creates an email send failed error.
func ErrEmailSendFailed(operation string, err error) *errors.Error {
	return errors.Infrastructure(operation, err).
		WithCode(CodeEmailSendFailed).
		WithMessage("failed to send email")
}

// ErrEmailRateLimited creates an email rate limited error.
func ErrEmailRateLimited(operation string, email string) *errors.Error {
	return errors.RateLimit(operation, "too many email requests").
		WithCode(CodeEmailRateLimited).
		WithMeta("email", email)
}

// ============================================================================
// Error Checkers
// ============================================================================

// IsMagicLinkNotFound checks if error is magic link not found.
func IsMagicLinkNotFound(err error) bool {
	return errors.GetCode(err) == CodeMagicLinkNotFound
}

// IsMagicLinkExpired checks if error is magic link expired.
func IsMagicLinkExpired(err error) bool {
	return errors.GetCode(err) == CodeMagicLinkExpired
}

// IsMagicLinkAlreadyUsed checks if error is magic link already used.
func IsMagicLinkAlreadyUsed(err error) bool {
	return errors.GetCode(err) == CodeMagicLinkAlreadyUsed
}

// IsMagicLinkInvalid checks if error is magic link invalid.
func IsMagicLinkInvalid(err error) bool {
	return errors.GetCode(err) == CodeMagicLinkInvalid
}

// IsEmailSendFailed checks if error is email send failed.
func IsEmailSendFailed(err error) bool {
	return errors.GetCode(err) == CodeEmailSendFailed
}

// IsEmailRateLimited checks if error is email rate limited.
func IsEmailRateLimited(err error) bool {
	return errors.GetCode(err) == CodeEmailRateLimited
}
