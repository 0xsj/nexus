package domain

// RelationshipType represents the type of social relationship.
type RelationshipType string

const (
	RelationshipTypeFollowing RelationshipType = "following"
	RelationshipTypeBlocking  RelationshipType = "blocking"
	RelationshipTypeMuting    RelationshipType = "muting"
)

// IsValid checks if the relationship type is valid.
func (rt RelationshipType) IsValid() bool {
	switch rt {
	case RelationshipTypeFollowing, RelationshipTypeBlocking, RelationshipTypeMuting:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (rt RelationshipType) String() string {
	return string(rt)
}

// RelationshipStatus represents the mutual relationship status between two users.
type RelationshipStatus string

const (
	StatusNotFollowing    RelationshipStatus = "not_following"
	StatusFollowing       RelationshipStatus = "following"
	StatusFollowedBy      RelationshipStatus = "followed_by"
	StatusMutualFollowing RelationshipStatus = "mutual"
	StatusBlocked         RelationshipStatus = "blocked"
	StatusBlockedBy       RelationshipStatus = "blocked_by"
)

// String returns the string representation.
func (rs RelationshipStatus) String() string {
	return string(rs)
}
