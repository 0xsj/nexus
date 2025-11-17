package application

import (
	"time"

	"github.com/0xsj/nexus/internal/users/domain"
)

// UserDTO is the application-layer representation of a user.
type UserDTO struct {
	ID            string
	Email         string
	Username      string
	DisplayName   *string
	AvatarURL     *string
	Bio           *string
	EmailVerified bool
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// FromDomain converts a domain User to a UserDTO.
func FromDomain(user *domain.User) UserDTO {
	return UserDTO{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		AvatarURL:     user.AvatarURL,
		Bio:           user.Bio,
		EmailVerified: user.EmailVerified,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
