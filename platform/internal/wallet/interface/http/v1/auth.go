package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/wallet/application/command"
	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Link Wallet Handler
// ============================================================================

// LinkWallet links a wallet to the authenticated user.
// POST /v1/wallets/link
func (h *Handler) LinkWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	var req LinkWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := ValidateLinkWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.LinkWallet{
		UserID:    userID,
		Address:   req.Address,
		ChainID:   req.ToChainID(),
		Signature: req.Signature,
		Message:   req.Message,
		Nonce:     req.Nonce,
		Label:     req.GetLabel(),
	})
	if err != nil {
		h.logger.Error("failed to link wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("address", req.Address),
			log.String("chain_id", req.ChainID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.LinkWalletResult)

	// Get chain name for response
	chainInfo := domain.GetChainInfo(domain.ChainID(data.ChainID))
	chainName := data.ChainID
	if !chainInfo.IsZero() {
		chainName = chainInfo.DisplayName
	}

	httpresponse.JSON(w, http.StatusCreated, &LinkWalletResponse{
		WalletID:  data.WalletID,
		UserID:    data.UserID,
		Address:   data.Address,
		ChainID:   data.ChainID,
		ChainName: chainName,
		DID:       data.DID,
		IsPrimary: data.IsPrimary,
		LinkedAt:  data.CreatedAt,
	})
}

// ============================================================================
// Reverify Wallet Handler
// ============================================================================

// ReverifyWallet re-verifies wallet ownership.
// POST /v1/wallets/{id}/verify
func (h *Handler) ReverifyWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		WriteBadRequest(w, "wallet id is required")
		return
	}

	var req ReverifyWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := ValidateReverifyWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	// Verify the challenge
	result, err := h.dispatchCommand(ctx, &command.VerifyChallenge{
		Nonce:     req.Nonce,
		Signature: req.Signature,
		Message:   req.Message,
	})
	if err != nil {
		h.logger.Error("failed to verify challenge for reverification",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.VerifyChallengeResult)

	if !data.Valid {
		WriteError(w, domain.ErrVerificationFailed("ReverifyWallet", "signature verification failed"))
		return
	}

	// Record the usage
	_, err = h.dispatchCommand(ctx, &command.RecordWalletUsage{
		WalletID:  walletID,
		Action:    "verify",
		IPAddress: getClientIPFromRequest(r),
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		// Log but don't fail the request
		h.logger.Warn("failed to record wallet usage",
			log.Err(err),
			log.String("wallet_id", walletID),
		)
	}

	httpresponse.JSON(w, http.StatusOK, &ReverifyWalletResponse{
		WalletID:   walletID,
		Address:    data.Address,
		Verified:   true,
		VerifiedAt: time.Now(),
	})
}

// ============================================================================
// Verify Signature Handler (Public)
// ============================================================================

// VerifySignature verifies a wallet signature without linking.
// POST /v1/wallet/verify
func (h *Handler) VerifySignature(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.dispatchCommand(ctx, &command.VerifyChallenge{
		Nonce:     req.Nonce,
		Signature: req.Signature,
		Message:   req.Message,
	})
	if err != nil {
		h.logger.Error("failed to verify signature",
			log.Err(err),
			log.String("nonce", req.Nonce),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.VerifyChallengeResult)

	httpresponse.JSON(w, http.StatusOK, &VerifyChallengeResponse{
		Valid:   data.Valid,
		Address: data.Address,
		ChainID: data.ChainID,
		DID:     "", // Would need to derive from address
	})
}

// ============================================================================
// Check Wallet Exists Handler (Public)
// ============================================================================

// CheckWalletExists checks if a wallet is already linked.
// GET /v1/wallet/exists?address=...&chain_id=...
func (h *Handler) CheckWalletExists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := r.URL.Query().Get("address")
	if address == "" {
		WriteBadRequest(w, "address is required")
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		WriteBadRequest(w, "chain_id is required")
		return
	}

	exists, err := dispatchBoolQuery(ctx, h.queryBus, &query.CheckWalletExists{
		Address: &query.AddressIdentifier{
			Address: address,
			ChainID: domain.ChainID(chainID),
		},
	})
	if err != nil {
		h.logger.Error("failed to check wallet exists",
			log.Err(err),
			log.String("address", address),
			log.String("chain_id", chainID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, map[string]bool{
		"exists": exists,
	})
}

// ============================================================================
// Supported Chains Handler (Public)
// ============================================================================

// GetSupportedChains returns the list of supported blockchains.
// GET /v1/wallet/chains
func (h *Handler) GetSupportedChains(w http.ResponseWriter, r *http.Request) {
	chainIDs := domain.SupportedChainIDs()

	response := &SupportedChainsResponse{
		Chains: make([]SupportedChainResponse, 0, len(chainIDs)),
	}

	for _, chainID := range chainIDs {
		chainInfo := domain.GetChainInfo(chainID)
		if chainInfo.IsZero() {
			continue
		}

		response.Chains = append(response.Chains, SupportedChainResponse{
			ChainID:     chainInfo.ID.String(),
			Name:        chainInfo.Name,
			DisplayName: chainInfo.DisplayName,
			Family:      chainInfo.Family.String(),
			IsTestnet:   chainInfo.IsTestnet,
		})
	}

	httpresponse.JSON(w, http.StatusOK, response)
}

// ============================================================================
// Helpers
// ============================================================================

// getClientIPFromRequest extracts the client IP from request.
func getClientIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		for i, c := range forwarded {
			if c == ',' {
				return forwarded[:i]
			}
		}
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}

	return addr
}
