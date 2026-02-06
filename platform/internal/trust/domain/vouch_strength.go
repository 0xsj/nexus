package domain

import (
	"fmt"
)

// VouchStrength represents the strength of a vouch on a scale of 1-10.
type VouchStrength int

// NewVouchStrength creates a new VouchStrength, validating the value is between 1 and 10.
func NewVouchStrength(val int) (VouchStrength, error) {
	s := VouchStrength(val)
	if !s.IsValid() {
		return 0, fmt.Errorf("vouch strength must be between 1 and 10, got %d", val)
	}
	return s, nil
}

// Value returns the integer value of the VouchStrength.
func (s VouchStrength) Value() int {
	return int(s)
}

// IsValid returns true if the VouchStrength is between 1 and 10.
func (s VouchStrength) IsValid() bool {
	return int(s) >= 1 && int(s) <= 10
}
