package v1

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	identitypb "github.com/0xsj/nexus-go/gen/proto/v1/identity"
	"github.com/0xsj/nexus-go/internal/identity/application/dto"
)

// ============================================================================
// User Mappers
// ============================================================================

// UserDTOToProto converts a UserDTO to proto User.
func UserDTOToProto(u *dto.UserDTO) *identitypb.User {
	// TODO: implement
	return nil
}

// UserSummaryDTOsToProto converts a slice of UserSummaryDTO to proto UserSummaries.
func UserSummaryDTOsToProto(users []*dto.UserSummaryDTO) []*identitypb.UserSummary {
	// TODO: implement
	return nil
}

// ============================================================================
// Session Mappers
// ============================================================================

// SessionDTOsToSummaryProtos converts a slice of SessionDTO to proto SessionSummaries.
func SessionDTOsToSummaryProtos(sessions []*dto.SessionDTO, currentSessionID string) []*identitypb.SessionSummary {
	// TODO: implement
	return nil
}

// ============================================================================
// Enum Mappers
// ============================================================================

// UserStatusStringToProto converts a status string to proto UserStatus.
func UserStatusStringToProto(s string) identitypb.UserStatus {
	// TODO: implement
	return identitypb.UserStatus_USER_STATUS_UNSPECIFIED
}

// UserRoleStringToProto converts a role string to proto UserRole.
func UserRoleStringToProto(r string) identitypb.UserRole {
	// TODO: implement
	return identitypb.UserRole_USER_ROLE_UNSPECIFIED
}

// ProtoToUserStatusString converts proto UserStatus to string.
func ProtoToUserStatusString(s identitypb.UserStatus) string {
	// TODO: implement
	return ""
}

// ProtoToUserRoleString converts proto UserRole to string.
func ProtoToUserRoleString(r identitypb.UserRole) string {
	// TODO: implement
	return ""
}

// ============================================================================
// Token Helpers
// ============================================================================

// TokenPairToProto creates a proto TokenPair.
func TokenPairToProto(accessToken, refreshToken string, expiresIn int) *identitypb.TokenPair {
	return &identitypb.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int32(expiresIn),
		TokenType:    "Bearer",
	}
}

// ============================================================================
// Timestamp Helpers
// ============================================================================

// TimestampToProto converts time.Time to proto Timestamp.
func TimestampToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// ProtoToTimestampPtr converts proto Timestamp to *time.Time.
func ProtoToTimestampPtr(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	result := t.AsTime()
	return &result
}
