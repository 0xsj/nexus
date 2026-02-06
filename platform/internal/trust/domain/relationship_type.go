package domain

import (
	"fmt"
)

// RelationshipType represents the type of relationship between voucher and vouchee.
type RelationshipType string

const (
	// RelationshipColleague indicates the voucher and vouchee worked together at the same level.
	RelationshipColleague RelationshipType = "colleague"

	// RelationshipManager indicates the voucher managed the vouchee.
	RelationshipManager RelationshipType = "manager"

	// RelationshipDirectReport indicates the vouchee managed the voucher.
	RelationshipDirectReport RelationshipType = "direct_report"

	// RelationshipClient indicates a professional client relationship.
	RelationshipClient RelationshipType = "client"

	// RelationshipCollaborator indicates the voucher and vouchee collaborated on a project.
	RelationshipCollaborator RelationshipType = "collaborator"

	// RelationshipMentor indicates a mentorship relationship.
	RelationshipMentor RelationshipType = "mentor"

	// RelationshipAcademic indicates an academic relationship (professor, advisor, classmate).
	RelationshipAcademic RelationshipType = "academic"

	// RelationshipPersonal indicates a personal acquaintance.
	RelationshipPersonal RelationshipType = "personal"
)

// String returns the string representation of the RelationshipType.
func (r RelationshipType) String() string {
	return string(r)
}

// IsValid returns true if the RelationshipType is a known, valid type.
func (r RelationshipType) IsValid() bool {
	switch r {
	case RelationshipColleague,
		RelationshipManager,
		RelationshipDirectReport,
		RelationshipClient,
		RelationshipCollaborator,
		RelationshipMentor,
		RelationshipAcademic,
		RelationshipPersonal:
		return true
	default:
		return false
	}
}

// ParseRelationshipType parses a string into a RelationshipType.
func ParseRelationshipType(s string) (RelationshipType, error) {
	r := RelationshipType(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid relationship type: %s", s)
	}
	return r, nil
}
