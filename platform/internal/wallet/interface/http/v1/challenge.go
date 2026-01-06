package v1

import (
	"encoding/json"
	"net/http"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Challenge Handlers
// ============================================================================

// CreateChallenge generates a new SIWE challenge for wallet authentication.
// POST /v1/wallet/challenge
func (h *Handler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := ValidateCreateChallengeRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	// Get domain and URI from request
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	requestDomain := r.Host
	uri := scheme + "://" + requestDomain

	// Create challenge params
	params := domain.CreateChallengeParams{
		Address: domain.NewAddressUnchecked(
			req.Address,
			req.Address, // Will be normalized by service
			req.ToChainID(),
			domain.GetChainFamily(req.ToChainID()),
		),
		Domain:    requestDomain,
		URI:       uri,
		Statement: "Sign in with your wallet to Nexus",
	}

	// Create challenge
	challenge, err := h.challengeSvc.CreateChallenge(ctx, params)
	if err != nil {
		h.logger.Error("failed to create challenge",
			log.Err(err),
			log.String("address", req.Address),
			log.String("chain_id", req.ChainID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, &ChallengeResponse{
		Nonce:     challenge.Nonce,
		Message:   challenge.Message,
		Address:   challenge.Address.Normalized(),
		ChainID:   challenge.Address.ChainID().String(),
		Domain:    challenge.Domain,
		URI:       challenge.URI,
		IssuedAt:  challenge.IssuedAt,
		ExpiresAt: challenge.ExpiresAt,
	})
}

// VerifyChallenge verifies a signed SIWE challenge.
// POST /v1/wallet/challenge/verify
func (h *Handler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req VerifyChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := ValidateVerifyChallengeRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	// Validate and consume the challenge
	challenge, err := h.challengeSvc.ValidateChallenge(ctx, req.Nonce)
	if err != nil {
		h.logger.Error("failed to validate challenge",
			log.Err(err),
			log.String("nonce", req.Nonce),
		)
		WriteError(w, err)
		return
	}

	// Verify the signature matches the challenge message
	if req.Message != challenge.Message {
		h.logger.Warn("message mismatch",
			log.String("nonce", req.Nonce),
		)
		WriteError(w, domain.ErrVerificationFailed("VerifyChallenge", "message does not match challenge"))
		return
	}

	// Note: Actual signature verification would be done here or delegated to a command
	// For now, we return the challenge details for the caller to complete verification

	httpresponse.JSON(w, http.StatusOK, &VerifyChallengeResponse{
		Valid:   true,
		Address: challenge.Address.Normalized(),
		ChainID: challenge.Address.ChainID().String(),
	})
}

// GetChallenge retrieves an existing challenge by nonce.
// GET /v1/wallet/challenge/{nonce}
func (h *Handler) GetChallenge(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	nonce := r.PathValue("nonce")
	if nonce == "" {
		WriteBadRequest(w, "nonce is required")
		return
	}

	// Note: This is a read-only operation, doesn't consume the challenge
	// Implementation depends on whether you want to expose this endpoint
	// For security, you might want to limit what information is returned

	h.logger.Debug("get challenge request",
		log.String("nonce", nonce),
	)

	// For now, return not implemented or use query bus
	WriteNotFound(w, "challenge lookup not available")
}
