package stubs

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Compile-time interface checks
// ============================================================================

var (
	_ domain.DIDGenerator = (*NullDIDGenerator)(nil)
	_ domain.WalletReader = (*NullWalletReader)(nil)
	_ domain.OAuthService = (*NullOAuthService)(nil)
	_ domain.EmailService = (*NullEmailService)(nil)
)

// ============================================================================
// NullDIDGenerator
// ============================================================================

// NullDIDGenerator generates placeholder DIDs for development.
type NullDIDGenerator struct{}

// NewNullDIDGenerator creates a new NullDIDGenerator.
func NewNullDIDGenerator() *NullDIDGenerator {
	return &NullDIDGenerator{}
}

// GenerateDIDKey returns a placeholder did:key using a new UUID.
func (g *NullDIDGenerator) GenerateDIDKey(_ context.Context) (string, error) {
	id := types.NewID()
	return fmt.Sprintf("did:key:%s", id.String()), nil
}

// DeriveDIDPKH returns a placeholder did:pkh from address and chain ID.
func (g *NullDIDGenerator) DeriveDIDPKH(_ context.Context, address string, chainID string) (string, error) {
	return fmt.Sprintf("did:pkh:eip155:%s:%s", chainID, address), nil
}

// ============================================================================
// NullWalletReader
// ============================================================================

// NullWalletReader returns errors indicating the wallet context is not available.
type NullWalletReader struct{}

// NewNullWalletReader creates a new NullWalletReader.
func NewNullWalletReader() *NullWalletReader {
	return &NullWalletReader{}
}

// VerifySignature returns an error indicating wallet context is not implemented.
func (r *NullWalletReader) VerifySignature(_ context.Context, _ domain.WalletVerificationRequest) (*domain.WalletVerificationResult, error) {
	return nil, fmt.Errorf("wallet context not yet implemented")
}

// GetDIDForWallet returns an error indicating wallet context is not implemented.
func (r *NullWalletReader) GetDIDForWallet(_ context.Context, _ domain.WalletAddress) (string, error) {
	return "", fmt.Errorf("wallet context not yet implemented")
}

// ============================================================================
// NullOAuthService
// ============================================================================

// NullOAuthService returns errors indicating OAuth is not configured.
type NullOAuthService struct{}

// NewNullOAuthService creates a new NullOAuthService.
func NewNullOAuthService() *NullOAuthService {
	return &NullOAuthService{}
}

// GetAuthorizationURL returns an error indicating OAuth is not configured.
func (s *NullOAuthService) GetAuthorizationURL(_ context.Context, provider string, _ string, _ string) (string, error) {
	return "", fmt.Errorf("oauth not configured for provider: %s", provider)
}

// ExchangeCode returns an error indicating OAuth is not configured.
func (s *NullOAuthService) ExchangeCode(_ context.Context, provider string, _ string, _ string) (*domain.OAuthProfile, error) {
	return nil, fmt.Errorf("oauth not configured for provider: %s", provider)
}

// ============================================================================
// NullEmailService
// ============================================================================

// NullEmailService logs email messages instead of sending them.
type NullEmailService struct {
	logger log.Logger
}

// NewNullEmailService creates a new NullEmailService.
func NewNullEmailService(logger log.Logger) *NullEmailService {
	return &NullEmailService{logger: logger}
}

// SendMagicLink logs the magic link instead of sending an email.
func (s *NullEmailService) SendMagicLink(_ context.Context, to types.Email, token string, expiresIn int) error {
	s.logger.Info("stub: magic link email",
		log.String("to", to.String()),
		log.String("token", token),
		log.Int("expires_in", expiresIn),
	)
	return nil
}

// SendWelcome logs the welcome message instead of sending an email.
func (s *NullEmailService) SendWelcome(_ context.Context, to types.Email, displayName string) error {
	s.logger.Info("stub: welcome email",
		log.String("to", to.String()),
		log.String("display_name", displayName),
	)
	return nil
}
