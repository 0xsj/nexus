package domain

import "time"

// Reputation represents a calculated reputation score for a user.
// This is a simple data struct, not an aggregate.
type Reputation struct {
	UserID           string
	OverallScore     int // 0-100
	CredentialScore  int // 0-40
	VouchScore       int // 0-40
	NetworkScore     int // 0-20
	VouchCount       int
	LastCalculatedAt time.Time
}
