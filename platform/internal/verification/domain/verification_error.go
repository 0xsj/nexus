package domain

import "fmt"

// VerificationError is a value object that captures error details
// when a verification attempt fails.
type VerificationError struct {
	code    string
	message string
	details map[string]any
}

// NewVerificationError creates a new VerificationError.
// A defensive copy is made of the details map.
func NewVerificationError(code, message string, details map[string]any) VerificationError {
	var detailsCopy map[string]any
	if details != nil {
		detailsCopy = make(map[string]any, len(details))
		for k, v := range details {
			detailsCopy[k] = v
		}
	}

	return VerificationError{
		code:    code,
		message: message,
		details: detailsCopy,
	}
}

// Code returns the error code.
func (e VerificationError) Code() string {
	return e.code
}

// Message returns the error message.
func (e VerificationError) Message() string {
	return e.message
}

// Details returns a defensive copy of the error details.
func (e VerificationError) Details() map[string]any {
	if e.details == nil {
		return nil
	}
	result := make(map[string]any, len(e.details))
	for k, v := range e.details {
		result[k] = v
	}
	return result
}

// String returns a string representation of the verification error.
func (e VerificationError) String() string {
	if e.IsZero() {
		return ""
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// IsZero returns true if the VerificationError is the zero value.
func (e VerificationError) IsZero() bool {
	return e.code == "" && e.message == ""
}
