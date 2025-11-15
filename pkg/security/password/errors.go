package password

import (
	"errors"
	"fmt"
)

var (
	ErrPasswordTooShort   = errors.New("password is too short")
	ErrPasswordTooWeak    = errors.New("password is too weak")
	ErrMissingUppercase   = errors.New("password must contain at least one uppercase letter")
	ErrMissingLowercase   = errors.New("password must contain at least one lowercase letter")
	ErrMissingNumber      = errors.New("password must contain at least one number")
	ErrMissingSpecialChar = errors.New("password must contain at least one special character")
	ErrPasswordMismatch   = errors.New("password does not match")
	ErrEmptyPassword      = errors.New("password cannot be empty")
)

// ErrHashingFailed is returned when password hashing fails.
type ErrHashingFailed struct {
	Err error
}

func (e ErrHashingFailed) Error() string {
	return fmt.Sprintf("password hashing failed: %v", e.Err)
}

func (e ErrHashingFailed) Unwrap() error {
	return e.Err
}

// ErrInvalidCost is returned when bcrypt cost is invalid.
type ErrInvalidCost struct {
	Cost int
}

func (e ErrInvalidCost) Error() string {
	return fmt.Sprintf("invalid bcrypt cost: %d (must be between 4 and 31)", e.Cost)
}
