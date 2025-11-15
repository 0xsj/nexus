package status

import (
	"github.com/0xsj/result"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPC converts a result.Error to a gRPC status error
func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	kind := result.KindOf(err)
	code := kindToCode(kind)
	message := err.Error()

	// Create gRPC status
	st := status.New(code, message)

	// TODO: Can add metadata as details here using google.rpc.ErrorInfo
	// if meta := result.MetaOf(err); meta != nil {
	//     st, _ = st.WithDetails(...)
	// }

	return st.Err()
}

// kindToCode maps result.Kind to gRPC codes.Code
func kindToCode(kind result.Kind) codes.Code {
	switch kind {
	case result.KindValidation:
		return codes.InvalidArgument
	case result.KindNotFound:
		return codes.NotFound
	case result.KindConflict:
		return codes.AlreadyExists
	case result.KindUnauthorized:
		return codes.Unauthenticated
	case result.KindForbidden:
		return codes.PermissionDenied
	case result.KindDomain:
		return codes.FailedPrecondition
	case result.KindInfrastructure:
		return codes.Unavailable
	case result.KindInternal:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// FromGRPC converts a gRPC status error to a result.Error
func FromGRPC(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return result.Internal("grpc.error", err)
	}

	kind := codeToKind(st.Code())

	return &result.Error{
		Kind:    kind,
		Message: st.Message(),
		Err:     err,
	}
}

// codeToKind maps gRPC codes.Code to result.Kind
func codeToKind(code codes.Code) result.Kind {
	switch code {
	case codes.InvalidArgument:
		return result.KindValidation
	case codes.NotFound:
		return result.KindNotFound
	case codes.AlreadyExists:
		return result.KindConflict
	case codes.Unauthenticated:
		return result.KindUnauthorized
	case codes.PermissionDenied:
		return result.KindForbidden
	case codes.FailedPrecondition:
		return result.KindDomain
	case codes.Unavailable:
		return result.KindInfrastructure
	default:
		return result.KindInternal
	}
}
