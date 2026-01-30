package domain

import (
	"errors"

	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
)

// ============================================================================
// Subject Error Codes
// ============================================================================

const (
	CodeInvalidSubjectID   pkgerrors.Code = "LEDGER_INVALID_SUBJECT_ID"
	CodeInvalidSubjectType pkgerrors.Code = "LEDGER_INVALID_SUBJECT_TYPE"
)

// ============================================================================
// Subject Sentinel Errors
// ============================================================================

var (
	ErrInvalidSubjectID   = errors.New("invalid subject ID")
	ErrInvalidSubjectType = errors.New("invalid subject type")
)

// ============================================================================
// Subject Error Constructors
// ============================================================================

// InvalidSubjectID creates an invalid subject ID error.
func InvalidSubjectID(operation string, value string, reason string) *pkgerrors.Error {
	return pkgerrors.Validation(operation, "invalid subject ID: "+reason).
		WithCode(CodeInvalidSubjectID).
		WithMeta("value", value)
}

// InvalidSubjectType creates an invalid subject type error.
func InvalidSubjectType(operation string, value string) *pkgerrors.Error {
	return pkgerrors.Validationf(operation, "invalid subject type: %s", value).
		WithCode(CodeInvalidSubjectType).
		WithMeta("value", value)
}

// ============================================================================
// SubjectType Value Object
// ============================================================================

// SubjectType represents the type of entity that an event is about.
type SubjectType int

const (
	// SubjectTypeUnknown indicates an unknown subject type.
	SubjectTypeUnknown SubjectType = iota

	// SubjectTypeUser indicates the subject is a user.
	SubjectTypeUser

	// SubjectTypeCredential indicates the subject is a credential.
	SubjectTypeCredential

	// SubjectTypePresentation indicates the subject is a presentation.
	SubjectTypePresentation

	// SubjectTypeVerification indicates the subject is a verification.
	SubjectTypeVerification

	// SubjectTypeWallet indicates the subject is a wallet.
	SubjectTypeWallet

	// SubjectTypeSession indicates the subject is a session.
	SubjectTypeSession

	// SubjectTypeOrganization indicates the subject is an organization.
	SubjectTypeOrganization

	// SubjectTypeDID indicates the subject is a DID.
	SubjectTypeDID

	// SubjectTypeVouch indicates the subject is a vouch/endorsement.
	SubjectTypeVouch
)

// String returns the string representation of the SubjectType.
func (t SubjectType) String() string {
	switch t {
	case SubjectTypeUser:
		return "user"
	case SubjectTypeCredential:
		return "credential"
	case SubjectTypePresentation:
		return "presentation"
	case SubjectTypeVerification:
		return "verification"
	case SubjectTypeWallet:
		return "wallet"
	case SubjectTypeSession:
		return "session"
	case SubjectTypeOrganization:
		return "organization"
	case SubjectTypeDID:
		return "did"
	case SubjectTypeVouch:
		return "vouch"
	default:
		return "unknown"
	}
}

// ParseSubjectType parses a string into a SubjectType.
func ParseSubjectType(s string) (SubjectType, error) {
	switch s {
	case "user":
		return SubjectTypeUser, nil
	case "credential":
		return SubjectTypeCredential, nil
	case "presentation":
		return SubjectTypePresentation, nil
	case "verification":
		return SubjectTypeVerification, nil
	case "wallet":
		return SubjectTypeWallet, nil
	case "session":
		return SubjectTypeSession, nil
	case "organization":
		return SubjectTypeOrganization, nil
	case "did":
		return SubjectTypeDID, nil
	case "vouch":
		return SubjectTypeVouch, nil
	case "unknown":
		return SubjectTypeUnknown, nil
	default:
		return SubjectTypeUnknown, InvalidSubjectType("domain.ParseSubjectType", s)
	}
}

// IsValid returns true if the SubjectType is a known, valid type.
func (t SubjectType) IsValid() bool {
	switch t {
	case SubjectTypeUser, SubjectTypeCredential, SubjectTypePresentation,
		SubjectTypeVerification, SubjectTypeWallet, SubjectTypeSession,
		SubjectTypeOrganization, SubjectTypeDID, SubjectTypeVouch:
		return true
	default:
		return false
	}
}

// ============================================================================
// SubjectID Value Object
// ============================================================================

// SubjectID identifies the specific entity that an event is about.
// The format depends on SubjectType:
//   - User: user UUID
//   - Credential: credential UUID
//   - Presentation: presentation UUID
//   - Verification: verification UUID
//   - Wallet: wallet address
//   - Session: session UUID
//   - Organization: organization UUID
//   - DID: DID string
//   - Vouch: vouch UUID
type SubjectID struct {
	value string
}

// NewSubjectID creates a new SubjectID.
func NewSubjectID(value string) (SubjectID, error) {
	if value == "" {
		return SubjectID{}, InvalidSubjectID("domain.NewSubjectID", value, "cannot be empty")
	}
	return SubjectID{value: value}, nil
}

// MustNewSubjectID creates a new SubjectID and panics if invalid.
// Only use for constants or tests.
func MustNewSubjectID(value string) SubjectID {
	id, err := NewSubjectID(value)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the string representation.
func (id SubjectID) String() string {
	return id.value
}

// IsZero returns true if the SubjectID is the zero value.
func (id SubjectID) IsZero() bool {
	return id.value == ""
}

// IsValid returns true if the SubjectID is valid (non-empty).
func (id SubjectID) IsValid() bool {
	return id.value != ""
}

// Equals checks if two SubjectIDs are equal.
func (id SubjectID) Equals(other SubjectID) bool {
	return id.value == other.value
}

// ============================================================================
// Subject Value Object (Composite)
// ============================================================================

// Subject represents what an event is about, combining type and identifier.
type Subject struct {
	subjectType SubjectType
	subjectID   SubjectID
}

// NewSubject creates a new Subject.
func NewSubject(subjectType SubjectType, subjectID SubjectID) (Subject, error) {
	if !subjectType.IsValid() {
		return Subject{}, InvalidSubjectType("domain.NewSubject", subjectType.String())
	}
	if subjectID.IsZero() {
		return Subject{}, InvalidSubjectID("domain.NewSubject", "", "subject ID cannot be empty")
	}
	return Subject{
		subjectType: subjectType,
		subjectID:   subjectID,
	}, nil
}

// MustNewSubject creates a new Subject and panics if invalid.
// Only use for constants or tests.
func MustNewSubject(subjectType SubjectType, subjectID SubjectID) Subject {
	subject, err := NewSubject(subjectType, subjectID)
	if err != nil {
		panic(err)
	}
	return subject
}

// Type returns the subject type.
func (s Subject) Type() SubjectType {
	return s.subjectType
}

// ID returns the subject ID.
func (s Subject) ID() SubjectID {
	return s.subjectID
}

// IsZero returns true if the Subject is the zero value.
func (s Subject) IsZero() bool {
	return !s.subjectType.IsValid() && s.subjectID.IsZero()
}

// Equals checks if two Subjects are equal.
func (s Subject) Equals(other Subject) bool {
	return s.subjectType == other.subjectType && s.subjectID.Equals(other.subjectID)
}
