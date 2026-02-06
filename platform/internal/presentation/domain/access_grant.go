package domain

import "time"

// AccessGrant is a simple data type (not an aggregate) that records
// each access to a presentation via a share link.
type AccessGrant struct {
	ShareLinkID     string
	VerifierDID     string
	AccessedAt      time.Time
	IPAddress       string
	DisclosedClaims map[string]any
}
