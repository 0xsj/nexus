package health

import (
	"encoding/json"
	"net/http"
)

// Handler provides HTTP handlers for health checks.
type Handler struct {
	checker *Checker
}

// NewHandler creates a new health handler.
func NewHandler(checker *Checker) *Handler {
	return &Handler{
		checker: checker,
	}
}

// ============================================================================
// HTTP Handlers
// ============================================================================

// Health returns a handler for full health checks.
// GET /health
func (h *Handler) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.checker.Check(ctx)

		h.writeJSON(w, response.HTTPStatus(), response)
	}
}

// Liveness returns a handler for Kubernetes liveness probes.
// GET /healthz or GET /health/live
func (h *Handler) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.checker.Liveness(ctx)

		status := http.StatusOK
		if response.Status != StatusUp {
			status = http.StatusServiceUnavailable
		}

		h.writeJSON(w, status, response)
	}
}

// Readiness returns a handler for Kubernetes readiness probes.
// GET /readyz or GET /health/ready
func (h *Handler) Readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.checker.Readiness(ctx)

		status := http.StatusOK
		if response.Status != StatusUp {
			status = http.StatusServiceUnavailable
		}

		h.writeJSON(w, status, response)
	}
}

// Startup returns a handler for Kubernetes startup probes.
// GET /startupz or GET /health/startup
func (h *Handler) Startup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.checker.Startup(ctx)

		status := http.StatusOK
		if response.Status != StatusUp {
			status = http.StatusServiceUnavailable
		}

		h.writeJSON(w, status, response)
	}
}

// Component returns a handler for checking a specific component.
// GET /health/{component}
func (h *Handler) Component() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		component := r.PathValue("component")

		if component == "" {
			h.writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "component name required",
			})
			return
		}

		result := h.checker.CheckComponent(ctx, component)

		status := http.StatusOK
		if result.Status != StatusUp {
			status = http.StatusServiceUnavailable
		}

		h.writeJSON(w, status, result)
	}
}

// ============================================================================
// Route Registration
// ============================================================================

// RegisterRoutes registers all health routes on a mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Full health check
	mux.HandleFunc("GET /health", h.Health())

	// Kubernetes-style probes
	mux.HandleFunc("GET /healthz", h.Liveness())
	mux.HandleFunc("GET /readyz", h.Readiness())
	mux.HandleFunc("GET /startupz", h.Startup())

	// Alternative paths
	mux.HandleFunc("GET /health/live", h.Liveness())
	mux.HandleFunc("GET /health/ready", h.Readiness())
	mux.HandleFunc("GET /health/startup", h.Startup())

	// Component-specific check
	mux.HandleFunc("GET /health/{component}", h.Component())
}

// RegisterRoutesWithPrefix registers health routes with a prefix.
func (h *Handler) RegisterRoutesWithPrefix(mux *http.ServeMux, prefix string) {
	// Ensure prefix starts with / and doesn't end with /
	if prefix == "" {
		prefix = ""
	} else {
		if prefix[0] != '/' {
			prefix = "/" + prefix
		}
		if prefix[len(prefix)-1] == '/' {
			prefix = prefix[:len(prefix)-1]
		}
	}

	// Full health check
	mux.HandleFunc("GET "+prefix+"/health", h.Health())

	// Kubernetes-style probes
	mux.HandleFunc("GET "+prefix+"/healthz", h.Liveness())
	mux.HandleFunc("GET "+prefix+"/readyz", h.Readiness())
	mux.HandleFunc("GET "+prefix+"/startupz", h.Startup())

	// Alternative paths
	mux.HandleFunc("GET "+prefix+"/health/live", h.Liveness())
	mux.HandleFunc("GET "+prefix+"/health/ready", h.Readiness())
	mux.HandleFunc("GET "+prefix+"/health/startup", h.Startup())

	// Component-specific check
	mux.HandleFunc("GET "+prefix+"/health/{component}", h.Component())
}

// ============================================================================
// Simple Handlers (without Handler struct)
// ============================================================================

// SimpleHealthHandler returns a basic health handler.
func SimpleHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	}
}

// SimpleLivenessHandler returns a minimal liveness handler.
func SimpleLivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"up"}`))
	}
}

// SimpleReadinessHandler returns a readiness handler with a ready flag.
func SimpleReadinessHandler(isReady func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if isReady() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"up"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"down"}`))
		}
	}
}

// ============================================================================
// Helpers
// ============================================================================

// writeJSON writes a JSON response.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but don't fail - headers already sent
		_ = err
	}
}