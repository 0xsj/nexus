package v1

import (
	"encoding/json"
	"net/http"

	"github.com/0xsj/nexus/platform/internal/identity/application/command"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Auth Handlers
// ============================================================================

// RequestChallenge generates a SIWE challenge for wallet authentication.
// POST /v1/auth/challenge
func (h *Handler) RequestChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if req.Address == "" {
		WriteValidationError(w, "address is required", nil)
		return
	}

	if req.Chain == "" {
		WriteValidationError(w, "chain is required", nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.RequestChallenge{
		Address: req.Address,
		Chain:   req.ToChain(),
		Domain:  r.Host,
		URI:     "https://" + r.Host,
	})
	if err != nil {
		h.logger.Error("failed to create challenge",
			log.Err(err),
			log.String("address", req.Address),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.RequestChallengeResult)

	httpresponse.JSON(w, http.StatusOK, &ChallengeResponse{
		Nonce:     data.Nonce,
		Message:   data.Message,
		Domain:    data.Domain,
		URI:       data.URI,
		IssuedAt:  data.IssuedAt,
		ExpiresAt: data.ExpiresAt,
	})
}

// RegisterWithWallet registers a new user with a wallet signature.
// POST /v1/auth/register/wallet
func (h *Handler) RegisterWithWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RegisterWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := validateRegisterWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.RegisterWithWallet{
		Address:   req.Address,
		Chain:     req.ToChain(),
		Signature: req.Signature,
		Nonce:     req.Nonce,
		Message:   req.Message,
		UserAgent: GetUserAgent(r),
		IPAddress: GetClientIP(r),
	})
	if err != nil {
		h.logger.Error("failed to register with wallet",
			log.Err(err),
			log.String("address", req.Address),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.AuthResult)

	httpresponse.JSON(w, http.StatusCreated, &AuthResponse{
		UserID:       data.UserID,
		DID:          data.DID,
		SessionID:    data.SessionID,
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		TokenType:    data.TokenType,
		ExpiresIn:    data.ExpiresIn,
		ExpiresAt:    data.ExpiresAt,
	})
}

// LoginWithWallet authenticates a user with a wallet signature.
// POST /v1/auth/login/wallet
func (h *Handler) LoginWithWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req LoginWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := validateLoginWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.AuthenticateWithWallet{
		Address:   req.Address,
		Chain:     req.ToChain(),
		Signature: req.Signature,
		Nonce:     req.Nonce,
		Message:   req.Message,
		UserAgent: GetUserAgent(r),
		IPAddress: GetClientIP(r),
	})
	if err != nil {
		h.logger.Error("failed to login with wallet",
			log.Err(err),
			log.String("address", req.Address),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.AuthResult)

	httpresponse.JSON(w, http.StatusOK, &AuthResponse{
		UserID:       data.UserID,
		DID:          data.DID,
		SessionID:    data.SessionID,
		AccessToken:  data.AccessToken,
		RefreshToken: data.RefreshToken,
		TokenType:    data.TokenType,
		ExpiresIn:    data.ExpiresIn,
		ExpiresAt:    data.ExpiresAt,
	})
}

// RefreshToken refreshes an access token using a refresh token.
// POST /v1/auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		WriteValidationError(w, "refresh_token is required", nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.RefreshToken{
		RefreshToken: req.RefreshToken,
		UserAgent:    GetUserAgent(r),
		IPAddress:    GetClientIP(r),
	})
	if err != nil {
		h.logger.Error("failed to refresh token", log.Err(err))
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.RefreshTokenResult)

	httpresponse.JSON(w, http.StatusOK, &RefreshResponse{
		AccessToken: data.AccessToken,
		TokenType:   data.TokenType,
		ExpiresIn:   data.ExpiresIn,
		ExpiresAt:   data.ExpiresAt,
	})
}

// Logout revokes the current session.
// POST /v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	sessionID, ok := SessionIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "session not found")
		return
	}

	_, err := h.dispatchCommand(ctx, &command.RevokeSession{
		UserID:    userID,
		SessionID: sessionID,
		Reason:    "user logout",
	})
	if err != nil {
		h.logger.Error("failed to logout",
			log.Err(err),
			log.String("user_id", userID),
			log.String("session_id", sessionID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, map[string]bool{"logged_out": true})
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateRegisterWalletRequest(req *RegisterWalletRequest) error {
	if req.Address == "" {
		return validationError("address is required")
	}
	if req.Chain == "" {
		return validationError("chain is required")
	}
	if req.Signature == "" {
		return validationError("signature is required")
	}
	if req.Message == "" {
		return validationError("message is required")
	}
	if req.Nonce == "" {
		return validationError("nonce is required")
	}
	return nil
}

func validateLoginWalletRequest(req *LoginWalletRequest) error {
	if req.Address == "" {
		return validationError("address is required")
	}
	if req.Chain == "" {
		return validationError("chain is required")
	}
	if req.Signature == "" {
		return validationError("signature is required")
	}
	if req.Message == "" {
		return validationError("message is required")
	}
	if req.Nonce == "" {
		return validationError("nonce is required")
	}
	return nil
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}
