package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID            string
	Email         string
	Username      string
	PasswordHash  string
	EmailVerified bool
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Optional profile fields (Nexus-specific)
	DisplayName *string
	AvatarURL   *string
	Bio         *string
}

// NewUser creates a new User entity with password.
func NewUser(email Email, username Username, password Password) *User {
	now := time.Now().UTC()

	return &User{
		ID:            uuid.New().String(),
		Email:         email.Value(),
		Username:      username.Value(),
		PasswordHash:  password.Hash(),
		EmailVerified: false,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// NewUserWithoutPassword creates a new user without a password (passwordless auth).
// Used for magic link and OAuth authentication where no password is required.
func NewUserWithoutPassword(email Email, username Username) *User {
	now := time.Now().UTC()

	return &User{
		ID:            uuid.New().String(),
		Email:         email.Value(),
		Username:      username.Value(),
		PasswordHash:  "", // No password for passwordless auth
		EmailVerified: false,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// VerifyEmail marks the user's email as verified.
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now().UTC()
}

// UpdateEmail updates the user's email address.
func (u *User) UpdateEmail(email Email) {
	u.Email = email.Value()
	u.EmailVerified = false // Reset verification when email changes
	u.UpdatedAt = time.Now().UTC()
}

// UpdateUsername updates the user's username.
func (u *User) UpdateUsername(username Username) {
	u.Username = username.Value()
	u.UpdatedAt = time.Now().UTC()
}

// ChangePassword updates the user's password.
func (u *User) ChangePassword(newPassword Password) {
	u.PasswordHash = newPassword.Hash()
	u.UpdatedAt = time.Now().UTC()
}

// UpdateProfile updates the user's profile information.
func (u *User) UpdateProfile(displayName, avatarURL, bio *string) {
	if displayName != nil {
		u.DisplayName = displayName
	}
	if avatarURL != nil {
		u.AvatarURL = avatarURL
	}
	if bio != nil {
		u.Bio = bio
	}
	u.UpdatedAt = time.Now().UTC()
}

// Deactivate marks the user as inactive (soft delete).
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now().UTC()
}

// Activate marks the user as active.
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now().UTC()
}

// IsEmailVerified returns true if the user's email is verified.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// CanLogin returns true if the user can log in.
func (u *User) CanLogin() bool {
	return u.IsActive
}

// RequiresEmailVerification returns true if email verification is required.
func (u *User) RequiresEmailVerification() bool {
	return !u.EmailVerified
}

// VerifyPassword checks if the provided password matches the user's password.
func (u *User) VerifyPassword(plainPassword string) bool {
	if u.PasswordHash == "" {
		return false // Passwordless accounts can't use password auth
	}

	password := NewPasswordFromHash(u.PasswordHash)
	return password.Compare(plainPassword).UnwrapOr(false)
}
