package domain

// PostVisibility represents who can see a post.
type PostVisibility string

const (
	VisibilityPublic    PostVisibility = "public"
	VisibilityFollowers PostVisibility = "followers"
	VisibilityPrivate   PostVisibility = "private"
)

// IsValid checks if the visibility setting is valid.
func (v PostVisibility) IsValid() bool {
	switch v {
	case VisibilityPublic, VisibilityFollowers, VisibilityPrivate:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (v PostVisibility) String() string {
	return string(v)
}

// IsPublic checks if the post is publicly visible.
func (v PostVisibility) IsPublic() bool {
	return v == VisibilityPublic
}

// IsPrivate checks if the post is private.
func (v PostVisibility) IsPrivate() bool {
	return v == VisibilityPrivate
}

// IsFollowersOnly checks if the post is visible to followers only.
func (v PostVisibility) IsFollowersOnly() bool {
	return v == VisibilityFollowers
}
