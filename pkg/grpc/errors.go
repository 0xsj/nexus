// pkg/grpc/errors.go

package grpc

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/0xsj/nexus-go/pkg/errors"
)

// ============================================================================
// Kind to gRPC Code Mapping
// ============================================================================

// KindToCode maps a domain error Kind to a gRPC status code.
//
// Mapping follows gRPC best practices:
//   - KindNotFound      → NotFound (5)
//   - KindValidation    → InvalidArgument (3)
//   - KindConflict      → AlreadyExists (6)
//   - KindUnauthorized  → Unauthenticated (16)
//   - KindForbidden     → PermissionDenied (7)
//   - KindDomain        → FailedPrecondition (9)
//   - KindInfrastructure → Unavailable (14)
//   - KindTimeout       → DeadlineExceeded (4)
//   - KindRateLimit     → ResourceExhausted (8)
//   - KindInternal      → Internal (13)
//   - KindOther         → Unknown (2)
func KindToCode(kind errors.Kind) codes.Code {
	switch kind {
	case errors.KindNotFound:
		return codes.NotFound
	case errors.KindValidation:
		return codes.InvalidArgument
	case errors.KindConflict:
		return codes.AlreadyExists
	case errors.KindUnauthorized:
		return codes.Unauthenticated
	case errors.KindForbidden:
		return codes.PermissionDenied
	case errors.KindDomain:
		return codes.FailedPrecondition
	case errors.KindInfrastructure:
		return codes.Unavailable
	case errors.KindTimeout:
		return codes.DeadlineExceeded
	case errors.KindRateLimit:
		return codes.ResourceExhausted
	case errors.KindInternal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// CodeToKind maps a gRPC status code to a domain error Kind.
// Used when receiving errors from gRPC services.
func CodeToKind(code codes.Code) errors.Kind {
	switch code {
	case codes.NotFound:
		return errors.KindNotFound
	case codes.InvalidArgument:
		return errors.KindValidation
	case codes.AlreadyExists:
		return errors.KindConflict
	case codes.Unauthenticated:
		return errors.KindUnauthorized
	case codes.PermissionDenied:
		return errors.KindForbidden
	case codes.FailedPrecondition:
		return errors.KindDomain
	case codes.Unavailable:
		return errors.KindInfrastructure
	case codes.DeadlineExceeded:
		return errors.KindTimeout
	case codes.Canceled:
		return errors.KindTimeout
	case codes.ResourceExhausted:
		return errors.KindRateLimit
	case codes.Internal:
		return errors.KindInternal
	default:
		return errors.KindOther
	}
}

// ============================================================================
// Domain Error to gRPC Status Conversion
// ============================================================================

// ToStatus converts a domain error to a gRPC status with rich details.
//
// The resulting status includes:
//   - Appropriate gRPC code based on error Kind
//   - Human-readable message
//   - ErrorInfo detail with domain, code, and metadata
//   - RetryInfo detail if error is retryable
//
// Example:
//
//	err := errors.NotFound("UserRepository.FindByID", "user not found")
//	st := grpc.ToStatus(err)
//	return nil, st.Err()
func ToStatus(err error) *status.Status {
	if err == nil {
		return status.New(codes.OK, "")
	}

	// Check if already a gRPC status error
	if st, ok := status.FromError(err); ok {
		return st
	}

	// Extract domain error details
	domainErr := errors.AsError(err)
	if domainErr == nil {
		// Wrap non-domain error as internal
		return status.New(codes.Internal, err.Error())
	}

	// Create base status
	code := KindToCode(domainErr.Kind)
	st := status.New(code, domainErr.Message)

	// Attach error details
	stWithDetails := attachErrorDetails(st, domainErr)

	return stWithDetails
}

// attachErrorDetails attaches rich error details to a gRPC status.
func attachErrorDetails(st *status.Status, err *errors.Error) *status.Status {
	// Build ErrorInfo
	errorInfo := &errdetails.ErrorInfo{
		Reason:   string(err.Code),
		Domain:   "nexus",
		Metadata: make(map[string]string),
	}

	if err.Operation != "" {
		errorInfo.Metadata["operation"] = err.Operation
	}

	if err.Severity != 0 {
		errorInfo.Metadata["severity"] = err.Severity.String()
	}

	for k, v := range err.Metadata {
		if str, ok := v.(string); ok {
			errorInfo.Metadata[k] = str
		}
	}

	// Attach ErrorInfo
	stWithDetails, detailErr := st.WithDetails(errorInfo)
	if detailErr != nil {
		return st
	}

	// Attach RetryInfo if retryable
	if err.Retryable {
		retryInfo := &errdetails.RetryInfo{}
		if updated, e := stWithDetails.WithDetails(retryInfo); e == nil {
			stWithDetails = updated
		}
	}

	// Attach BadRequest for validation errors
	if err.Kind == errors.KindValidation {
		if badRequest := buildBadRequest(err); badRequest != nil {
			if updated, e := stWithDetails.WithDetails(badRequest); e == nil {
				stWithDetails = updated
			}
		}
	}

	return stWithDetails
}

// buildBadRequest creates a BadRequest detail for validation errors.
func buildBadRequest(err *errors.Error) *errdetails.BadRequest {
	var violations []*errdetails.BadRequest_FieldViolation

	// Check for "field" in metadata
	if field, ok := err.Metadata["field"].(string); ok {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: err.Message,
		})
	}

	// Check for "fields" map in metadata
	if fields, ok := err.Metadata["fields"].(map[string]string); ok {
		for field, desc := range fields {
			violations = append(violations, &errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: desc,
			})
		}
	}

	if len(violations) == 0 {
		return nil
	}

	return &errdetails.BadRequest{
		FieldViolations: violations,
	}
}

// ============================================================================
// gRPC Status to Domain Error Conversion
// ============================================================================

// FromStatus converts a gRPC status to a domain error.
// Used when receiving errors from gRPC services.
//
// Example:
//
//	resp, err := client.GetUser(ctx, req)
//	if err != nil {
//	    domainErr := grpc.FromStatus(status.Convert(err))
//	    // Handle domain error
//	}
func FromStatus(st *status.Status) *errors.Error {
	if st == nil || st.Code() == codes.OK {
		return nil
	}

	// Create base error
	kind := CodeToKind(st.Code())
	err := &errors.Error{
		Kind:     kind,
		Message:  st.Message(),
		Metadata: make(map[string]any),
	}

	// Extract details
	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			err.Code = errors.Code(d.Reason)
			if op, ok := d.Metadata["operation"]; ok {
				err.Operation = op
			}
			for k, v := range d.Metadata {
				if k != "operation" && k != "severity" {
					err.Metadata[k] = v
				}
			}

		case *errdetails.RetryInfo:
			err.Retryable = true

		case *errdetails.BadRequest:
			if len(d.FieldViolations) > 0 {
				fields := make(map[string]string)
				for _, v := range d.FieldViolations {
					fields[v.Field] = v.Description
				}
				err.Metadata["fields"] = fields
			}
		}
	}

	return err
}

// ============================================================================
// Helper Functions
// ============================================================================

// Error converts an error to a gRPC error.
// Convenience function for use in gRPC handlers.
//
// Example:
//
//	func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
//	    user, err := s.query.Handle(ctx, req.Id)
//	    if err != nil {
//	        return nil, grpc.Error(err)
//	    }
//	    return toProto(user), nil
//	}
func Error(err error) error {
	if err == nil {
		return nil
	}
	return ToStatus(err).Err()
}

// IsRetryable checks if a gRPC error indicates a retryable condition.
func IsRetryable(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// Check for RetryInfo in details
	for _, detail := range st.Details() {
		if _, ok := detail.(*errdetails.RetryInfo); ok {
			return true
		}
	}

	// Fall back to code-based check
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// Code extracts the gRPC status code from an error.
func Code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	st, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return st.Code()
}

// Message extracts the message from a gRPC error.
func Message(err error) string {
	if err == nil {
		return ""
	}
	st, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}
	return st.Message()
}
