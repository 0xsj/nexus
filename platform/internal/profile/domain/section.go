package domain

import (
	"fmt"
)

// SectionType represents the type of a profile section.
type SectionType string

const (
	// SectionTypeOverview is the bio, headline, and key badges section.
	SectionTypeOverview SectionType = "overview"

	// SectionTypeExperience is the employment history section.
	SectionTypeExperience SectionType = "experience"

	// SectionTypeSkills is the technical skills section.
	SectionTypeSkills SectionType = "skills"

	// SectionTypeCertifications is the professional certifications section.
	SectionTypeCertifications SectionType = "certifications"

	// SectionTypeEducation is the degrees and courses section.
	SectionTypeEducation SectionType = "education"

	// SectionTypeContributions is the open source contributions section.
	SectionTypeContributions SectionType = "contributions"

	// SectionTypeVouches is the peer endorsements section.
	SectionTypeVouches SectionType = "vouches"

	// SectionTypeCustom is a user-defined section.
	SectionTypeCustom SectionType = "custom"
)

// ParseSectionType parses a string into a SectionType.
func ParseSectionType(s string) (SectionType, error) {
	switch s {
	case "overview":
		return SectionTypeOverview, nil
	case "experience":
		return SectionTypeExperience, nil
	case "skills":
		return SectionTypeSkills, nil
	case "certifications":
		return SectionTypeCertifications, nil
	case "education":
		return SectionTypeEducation, nil
	case "contributions":
		return SectionTypeContributions, nil
	case "vouches":
		return SectionTypeVouches, nil
	case "custom":
		return SectionTypeCustom, nil
	default:
		return "", fmt.Errorf("unknown section type: %s", s)
	}
}

// String returns the string representation of the SectionType.
func (t SectionType) String() string {
	return string(t)
}

// IsValid returns true if the SectionType is a known, valid type.
func (t SectionType) IsValid() bool {
	switch t {
	case SectionTypeOverview,
		SectionTypeExperience,
		SectionTypeSkills,
		SectionTypeCertifications,
		SectionTypeEducation,
		SectionTypeContributions,
		SectionTypeVouches,
		SectionTypeCustom:
		return true
	default:
		return false
	}
}

// ============================================================================
// Section Value Object
// ============================================================================

// Section is a value object representing a configurable section of a profile.
// Each section has its own type, title, visibility settings, and sort order.
type Section struct {
	sectionType SectionType
	title       string
	visibility  Visibility
	sortOrder   int
}

// NewSection creates a new Section value object.
func NewSection(sectionType SectionType, title string, visibility Visibility, sortOrder int) Section {
	return Section{
		sectionType: sectionType,
		title:       title,
		visibility:  visibility,
		sortOrder:   sortOrder,
	}
}

// SectionType returns the section's type.
func (s Section) SectionType() SectionType {
	return s.sectionType
}

// Title returns the section's display title.
func (s Section) Title() string {
	return s.title
}

// Visibility returns the section's visibility setting.
func (s Section) Visibility() Visibility {
	return s.visibility
}

// SortOrder returns the section's sort order.
func (s Section) SortOrder() int {
	return s.sortOrder
}

// WithTitle returns a new Section with the title updated.
func (s Section) WithTitle(title string) Section {
	s.title = title
	return s
}

// WithVisibility returns a new Section with the visibility updated.
func (s Section) WithVisibility(visibility Visibility) Section {
	s.visibility = visibility
	return s
}

// WithSortOrder returns a new Section with the sort order updated.
func (s Section) WithSortOrder(sortOrder int) Section {
	s.sortOrder = sortOrder
	return s
}
