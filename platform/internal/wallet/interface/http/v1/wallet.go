package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/wallet/application/command"
	"github.com/0xsj/nexus/platform/internal/wallet/application/query"
	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// List & Get Handlers
// ============================================================================

// ListWallets returns all wallets for the authenticated user.
// GET /v1/wallets
func (h *Handler) ListWallets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	params := parseWalletListParams(r)

	result, err := dispatchWalletListQuery(ctx, h.queryBus, &query.ListWalletsByUser{
		UserID:    userID,
		Status:    params.ToStatus(),
		Limit:     params.Limit,
		Offset:    params.Offset,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	})
	if err != nil {
		h.logger.Error("failed to list wallets",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromWalletListView(result))
}

// GetWallet returns a specific wallet by ID.
// GET /v1/wallets/{id}
func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
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

	result, err := dispatchWalletQuery(ctx, h.queryBus, &query.GetWalletByID{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		h.logger.Error("failed to get wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromWalletView(result))
}

// GetPrimaryWallet returns the user's primary wallet.
// GET /v1/wallets/primary
func (h *Handler) GetPrimaryWallet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	result, err := dispatchWalletQuery(ctx, h.queryBus, &query.GetPrimaryWallet{
		UserID: userID,
	})
	if err != nil {
		h.logger.Error("failed to get primary wallet",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromWalletView(result))
}

// GetWalletByAddress returns a wallet by address and chain.
// GET /v1/wallets/address/{address}?chain_id=...
func (h *Handler) GetWalletByAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	address := chi.URLParam(r, "address")
	if address == "" {
		WriteBadRequest(w, "address is required")
		return
	}

	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		WriteBadRequest(w, "chain_id is required")
		return
	}

	result, err := dispatchWalletQuery(ctx, h.queryBus, &query.GetWalletByAddress{
		Address: address,
		ChainID: domain.ChainID(chainID),
	})
	if err != nil {
		h.logger.Error("failed to get wallet by address",
			log.Err(err),
			log.String("address", address),
			log.String("chain_id", chainID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromWalletView(result))
}

// ============================================================================
// Update Handlers
// ============================================================================

// UpdateWallet updates a wallet's label.
// PATCH /v1/wallets/{id}
func (h *Handler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteBadRequest(w, "invalid request body")
		return
	}

	if err := ValidateUpdateWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	// Handle label update
	if req.Label != nil {
		result, err := h.dispatchCommand(ctx, &command.UpdateLabel{
			WalletID: walletID,
			UserID:   userID,
			Label:    *req.Label,
		})
		if err != nil {
			h.logger.Error("failed to update wallet label",
				log.Err(err),
				log.String("user_id", userID),
				log.String("wallet_id", walletID),
			)
			WriteError(w, err)
			return
		}

		data := result.Data.(*command.UpdateLabelResult)

		httpresponse.JSON(w, http.StatusOK, &UpdateWalletResponse{
			WalletID:  walletID,
			Label:     data.NewLabel,
			IsPrimary: false, // Would need to fetch current value
			UpdatedAt: time.Now(),
		})
		return
	}

	// Handle primary update
	if req.IsPrimary != nil && *req.IsPrimary {
		result, err := h.dispatchCommand(ctx, &command.SetPrimary{
			WalletID: walletID,
			UserID:   userID,
		})
		if err != nil {
			h.logger.Error("failed to set primary wallet",
				log.Err(err),
				log.String("user_id", userID),
				log.String("wallet_id", walletID),
			)
			WriteError(w, err)
			return
		}

		data := result.Data.(*command.SetPrimaryResult)

		httpresponse.JSON(w, http.StatusOK, &SetPrimaryWalletResponse{
			WalletID:        data.WalletID,
			IsPrimary:       true,
			PreviousPrimary: data.PreviousPrimaryID,
		})
		return
	}

	WriteBadRequest(w, "no valid update fields provided")
}

// SetPrimaryWallet sets a wallet as the user's primary wallet.
// PATCH /v1/wallets/{id}/primary
func (h *Handler) SetPrimaryWallet(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.dispatchCommand(ctx, &command.SetPrimary{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		h.logger.Error("failed to set primary wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.SetPrimaryResult)

	httpresponse.JSON(w, http.StatusOK, &SetPrimaryWalletResponse{
		WalletID:        data.WalletID,
		IsPrimary:       true,
		PreviousPrimary: data.PreviousPrimaryID,
	})
}

// ============================================================================
// Delete Handler
// ============================================================================

// DeleteWallet removes a wallet from the user's account.
// DELETE /v1/wallets/{id}
func (h *Handler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.dispatchCommand(ctx, &command.UnlinkWallet{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		h.logger.Error("failed to delete wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.UnlinkWalletResult)

	httpresponse.JSON(w, http.StatusOK, &UnlinkWalletResponse{
		WalletID:   data.WalletID,
		Address:    data.Address,
		ChainID:    data.ChainID,
		UnlinkedAt: time.Now(),
	})
}

// ============================================================================
// Status Handlers
// ============================================================================

// ActivateWallet activates a wallet.
// POST /v1/wallets/{id}/activate
func (h *Handler) ActivateWallet(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.dispatchCommand(ctx, &command.ActivateWallet{
		WalletID: walletID,
		UserID:   userID,
	})
	if err != nil {
		h.logger.Error("failed to activate wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.ActivateWalletResult)

	httpresponse.JSON(w, http.StatusOK, &ActivateWalletResponse{
		WalletID:    data.WalletID,
		Status:      data.CurrentStatus,
		ActivatedAt: time.Now(),
	})
}

// DeactivateWallet deactivates a wallet.
// POST /v1/wallets/{id}/deactivate
func (h *Handler) DeactivateWallet(w http.ResponseWriter, r *http.Request) {
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

	var req DeactivateWalletRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteBadRequest(w, "invalid request body")
			return
		}
	}

	if err := ValidateDeactivateWalletRequest(&req); err != nil {
		WriteValidationError(w, err.Error(), nil)
		return
	}

	result, err := h.dispatchCommand(ctx, &command.DeactivateWallet{
		WalletID: walletID,
		UserID:   userID,
		Reason:   req.Reason,
	})
	if err != nil {
		h.logger.Error("failed to deactivate wallet",
			log.Err(err),
			log.String("user_id", userID),
			log.String("wallet_id", walletID),
		)
		WriteError(w, err)
		return
	}

	data := result.Data.(*command.DeactivateWalletResult)

	httpresponse.JSON(w, http.StatusOK, &DeactivateWalletResponse{
		WalletID:      data.WalletID,
		Status:        data.CurrentStatus,
		DeactivatedAt: time.Now(),
	})
}

// ============================================================================
// Stats Handlers
// ============================================================================

// GetWalletStats returns wallet statistics for the authenticated user.
// GET /v1/wallets/stats
func (h *Handler) GetWalletStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		WriteUnauthorized(w, "authentication required")
		return
	}

	result, err := dispatchUserWalletStatsQuery(ctx, h.queryBus, &query.CountWalletsByUser{
		UserID:     userID,
		ActiveOnly: false,
	})
	if err != nil {
		h.logger.Error("failed to get wallet stats",
			log.Err(err),
			log.String("user_id", userID),
		)
		WriteError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, FromUserWalletStatsView(result))
}

// ============================================================================
// Helpers
// ============================================================================

func parseWalletListParams(r *http.Request) WalletListParams {
	q := r.URL.Query()

	params := WalletListParams{}
	params.Limit, _ = strconv.Atoi(q.Get("limit"))
	params.Offset, _ = strconv.Atoi(q.Get("offset"))
	params.ChainID = q.Get("chain_id")
	params.Status = q.Get("status")
	params.ActiveOnly = q.Get("active_only") == "true"
	params.SortBy = q.Get("sort_by")
	params.SortOrder = q.Get("sort_order")

	params.WithDefaults()

	return params
}
