package v1

import (
	"net/http"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/errors"
	httpresponse "github.com/0xsj/nexus/platform/pkg/http/response"
)

// ============================================================================
// Error Response
// ============================================================================

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ============================================================================
// Error Mapper
// ============================================================================

// MapError maps domain/application errors to HTTP responses.
func MapError(err error) (int, *ErrorResponse) {
	if err == nil {
		return http.StatusInternalServerError, &ErrorResponse{
			Error: "unknown error",
			Code:  "INTERNAL_ERROR",
		}
	}

	// Check for domain errors first
	if domainErr := mapDomainError(err); domainErr != nil {
		return domainErr.status, domainErr.response
	}

	// Check for pkg/errors types
	if pkgErr := mapPkgError(err); pkgErr != nil {
		return pkgErr.status, pkgErr.response
	}

	// Default to internal error
	return http.StatusInternalServerError, &ErrorResponse{
		Error:   "internal server error",
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}
}

type mappedError struct {
	status   int
	response *ErrorResponse
}

func mapDomainError(err error) *mappedError {
	// Wallet errors
	if domain.IsWalletNotFound(err) {
		return &mappedError{
			status: http.StatusNotFound,
			response: &ErrorResponse{
				Error: "wallet not found",
				Code:  "WALLET_NOT_FOUND",
			},
		}
	}

	if domain.IsWalletAlreadyExists(err) {
		return &mappedError{
			status: http.StatusConflict,
			response: &ErrorResponse{
				Error: "wallet already exists",
				Code:  "WALLET_ALREADY_EXISTS",
			},
		}
	}

	// Address errors
	if domain.IsInvalidAddress(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "invalid wallet address",
				Code:  "INVALID_ADDRESS",
			},
		}
	}

	if domain.IsAddressMismatch(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "address mismatch",
				Code:  "ADDRESS_MISMATCH",
			},
		}
	}

	// Chain errors
	if domain.IsInvalidChain(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "invalid chain",
				Code:  "INVALID_CHAIN",
			},
		}
	}

	if domain.IsChainNotSupported(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "chain not supported",
				Code:  "CHAIN_NOT_SUPPORTED",
			},
		}
	}

	if domain.IsChainMismatch(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "chain mismatch",
				Code:  "CHAIN_MISMATCH",
			},
		}
	}

	// Signature errors
	if domain.IsInvalidSignature(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "invalid signature",
				Code:  "INVALID_SIGNATURE",
			},
		}
	}

	if domain.IsSignatureExpired(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "signature expired",
				Code:  "SIGNATURE_EXPIRED",
			},
		}
	}

	if domain.IsSignatureUsed(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "signature already used",
				Code:  "SIGNATURE_USED",
			},
		}
	}

	// Nonce errors
	if domain.IsInvalidNonce(err) {
		return &mappedError{
			status: http.StatusBadRequest,
			response: &ErrorResponse{
				Error: "invalid or expired nonce",
				Code:  "INVALID_NONCE",
			},
		}
	}

	// Verification errors
	if domain.IsVerificationFailed(err) {
		return &mappedError{
			status: http.StatusUnauthorized,
			response: &ErrorResponse{
				Error: "verification failed",
				Code:  "VERIFICATION_FAILED",
			},
		}
	}

	return nil
}

func mapPkgError(err error) *mappedError {
	var e *errors.Error
	if !errors.As(err, &e) {
		return nil
	}

	status := e.Kind.HTTPStatus()
	message := e.Message

	var code string
	var errorText string

	switch e.Kind {
	case errors.KindValidation:
		code = "VALIDATION_ERROR"
		errorText = "validation error"
	case errors.KindNotFound:
		code = "NOT_FOUND"
		errorText = "not found"
	case errors.KindConflict:
		code = "CONFLICT"
		errorText = "conflict"
	case errors.KindUnauthorized:
		code = "UNAUTHORIZED"
		errorText = "unauthorized"
	case errors.KindForbidden:
		code = "FORBIDDEN"
		errorText = "forbidden"
	case errors.KindRateLimit:
		code = "RATE_LIMIT_EXCEEDED"
		errorText = "rate limit exceeded"
	case errors.KindDomain:
		code = "DOMAIN_ERROR"
		errorText = "domain error"
	case errors.KindInfrastructure:
		code = "SERVICE_UNAVAILABLE"
		errorText = "service unavailable"
	case errors.KindTimeout:
		code = "TIMEOUT"
		errorText = "request timeout"
	case errors.KindInternal:
		code = "INTERNAL_ERROR"
		errorText = "internal server error"
		message = "" // Don't expose internal error details
	default:
		code = "INTERNAL_ERROR"
		errorText = "internal server error"
		message = ""
	}

	return &mappedError{
		status: status,
		response: &ErrorResponse{
			Error:   errorText,
			Code:    code,
			Message: message,
		},
	}
}

// ============================================================================
// Response Helpers
// ============================================================================

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, err error) {
	status, errResp := MapError(err)
	httpresponse.JSON(w, status, errResp)
}

// WriteValidationError writes a validation error response.
func WriteValidationError(w http.ResponseWriter, message string, details map[string]string) {
	httpresponse.JSON(w, http.StatusBadRequest, &ErrorResponse{
		Error:   "validation error",
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
	})
}

// WriteBadRequest writes a bad request error.
func WriteBadRequest(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusBadRequest, &ErrorResponse{
		Error:   "bad request",
		Code:    "BAD_REQUEST",
		Message: message,
	})
}

// WriteUnauthorized writes an unauthorized error.
func WriteUnauthorized(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusUnauthorized, &ErrorResponse{
		Error:   "unauthorized",
		Code:    "UNAUTHORIZED",
		Message: message,
	})
}

// WriteForbidden writes a forbidden error.
func WriteForbidden(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusForbidden, &ErrorResponse{
		Error:   "forbidden",
		Code:    "FORBIDDEN",
		Message: message,
	})
}

// WriteNotFound writes a not found error.
func WriteNotFound(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusNotFound, &ErrorResponse{
		Error:   "not found",
		Code:    "NOT_FOUND",
		Message: message,
	})
}

// WriteConflict writes a conflict error.
func WriteConflict(w http.ResponseWriter, message string) {
	httpresponse.JSON(w, http.StatusConflict, &ErrorResponse{
		Error:   "conflict",
		Code:    "CONFLICT",
		Message: message,
	})
}

// WriteInternalError writes an internal server error.
func WriteInternalError(w http.ResponseWriter) {
	httpresponse.JSON(w, http.StatusInternalServerError, &ErrorResponse{
		Error: "internal server error",
		Code:  "INTERNAL_ERROR",
	})
}

// ============================================================================
// Validation Helper
// ============================================================================

type validationErr struct {
	message string
}

func (e *validationErr) Error() string {
	return e.message
}

func validationError(message string) error {
	return &validationErr{message: message}
}
