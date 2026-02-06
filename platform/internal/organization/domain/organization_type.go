package domain

import "fmt"

// OrganizationType represents the type of organization.
type OrganizationType string

const (
	OrganizationTypeCompany     OrganizationType = "Company"
	OrganizationTypeAgency      OrganizationType = "Agency"
	OrganizationTypeInstitution OrganizationType = "Institution"
	OrganizationTypeCommunity   OrganizationType = "Community"
)

// ParseOrganizationType parses a string into an OrganizationType.
func ParseOrganizationType(s string) (OrganizationType, error) {
	switch s {
	case string(OrganizationTypeCompany):
		return OrganizationTypeCompany, nil
	case string(OrganizationTypeAgency):
		return OrganizationTypeAgency, nil
	case string(OrganizationTypeInstitution):
		return OrganizationTypeInstitution, nil
	case string(OrganizationTypeCommunity):
		return OrganizationTypeCommunity, nil
	default:
		return "", fmt.Errorf("unknown organization type: %s", s)
	}
}

// String returns the string representation.
func (t OrganizationType) String() string { return string(t) }

// IsValid returns true if the type is a known, valid type.
func (t OrganizationType) IsValid() bool {
	switch t {
	case OrganizationTypeCompany, OrganizationTypeAgency,
		OrganizationTypeInstitution, OrganizationTypeCommunity:
		return true
	default:
		return false
	}
}
