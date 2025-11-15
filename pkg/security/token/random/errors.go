package random

import "fmt"

// ErrInvalidLength is returned when token length is invalid.
type ErrInvalidLength struct {
	Length int
}

func (e ErrInvalidLength) Error() string {
	return fmt.Sprintf("invalid token length: %d (must be > 0)", e.Length)
}

// ErrGenerationFailed is returned when token generation fails.
type ErrGenerationFailed struct {
	Err error
}

func (e ErrGenerationFailed) Error() string {
	return fmt.Sprintf("token generation failed: %v", e.Err)
}

func (e ErrGenerationFailed) Unwrap() error {
	return e.Err
}
