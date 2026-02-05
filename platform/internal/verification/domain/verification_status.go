package domain

import (
	"fmt"
)

// VerificationStatus represents the current status of a verification attempt.
type VerificationStatus int

const (
	// VerificationStatusPending indicates the verification has been initiated.
	VerificationStatusPending VerificationStatus = 1

	// VerificationStatusOAuthStarted indicates the user has been redirected to the provider.
	VerificationStatusOAuthStarted VerificationStatus = 2

	// VerificationStatusOAuthCompleted indicates the OAuth callback has been received.
	VerificationStatusOAuthCompleted VerificationStatus = 3

	// VerificationStatusDataFetched indicates provider data has been retrieved.
	VerificationStatusDataFetched VerificationStatus = 4

	// VerificationStatusCredentialIssued indicates a credential has been issued.
	VerificationStatusCredentialIssued VerificationStatus = 5

	// VerificationStatusFailed indicates the verification has failed.
	VerificationStatusFailed VerificationStatus = 6
)

// String returns the string representation of the verification status.
func (s VerificationStatus) String() string {
	switch s {
	case VerificationStatusPending:
		return "pending"
	case VerificationStatusOAuthStarted:
		return "oauth_started"
	case VerificationStatusOAuthCompleted:
		return "oauth_completed"
	case VerificationStatusDataFetched:
		return "data_fetched"
	case VerificationStatusCredentialIssued:
		return "credential_issued"
	case VerificationStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ParseVerificationStatus parses a string into a VerificationStatus.
func ParseVerificationStatus(str string) (VerificationStatus, error) {
	switch str {
	case "pending":
		return VerificationStatusPending, nil
	case "oauth_started":
		return VerificationStatusOAuthStarted, nil
	case "oauth_completed":
		return VerificationStatusOAuthCompleted, nil
	case "data_fetched":
		return VerificationStatusDataFetched, nil
	case "credential_issued":
		return VerificationStatusCredentialIssued, nil
	case "failed":
		return VerificationStatusFailed, nil
	default:
		return 0, fmt.Errorf("invalid verification status: %s", str)
	}
}

// IsValid returns true if the status is a known, valid status.
func (s VerificationStatus) IsValid() bool {
	switch s {
	case VerificationStatusPending, VerificationStatusOAuthStarted,
		VerificationStatusOAuthCompleted, VerificationStatusDataFetched,
		VerificationStatusCredentialIssued, VerificationStatusFailed:
		return true
	default:
		return false
	}
}

// IsPending returns true if the verification is pending.
func (s VerificationStatus) IsPending() bool {
	return s == VerificationStatusPending
}

// IsOAuthStarted returns true if the OAuth flow has started.
func (s VerificationStatus) IsOAuthStarted() bool {
	return s == VerificationStatusOAuthStarted
}

// IsOAuthCompleted returns true if the OAuth callback has been received.
func (s VerificationStatus) IsOAuthCompleted() bool {
	return s == VerificationStatusOAuthCompleted
}

// IsDataFetched returns true if provider data has been fetched.
func (s VerificationStatus) IsDataFetched() bool {
	return s == VerificationStatusDataFetched
}

// IsCredentialIssued returns true if a credential has been issued.
func (s VerificationStatus) IsCredentialIssued() bool {
	return s == VerificationStatusCredentialIssued
}

// IsFailed returns true if the verification has failed.
func (s VerificationStatus) IsFailed() bool {
	return s == VerificationStatusFailed
}

// IsTerminal returns true if the status is terminal (CredentialIssued or Failed).
func (s VerificationStatus) IsTerminal() bool {
	return s == VerificationStatusCredentialIssued || s == VerificationStatusFailed
}

// CanTransitionTo returns true if transitioning to the target status is allowed.
// Linear progression: Pending -> OAuthStarted -> OAuthCompleted -> DataFetched -> CredentialIssued.
// Failed can be reached from any non-terminal status.
func (s VerificationStatus) CanTransitionTo(target VerificationStatus) bool {
	// Cannot transition from terminal states.
	if s.IsTerminal() {
		return false
	}

	// Any non-terminal status can transition to Failed.
	if target == VerificationStatusFailed {
		return true
	}

	// Linear progression only.
	switch s {
	case VerificationStatusPending:
		return target == VerificationStatusOAuthStarted
	case VerificationStatusOAuthStarted:
		return target == VerificationStatusOAuthCompleted
	case VerificationStatusOAuthCompleted:
		return target == VerificationStatusDataFetched
	case VerificationStatusDataFetched:
		return target == VerificationStatusCredentialIssued
	default:
		return false
	}
}
