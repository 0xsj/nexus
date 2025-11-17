package grpc

import (
	"context"

	commonv1 "github.com/0xsj/nexus/api/common/v1"
	usersv1 "github.com/0xsj/nexus/api/users/v1"
	"github.com/0xsj/nexus/internal/users/application/commands"
	"github.com/0xsj/nexus/internal/users/application/queries"
	"github.com/0xsj/nexus/pkg/observability/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserHandler implements the gRPC UserService.
type UserHandler struct {
	usersv1.UnimplementedUserServiceServer
	createUserHandler *commands.CreateUserHandler
	getUserHandler    *queries.GetUserHandler
	logger            logger.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(
	createUserHandler *commands.CreateUserHandler,
	getUserHandler *queries.GetUserHandler,
	log logger.Logger,
) *UserHandler {
	return &UserHandler{
		createUserHandler: createUserHandler,
		getUserHandler:    getUserHandler,
		logger:            log,
	}
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	// Map request to command
	cmd := commands.CreateUserCommand{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// Execute command
	result := h.createUserHandler.Handle(ctx, cmd)

	// Handle result
	if result.IsErr() {
		return nil, mapErrorToGRPC(result.UnwrapErr())
	}

	dto := result.Unwrap()

	// Map DTO to protobuf
	return &usersv1.CreateUserResponse{
		User: &usersv1.User{
			Id:            dto.ID,
			Email:         dto.Email,
			Username:      dto.Username,
			DisplayName:   dto.DisplayName,
			AvatarUrl:     dto.AvatarURL,
			Bio:           dto.Bio,
			EmailVerified: dto.EmailVerified,
			IsActive:      dto.IsActive,
			CreatedAt:     timestamppb.New(dto.CreatedAt),
			UpdatedAt:     timestamppb.New(dto.UpdatedAt),
		},
		Metadata: &commonv1.ResponseMetadata{
			RequestId:  "req-123", // TODO: extract from context
			ServerTime: timestamppb.Now(),
		},
	}, nil
}

// GetUser retrieves a user by ID.
func (h *UserHandler) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	// Map request to query
	query := queries.GetUserQuery{
		UserID: req.UserId,
	}

	// Execute query
	result := h.getUserHandler.Handle(ctx, query)

	// Handle result
	if result.IsErr() {
		return nil, mapErrorToGRPC(result.UnwrapErr())
	}

	dto := result.Unwrap()

	// Map DTO to protobuf
	return &usersv1.GetUserResponse{
		User: &usersv1.User{
			Id:            dto.ID,
			Email:         dto.Email,
			Username:      dto.Username,
			DisplayName:   dto.DisplayName,
			AvatarUrl:     dto.AvatarURL,
			Bio:           dto.Bio,
			EmailVerified: dto.EmailVerified,
			IsActive:      dto.IsActive,
			CreatedAt:     timestamppb.New(dto.CreatedAt),
			UpdatedAt:     timestamppb.New(dto.UpdatedAt),
		},
		Metadata: &commonv1.ResponseMetadata{
			RequestId:  "req-123",
			ServerTime: timestamppb.Now(),
		},
	}, nil
}

// UpdateUser updates a user (stub).
func (h *UserHandler) UpdateUser(ctx context.Context, req *usersv1.UpdateUserRequest) (*usersv1.UpdateUserResponse, error) {
	// TODO: Implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// DeleteUser deletes a user (stub).
func (h *UserHandler) DeleteUser(ctx context.Context, req *usersv1.DeleteUserRequest) (*usersv1.DeleteUserResponse, error) {
	// TODO: Implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ListUsers lists users (stub).
func (h *UserHandler) ListUsers(ctx context.Context, req *usersv1.ListUsersRequest) (*usersv1.ListUsersResponse, error) {
	// TODO: Implement
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
