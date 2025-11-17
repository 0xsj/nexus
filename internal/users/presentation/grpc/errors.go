package grpc

import (
	"github.com/0xsj/result"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mapErrorToGRPC maps domain errors to gRPC status codes.
func mapErrorToGRPC(err error) error {
	// Use result.MatchKind for mapping
	var grpcErr error

	result.Err[any](err).MatchKind(
		map[result.Kind]func(error){
			result.KindValidation: func(e error) {
				grpcErr = status.Error(codes.InvalidArgument, e.Error())
			},
			result.KindNotFound: func(e error) {
				grpcErr = status.Error(codes.NotFound, "user not found")
			},
			result.KindConflict: func(e error) {
				grpcErr = status.Error(codes.AlreadyExists, e.Error())
			},
			result.KindDomain: func(e error) {
				grpcErr = mapDomainError(e)
			},
		},
		func(e error) {
			grpcErr = status.Error(codes.Internal, "internal server error")
		},
	)

	return grpcErr
}

// mapDomainError maps specific domain errors to gRPC codes.
func mapDomainError(err error) error {
	// Check for specific domain error types
	switch err.(type) {
	default:
		return status.Error(codes.FailedPrecondition, err.Error())
	}
}
