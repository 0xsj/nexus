package domain

import (
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/types"
)

// OrganizationID uniquely identifies an organization.
type OrganizationID struct {
	value types.ID
}

// NewOrganizationID generates a new unique OrganizationID.
func NewOrganizationID() OrganizationID {
	return OrganizationID{value: types.NewID()}
}

// ParseOrganizationID parses a string into an OrganizationID.
func ParseOrganizationID(s string) (OrganizationID, error) {
	id, err := types.ParseID(s)
	if err != nil {
		return OrganizationID{}, pkgerrors.Validation("OrganizationID.Parse", "invalid organization ID format").
			WithMeta("value", s)
	}
	return OrganizationID{value: id}, nil
}

// String returns the string representation.
func (id OrganizationID) String() string { return id.value.String() }

// IsZero returns true if the ID is the zero value.
func (id OrganizationID) IsZero() bool { return id.value.IsZero() }

// IsValid returns true if the ID is valid (non-zero).
func (id OrganizationID) IsValid() bool { return id.value.IsValid() }

// Equals checks if two IDs are equal.
func (id OrganizationID) Equals(other OrganizationID) bool { return id.value.Equals(other.value) }

// TypesID returns the underlying types.ID.
func (id OrganizationID) TypesID() types.ID { return id.value }
