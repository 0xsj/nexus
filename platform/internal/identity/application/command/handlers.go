package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Command Handlers
// ============================================================================

// Handlers handles all identity commands.
type Handlers struct {
	// Repositories
	users       domain.UserRepository
	sessions    domain.SessionRepository
	connections domain.ConnectionRepository
	apiKeys     domain.APIKeyRepository
	challenges  domain.ChallengeRepository
	tokens      domain.TokenRepository

	// Domain services
	signatureVerifier domain.SignatureVerifier
	tokenService      domain.TokenService
	challengeService  domain.ChallengeService
	didService        domain.DIDService
	oauthService      domain.OAuthService
	rateLimiter       domain.RateLimiter
	auditLogger       domain.AuditLogger

	// Optional services
	notificationService domain.IdentityNotificationService

	// Config
	config HandlersConfig

	// Logger
	logger log.Logger
}

// HandlersConfig contains configuration for command handlers.
type HandlersConfig struct {
	// Token expiration
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	// Session limits
	MaxSessionsPerUser int

	// Challenge
	ChallengeTTL    time.Duration
	ChallengeDomain string
	ChallengeURI    string

	// API Keys
	MaxAPIKeysPerUser int
	DefaultAPIKeyTTL  time.Duration
}

// DefaultHandlersConfig returns default configuration.
func DefaultHandlersConfig() HandlersConfig {
	return HandlersConfig{
		AccessTokenTTL:     15 * time.Minute,
		RefreshTokenTTL:    7 * 24 * time.Hour,
		MaxSessionsPerUser: 10,
		ChallengeTTL:       5 * time.Minute,
		ChallengeDomain:    "proof.io",
		ChallengeURI:       "https://proof.io",
		MaxAPIKeysPerUser:  10,
		DefaultAPIKeyTTL:   90 * 24 * time.Hour, // 90 days
	}
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	connections domain.ConnectionRepository,
	apiKeys domain.APIKeyRepository,
	challenges domain.ChallengeRepository,
	tokens domain.TokenRepository,
	signatureVerifier domain.SignatureVerifier,
	tokenService domain.TokenService,
	challengeService domain.ChallengeService,
	didService domain.DIDService,
	oauthService domain.OAuthService,
	rateLimiter domain.RateLimiter,
	auditLogger domain.AuditLogger,
	logger log.Logger,
	config HandlersConfig,
) *Handlers {
	return &Handlers{
		users:             users,
		sessions:          sessions,
		connections:       connections,
		apiKeys:           apiKeys,
		challenges:        challenges,
		tokens:            tokens,
		signatureVerifier: signatureVerifier,
		tokenService:      tokenService,
		challengeService:  challengeService,
		didService:        didService,
		oauthService:      oauthService,
		rateLimiter:       rateLimiter,
		auditLogger:       auditLogger,
		logger:            logger,
		config:            config,
	}
}

// WithNotificationService sets the notification service.
func (h *Handlers) WithNotificationService(svc domain.IdentityNotificationService) *Handlers {
	h.notificationService = svc
	return h
}

// ============================================================================
// Challenge
// ============================================================================

// RequestChallenge creates a new authentication challenge.
func (h *Handlers) RequestChallenge(ctx context.Context, cmd RequestChallenge) (*ChallengeResult, error) {
	const op = "command.Handlers.RequestChallenge"

	// Rate limit
	if h.rateLimiter != nil {
		key := "challenge:" + cmd.Address
		allowed, err := h.rateLimiter.Allow(ctx, key, domain.RateLimitChallenge)
		if err != nil {
			h.logger.Error("rate limiter error", log.Err(err), log.String("op", op))
		}
		if !allowed {
			return nil, errors.RateLimit(op, "too many challenge requests")
		}
	}

	// Create challenge
	challenge, err := h.challengeService.CreateChallenge(ctx, domain.CreateChallengeParams{
		Address: cmd.Address,
		Chain:   cmd.Chain,
		Domain:  h.config.ChallengeDomain,
		URI:     h.config.ChallengeURI,
		TTL:     h.config.ChallengeTTL,
	})
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &ChallengeResult{
		Nonce:     challenge.Nonce,
		Message:   challenge.SIWE(),
		Domain:    challenge.Domain,
		URI:       challenge.URI,
		IssuedAt:  challenge.IssuedAt,
		ExpiresAt: challenge.ExpiresAt,
	}, nil
}

// ============================================================================
// Registration
// ============================================================================

// RegisterWithWallet registers a new user with a wallet signature.
func (h *Handlers) RegisterWithWallet(ctx context.Context, cmd RegisterWithWallet) (*RegisterResult, error) {
	const op = "command.Handlers.RegisterWithWallet"

	// Rate limit
	if h.rateLimiter != nil {
		key := "register:" + cmd.Address
		allowed, err := h.rateLimiter.Allow(ctx, key, domain.RateLimitRegister)
		if err != nil {
			h.logger.Error("rate limiter error", log.Err(err), log.String("op", op))
		}
		if !allowed {
			return nil, errors.RateLimit(op, "too many registration attempts")
		}
	}

	// Log attempt
	if h.auditLogger != nil {
		h.auditLogger.LogAuthAttempt(ctx, domain.AuthAttemptEvent{
			Timestamp: time.Now(),
			Method:    domain.AuthMethodWallet,
			Identity:  cmd.Address,
			IPAddress: cmd.IPAddress,
			UserAgent: cmd.UserAgent,
		})
	}

	// Validate challenge
	challenge, err := h.challengeService.ValidateChallenge(ctx, cmd.Nonce)
	if err != nil {
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "invalid challenge")
		return nil, errors.Wrap(err, op)
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
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "invalid signature")
		return nil, domain.ErrInvalidCredentials(op)
	}

	// Check if wallet already registered
	exists, err := h.users.ExistsByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	if exists {
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "wallet already registered")
		return nil, domain.ErrUserAlreadyExists(op, cmd.Address)
	}

	// Create wallet address
	wallet := domain.NewWalletAddress(cmd.Address, cmd.Chain)

	// Generate user ID
	userID, err := generateID()
	if err != nil {
		return nil, errors.Internal(op, err)
	}

	// Create user
	user, err := domain.NewUserFromWallet(userID, wallet)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save user
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Create session and tokens
	authResult, err := h.createSession(ctx, user, domain.AuthMethodWallet, cmd.UserAgent, cmd.IPAddress)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log success
	if h.auditLogger != nil {
		h.auditLogger.LogAuthSuccess(ctx, domain.AuthSuccessEvent{
			AuthAttemptEvent: domain.AuthAttemptEvent{
				Timestamp: time.Now(),
				Method:    domain.AuthMethodWallet,
				Identity:  cmd.Address,
				IPAddress: cmd.IPAddress,
				UserAgent: cmd.UserAgent,
			},
			UserID:    user.ID(),
			SessionID: authResult.SessionID,
		})
	}

	// Invalidate challenge
	_ = h.challengeService.InvalidateChallenge(ctx, challenge.Nonce)

	h.logger.Info("user registered with wallet",
		log.String("user_id", user.ID()),
		log.String("did", user.PrimaryDID().String()),
		log.String("chain", cmd.Chain.String()),
	)

	return &RegisterResult{
		UserID:       user.ID(),
		DID:          user.PrimaryDID().String(),
		SessionID:    authResult.SessionID,
		AccessToken:  authResult.AccessToken,
		RefreshToken: authResult.RefreshToken,
		TokenType:    authResult.TokenType,
		ExpiresIn:    authResult.ExpiresIn,
		ExpiresAt:    authResult.ExpiresAt,
	}, nil
}

// ============================================================================
// Authentication
// ============================================================================

// AuthenticateWithWallet authenticates a user with a wallet signature.
func (h *Handlers) AuthenticateWithWallet(ctx context.Context, cmd AuthenticateWithWallet) (*AuthResult, error) {
	const op = "command.Handlers.AuthenticateWithWallet"

	// Rate limit
	if h.rateLimiter != nil {
		key := "login:" + cmd.Address
		allowed, err := h.rateLimiter.Allow(ctx, key, domain.RateLimitLogin)
		if err != nil {
			h.logger.Error("rate limiter error", log.Err(err), log.String("op", op))
		}
		if !allowed {
			return nil, errors.RateLimit(op, "too many login attempts")
		}
	}

	// Log attempt
	if h.auditLogger != nil {
		h.auditLogger.LogAuthAttempt(ctx, domain.AuthAttemptEvent{
			Timestamp: time.Now(),
			Method:    domain.AuthMethodWallet,
			Identity:  cmd.Address,
			IPAddress: cmd.IPAddress,
			UserAgent: cmd.UserAgent,
		})
	}

	// Validate challenge
	_, err := h.challengeService.ValidateChallenge(ctx, cmd.Nonce)
	if err != nil {
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "invalid challenge")
		return nil, errors.Wrap(err, op)
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
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "invalid signature")
		return nil, domain.ErrInvalidCredentials(op)
	}

	// Find user by wallet
	user, err := h.users.FindByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		if domain.IsUserNotFound(err) {
			h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "user not found")
		}
		return nil, errors.Wrap(err, op)
	}

	// Check if user can authenticate
	if !user.CanAuthenticate() {
		h.logAuthFailure(ctx, domain.AuthMethodWallet, cmd.Address, cmd.IPAddress, cmd.UserAgent, "user suspended")
		return nil, domain.ErrUserDisabled(op, user.ID())
	}

	// Record login
	user.RecordLogin(domain.AuthMethodWallet)
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Check session limit and cleanup if needed
	if err := h.enforceSessionLimit(ctx, user.ID()); err != nil {
		h.logger.Warn("failed to enforce session limit", log.Err(err))
	}

	// Create session and tokens
	authResult, err := h.createSession(ctx, user, domain.AuthMethodWallet, cmd.UserAgent, cmd.IPAddress)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log success
	if h.auditLogger != nil {
		h.auditLogger.LogAuthSuccess(ctx, domain.AuthSuccessEvent{
			AuthAttemptEvent: domain.AuthAttemptEvent{
				Timestamp: time.Now(),
				Method:    domain.AuthMethodWallet,
				Identity:  cmd.Address,
				IPAddress: cmd.IPAddress,
				UserAgent: cmd.UserAgent,
			},
			UserID:    user.ID(),
			SessionID: authResult.SessionID,
		})
	}

	// Invalidate challenge
	_ = h.challengeService.InvalidateChallenge(ctx, cmd.Nonce)

	h.logger.Info("user authenticated with wallet",
		log.String("user_id", user.ID()),
		log.String("chain", cmd.Chain.String()),
	)

	return authResult, nil
}

// RefreshToken refreshes an access token.
func (h *Handlers) RefreshToken(ctx context.Context, cmd RefreshToken) (*RefreshResult, error) {
	const op = "command.Handlers.RefreshToken"

	// Validate refresh token
	claims, err := h.tokenService.ValidateRefreshToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Find session
	session, err := h.sessions.FindByID(ctx, claims.SessionID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Validate session
	if err := session.Validate(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Find user
	user, err := h.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Check if user can authenticate
	if !user.CanAuthenticate() {
		return nil, domain.ErrUserDisabled(op, user.ID())
	}

	// Touch session
	session.Touch()
	if err := h.sessions.Save(ctx, session); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Generate new access token
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: session.ID(),
		ExpiresIn: h.config.AccessTokenTTL,
	})
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	expiresAt := time.Now().Add(h.config.AccessTokenTTL)

	return &RefreshResult{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.config.AccessTokenTTL.Seconds()),
		ExpiresAt:   expiresAt,
	}, nil
}

// AuthenticateWithAPIKey authenticates a request with an API key.
func (h *Handlers) AuthenticateWithAPIKey(ctx context.Context, cmd AuthenticateWithAPIKey) (*APIKeyAuthResult, error) {
	const op = "command.Handlers.AuthenticateWithAPIKey"

	// Hash the key
	keyHash := domain.HashAPIKey(cmd.RawKey)

	// Find API key
	apiKey, err := h.apiKeys.FindByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Validate key
	if err := apiKey.ValidateKey(cmd.RawKey); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Find user
	user, err := h.users.FindByID(ctx, apiKey.UserID())
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Check if user can authenticate
	if !user.CanAuthenticate() {
		return nil, domain.ErrUserDisabled(op, user.ID())
	}

	// Record usage
	apiKey.RecordUsage()
	if err := h.apiKeys.Save(ctx, apiKey); err != nil {
		h.logger.Warn("failed to record API key usage", log.Err(err))
	}

	return &APIKeyAuthResult{
		UserID:  user.ID(),
		DID:     user.PrimaryDID().String(),
		KeyID:   apiKey.ID(),
		KeyName: apiKey.Name(),
		Scopes:  apiKey.Scopes().Strings(),
	}, nil
}

// ============================================================================
// Session Management
// ============================================================================

// RevokeSession revokes a single session.
func (h *Handlers) RevokeSession(ctx context.Context, cmd RevokeSession) (*RevokeSessionResult, error) {
	const op = "command.Handlers.RevokeSession"

	// Find session
	session, err := h.sessions.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Verify ownership
	if session.UserID() != cmd.UserID {
		return nil, domain.ErrSessionNotFound(op, cmd.SessionID)
	}

	// Revoke session
	if err := session.Revoke(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save session
	if err := h.sessions.Save(ctx, session); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventSessionRevoked,
			Details:   map[string]any{"session_id": cmd.SessionID, "reason": cmd.Reason},
		})
	}

	h.logger.Info("session revoked",
		log.String("user_id", cmd.UserID),
		log.String("session_id", cmd.SessionID),
	)

	return &RevokeSessionResult{
		SessionID: cmd.SessionID,
		Revoked:   true,
	}, nil
}

// RevokeAllSessions revokes all sessions for a user.
func (h *Handlers) RevokeAllSessions(ctx context.Context, cmd RevokeAllSessions) (*RevokeAllSessionsResult, error) {
	const op = "command.Handlers.RevokeAllSessions"

	// Find all sessions
	sessions, err := h.sessions.FindActiveByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	var revokedIDs []string
	for _, session := range sessions {
		// Skip current session if requested
		if cmd.ExceptCurrent && session.ID() == cmd.CurrentSession {
			continue
		}

		if err := session.Revoke(); err != nil {
			continue
		}

		if err := h.sessions.Save(ctx, session); err != nil {
			h.logger.Warn("failed to save revoked session", log.Err(err), log.String("session_id", session.ID()))
			continue
		}

		revokedIDs = append(revokedIDs, session.ID())
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventAllSessionsRevoked,
			Details:   map[string]any{"revoked_count": len(revokedIDs), "reason": cmd.Reason},
		})
	}

	h.logger.Info("all sessions revoked",
		log.String("user_id", cmd.UserID),
		log.Int("revoked_count", len(revokedIDs)),
	)

	return &RevokeAllSessionsResult{
		RevokedCount: len(revokedIDs),
		SessionIDs:   revokedIDs,
	}, nil
}

// ============================================================================
// API Key Management
// ============================================================================

// CreateAPIKey creates a new API key.
func (h *Handlers) CreateAPIKey(ctx context.Context, cmd CreateAPIKey) (*CreateAPIKeyResult, error) {
	const op = "command.Handlers.CreateAPIKey"

	// Check limit
	count, err := h.apiKeys.CountActiveByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	if count >= h.config.MaxAPIKeysPerUser {
		return nil, errors.Validation(op, "maximum API keys limit reached")
	}

	// Generate ID
	keyID, err := generateID()
	if err != nil {
		return nil, errors.Internal(op, err)
	}

	// Determine expiration
	var expiresAt time.Time
	if cmd.ExpiresIn > 0 {
		expiresAt = time.Now().Add(cmd.ExpiresIn)
	} else if h.config.DefaultAPIKeyTTL > 0 {
		expiresAt = time.Now().Add(h.config.DefaultAPIKeyTTL)
	}

	// Create API key
	result, err := domain.NewAPIKey(keyID, cmd.UserID, cmd.Name, cmd.Scopes, expiresAt, cmd.Description)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save API key
	if err := h.apiKeys.Save(ctx, result.APIKey); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventAPIKeyCreated,
			Details:   map[string]any{"key_id": keyID, "name": cmd.Name, "scopes": cmd.Scopes},
		})
	}

	// Send notification
	if h.notificationService != nil {
		_ = h.notificationService.SendAPIKeyCreated(ctx, cmd.UserID, cmd.Name)
	}

	h.logger.Info("API key created",
		log.String("user_id", cmd.UserID),
		log.String("key_id", keyID),
		log.String("name", cmd.Name),
	)

	var expiresAtPtr *time.Time
	if !expiresAt.IsZero() {
		expiresAtPtr = &expiresAt
	}

	return &CreateAPIKeyResult{
		KeyID:     keyID,
		RawKey:    result.RawKey,
		Name:      cmd.Name,
		Prefix:    result.APIKey.KeyPrefix(),
		Scopes:    result.APIKey.Scopes().Strings(),
		ExpiresAt: expiresAtPtr,
		CreatedAt: result.APIKey.CreatedAt(),
	}, nil
}

// RevokeAPIKey revokes an API key.
func (h *Handlers) RevokeAPIKey(ctx context.Context, cmd RevokeAPIKey) (*RevokeAPIKeyResult, error) {
	const op = "command.Handlers.RevokeAPIKey"

	// Find API key
	apiKey, err := h.apiKeys.FindByID(ctx, cmd.KeyID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Verify ownership
	if apiKey.UserID() != cmd.UserID {
		return nil, domain.ErrAPIKeyNotFound(op, cmd.KeyID)
	}

	// Revoke
	if err := apiKey.Revoke(cmd.UserID, cmd.Reason); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save
	if err := h.apiKeys.Save(ctx, apiKey); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventAPIKeyRevoked,
			Details:   map[string]any{"key_id": cmd.KeyID, "reason": cmd.Reason},
		})
	}

	h.logger.Info("API key revoked",
		log.String("user_id", cmd.UserID),
		log.String("key_id", cmd.KeyID),
	)

	return &RevokeAPIKeyResult{
		KeyID:     cmd.KeyID,
		Revoked:   true,
		RevokedAt: apiKey.RevokedAt(),
	}, nil
}

// ============================================================================
// Wallet Management
// ============================================================================

// LinkWallet links a wallet to an existing user.
func (h *Handlers) LinkWallet(ctx context.Context, cmd LinkWallet) (*LinkWalletResult, error) {
	const op = "command.Handlers.LinkWallet"

	// Verify signature
	err := h.signatureVerifier.VerifyWalletSignature(ctx, domain.WalletSignatureParams{
		Address:   cmd.Address,
		Chain:     cmd.Chain,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		Nonce:     cmd.Nonce,
	})
	if err != nil {
		return nil, domain.ErrInvalidCredentials(op)
	}

	// Check if wallet already linked to another user
	exists, err := h.users.ExistsByWallet(ctx, cmd.Address, cmd.Chain)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	if exists {
		return nil, domain.ErrConnectionAlreadyExists(op, cmd.Chain.String(), cmd.UserID)
	}

	// Find user
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Link wallet
	if err := user.LinkWallet(cmd.Address, cmd.Chain); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save user
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventWalletLinked,
			Details:   map[string]any{"address": cmd.Address, "chain": cmd.Chain.String()},
		})
	}

	wallet := domain.NewWalletAddress(cmd.Address, cmd.Chain)

	h.logger.Info("wallet linked",
		log.String("user_id", cmd.UserID),
		log.String("address", cmd.Address),
		log.String("chain", cmd.Chain.String()),
	)

	return &LinkWalletResult{
		UserID:  cmd.UserID,
		Address: cmd.Address,
		Chain:   cmd.Chain.String(),
		DID:     wallet.ToDID(),
	}, nil
}

// UnlinkWallet removes a wallet from a user.
func (h *Handlers) UnlinkWallet(ctx context.Context, cmd UnlinkWallet) (*UnlinkWalletResult, error) {
	const op = "command.Handlers.UnlinkWallet"

	// Find user
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Unlink wallet
	if err := user.UnlinkWallet(cmd.Address, cmd.Chain); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save user
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventWalletUnlinked,
			Details:   map[string]any{"address": cmd.Address, "chain": cmd.Chain.String()},
		})
	}

	h.logger.Info("wallet unlinked",
		log.String("user_id", cmd.UserID),
		log.String("address", cmd.Address),
		log.String("chain", cmd.Chain.String()),
	)

	return &UnlinkWalletResult{
		UserID:  cmd.UserID,
		Address: cmd.Address,
		Chain:   cmd.Chain.String(),
	}, nil
}

// ============================================================================
// User Management
// ============================================================================

// SuspendUser suspends a user account.
func (h *Handlers) SuspendUser(ctx context.Context, cmd SuspendUser) (*SuspendUserResult, error) {
	const op = "command.Handlers.SuspendUser"

	// Find user
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Suspend user
	if err := user.Suspend(cmd.Reason); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save user
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Revoke all sessions
	_, _ = h.RevokeAllSessions(ctx, RevokeAllSessions{
		UserID: cmd.UserID,
		Reason: "user suspended: " + cmd.Reason,
	})

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventUserSuspended,
			Details:   map[string]any{"reason": cmd.Reason, "suspended_by": cmd.SuspendedBy},
		})
	}

	h.logger.Info("user suspended",
		log.String("user_id", cmd.UserID),
		log.String("reason", cmd.Reason),
	)

	return &SuspendUserResult{
		UserID:      cmd.UserID,
		Suspended:   true,
		SuspendedAt: time.Now(),
	}, nil
}

// ActivateUser reactivates a suspended user.
func (h *Handlers) ActivateUser(ctx context.Context, cmd ActivateUser) (*ActivateUserResult, error) {
	const op = "command.Handlers.ActivateUser"

	// Find user
	user, err := h.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Activate user
	if err := user.Activate(); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Save user
	if err := h.users.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Log security event
	if h.auditLogger != nil {
		h.auditLogger.LogSecurityEvent(ctx, domain.SecurityEvent{
			Timestamp: time.Now(),
			UserID:    cmd.UserID,
			EventType: domain.SecurityEventUserActivated,
			Details:   map[string]any{"activated_by": cmd.ActivatedBy},
		})
	}

	h.logger.Info("user activated",
		log.String("user_id", cmd.UserID),
	)

	return &ActivateUserResult{
		UserID:      cmd.UserID,
		Activated:   true,
		ActivatedAt: time.Now(),
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// createSession creates a new session and generates tokens.
func (h *Handlers) createSession(
	ctx context.Context,
	user *domain.User,
	method domain.AuthMethod,
	userAgent string,
	ipAddress string,
) (*AuthResult, error) {
	const op = "command.Handlers.createSession"

	// Generate session ID
	sessionID, err := domain.GenerateSessionID()
	if err != nil {
		return nil, errors.Internal(op, err)
	}

	// Generate session token
	sessionToken, err := domain.GenerateSessionToken()
	if err != nil {
		return nil, errors.Internal(op, err)
	}

	// Hash session token
	tokenHash := domain.HashAPIKey(sessionToken) // Reuse hash function

	// Create session
	expiresAt := time.Now().Add(h.config.RefreshTokenTTL)
	session := domain.NewSession(sessionID, user.ID(), method, tokenHash, expiresAt, userAgent, ipAddress)

	// Save session
	if err := h.sessions.Save(ctx, session); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Generate access token
	accessToken, err := h.tokenService.GenerateAccessToken(ctx, domain.AccessTokenParams{
		UserID:    user.ID(),
		DID:       user.PrimaryDID().String(),
		SessionID: sessionID,
		ExpiresIn: h.config.AccessTokenTTL,
	})
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Generate refresh token
	refreshToken, err := h.tokenService.GenerateRefreshToken(ctx, domain.RefreshTokenParams{
		UserID:    user.ID(),
		SessionID: sessionID,
		ExpiresIn: h.config.RefreshTokenTTL,
	})
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	return &AuthResult{
		UserID:       user.ID(),
		DID:          user.PrimaryDID().String(),
		SessionID:    sessionID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.config.AccessTokenTTL.Seconds()),
		ExpiresAt:    time.Now().Add(h.config.AccessTokenTTL),
	}, nil
}

// enforceSessionLimit removes oldest sessions if limit is exceeded.
func (h *Handlers) enforceSessionLimit(ctx context.Context, userID string) error {
	count, err := h.sessions.CountActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if count < h.config.MaxSessionsPerUser {
		return nil
	}

	// Get all sessions and revoke oldest
	sessions, err := h.sessions.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Sessions should be ordered by creation time
	// Revoke excess sessions (oldest first)
	excess := count - h.config.MaxSessionsPerUser + 1
	for i := 0; i < excess && i < len(sessions); i++ {
		_ = sessions[i].Revoke()
		_ = h.sessions.Save(ctx, sessions[i])
	}

	return nil
}

// logAuthFailure logs an authentication failure.
func (h *Handlers) logAuthFailure(ctx context.Context, method domain.AuthMethod, identity, ipAddress, userAgent, reason string) {
	if h.auditLogger != nil {
		h.auditLogger.LogAuthFailure(ctx, domain.AuthFailureEvent{
			AuthAttemptEvent: domain.AuthAttemptEvent{
				Timestamp: time.Now(),
				Method:    method,
				Identity:  identity,
				IPAddress: ipAddress,
				UserAgent: userAgent,
			},
			Reason: reason,
		})
	}
}

// generateID generates a unique ID.
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := randomRead(bytes); err != nil {
		return "", err
	}
	return hexEncode(bytes), nil
}

// randomRead reads random bytes (abstracted for testing).
var randomRead = func(b []byte) (int, error) {
	return len(b), nil // Will be replaced with crypto/rand.Read
}

// hexEncode encodes bytes to hex string.
func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	result := make([]byte, len(b)*2)
	for i, v := range b {
		result[i*2] = hex[v>>4]
		result[i*2+1] = hex[v&0x0f]
	}
	return string(result)
}
