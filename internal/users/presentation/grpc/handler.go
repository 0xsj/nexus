package grpc

import (
	"context"
	"time"

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

// CreateUser creates a new user (stub - returns fake response).
func (h *UserHandler) CreateUser(ctx context.Context, req *usersv1.CreateUserRequest) (*usersv1.CreateUserResponse, error) {
	h.logger.Info("CreateUser called",
		logger.String("email", req.Email),
		logger.String("username", req.Username),
	)

	// Return stub response
	now := timestamppb.Now()
	return &usersv1.CreateUserResponse{
		User: &usersv1.User{
			Id:            "user-stub-123",
			Email:         req.Email,
			Username:      req.Username,
			EmailVerified: false,
			IsActive:      true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		Metadata: &commonv1.ResponseMetadata{
			RequestId:  "req-stub-123",
			ServerTime: now,
		},
	}, nil
}

// GetUser retrieves a user by ID (stub - returns fake response).
func (h *UserHandler) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	h.logger.Info("GetUser called",
		logger.String("user_id", req.UserId),
	)

	// Return stub response
	now := timestamppb.Now()
	return &usersv1.GetUserResponse{
		User: &usersv1.User{
			Id:            req.UserId,
			Email:         "stub@example.com",
			Username:      "stubuser",
			EmailVerified: true,
			IsActive:      true,
			CreatedAt:     timestamppb.New(time.Now().Add(-24 * time.Hour)),
			UpdatedAt:     now,
		},
		Metadata: &commonv1.ResponseMetadata{
			RequestId:  "req-stub-123",
			ServerTime: now,
		},
	}, nil
}

// UpdateUser updates a user (stub).
func (h *UserHandler) UpdateUser(ctx context.Context, req *usersv1.UpdateUserRequest) (*usersv1.UpdateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// DeleteUser deletes a user (stub).
func (h *UserHandler) DeleteUser(ctx context.Context, req *usersv1.DeleteUserRequest) (*usersv1.DeleteUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// ListUsers lists users (stub).
func (h *UserHandler) ListUsers(ctx context.Context, req *usersv1.ListUsersRequest) (*usersv1.ListUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
