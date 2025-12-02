package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identitypb "github.com/0xsj/nexus-go/gen/proto/v1/identity"
	"github.com/0xsj/nexus-go/internal/identity/application/command"
	"github.com/0xsj/nexus-go/internal/identity/application/query"
	pkgerrors "github.com/0xsj/nexus-go/pkg/errors"
	"github.com/0xsj/nexus-go/pkg/observability/logger"
)

// IdentityServer implements the gRPC IdentityService.
type IdentityServer struct {
	identitypb.UnimplementedIdentityServiceServer

	// Commands
	loginCmd          *command.LoginCommand
	registerCmd       *command.RegisterUserCommand
	refreshTokenCmd   *command.RefreshTokenCommand
	logoutCmd         *command.LogoutCommand
	verifyEmailCmd    *command.VerifyEmailCommand
	changePasswordCmd *command.ChangePasswordCommand
	// TODO: add remaining commands

	// Queries
	getUserQuery        *query.GetUserQuery
	getCurrentUserQuery *query.GetCurrentUserQuery
	listUsersQuery      *query.ListUsersQuery
	listSessionsQuery   *query.ListSessionsQuery

	logger logger.Logger
}

// NewIdentityServer creates a new IdentityServer.
func NewIdentityServer(
	loginCmd *command.LoginCommand,
	registerCmd *command.RegisterUserCommand,
	// TODO: add remaining dependencies
	logger logger.Logger,
) *IdentityServer {
	return &IdentityServer{
		loginCmd:    loginCmd,
		registerCmd: registerCmd,
		logger:      logger,
	}
}

// ============================================================================
// Authentication
// ============================================================================

func (s *IdentityServer) Login(ctx context.Context, req *identitypb.LoginRequest) (*identitypb.LoginResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) Register(ctx context.Context, req *identitypb.RegisterRequest) (*identitypb.RegisterResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) RefreshToken(ctx context.Context, req *identitypb.RefreshTokenRequest) (*identitypb.RefreshTokenResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) Logout(ctx context.Context, req *identitypb.LogoutRequest) (*identitypb.LogoutResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// Email Verification
// ============================================================================

func (s *IdentityServer) VerifyEmail(ctx context.Context, req *identitypb.VerifyEmailRequest) (*identitypb.VerifyEmailResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ResendVerification(ctx context.Context, req *identitypb.ResendVerificationRequest) (*identitypb.ResendVerificationResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// Password Management
// ============================================================================

func (s *IdentityServer) RequestPasswordReset(ctx context.Context, req *identitypb.RequestPasswordResetRequest) (*identitypb.RequestPasswordResetResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ResetPassword(ctx context.Context, req *identitypb.ResetPasswordRequest) (*identitypb.ResetPasswordResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ChangePassword(ctx context.Context, req *identitypb.ChangePasswordRequest) (*identitypb.ChangePasswordResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// User Queries
// ============================================================================

func (s *IdentityServer) GetUser(ctx context.Context, req *identitypb.GetUserRequest) (*identitypb.GetUserResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) GetCurrentUser(ctx context.Context, req *identitypb.GetCurrentUserRequest) (*identitypb.GetCurrentUserResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ListUsers(ctx context.Context, req *identitypb.ListUsersRequest) (*identitypb.ListUsersResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// User Management (Admin)
// ============================================================================

func (s *IdentityServer) SuspendUser(ctx context.Context, req *identitypb.SuspendUserRequest) (*identitypb.SuspendUserResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ReactivateUser(ctx context.Context, req *identitypb.ReactivateUserRequest) (*identitypb.ReactivateUserResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) DeleteUser(ctx context.Context, req *identitypb.DeleteUserRequest) (*identitypb.DeleteUserResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) ChangeUserRole(ctx context.Context, req *identitypb.ChangeUserRoleRequest) (*identitypb.ChangeUserRoleResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// Session Management
// ============================================================================

func (s *IdentityServer) ListSessions(ctx context.Context, req *identitypb.ListSessionsRequest) (*identitypb.ListSessionsResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *IdentityServer) RevokeSession(ctx context.Context, req *identitypb.RevokeSessionRequest) (*identitypb.RevokeSessionResponse, error) {
	// TODO: implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ============================================================================
// Helpers
// ============================================================================

// toGRPCError converts domain errors to gRPC status errors.
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	appErr := pkgerrors.AsError(err)
	if appErr == nil {
		return status.Error(codes.Internal, err.Error())
	}

	var code codes.Code
	switch appErr.Kind {
	case pkgerrors.KindNotFound:
		code = codes.NotFound
	case pkgerrors.KindValidation:
		code = codes.InvalidArgument
	case pkgerrors.KindUnauthorized:
		code = codes.Unauthenticated
	case pkgerrors.KindForbidden:
		code = codes.PermissionDenied
	case pkgerrors.KindConflict:
		code = codes.AlreadyExists
	case pkgerrors.KindRateLimit:
		code = codes.ResourceExhausted
	default:
		code = codes.Internal
	}

	return status.Error(code, appErr.Message)
}
