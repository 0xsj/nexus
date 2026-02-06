package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/0xsj/nexus/platform/internal/wallet/app/command"
	"github.com/0xsj/nexus/platform/internal/wallet/app/query"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// ============================================================================
// Handler
// ============================================================================

// Handler handles HTTP requests for the Wallet context.
type Handler struct {
	commands *command.Handlers
	queries  *query.Handlers
}

// NewHandler creates a new Handler.
func NewHandler(commands *command.Handlers, queries *query.Handlers) *Handler {
	return &Handler{
		commands: commands,
		queries:  queries,
	}
}

// ============================================================================
// Wallet Linking
// ============================================================================

// LinkWallet handles POST /api/v1/wallets
func (h *Handler) LinkWallet(w http.ResponseWriter, r *http.Request) {
	var req LinkWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	// Validate required fields
	if req.Address == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("address", "address is required"))
		return
	}
	if req.ChainID <= 0 {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("chain_id", "chain_id must be positive"))
		return
	}

	// TODO: Extract user_id from auth context in production
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("user_id", "user_id is required"))
		return
	}

	// Build command
	cmd := command.LinkWallet{
		UserID:  userID,
		Address: req.Address,
		ChainID: req.ChainID,
		Label:   req.Label,
	}

	// Execute command
	result, err := h.commands.HandleLinkWallet(r.Context(), cmd)
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	// Map result
	data := result.Data.(command.LinkWalletResult)
	h.writeJSON(w, http.StatusCreated, WalletLinkedResponse{
		WalletID: data.WalletID,
		Address:  data.Address,
		Status:   data.Status,
		LinkedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Wallet Retrieval
// ============================================================================

// GetWallet handles GET /api/v1/wallets/{walletId}
func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletId")
	if walletID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("wallet_id is required"))
		return
	}

	id, err := types.ParseID(walletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("wallet_id", "invalid wallet_id format"))
		return
	}

	result, err := h.queries.HandleGetWallet(r.Context(), query.GetWallet{
		WalletID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toWalletResponse(result))
}

// ============================================================================
// Wallet Lists
// ============================================================================

// ListWalletsByUser handles GET /api/v1/wallets/user/{userId}
func (h *Handler) ListWalletsByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("user_id is required"))
		return
	}

	limit, offset := h.parsePagination(r)

	result, err := h.queries.HandleListWalletsByUser(r.Context(), query.ListWalletsByUser{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	h.writeJSON(w, http.StatusOK, toWalletListResponse(result))
}

// ============================================================================
// Wallet Verification
// ============================================================================

// VerifyWallet handles POST /api/v1/wallets/{walletId}/verify
func (h *Handler) VerifyWallet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletId")
	if walletID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("wallet_id is required"))
		return
	}

	id, err := types.ParseID(walletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("wallet_id", "invalid wallet_id format"))
		return
	}

	var req VerifyWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Signature == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("signature", "signature is required"))
		return
	}
	if req.Message == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("message", "message is required"))
		return
	}

	result, err := h.commands.HandleVerifyWallet(r.Context(), command.VerifyWallet{
		WalletID:  id,
		Signature: req.Signature,
		Message:   req.Message,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.VerifyWalletResult)
	h.writeJSON(w, http.StatusOK, WalletVerifiedResponse{
		WalletID:   data.WalletID,
		Status:     data.Status,
		VerifiedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Wallet Unlinking
// ============================================================================

// UnlinkWallet handles POST /api/v1/wallets/{walletId}/unlink
func (h *Handler) UnlinkWallet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletId")
	if walletID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("wallet_id is required"))
		return
	}

	id, err := types.ParseID(walletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("wallet_id", "invalid wallet_id format"))
		return
	}

	result, err := h.commands.HandleUnlinkWallet(r.Context(), command.UnlinkWallet{
		WalletID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UnlinkWalletResult)
	h.writeJSON(w, http.StatusOK, WalletUnlinkedResponse{
		WalletID:   data.WalletID,
		Status:     data.Status,
		UnlinkedAt: time.Now().UTC(),
	})
}

// ============================================================================
// Wallet Primary
// ============================================================================

// SetPrimaryWallet handles POST /api/v1/wallets/{walletId}/primary
func (h *Handler) SetPrimaryWallet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletId")
	if walletID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("wallet_id is required"))
		return
	}

	id, err := types.ParseID(walletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("wallet_id", "invalid wallet_id format"))
		return
	}

	result, err := h.commands.HandleSetPrimaryWallet(r.Context(), command.SetPrimaryWallet{
		WalletID: id,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.SetPrimaryWalletResult)
	h.writeJSON(w, http.StatusOK, WalletPrimarySetResponse{
		WalletID: data.WalletID,
		Status:   data.Status,
	})
}

// ============================================================================
// Wallet Label
// ============================================================================

// UpdateWalletLabel handles PATCH /api/v1/wallets/{walletId}/label
func (h *Handler) UpdateWalletLabel(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "walletId")
	if walletID == "" {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("wallet_id is required"))
		return
	}

	id, err := types.ParseID(walletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("wallet_id", "invalid wallet_id format"))
		return
	}

	var req UpdateWalletLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, BadRequestResponse("invalid request body"))
		return
	}

	if req.Label == "" {
		h.writeError(w, http.StatusBadRequest, ValidationErrorResponse("label", "label is required"))
		return
	}

	result, err := h.commands.HandleUpdateWalletLabel(r.Context(), command.UpdateWalletLabel{
		WalletID: id,
		Label:    req.Label,
	})
	if err != nil {
		status, errResp := MapError(err)
		h.writeError(w, status, errResp)
		return
	}

	data := result.Data.(command.UpdateWalletLabelResult)
	h.writeJSON(w, http.StatusOK, WalletLabelUpdatedResponse{
		WalletID: data.WalletID,
		Label:    data.Label,
	})
}

// ============================================================================
// Request Parsing Helpers
// ============================================================================

func (h *Handler) parsePagination(r *http.Request) (int, int) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

func toWalletResponse(v *query.WalletView) WalletResponse {
	return WalletResponse{
		WalletID:   v.WalletID,
		UserID:     v.UserID,
		Address:    v.Address,
		Chain:      v.Chain,
		Label:      v.Label,
		Status:     v.Status,
		IsPrimary:  v.IsPrimary,
		DID:        v.DID,
		LinkedAt:   v.LinkedAt,
		VerifiedAt: v.VerifiedAt,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

func toWalletListResponse(v *query.WalletListView) WalletListResponse {
	summaries := make([]WalletSummaryResponse, len(v.Wallets))
	for i, w := range v.Wallets {
		summaries[i] = WalletSummaryResponse{
			WalletID:  w.WalletID,
			Address:   w.Address,
			Chain:     w.Chain,
			Label:     w.Label,
			Status:    w.Status,
			IsPrimary: w.IsPrimary,
		}
	}

	return WalletListResponse{
		Wallets:    summaries,
		TotalCount: v.TotalCount,
		Limit:      v.Limit,
		Offset:     v.Offset,
		HasMore:    v.HasMore,
	}
}

// ============================================================================
// Response Writing Helpers
// ============================================================================

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, errResp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errResp)
}
