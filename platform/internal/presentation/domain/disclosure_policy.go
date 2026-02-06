package domain

import (
	"fmt"
)

// Valid disclosure policy types.
const (
	PolicyTypeAll       = "all"
	PolicyTypeAllowlist = "allowlist"
	PolicyTypeBlocklist = "blocklist"
)

// DisclosurePolicy is an immutable value object that defines which claims
// are included or excluded from a presentation.
type DisclosurePolicy struct {
	policyType    string
	allowedClaims []string
	blockedClaims []string
}

// NewDisclosurePolicy creates a new DisclosurePolicy value object.
// Returns an error if the policyType is invalid.
func NewDisclosurePolicy(policyType string, allowedClaims, blockedClaims []string) (DisclosurePolicy, error) {
	switch policyType {
	case PolicyTypeAll:
		// No claims needed
	case PolicyTypeAllowlist:
		if len(allowedClaims) == 0 {
			return DisclosurePolicy{}, fmt.Errorf("allowlist policy requires at least one allowed claim")
		}
	case PolicyTypeBlocklist:
		if len(blockedClaims) == 0 {
			return DisclosurePolicy{}, fmt.Errorf("blocklist policy requires at least one blocked claim")
		}
	default:
		return DisclosurePolicy{}, fmt.Errorf("invalid disclosure policy type: %s", policyType)
	}

	// Defensive copies
	var allowed []string
	if len(allowedClaims) > 0 {
		allowed = make([]string, len(allowedClaims))
		copy(allowed, allowedClaims)
	}

	var blocked []string
	if len(blockedClaims) > 0 {
		blocked = make([]string, len(blockedClaims))
		copy(blocked, blockedClaims)
	}

	return DisclosurePolicy{
		policyType:    policyType,
		allowedClaims: allowed,
		blockedClaims: blocked,
	}, nil
}

// PolicyType returns the disclosure policy type.
func (dp DisclosurePolicy) PolicyType() string {
	return dp.policyType
}

// AllowedClaims returns a copy of the allowed claims list.
func (dp DisclosurePolicy) AllowedClaims() []string {
	if dp.allowedClaims == nil {
		return nil
	}
	copied := make([]string, len(dp.allowedClaims))
	copy(copied, dp.allowedClaims)
	return copied
}

// BlockedClaims returns a copy of the blocked claims list.
func (dp DisclosurePolicy) BlockedClaims() []string {
	if dp.blockedClaims == nil {
		return nil
	}
	copied := make([]string, len(dp.blockedClaims))
	copy(copied, dp.blockedClaims)
	return copied
}

// ToMap returns the disclosure policy as a map for JSON serialization.
func (dp DisclosurePolicy) ToMap() map[string]any {
	m := map[string]any{
		"policy_type": dp.policyType,
	}
	if len(dp.allowedClaims) > 0 {
		m["allowed_claims"] = dp.AllowedClaims()
	}
	if len(dp.blockedClaims) > 0 {
		m["blocked_claims"] = dp.BlockedClaims()
	}
	return m
}

// IsZero returns true if the DisclosurePolicy is the zero value.
func (dp DisclosurePolicy) IsZero() bool {
	return dp.policyType == ""
}

// DisclosurePolicyFromMap reconstructs a DisclosurePolicy from a map.
// Used when replaying events from the event store.
func DisclosurePolicyFromMap(m map[string]any) (DisclosurePolicy, error) {
	if m == nil {
		return DisclosurePolicy{}, fmt.Errorf("disclosure policy map cannot be nil")
	}

	policyType, ok := m["policy_type"].(string)
	if !ok {
		return DisclosurePolicy{}, fmt.Errorf("disclosure policy map missing policy_type")
	}

	var allowedClaims []string
	if raw, ok := m["allowed_claims"]; ok {
		if arr, ok := raw.([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allowedClaims = append(allowedClaims, s)
				}
			}
		}
	}

	var blockedClaims []string
	if raw, ok := m["blocked_claims"]; ok {
		if arr, ok := raw.([]any); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					blockedClaims = append(blockedClaims, s)
				}
			}
		}
	}

	return NewDisclosurePolicy(policyType, allowedClaims, blockedClaims)
}
