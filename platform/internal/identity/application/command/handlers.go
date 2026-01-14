package command

import (
	"context"
	"fmt"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	"github.com/0xsj/nexus/platform/pkg/id"
)

// ============================================================================
// Request Challenge Handler
// ============================================================================

// RequestChallengeHandler handles RequestChallenge commands.
type RequestChallengeHandler struct {
	challengeService domain.ChallengeService
}

// NewRequestChallengeHandler creates a new RequestChallengeHandler.
func NewRequestChallengeHandler(challengeService domain.ChallengeService) *RequestChallengeHandler {
	return &RequestChallengeHandler{
		challengeService: challengeService,
	}
}

// Handle handles the RequestChallenge command.
func (h *RequestChallengeHandler) Handle(ctx context.Context, cmd *RequestChallenge) (*cqrs.CommandResult, error) {
	challenge, err := h.challengeService.CreateChallenge(ctx, domain.CreateChallengeParams{
		Address: cmd.Address,
		Chain:   cmd.Chain,
		Domain:  cmd.Domain,
		URI:     cmd.URI,
	})
	if err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: challenge.Nonce,
		Data: &RequestChallengeResult{
			Nonce:     challenge.Nonce,
			Message:   challenge.Message,
			Domain:    challenge.Domain,
			URI:       challenge.URI,
			IssuedAt:  challenge.IssuedAt,
			ExpiresAt: challenge.ExpiresAt,
		},
	}, nil
}

// ============================================================================
// Register With Wallet Handler
// ============================================================================

// RegisterWithWalletHandler handles RegisterWithWallet commands.
type RegisterWithWalletHandler struct {
	userRepo          domain.UserRepository
	sessionRepo       domain.SessionRepository
	challengeService  domain.ChallengeService
	tokenService      domain.TokenService
	signatureVerifier domain.SignatureVerifier
	idGenerator       id.Generator
}

// NewRegisterWithWalletHandler creates a new RegisterWithWalletHandler.
func NewRegisterWithWalletHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	challengeService domain.ChallengeService,
	tokenService domain.TokenService,
	signatureVerifier domain.SignatureVerifier,
	idGenerator id.Generator,
) *RegisterWithWalletHandler {
	return &RegisterWithWalletHandler{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		challengeService:  challengeService,
		tokenService:      tokenService,
		signatureVerifier: signatureVerifier,
		idGenerator:       idGenerator,
	}
}

// Handle handles the RegisterWithWallet command.
func (h *RegisterWithWalletHandler) Handle(ctx context.Context, cmd *RegisterWithWallet) (*cqrs.CommandResult, error) {
	const op = "RegisterWithWalletHandler.Handle"

	// Validate challenge
	challenge, err := h.challengeService.ValidateChallenge(ctx, cmd.Nonce)
	if err != nil {
		return nil, err
	}

	// Verify signature
	err = h.signatureVerifier.VerifyWalletSignature(ctx, domain.WalletSignatureParams{
		Address:   cmd.Address,
		Chain:     cmd.Chain,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		Nonce:     cmd.Nonce,
	})
	if err != nil {
		return nil, domain.ErrInvalidSignature(op)
	}

	// Check if user already exists
	exists, err := h.userRepo.ExistsByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrWalletAlreadyLinked(op, cmd.Address)
	}

	// Create user
	userID := h.idGenerator.Generate().String()
	linkedDIDID := h.idGenerator.Generate().String()
	wallet := domain.NewWalletAddress(cmd.Address, cmd.Chain)

	user, err := domain.NewUserFromWallet(userID, linkedDIDID, wallet)
	if err != nil {
		return nil, err
	}

	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// Invalidate challenge
	_ = h.challengeService.InvalidateChallenge(ctx, challenge.Nonce)

	// Create session
	sessionID := h.idGenerator.Generate().String()
	tokenHash, _ := domain.GenerateSessionToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := domain.NewSession(
		sessionID,
		user.ID(),
		domain.AuthMethodWallet,
		tokenHash,
		expiresAt,
		cmd.UserAgent,
		cmd.IPAddress,
	)

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(ctx, domain.RefreshTokenParams{
		UserID:    user.ID(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	// Default token expiration
	accessExpiresAt := time.Now().Add(15 * time.Minute)

	return &cqrs.CommandResult{
		ID: user.ID(),
		Data: &AuthResult{
			UserID:       user.ID(),
			DID:          user.PrimaryDID().String(),
			SessionID:    session.ID(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(15 * 60),
			ExpiresAt:    accessExpiresAt,
		},
	}, nil
}

// ============================================================================
// Authenticate With Wallet Handler
// ============================================================================

// AuthenticateWithWalletHandler handles AuthenticateWithWallet commands.
type AuthenticateWithWalletHandler struct {
	userRepo          domain.UserRepository
	sessionRepo       domain.SessionRepository
	challengeService  domain.ChallengeService
	tokenService      domain.TokenService
	signatureVerifier domain.SignatureVerifier
	idGenerator       id.Generator
}

// NewAuthenticateWithWalletHandler creates a new AuthenticateWithWalletHandler.
func NewAuthenticateWithWalletHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	challengeService domain.ChallengeService,
	tokenService domain.TokenService,
	signatureVerifier domain.SignatureVerifier,
	idGenerator id.Generator,
) *AuthenticateWithWalletHandler {
	return &AuthenticateWithWalletHandler{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		challengeService:  challengeService,
		tokenService:      tokenService,
		signatureVerifier: signatureVerifier,
		idGenerator:       idGenerator,
	}
}

// Handle handles the AuthenticateWithWallet command.
func (h *AuthenticateWithWalletHandler) Handle(ctx context.Context, cmd *AuthenticateWithWallet) (*cqrs.CommandResult, error) {
	const op = "AuthenticateWithWalletHandler.Handle"

	// Validate challenge
	challenge, err := h.challengeService.ValidateChallenge(ctx, cmd.Nonce)
	if err != nil {
		return nil, err
	}

	// Verify signature
	err = h.signatureVerifier.VerifyWalletSignature(ctx, domain.WalletSignatureParams{
		Address:   cmd.Address,
		Chain:     cmd.Chain,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		Nonce:     cmd.Nonce,
	})
	if err != nil {
		return nil, domain.ErrInvalidSignature(op)
	}

	// Find user by wallet
	user, err := h.userRepo.FindByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		return nil, err
	}

	// Check user status
	if !user.CanAuthenticate() {
		return nil, domain.ErrUserDisabled(op, user.ID())
	}

	// Invalidate challenge
	_ = h.challengeService.InvalidateChallenge(ctx, challenge.Nonce)

	// Update last login
	user.RecordLogin(domain.AuthMethodWallet)
	_ = h.userRepo.Save(ctx, user)

	// Create session
	sessionID := h.idGenerator.Generate().String()
	tokenHash, _ := domain.GenerateSessionToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := domain.NewSession(
		sessionID,
		user.ID(),
		domain.AuthMethodWallet,
		tokenHash,
		expiresAt,
		cmd.UserAgent,
		cmd.IPAddress,
	)

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(ctx, domain.RefreshTokenParams{
		UserID:    user.ID(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	accessExpiresAt := time.Now().Add(15 * time.Minute)

	return &cqrs.CommandResult{
		ID: user.ID(),
		Data: &AuthResult{
			UserID:       user.ID(),
			DID:          user.PrimaryDID().String(),
			SessionID:    session.ID(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(15 * 60),
			ExpiresAt:    accessExpiresAt,
		},
	}, nil
}

// ============================================================================
// Refresh Token Handler
// ============================================================================

// RefreshTokenHandler handles RefreshToken commands.
type RefreshTokenHandler struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenService domain.TokenService
}

// NewRefreshTokenHandler creates a new RefreshTokenHandler.
func NewRefreshTokenHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenService domain.TokenService,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenService: tokenService,
	}
}

// Handle handles the RefreshToken command.
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd *RefreshToken) (*cqrs.CommandResult, error) {
	const op = "RefreshTokenHandler.Handle"

	// Validate refresh token
	claims, err := h.tokenService.ValidateRefreshToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Load session
	session, err := h.sessionRepo.FindByID(ctx, claims.SessionID)
	if err != nil {
		return nil, err
	}

	// Check session is active
	if !session.IsActive() {
		return nil, domain.ErrSessionRevoked(op, session.ID())
	}

	// Load user
	user, err := h.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check user status
	if !user.CanAuthenticate() {
		return nil, domain.ErrUserDisabled(op, user.ID())
	}

	// Update session activity
	session.Touch()
	_ = h.sessionRepo.Save(ctx, session)

	// Generate new access token
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	accessExpiresAt := time.Now().Add(15 * time.Minute)

	return &cqrs.CommandResult{
		ID: session.ID(),
		Data: &RefreshTokenResult{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   int64(15 * 60),
			ExpiresAt:   accessExpiresAt,
		},
	}, nil
}

// ============================================================================
// Revoke Session Handler
// ============================================================================

// RevokeSessionHandler handles RevokeSession commands.
type RevokeSessionHandler struct {
	sessionRepo domain.SessionRepository
}

// NewRevokeSessionHandler creates a new RevokeSessionHandler.
func NewRevokeSessionHandler(sessionRepo domain.SessionRepository) *RevokeSessionHandler {
	return &RevokeSessionHandler{
		sessionRepo: sessionRepo,
	}
}

// Handle handles the RevokeSession command.
func (h *RevokeSessionHandler) Handle(ctx context.Context, cmd *RevokeSession) (*cqrs.CommandResult, error) {
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.UserID() != cmd.UserID {
		return nil, domain.ErrSessionNotFound("RevokeSessionHandler.Handle", cmd.SessionID)
	}

	_ = session.Revoke()

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.SessionID,
		Data: &RevokeSessionResult{
			SessionID: cmd.SessionID,
			Revoked:   true,
		},
	}, nil
}

// ============================================================================
// Revoke All Sessions Handler
// ============================================================================

// RevokeAllSessionsHandler handles RevokeAllSessions commands.
type RevokeAllSessionsHandler struct {
	sessionRepo domain.SessionRepository
}

// NewRevokeAllSessionsHandler creates a new RevokeAllSessionsHandler.
func NewRevokeAllSessionsHandler(sessionRepo domain.SessionRepository) *RevokeAllSessionsHandler {
	return &RevokeAllSessionsHandler{
		sessionRepo: sessionRepo,
	}
}

// Handle handles the RevokeAllSessions command.
func (h *RevokeAllSessionsHandler) Handle(ctx context.Context, cmd *RevokeAllSessions) (*cqrs.CommandResult, error) {
	sessions, err := h.sessionRepo.FindActiveByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	var revokedIDs []string
	for _, session := range sessions {
		if cmd.ExceptCurrent && session.ID() == cmd.CurrentSession {
			continue
		}

		_ = session.Revoke()
		if err := h.sessionRepo.Save(ctx, session); err != nil {
			continue
		}
		revokedIDs = append(revokedIDs, session.ID())
	}

	return &cqrs.CommandResult{
		ID: cmd.UserID,
		Data: &RevokeAllSessionsResult{
			RevokedCount: len(revokedIDs),
			SessionIDs:   revokedIDs,
		},
	}, nil
}

// ============================================================================
// Create API Key Handler
// ============================================================================

// CreateAPIKeyHandler handles CreateAPIKey commands.
type CreateAPIKeyHandler struct {
	apiKeyRepo  domain.APIKeyRepository
	idGenerator id.Generator
}

// NewCreateAPIKeyHandler creates a new CreateAPIKeyHandler.
func NewCreateAPIKeyHandler(apiKeyRepo domain.APIKeyRepository, idGenerator id.Generator) *CreateAPIKeyHandler {
	return &CreateAPIKeyHandler{
		apiKeyRepo:  apiKeyRepo,
		idGenerator: idGenerator,
	}
}

// Handle handles the CreateAPIKey command.
func (h *CreateAPIKeyHandler) Handle(ctx context.Context, cmd *CreateAPIKey) (*cqrs.CommandResult, error) {
	keyID := h.idGenerator.Generate().String()

	var expiresAt time.Time
	if cmd.ExpiresIn > 0 {
		expiresAt = time.Now().Add(cmd.ExpiresIn)
	}

	result, err := domain.NewAPIKey(
		keyID,
		cmd.UserID,
		cmd.Name,
		cmd.Scopes,
		expiresAt,
		cmd.Description,
	)
	if err != nil {
		return nil, err
	}

	if err := h.apiKeyRepo.Save(ctx, result.APIKey); err != nil {
		return nil, err
	}

	var expiresAtPtr *time.Time
	if !expiresAt.IsZero() {
		expiresAtPtr = &expiresAt
	}

	return &cqrs.CommandResult{
		ID: result.APIKey.ID(),
		Data: &CreateAPIKeyResult{
			KeyID:     result.APIKey.ID(),
			Name:      result.APIKey.Name(),
			RawKey:    result.RawKey,
			Prefix:    result.APIKey.KeyPrefix(),
			Scopes:    result.APIKey.Scopes().Strings(),
			ExpiresAt: expiresAtPtr,
			CreatedAt: result.APIKey.CreatedAt(),
		},
	}, nil
}

// ============================================================================
// Revoke API Key Handler
// ============================================================================

// RevokeAPIKeyHandler handles RevokeAPIKey commands.
type RevokeAPIKeyHandler struct {
	apiKeyRepo domain.APIKeyRepository
}

// NewRevokeAPIKeyHandler creates a new RevokeAPIKeyHandler.
func NewRevokeAPIKeyHandler(apiKeyRepo domain.APIKeyRepository) *RevokeAPIKeyHandler {
	return &RevokeAPIKeyHandler{
		apiKeyRepo: apiKeyRepo,
	}
}

// Handle handles the RevokeAPIKey command.
func (h *RevokeAPIKeyHandler) Handle(ctx context.Context, cmd *RevokeAPIKey) (*cqrs.CommandResult, error) {
	apiKey, err := h.apiKeyRepo.FindByID(ctx, cmd.KeyID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if apiKey.UserID() != cmd.UserID {
		return nil, domain.ErrAPIKeyNotFound("RevokeAPIKeyHandler.Handle", cmd.KeyID)
	}

	_ = apiKey.Revoke(cmd.UserID, cmd.Reason)

	if err := h.apiKeyRepo.Save(ctx, apiKey); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.KeyID,
		Data: &RevokeAPIKeyResult{
			KeyID:     cmd.KeyID,
			Revoked:   true,
			RevokedAt: apiKey.RevokedAt(),
		},
	}, nil
}

// ============================================================================
// Link Wallet Handler
// ============================================================================

// LinkWalletHandler handles LinkWallet commands.
type LinkWalletHandler struct {
	userRepo          domain.UserRepository
	challengeService  domain.ChallengeService
	signatureVerifier domain.SignatureVerifier
	idGenerator       id.Generator
}

// NewLinkWalletHandler creates a new LinkWalletHandler.
func NewLinkWalletHandler(
	userRepo domain.UserRepository,
	challengeService domain.ChallengeService,
	signatureVerifier domain.SignatureVerifier,
	idGenerator id.Generator,
) *LinkWalletHandler {
	return &LinkWalletHandler{
		userRepo:          userRepo,
		challengeService:  challengeService,
		signatureVerifier: signatureVerifier,
		idGenerator:       idGenerator,
	}
}

// Handle handles the LinkWallet command.
func (h *LinkWalletHandler) Handle(ctx context.Context, cmd *LinkWallet) (*cqrs.CommandResult, error) {
	const op = "LinkWalletHandler.Handle"

	// Validate challenge
	challenge, err := h.challengeService.ValidateChallenge(ctx, cmd.Nonce)
	if err != nil {
		return nil, err
	}

	// Verify signature
	err = h.signatureVerifier.VerifyWalletSignature(ctx, domain.WalletSignatureParams{
		Address:   cmd.Address,
		Chain:     cmd.Chain,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		Nonce:     cmd.Nonce,
	})
	if err != nil {
		return nil, domain.ErrInvalidSignature(op)
	}

	// Check wallet not already linked
	exists, err := h.userRepo.ExistsByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrWalletAlreadyLinked(op, cmd.Address)
	}

	// Load user
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// Generate linkedDIDID for the wallet DID
	linkedDIDID := h.idGenerator.Generate().String()

	// Link wallet (with linkedDIDID to create the wallet DID)
	if err := user.LinkWallet(cmd.Address, cmd.Chain, linkedDIDID); err != nil {
		return nil, err
	}

	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	// Invalidate challenge
	_ = h.challengeService.InvalidateChallenge(ctx, challenge.Nonce)

	wallet := domain.NewWalletAddress(cmd.Address, cmd.Chain)

	return &cqrs.CommandResult{
		ID: cmd.UserID,
		Data: &LinkWalletResult{
			UserID:  cmd.UserID,
			Address: cmd.Address,
			Chain:   cmd.Chain.String(),
			DID:     wallet.ToDID(),
		},
	}, nil
}

// ============================================================================
// Unlink Wallet Handler
// ============================================================================

// UnlinkWalletHandler handles UnlinkWallet commands.
type UnlinkWalletHandler struct {
	userRepo domain.UserRepository
}

// NewUnlinkWalletHandler creates a new UnlinkWalletHandler.
func NewUnlinkWalletHandler(userRepo domain.UserRepository) *UnlinkWalletHandler {
	return &UnlinkWalletHandler{
		userRepo: userRepo,
	}
}

// Handle handles the UnlinkWallet command.
func (h *UnlinkWalletHandler) Handle(ctx context.Context, cmd *UnlinkWallet) (*cqrs.CommandResult, error) {
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if err := user.UnlinkWallet(cmd.Address, cmd.Chain); err != nil {
		return nil, err
	}

	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return &cqrs.CommandResult{
		ID: cmd.UserID,
		Data: &UnlinkWalletResult{
			UserID:  cmd.UserID,
			Address: cmd.Address,
			Chain:   cmd.Chain.String(),
		},
	}, nil
}

// ============================================================================
// Request Magic Link Handler
// ============================================================================

// RequestMagicLinkHandler handles RequestMagicLink commands.
type RequestMagicLinkHandler struct {
	userRepo         domain.UserRepository
	magicLinkService domain.MagicLinkService
	emailService     domain.EmailService
}

// NewRequestMagicLinkHandler creates a new RequestMagicLinkHandler.
func NewRequestMagicLinkHandler(
	userRepo domain.UserRepository,
	magicLinkService domain.MagicLinkService,
	emailService domain.EmailService,
) *RequestMagicLinkHandler {
	return &RequestMagicLinkHandler{
		userRepo:         userRepo,
		magicLinkService: magicLinkService,
		emailService:     emailService,
	}
}

// Handle handles the RequestMagicLink command.
func (h *RequestMagicLinkHandler) Handle(ctx context.Context, cmd *RequestMagicLink) (*cqrs.CommandResult, error) {
	fmt.Printf("[DEBUG] RequestMagicLink: email=%s\n", cmd.Email)

	// Determine purpose based on whether user exists
	purpose := cmd.Purpose
	if purpose == "" {
		// Auto-detect: login if user exists, register if not
		_, err := h.userRepo.FindByEmail(ctx, cmd.Email)
		if err != nil {
			if domain.IsUserNotFound(err) {
				purpose = domain.MagicLinkPurposeRegister
				fmt.Printf("[DEBUG] User not found, purpose=register\n")
			} else {
				fmt.Printf("[DEBUG] FindByEmail error: %v\n", err)
				return &cqrs.CommandResult{
					Data: &RequestMagicLinkResult{
						Success: true,
						Message: "If this email is registered, you will receive a magic link shortly.",
					},
				}, nil
			}
		} else {
			purpose = domain.MagicLinkPurposeLogin
			fmt.Printf("[DEBUG] User found, purpose=login\n")
		}
	}

	// Revoke any existing pending tokens for this email
	_ = h.magicLinkService.RevokeAllForEmail(ctx, cmd.Email)

	// Create magic link token
	token, err := h.magicLinkService.CreateToken(ctx, domain.CreateMagicLinkParams{
		Email:     cmd.Email,
		Purpose:   purpose,
		IPAddress: cmd.IPAddress,
		UserAgent: cmd.UserAgent,
	})
	if err != nil {
		fmt.Printf("[DEBUG] CreateToken error: %v\n", err)
		return &cqrs.CommandResult{
			Data: &RequestMagicLinkResult{
				Success: true,
				Message: "If this email is registered, you will receive a magic link shortly.",
			},
		}, nil
	}

	fmt.Printf("[DEBUG] Token created: %s\n", token.Token())

	// Send magic link email
	err = h.emailService.SendMagicLink(ctx, domain.SendMagicLinkParams{
		To:        cmd.Email,
		Token:     token.Token(),
		Purpose:   purpose,
		ExpiresAt: token.ExpiresAt(),
		IPAddress: cmd.IPAddress,
		UserAgent: cmd.UserAgent,
	})
	if err != nil {
		fmt.Printf("[DEBUG] SendMagicLink error: %v\n", err)
		return &cqrs.CommandResult{
			Data: &RequestMagicLinkResult{
				Success: true,
				Message: "If this email is registered, you will receive a magic link shortly.",
			},
		}, nil
	}

	fmt.Printf("[DEBUG] Email sent successfully\n")

	return &cqrs.CommandResult{
		Data: &RequestMagicLinkResult{
			Success: true,
			Message: "If this email is registered, you will receive a magic link shortly.",
		},
	}, nil
}

// ============================================================================
// Verify Magic Link Handler
// ============================================================================

// VerifyMagicLinkHandler handles VerifyMagicLink commands.
type VerifyMagicLinkHandler struct {
	userRepo             domain.UserRepository
	sessionRepo          domain.SessionRepository
	magicLinkService     domain.MagicLinkService
	tokenService         domain.TokenService
	emailService         domain.EmailService
	didGenerationService domain.DIDGenerationService
	idGenerator          id.Generator
}

// NewVerifyMagicLinkHandler creates a new VerifyMagicLinkHandler.
func NewVerifyMagicLinkHandler(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	magicLinkService domain.MagicLinkService,
	tokenService domain.TokenService,
	emailService domain.EmailService,
	didGenerationService domain.DIDGenerationService,
	idGenerator id.Generator,
) *VerifyMagicLinkHandler {
	return &VerifyMagicLinkHandler{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		magicLinkService:     magicLinkService,
		tokenService:         tokenService,
		emailService:         emailService,
		didGenerationService: didGenerationService,
		idGenerator:          idGenerator,
	}
}

// Handle handles the VerifyMagicLink command.
func (h *VerifyMagicLinkHandler) Handle(ctx context.Context, cmd *VerifyMagicLink) (*cqrs.CommandResult, error) {
	const op = "VerifyMagicLinkHandler.Handle"

	// Validate and consume the magic link token
	magicLink, err := h.magicLinkService.ValidateToken(ctx, cmd.Token)
	if err != nil {
		return nil, err
	}

	var user *domain.User

	switch magicLink.Purpose() {
	case domain.MagicLinkPurposeRegister:
		// Create new user
		user, err = h.createUserFromEmail(ctx, magicLink.Email())
		if err != nil {
			return nil, err
		}

		// Send welcome email
		_ = h.emailService.SendWelcome(ctx, domain.SendWelcomeParams{
			To:       magicLink.Email(),
			Username: magicLink.Email(),
		})

	case domain.MagicLinkPurposeLogin:
		// Find existing user
		user, err = h.userRepo.FindByEmail(ctx, magicLink.Email())
		if err != nil {
			return nil, err
		}

		// Check user status
		if !user.CanAuthenticate() {
			return nil, domain.ErrUserDisabled(op, user.ID())
		}

		// Update last login
		user.RecordLogin(domain.AuthMethodEmail)
		_ = h.userRepo.Save(ctx, user)

	case domain.MagicLinkPurposeVerify:
		// Find existing user and verify email
		user, err = h.userRepo.FindByEmail(ctx, magicLink.Email())
		if err != nil {
			return nil, err
		}

		user.VerifyPrimaryEmail()
		_ = h.userRepo.Save(ctx, user)

	default:
		return nil, domain.ErrMagicLinkInvalid(op, "unknown purpose")
	}

	// Create session
	sessionID := h.idGenerator.Generate().String()
	tokenHash, _ := domain.GenerateSessionToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := domain.NewSession(
		sessionID,
		user.ID(),
		domain.AuthMethodEmail,
		tokenHash,
		expiresAt,
		cmd.UserAgent,
		cmd.IPAddress,
	)

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(ctx, domain.RefreshTokenParams{
		UserID:    user.ID(),
		SessionID: session.ID(),
	})
	if err != nil {
		return nil, err
	}

	accessExpiresAt := time.Now().Add(15 * time.Minute)

	return &cqrs.CommandResult{
		ID: user.ID(),
		Data: &AuthResult{
			UserID:       user.ID(),
			DID:          user.PrimaryDID().String(),
			SessionID:    session.ID(),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(15 * 60),
			ExpiresAt:    accessExpiresAt,
		},
	}, nil
}

// createUserFromEmail creates a new user from an email address.
func (h *VerifyMagicLinkHandler) createUserFromEmail(ctx context.Context, email string) (*domain.User, error) {
	userID := h.idGenerator.Generate().String()
	// Don't generate linkedDIDID here - get it from GenerateCustodialDID

	custodialDID, linkedDIDID, err := h.didGenerationService.GenerateCustodialDID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUserFromEmail(userID, linkedDIDID, email, custodialDID)
	if err != nil {
		return nil, err
	}

	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ============================================================================
// Handler Registration
// ============================================================================

// HandlerDependencies contains all dependencies needed for command handlers.
type HandlerDependencies struct {
	UserRepo             domain.UserRepository
	SessionRepo          domain.SessionRepository
	APIKeyRepo           domain.APIKeyRepository
	ChallengeService     domain.ChallengeService
	TokenService         domain.TokenService
	SignatureVerifier    domain.SignatureVerifier
	MagicLinkService     domain.MagicLinkService
	EmailService         domain.EmailService
	DIDGenerationService domain.DIDGenerationService
	IDGenerator          id.Generator
}

// RegisterHandlers registers all identity command handlers with the command bus.
func RegisterHandlers(bus *cqrs.InMemoryCommandBus, deps HandlerDependencies) error {
	handlers := map[string]any{
		TypeRequestChallenge: NewRequestChallengeHandler(deps.ChallengeService),
		TypeRegisterWithWallet: NewRegisterWithWalletHandler(
			deps.UserRepo,
			deps.SessionRepo,
			deps.ChallengeService,
			deps.TokenService,
			deps.SignatureVerifier,
			deps.IDGenerator,
		),
		TypeAuthenticateWithWallet: NewAuthenticateWithWalletHandler(
			deps.UserRepo,
			deps.SessionRepo,
			deps.ChallengeService,
			deps.TokenService,
			deps.SignatureVerifier,
			deps.IDGenerator,
		),
		TypeRefreshToken: NewRefreshTokenHandler(
			deps.UserRepo,
			deps.SessionRepo,
			deps.TokenService,
		),
		TypeRevokeSession:     NewRevokeSessionHandler(deps.SessionRepo),
		TypeRevokeAllSessions: NewRevokeAllSessionsHandler(deps.SessionRepo),
		TypeCreateAPIKey:      NewCreateAPIKeyHandler(deps.APIKeyRepo, deps.IDGenerator),
		TypeRevokeAPIKey:      NewRevokeAPIKeyHandler(deps.APIKeyRepo),
		TypeLinkWallet: NewLinkWalletHandler(
			deps.UserRepo,
			deps.ChallengeService,
			deps.SignatureVerifier,
			deps.IDGenerator,
		),
		TypeUnlinkWallet: NewUnlinkWalletHandler(deps.UserRepo),
		TypeRequestMagicLink: NewRequestMagicLinkHandler(
			deps.UserRepo,
			deps.MagicLinkService,
			deps.EmailService,
		),
		TypeVerifyMagicLink: NewVerifyMagicLinkHandler(
			deps.UserRepo,
			deps.SessionRepo,
			deps.MagicLinkService,
			deps.TokenService,
			deps.EmailService,
			deps.DIDGenerationService,
			deps.IDGenerator,
		),
	}

	for cmdType, handler := range handlers {
		if err := bus.Register(cmdType, handler); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var (
	_ cqrs.CommandHandler[*RequestChallenge]       = (*RequestChallengeHandler)(nil)
	_ cqrs.CommandHandler[*RegisterWithWallet]     = (*RegisterWithWalletHandler)(nil)
	_ cqrs.CommandHandler[*AuthenticateWithWallet] = (*AuthenticateWithWalletHandler)(nil)
	_ cqrs.CommandHandler[*RefreshToken]           = (*RefreshTokenHandler)(nil)
	_ cqrs.CommandHandler[*RevokeSession]          = (*RevokeSessionHandler)(nil)
	_ cqrs.CommandHandler[*RevokeAllSessions]      = (*RevokeAllSessionsHandler)(nil)
	_ cqrs.CommandHandler[*CreateAPIKey]           = (*CreateAPIKeyHandler)(nil)
	_ cqrs.CommandHandler[*RevokeAPIKey]           = (*RevokeAPIKeyHandler)(nil)
	_ cqrs.CommandHandler[*LinkWallet]             = (*LinkWalletHandler)(nil)
	_ cqrs.CommandHandler[*UnlinkWallet]           = (*UnlinkWalletHandler)(nil)
	_ cqrs.CommandHandler[*RequestMagicLink]       = (*RequestMagicLinkHandler)(nil)
	_ cqrs.CommandHandler[*VerifyMagicLink]        = (*VerifyMagicLinkHandler)(nil)
)
