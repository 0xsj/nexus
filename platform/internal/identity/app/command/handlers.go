package command

import (
	"context"
	"time"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Identity context.
type Handlers struct {
	userRepo       domain.UserRepository
	sessionRepo    domain.SessionRepository
	magicLinkRepo  domain.MagicLinkRepository
	oauthStateRepo domain.OAuthStateRepository
	userLookup     domain.UserLookup
	sessionLookup  domain.SessionLookup
	didGenerator   domain.DIDGenerator
	walletReader   domain.WalletReader
	oauthService   domain.OAuthService
	emailService   domain.EmailService
	publisher      domain.EventPublisher
	logger         log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	magicLinkRepo domain.MagicLinkRepository,
	oauthStateRepo domain.OAuthStateRepository,
	userLookup domain.UserLookup,
	sessionLookup domain.SessionLookup,
	didGenerator domain.DIDGenerator,
	walletReader domain.WalletReader,
	oauthService domain.OAuthService,
	emailService domain.EmailService,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		magicLinkRepo:  magicLinkRepo,
		oauthStateRepo: oauthStateRepo,
		userLookup:     userLookup,
		sessionLookup:  sessionLookup,
		didGenerator:   didGenerator,
		walletReader:   walletReader,
		oauthService:   oauthService,
		emailService:   emailService,
		publisher:      publisher,
		logger:         logger,
	}
}

// ============================================================================
// Registration Handlers
// ============================================================================

// HandleRegisterWithEmail handles the RegisterWithEmail command.
func (h *Handlers) HandleRegisterWithEmail(ctx context.Context, cmd RegisterWithEmail) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRegisterWithEmail"

	// 1. Check if email already exists
	exists, err := h.userLookup.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.EmailAlreadyRegistered(op, cmd.Email.String())
	}

	// 2. Generate DID for custodial user
	did, err := h.didGenerator.GenerateDIDKey(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Create display name value object
	displayName, err := domain.NewDisplayName(cmd.DisplayName)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Create user aggregate
	userID := domain.NewUserID()
	user, err := domain.NewUser(
		userID,
		cmd.Email,
		displayName,
		domain.AuthMethodMagicLink,
		did,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	// 7. Send welcome email (fire-and-forget)
	go func() {
		if err := h.emailService.SendWelcome(context.Background(), cmd.Email, cmd.DisplayName); err != nil {
			h.logger.Error("failed to send welcome email",
				log.String("op", op),
				log.Err(err),
			)
		}
	}()

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: RegisterWithEmailResult{
			UserID:    userID.String(),
			Email:     cmd.Email.String(),
			CreatedAt: user.CreatedAt(),
		},
	}, nil
}

// HandleRegisterWithWallet handles the RegisterWithWallet command.
func (h *Handlers) HandleRegisterWithWallet(ctx context.Context, cmd RegisterWithWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRegisterWithWallet"

	// 1. Verify wallet signature
	verifyResult, err := h.walletReader.VerifySignature(ctx, domain.WalletVerificationRequest{
		Address:   cmd.Address,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		ChainID:   cmd.ChainID,
	})
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if !verifyResult.Valid {
		return nil, domain.InvalidTokenError(op, "invalid wallet signature")
	}

	// 2. Check if DID already exists
	exists, err := h.userLookup.ExistsByDID(ctx, verifyResult.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.DIDAlreadyLinked(op, verifyResult.DID)
	}

	// 3. Create display name (use address if not provided)
	var displayName domain.DisplayName
	if cmd.DisplayName != nil && *cmd.DisplayName != "" {
		displayName, err = domain.NewDisplayName(*cmd.DisplayName)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	} else {
		// Use truncated address as default display name
		truncated := cmd.Address
		if len(truncated) > 10 {
			truncated = truncated[:6] + "..." + truncated[len(truncated)-4:]
		}
		displayName, _ = domain.NewDisplayName(truncated)
	}

	// 4. Create user aggregate (no email for wallet users)
	userID := domain.NewUserID()
	user, err := domain.NewUser(
		userID,
		types.Email{}, // Empty email for wallet users
		displayName,
		domain.AuthMethodWallet,
		verifyResult.DID,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: RegisterWithWalletResult{
			UserID:    userID.String(),
			DID:       verifyResult.DID,
			CreatedAt: user.CreatedAt(),
		},
	}, nil
}

// HandleRegisterWithOAuth handles the RegisterWithOAuth command.
func (h *Handlers) HandleRegisterWithOAuth(ctx context.Context, cmd RegisterWithOAuth) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRegisterWithOAuth"

	// 1. Validate provider
	provider, err := domain.ParseOAuthProvider(cmd.Provider)
	if err != nil {
		return nil, domain.OAuthProviderNotSupported(op, cmd.Provider)
	}

	// 2. Validate OAuth state
	stateRecord, err := h.oauthStateRepo.GetByState(ctx, cmd.State)
	if err != nil {
		return nil, domain.OAuthStateMismatch(op)
	}
	if stateRecord.Provider != cmd.Provider {
		return nil, domain.OAuthStateMismatch(op)
	}
	if time.Now().Unix() > stateRecord.ExpiresAt {
		return nil, domain.OAuthStateMismatch(op)
	}

	// 3. Delete used state
	_ = h.oauthStateRepo.Delete(ctx, cmd.State)

	// 4. Exchange code for profile
	profile, err := h.oauthService.ExchangeCode(ctx, cmd.Provider, cmd.Code, stateRecord.RedirectURL)
	if err != nil {
		return nil, domain.OAuthCodeExchangeFailed(op, err.Error())
	}

	// 5. Create OAuth subject
	oauthSubject, err := domain.NewOAuthSubject(provider, profile.ExternalID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Check if OAuth account already exists
	exists, err := h.userLookup.ExistsByOAuthSubject(ctx, oauthSubject)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.OAuthAccountLinked(op, cmd.Provider)
	}

	// 7. Check if email already exists (if provided)
	if profile.Email != "" {
		email, err := types.NewEmail(profile.Email)
		if err == nil {
			emailExists, err := h.userLookup.ExistsByEmail(ctx, email)
			if err != nil {
				return nil, pkgerrors.Wrap(err, op)
			}
			if emailExists {
				return nil, domain.EmailAlreadyRegistered(op, profile.Email)
			}
		}
	}

	// 8. Create user
	user, err := h.createUserFromOAuth(ctx, profile, oauthSubject)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 9. Override display name if provided
	if cmd.DisplayName != nil && *cmd.DisplayName != "" {
		newDisplayName, err := domain.NewDisplayName(*cmd.DisplayName)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := user.ChangeDisplayName(newDisplayName); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := h.userRepo.Save(ctx, user); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 10. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      user.ID().String(),
		Version: user.Version(),
		Data: RegisterWithOAuthResult{
			UserID:    user.ID().String(),
			Email:     profile.Email,
			Provider:  cmd.Provider,
			CreatedAt: user.CreatedAt(),
		},
	}, nil
}

// ============================================================================
// Magic Link Handlers
// ============================================================================

// HandleRequestMagicLink handles the RequestMagicLink command.
func (h *Handlers) HandleRequestMagicLink(ctx context.Context, cmd RequestMagicLink) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRequestMagicLink"

	// 1. Generate magic link token
	token, err := domain.NewMagicLinkToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Store magic link record
	record := &domain.MagicLinkRecord{
		TokenHash: token.HashString(),
		Email:     cmd.Email.String(),
		ExpiresAt: token.ExpiresAt().Unix(),
		Used:      false,
		CreatedAt: time.Now().UTC().Unix(),
	}
	if err := h.magicLinkRepo.Save(ctx, record); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Send magic link email
	expiresInMinutes := int(token.TTL().Minutes())
	if err := h.emailService.SendMagicLink(ctx, cmd.Email, token.String(), expiresInMinutes); err != nil {
		h.logger.Error("failed to send magic link email",
			log.String("op", op),
			log.Err(err),
		)
		// Don't return error to prevent email enumeration
	}

	// Always return success to prevent email enumeration
	return &cqrs.CommandResult{
		ID: "",
		Data: RequestMagicLinkResult{
			Email:     cmd.Email.String(),
			ExpiresAt: token.ExpiresAt(),
			Success:   true,
		},
	}, nil
}

// HandleVerifyMagicLink handles the VerifyMagicLink command.
func (h *Handlers) HandleVerifyMagicLink(ctx context.Context, cmd VerifyMagicLink) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleVerifyMagicLink"

	// 1. Parse token
	token, err := domain.ParseMagicLinkToken(cmd.Token)
	if err != nil {
		return nil, domain.MagicLinkInvalidError(op, "invalid token format")
	}

	// 2. Get magic link record by hash
	record, err := h.magicLinkRepo.GetByTokenHash(ctx, token.HashString())
	if err != nil {
		return nil, domain.MagicLinkNotFoundError(op)
	}

	// 3. Validate magic link
	if record.Used {
		return nil, domain.MagicLinkUsedError(op)
	}
	if time.Now().Unix() > record.ExpiresAt {
		return nil, domain.MagicLinkExpiredError(op)
	}

	// 4. Mark magic link as used
	if err := h.magicLinkRepo.MarkUsed(ctx, token.HashString()); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Find or create user
	email, err := types.NewEmail(record.Email)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	var user *domain.User
	var isNewUser bool

	userID, err := h.userLookup.GetUserIDByEmail(ctx, email)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.Wrap(err, op)
		}
		// User doesn't exist, create new user
		isNewUser = true
		user, err = h.createUserFromMagicLink(ctx, email)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	} else {
		// Load existing user
		user, err = h.userRepo.Get(ctx, userID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 6. Activate user if pending
	if user.IsPending() {
		if err := user.Activate(); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := h.userRepo.Save(ctx, user); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 7. Create session
	sessionToken, err := domain.NewToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	sessionID := domain.NewSessionID()
	session, err := domain.NewSession(
		sessionID,
		user.ID(),
		sessionToken,
		domain.AuthMethodMagicLink,
		cmd.IPAddress,
		cmd.UserAgent,
		domain.DefaultSessionDuration,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 8. Publish events
	h.publishEvents(ctx, op, user.Changes())
	h.publishEvents(ctx, op, session.Changes())

	return &cqrs.CommandResult{
		ID:      sessionID.String(),
		Version: session.Version(),
		Data: VerifyMagicLinkResult{
			UserID:      user.ID().String(),
			SessionID:   sessionID.String(),
			AccessToken: sessionToken.String(),
			ExpiresAt:   session.ExpiresAt(),
			IsNewUser:   isNewUser,
		},
	}, nil
}

// ============================================================================
// Wallet Authentication Handlers
// ============================================================================

// HandleAuthenticateWithWallet handles the AuthenticateWithWallet command.
func (h *Handlers) HandleAuthenticateWithWallet(ctx context.Context, cmd AuthenticateWithWallet) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAuthenticateWithWallet"

	// 1. Verify wallet signature
	verifyResult, err := h.walletReader.VerifySignature(ctx, domain.WalletVerificationRequest{
		Address:   cmd.Address,
		Message:   cmd.Message,
		Signature: cmd.Signature,
		ChainID:   cmd.ChainID,
	})
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if !verifyResult.Valid {
		return nil, domain.InvalidTokenError(op, "invalid wallet signature")
	}

	// 2. Find or create user
	var user *domain.User
	var isNewUser bool

	userID, err := h.userLookup.GetUserIDByDID(ctx, verifyResult.DID)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.Wrap(err, op)
		}
		// User doesn't exist, create new user
		isNewUser = true
		user, err = h.createUserFromWallet(ctx, cmd.Address, verifyResult.DID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	} else {
		// Load existing user
		user, err = h.userRepo.Get(ctx, userID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 3. Check user can authenticate
	if !user.CanAuthenticate() {
		return nil, domain.UserNotActive(op, user.ID().String())
	}

	// 4. Activate user if pending
	if user.IsPending() {
		if err := user.Activate(); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := h.userRepo.Save(ctx, user); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 5. Create session
	sessionToken, err := domain.NewToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	sessionID := domain.NewSessionID()
	session, err := domain.NewSession(
		sessionID,
		user.ID(),
		sessionToken,
		domain.AuthMethodWallet,
		cmd.IPAddress,
		cmd.UserAgent,
		domain.DefaultSessionDuration,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())
	h.publishEvents(ctx, op, session.Changes())

	return &cqrs.CommandResult{
		ID:      sessionID.String(),
		Version: session.Version(),
		Data: AuthenticateWithWalletResult{
			UserID:      user.ID().String(),
			SessionID:   sessionID.String(),
			AccessToken: sessionToken.String(),
			DID:         verifyResult.DID,
			ExpiresAt:   session.ExpiresAt(),
			IsNewUser:   isNewUser,
		},
	}, nil
}

// ============================================================================
// OAuth Authentication Handlers
// ============================================================================

// HandleInitiateOAuth handles the InitiateOAuth command.
func (h *Handlers) HandleInitiateOAuth(ctx context.Context, cmd InitiateOAuth) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleInitiateOAuth"

	// 1. Validate provider
	_, err := domain.ParseOAuthProvider(cmd.Provider)
	if err != nil {
		return nil, domain.OAuthProviderNotSupported(op, cmd.Provider)
	}

	// 2. Generate state token
	stateToken, err := domain.NewToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Store OAuth state
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	record := &domain.OAuthStateRecord{
		State:       stateToken.String(),
		Provider:    cmd.Provider,
		RedirectURL: cmd.RedirectURL,
		ExpiresAt:   expiresAt.Unix(),
		CreatedAt:   time.Now().UTC().Unix(),
	}
	if err := h.oauthStateRepo.Save(ctx, record); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Get authorization URL from OAuth service
	authURL, err := h.oauthService.GetAuthorizationURL(ctx, cmd.Provider, stateToken.String(), cmd.RedirectURL)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return &cqrs.CommandResult{
		ID: stateToken.String(),
		Data: InitiateOAuthResult{
			AuthURL:   authURL,
			State:     stateToken.String(),
			ExpiresAt: expiresAt,
		},
	}, nil
}

// HandleAuthenticateWithOAuth handles the AuthenticateWithOAuth command.
func (h *Handlers) HandleAuthenticateWithOAuth(ctx context.Context, cmd AuthenticateWithOAuth) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAuthenticateWithOAuth"

	// 1. Validate provider
	provider, err := domain.ParseOAuthProvider(cmd.Provider)
	if err != nil {
		return nil, domain.OAuthProviderNotSupported(op, cmd.Provider)
	}

	// 2. Validate OAuth state
	stateRecord, err := h.oauthStateRepo.GetByState(ctx, cmd.State)
	if err != nil {
		return nil, domain.OAuthStateMismatch(op)
	}
	if stateRecord.Provider != cmd.Provider {
		return nil, domain.OAuthStateMismatch(op)
	}
	if time.Now().Unix() > stateRecord.ExpiresAt {
		return nil, domain.OAuthStateMismatch(op)
	}

	// 3. Delete used state
	_ = h.oauthStateRepo.Delete(ctx, cmd.State)

	// 4. Exchange code for profile
	profile, err := h.oauthService.ExchangeCode(ctx, cmd.Provider, cmd.Code, stateRecord.RedirectURL)
	if err != nil {
		return nil, domain.OAuthCodeExchangeFailed(op, err.Error())
	}

	// 5. Create OAuth subject
	oauthSubject, err := domain.NewOAuthSubject(provider, profile.ExternalID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Find or create user
	var user *domain.User
	var isNewUser bool

	userID, err := h.userLookup.GetUserIDByOAuthSubject(ctx, oauthSubject)
	if err != nil {
		if !pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.Wrap(err, op)
		}
		// User doesn't exist, create new user
		isNewUser = true
		user, err = h.createUserFromOAuth(ctx, profile, oauthSubject)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	} else {
		// Load existing user
		user, err = h.userRepo.Get(ctx, userID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 7. Check user can authenticate
	if !user.CanAuthenticate() {
		return nil, domain.UserNotActive(op, user.ID().String())
	}

	// 8. Activate user if pending
	if user.IsPending() {
		if err := user.Activate(); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if err := h.userRepo.Save(ctx, user); err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
	}

	// 9. Create session
	sessionToken, err := domain.NewToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	sessionID := domain.NewSessionID()
	session, err := domain.NewSession(
		sessionID,
		user.ID(),
		sessionToken,
		domain.AuthMethodOAuth,
		cmd.IPAddress,
		cmd.UserAgent,
		domain.DefaultSessionDuration,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 10. Publish events
	h.publishEvents(ctx, op, user.Changes())
	h.publishEvents(ctx, op, session.Changes())

	return &cqrs.CommandResult{
		ID:      sessionID.String(),
		Version: session.Version(),
		Data: AuthenticateWithOAuthResult{
			UserID:      user.ID().String(),
			SessionID:   sessionID.String(),
			AccessToken: sessionToken.String(),
			Email:       profile.Email,
			Provider:    cmd.Provider,
			ExpiresAt:   session.ExpiresAt(),
			IsNewUser:   isNewUser,
		},
	}, nil
}

// ============================================================================
// Session Handlers
// ============================================================================

// HandleRefreshSession handles the RefreshSession command.
func (h *Handlers) HandleRefreshSession(ctx context.Context, cmd RefreshSession) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRefreshSession"

	// 1. Parse session ID
	sessionID := domain.SessionIDFromTypesID(cmd.SessionID)

	// 2. Load session
	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Verify current token
	currentToken, err := domain.ParseToken(cmd.Token)
	if err != nil {
		return nil, domain.InvalidTokenError(op, "invalid token format")
	}
	if !session.VerifyToken(currentToken) {
		return nil, domain.InvalidTokenError(op, "token mismatch")
	}

	// 4. Generate new token
	newToken, err := domain.NewToken()
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Refresh session
	if err := session.Refresh(newToken, domain.DefaultSessionDuration); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Save session
	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 7. Publish events
	h.publishEvents(ctx, op, session.Changes())

	return &cqrs.CommandResult{
		ID:      sessionID.String(),
		Version: session.Version(),
		Data: RefreshSessionResult{
			SessionID:   sessionID.String(),
			AccessToken: newToken.String(),
			ExpiresAt:   session.ExpiresAt(),
		},
	}, nil
}

// HandleRevokeSession handles the RevokeSession command.
func (h *Handlers) HandleRevokeSession(ctx context.Context, cmd RevokeSession) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokeSession"

	// 1. Parse session ID
	sessionID := domain.SessionIDFromTypesID(cmd.SessionID)

	// 2. Load session
	session, err := h.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Revoke session
	reason := ""
	if cmd.Reason != nil {
		reason = *cmd.Reason
	}
	if err := session.Revoke(reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save session
	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, session.Changes())

	return &cqrs.CommandResult{
		ID:      sessionID.String(),
		Version: session.Version(),
		Data: RevokeSessionResult{
			SessionID: sessionID.String(),
			RevokedAt: time.Now().UTC(),
		},
	}, nil
}

// HandleRevokeAllUserSessions handles the RevokeAllUserSessions command.
func (h *Handlers) HandleRevokeAllUserSessions(ctx context.Context, cmd RevokeAllUserSessions) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRevokeAllUserSessions"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Get all active sessions for user
	sessions, err := h.sessionLookup.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Revoke each session
	reason := ""
	if cmd.Reason != nil {
		reason = *cmd.Reason
	}

	revokedCount := 0
	for _, session := range sessions {
		if err := session.Revoke(reason); err != nil {
			h.logger.Error("failed to revoke session",
				log.String("op", op),
				log.String("session_id", session.ID().String()),
				log.Err(err),
			)
			continue
		}
		if err := h.sessionRepo.Save(ctx, session); err != nil {
			h.logger.Error("failed to save revoked session",
				log.String("op", op),
				log.String("session_id", session.ID().String()),
				log.Err(err),
			)
			continue
		}
		h.publishEvents(ctx, op, session.Changes())
		revokedCount++
	}

	return &cqrs.CommandResult{
		ID: userID.String(),
		Data: RevokeAllUserSessionsResult{
			UserID:       userID.String(),
			RevokedCount: revokedCount,
			RevokedAt:    time.Now().UTC(),
		},
	}, nil
}

// ============================================================================
// Profile Handlers
// ============================================================================

// HandleChangeDisplayName handles the ChangeDisplayName command.
func (h *Handlers) HandleChangeDisplayName(ctx context.Context, cmd ChangeDisplayName) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleChangeDisplayName"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Create display name value object
	newDisplayName, err := domain.NewDisplayName(cmd.NewDisplayName)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Change display name
	if err := user.ChangeDisplayName(newDisplayName); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: ChangeDisplayNameResult{
			UserID:      userID.String(),
			DisplayName: newDisplayName.String(),
			ChangedAt:   user.UpdatedAt(),
		},
	}, nil
}

// HandleChangeEmail handles the ChangeEmail command.
func (h *Handlers) HandleChangeEmail(ctx context.Context, cmd ChangeEmail) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleChangeEmail"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Check if new email is already registered
	exists, err := h.userLookup.ExistsByEmail(ctx, cmd.NewEmail)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.EmailAlreadyRegistered(op, cmd.NewEmail.String())
	}

	// 3. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Change email
	if err := user.ChangeEmail(cmd.NewEmail); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: ChangeEmailResult{
			UserID:    userID.String(),
			Email:     cmd.NewEmail.String(),
			ChangedAt: user.UpdatedAt(),
		},
	}, nil
}

// ============================================================================
// User Status Handlers
// ============================================================================

// HandleActivateUser handles the ActivateUser command.
func (h *Handlers) HandleActivateUser(ctx context.Context, cmd ActivateUser) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleActivateUser"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Activate user
	if err := user.Activate(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: ActivateUserResult{
			UserID:      userID.String(),
			ActivatedAt: user.UpdatedAt(),
		},
	}, nil
}

// HandleSuspendUser handles the SuspendUser command.
func (h *Handlers) HandleSuspendUser(ctx context.Context, cmd SuspendUser) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleSuspendUser"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Suspend user
	if err := user.Suspend(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Revoke all active sessions
	go func() {
		revokeCmd := RevokeAllUserSessions{
			UserID: cmd.UserID,
			Reason: &cmd.Reason,
		}
		if _, err := h.HandleRevokeAllUserSessions(context.Background(), revokeCmd); err != nil {
			h.logger.Error("failed to revoke sessions after suspension",
				log.String("op", op),
				log.String("user_id", userID.String()),
				log.Err(err),
			)
		}
	}()

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: SuspendUserResult{
			UserID:      userID.String(),
			SuspendedAt: user.UpdatedAt(),
		},
	}, nil
}

// HandleReactivateUser handles the ReactivateUser command.
func (h *Handlers) HandleReactivateUser(ctx context.Context, cmd ReactivateUser) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleReactivateUser"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Reactivate user
	if err := user.Reactivate(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: ReactivateUserResult{
			UserID:        userID.String(),
			ReactivatedAt: user.UpdatedAt(),
		},
	}, nil
}

// HandleDeleteUser handles the DeleteUser command.
func (h *Handlers) HandleDeleteUser(ctx context.Context, cmd DeleteUser) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleDeleteUser"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Delete user
	if err := user.Delete(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Revoke all active sessions
	reason := "user deleted"
	go func() {
		revokeCmd := RevokeAllUserSessions{
			UserID: cmd.UserID,
			Reason: &reason,
		}
		if _, err := h.HandleRevokeAllUserSessions(context.Background(), revokeCmd); err != nil {
			h.logger.Error("failed to revoke sessions after deletion",
				log.String("op", op),
				log.String("user_id", userID.String()),
				log.Err(err),
			)
		}
	}()

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: DeleteUserResult{
			UserID:    userID.String(),
			DeletedAt: user.UpdatedAt(),
		},
	}, nil
}

// ============================================================================
// DID Management Handlers
// ============================================================================

// HandleAddDID handles the AddDID command.
func (h *Handlers) HandleAddDID(ctx context.Context, cmd AddDID) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAddDID"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Check if DID is already linked to another user
	exists, err := h.userLookup.ExistsByDID(ctx, cmd.DID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.DIDAlreadyLinked(op, cmd.DID)
	}

	// 3. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Add DID
	if err := user.AddDID(cmd.DID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: AddDIDResult{
			UserID:  userID.String(),
			DID:     cmd.DID,
			AddedAt: user.UpdatedAt(),
		},
	}, nil
}

// HandleRemoveDID handles the RemoveDID command.
func (h *Handlers) HandleRemoveDID(ctx context.Context, cmd RemoveDID) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRemoveDID"

	// 1. Parse user ID
	userID := domain.UserIDFromTypesID(cmd.UserID)

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Remove DID
	if err := user.RemoveDID(cmd.DID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: RemoveDIDResult{
			UserID:    userID.String(),
			DID:       cmd.DID,
			RemovedAt: user.UpdatedAt(),
		},
	}, nil
}

// ============================================================================
// OAuth Linking Handlers
// ============================================================================

// HandleLinkOAuthAccount handles the LinkOAuthAccount command.
func (h *Handlers) HandleLinkOAuthAccount(ctx context.Context, cmd LinkOAuthAccount) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleLinkOAuthAccount"

	// 1. Parse user ID and provider
	userID := domain.UserIDFromTypesID(cmd.UserID)
	provider, err := domain.ParseOAuthProvider(cmd.Provider)
	if err != nil {
		return nil, domain.OAuthProviderNotSupported(op, cmd.Provider)
	}

	// 2. Validate OAuth state
	stateRecord, err := h.oauthStateRepo.GetByState(ctx, cmd.State)
	if err != nil {
		return nil, domain.OAuthStateMismatch(op)
	}
	if stateRecord.Provider != cmd.Provider {
		return nil, domain.OAuthStateMismatch(op)
	}
	if time.Now().Unix() > stateRecord.ExpiresAt {
		return nil, domain.OAuthStateMismatch(op)
	}

	// 3. Delete used state
	_ = h.oauthStateRepo.Delete(ctx, cmd.State)

	// 4. Exchange code for profile
	profile, err := h.oauthService.ExchangeCode(ctx, cmd.Provider, cmd.Code, stateRecord.RedirectURL)
	if err != nil {
		return nil, domain.OAuthCodeExchangeFailed(op, err.Error())
	}

	// 5. Check if OAuth account is already linked to another user
	oauthSubject, err := domain.NewOAuthSubject(provider, profile.ExternalID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	existingUserID, err := h.userLookup.GetUserIDByOAuthSubject(ctx, oauthSubject)
	if err == nil && !existingUserID.Equals(userID) {
		return nil, domain.OAuthAccountLinked(op, cmd.Provider)
	}

	// 6. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 7. Link OAuth account
	if err := user.LinkOAuth(oauthSubject, profile.Email); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 8. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 9. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: LinkOAuthAccountResult{
			UserID:     userID.String(),
			Provider:   cmd.Provider,
			ExternalID: profile.ExternalID,
			LinkedAt:   user.UpdatedAt(),
		},
	}, nil
}

// HandleUnlinkOAuthAccount handles the UnlinkOAuthAccount command.
func (h *Handlers) HandleUnlinkOAuthAccount(ctx context.Context, cmd UnlinkOAuthAccount) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUnlinkOAuthAccount"

	// 1. Parse user ID and provider
	userID := domain.UserIDFromTypesID(cmd.UserID)
	provider, err := domain.ParseOAuthProvider(cmd.Provider)
	if err != nil {
		return nil, domain.OAuthProviderNotSupported(op, cmd.Provider)
	}

	// 2. Load user
	user, err := h.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Unlink OAuth account
	if err := user.UnlinkOAuth(provider); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, user.Changes())

	return &cqrs.CommandResult{
		ID:      userID.String(),
		Version: user.Version(),
		Data: UnlinkOAuthAccountResult{
			UserID:     userID.String(),
			Provider:   cmd.Provider,
			UnlinkedAt: user.UpdatedAt(),
		},
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// createUserFromMagicLink creates a new user from a magic link verification.
func (h *Handlers) createUserFromMagicLink(ctx context.Context, email types.Email) (*domain.User, error) {
	const op = "Handlers.createUserFromMagicLink"

	// Generate DID
	did, err := h.didGenerator.GenerateDIDKey(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Create default display name from email local part
	displayName, _ := domain.NewDisplayName(email.Local())

	// Create user
	userID := domain.NewUserID()
	user, err := domain.NewUser(
		userID,
		email,
		displayName,
		domain.AuthMethodMagicLink,
		did,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return user, nil
}

// createUserFromWallet creates a new user from wallet authentication.
func (h *Handlers) createUserFromWallet(ctx context.Context, address string, did string) (*domain.User, error) {
	const op = "Handlers.createUserFromWallet"

	// Create default display name from address
	truncated := address
	if len(truncated) > 10 {
		truncated = truncated[:6] + "..." + truncated[len(truncated)-4:]
	}
	displayName, _ := domain.NewDisplayName(truncated)

	// Create user (no email for wallet users)
	userID := domain.NewUserID()
	user, err := domain.NewUser(
		userID,
		types.Email{}, // Empty email
		displayName,
		domain.AuthMethodWallet,
		did,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return user, nil
}

// createUserFromOAuth creates a new user from OAuth authentication.
func (h *Handlers) createUserFromOAuth(ctx context.Context, profile *domain.OAuthProfile, subject domain.OAuthSubject) (*domain.User, error) {
	const op = "Handlers.createUserFromOAuth"

	// Generate DID
	did, err := h.didGenerator.GenerateDIDKey(ctx)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Create display name from profile name or email
	displayNameStr := profile.Name
	if displayNameStr == "" && profile.Email != "" {
		email, err := types.NewEmail(profile.Email)
		if err == nil {
			displayNameStr = email.Local()
		}
	}
	if displayNameStr == "" {
		displayNameStr = "User"
	}
	displayName, _ := domain.NewDisplayName(displayNameStr)

	// Create email if provided
	var email types.Email
	if profile.Email != "" {
		email, _ = types.NewEmail(profile.Email)
	}

	// Create user
	userID := domain.NewUserID()
	user, err := domain.NewUser(
		userID,
		email,
		displayName,
		domain.AuthMethodOAuth,
		did,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Save user first
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Activate the user (OAuth users are pre-verified)
	if err := user.Activate(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Link OAuth account
	if err := user.LinkOAuth(subject, profile.Email); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// Save again with OAuth link
	if err := h.userRepo.Save(ctx, user); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	return user, nil
}

// publishEvents publishes domain events (fire-and-forget with logging).
func (h *Handlers) publishEvents(ctx context.Context, op string, events []eventsourcing.Event) {
	if len(events) == 0 {
		return
	}

	domainEvents := make([]domain.DomainEvent, 0, len(events))
	for _, e := range events {
		if de, ok := e.(domain.DomainEvent); ok {
			domainEvents = append(domainEvents, de)
		}
	}

	if len(domainEvents) == 0 {
		return
	}

	if err := h.publisher.Publish(ctx, domainEvents...); err != nil {
		h.logger.Error("failed to publish events",
			log.String("op", op),
			log.Err(err),
		)
	}
}

// ============================================================================
// Command Registration
// ============================================================================

// RegisterCommands registers all Identity command handlers with the command bus.
func RegisterCommands(bus *cqrs.InMemoryCommandBus, handlers *Handlers) error {
	registrations := []struct {
		name    string
		handler any
	}{
		{CommandRegisterWithEmail, handlers},
		{CommandRegisterWithWallet, handlers},
		{CommandRegisterWithOAuth, handlers},
		{CommandRequestMagicLink, handlers},
		{CommandVerifyMagicLink, handlers},
		{CommandAuthenticateWithWallet, handlers},
		{CommandInitiateOAuth, handlers},
		{CommandAuthenticateWithOAuth, handlers},
		{CommandRefreshSession, handlers},
		{CommandRevokeSession, handlers},
		{CommandRevokeAllUserSessions, handlers},
		{CommandChangeDisplayName, handlers},
		{CommandChangeEmail, handlers},
		{CommandActivateUser, handlers},
		{CommandSuspendUser, handlers},
		{CommandReactivateUser, handlers},
		{CommandDeleteUser, handlers},
		{CommandAddDID, handlers},
		{CommandRemoveDID, handlers},
		{CommandLinkOAuthAccount, handlers},
		{CommandUnlinkOAuthAccount, handlers},
	}

	for _, r := range registrations {
		if err := bus.Register(r.name, r.handler); err != nil {
			return err
		}
	}

	return nil
}
