package domain

import (
	"testing"
)

// ============================================================================
// SubjectType Tests
// ============================================================================

func TestSubjectType_String(t *testing.T) {
	tests := []struct {
		subjectType SubjectType
		want        string
	}{
		{SubjectTypeUnknown, "unknown"},
		{SubjectTypeUser, "user"},
		{SubjectTypeCredential, "credential"},
		{SubjectTypePresentation, "presentation"},
		{SubjectTypeVerification, "verification"},
		{SubjectTypeWallet, "wallet"},
		{SubjectTypeSession, "session"},
		{SubjectTypeOrganization, "organization"},
		{SubjectTypeDID, "did"},
		{SubjectTypeVouch, "vouch"},
		{SubjectType(999), "unknown"}, // invalid value
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.subjectType.String()
			if got != tt.want {
				t.Errorf("SubjectType(%d).String() = %v, want %v", tt.subjectType, got, tt.want)
			}
		})
	}
}

func TestParseSubjectType_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  SubjectType
	}{
		{"user", SubjectTypeUser},
		{"credential", SubjectTypeCredential},
		{"presentation", SubjectTypePresentation},
		{"verification", SubjectTypeVerification},
		{"wallet", SubjectTypeWallet},
		{"session", SubjectTypeSession},
		{"organization", SubjectTypeOrganization},
		{"did", SubjectTypeDID},
		{"vouch", SubjectTypeVouch},
		{"unknown", SubjectTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSubjectType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseSubjectType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSubjectType_Invalid(t *testing.T) {
	invalidTypes := []string{
		"invalid",
		"USER",
		"Credential",
		"",
		"credentials",
		"users",
	}

	for _, input := range invalidTypes {
		t.Run(input, func(t *testing.T) {
			_, err := ParseSubjectType(input)
			if err == nil {
				t.Errorf("expected error for ParseSubjectType(%q)", input)
			}
		})
	}
}

func TestSubjectType_IsValid(t *testing.T) {
	tests := []struct {
		subjectType SubjectType
		want        bool
	}{
		{SubjectTypeUnknown, false},
		{SubjectTypeUser, true},
		{SubjectTypeCredential, true},
		{SubjectTypePresentation, true},
		{SubjectTypeVerification, true},
		{SubjectTypeWallet, true},
		{SubjectTypeSession, true},
		{SubjectTypeOrganization, true},
		{SubjectTypeDID, true},
		{SubjectTypeVouch, true},
		{SubjectType(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.subjectType.String(), func(t *testing.T) {
			got := tt.subjectType.IsValid()
			if got != tt.want {
				t.Errorf("SubjectType(%d).IsValid() = %v, want %v", tt.subjectType, got, tt.want)
			}
		})
	}
}

// ============================================================================
// SubjectID Tests
// ============================================================================

func TestNewSubjectID_Valid(t *testing.T) {
	tests := []string{
		"cred-123",
		"user-456",
		"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		"0x1234567890abcdef",
		"123e4567-e89b-12d3-a456-426614174000",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			id, err := NewSubjectID(input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id.String() != input {
				t.Errorf("SubjectID.String() = %v, want %v", id.String(), input)
			}
			if id.IsZero() {
				t.Error("expected non-zero SubjectID")
			}
			if !id.IsValid() {
				t.Error("expected valid SubjectID")
			}
		})
	}
}

func TestNewSubjectID_Empty(t *testing.T) {
	_, err := NewSubjectID("")

	if err == nil {
		t.Error("expected error for empty SubjectID")
	}
}

func TestMustNewSubjectID_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	id := MustNewSubjectID("cred-123")
	if id.String() != "cred-123" {
		t.Errorf("unexpected SubjectID: %v", id.String())
	}
}

func TestMustNewSubjectID_Empty_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty SubjectID")
		}
	}()

	MustNewSubjectID("")
}

func TestSubjectID_IsZero(t *testing.T) {
	var zeroID SubjectID

	if !zeroID.IsZero() {
		t.Error("expected zero value to be zero")
	}
	if zeroID.IsValid() {
		t.Error("expected zero value to be invalid")
	}
}

func TestSubjectID_Equals(t *testing.T) {
	id1, _ := NewSubjectID("cred-123")
	id2, _ := NewSubjectID("cred-123")
	id3, _ := NewSubjectID("cred-456")

	if !id1.Equals(id2) {
		t.Error("expected equal SubjectIDs to be equal")
	}
	if id1.Equals(id3) {
		t.Error("expected different SubjectIDs to not be equal")
	}
}

// ============================================================================
// Subject Tests
// ============================================================================

func TestNewSubject_Valid(t *testing.T) {
	subjectID, _ := NewSubjectID("cred-123")
	subject, err := NewSubject(SubjectTypeCredential, subjectID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subject.Type() != SubjectTypeCredential {
		t.Errorf("Subject.Type() = %v, want %v", subject.Type(), SubjectTypeCredential)
	}
	if !subject.ID().Equals(subjectID) {
		t.Errorf("Subject.ID() = %v, want %v", subject.ID(), subjectID)
	}
	if subject.IsZero() {
		t.Error("expected non-zero Subject")
	}
}

func TestNewSubject_InvalidType(t *testing.T) {
	subjectID, _ := NewSubjectID("cred-123")
	_, err := NewSubject(SubjectTypeUnknown, subjectID)

	if err == nil {
		t.Error("expected error for invalid SubjectType")
	}
}

func TestNewSubject_InvalidID(t *testing.T) {
	var zeroID SubjectID
	_, err := NewSubject(SubjectTypeCredential, zeroID)

	if err == nil {
		t.Error("expected error for zero SubjectID")
	}
}

func TestMustNewSubject_Valid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	subjectID := MustNewSubjectID("cred-123")
	subject := MustNewSubject(SubjectTypeCredential, subjectID)

	if subject.Type() != SubjectTypeCredential {
		t.Errorf("unexpected SubjectType: %v", subject.Type())
	}
}

func TestMustNewSubject_InvalidType_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid SubjectType")
		}
	}()

	subjectID := MustNewSubjectID("cred-123")
	MustNewSubject(SubjectTypeUnknown, subjectID)
}

func TestSubject_IsZero(t *testing.T) {
	var zeroSubject Subject

	if !zeroSubject.IsZero() {
		t.Error("expected zero value to be zero")
	}
}

func TestSubject_Equals(t *testing.T) {
	subjectID1, _ := NewSubjectID("cred-123")
	subjectID2, _ := NewSubjectID("cred-456")

	subject1, _ := NewSubject(SubjectTypeCredential, subjectID1)
	subject2, _ := NewSubject(SubjectTypeCredential, subjectID1)
	subject3, _ := NewSubject(SubjectTypeCredential, subjectID2)
	subject4, _ := NewSubject(SubjectTypeUser, subjectID1)

	if !subject1.Equals(subject2) {
		t.Error("expected equal Subjects to be equal")
	}
	if subject1.Equals(subject3) {
		t.Error("expected Subjects with different IDs to not be equal")
	}
	if subject1.Equals(subject4) {
		t.Error("expected Subjects with different types to not be equal")
	}
}

// ============================================================================
// Subject Roundtrip Tests
// ============================================================================

func TestSubject_TypeRoundtrip(t *testing.T) {
	types := []SubjectType{
		SubjectTypeUser,
		SubjectTypeCredential,
		SubjectTypePresentation,
		SubjectTypeVerification,
		SubjectTypeWallet,
		SubjectTypeSession,
		SubjectTypeOrganization,
		SubjectTypeDID,
		SubjectTypeVouch,
	}

	for _, st := range types {
		t.Run(st.String(), func(t *testing.T) {
			// String -> Parse -> String
			str := st.String()
			parsed, err := ParseSubjectType(str)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed != st {
				t.Errorf("roundtrip failed: got %v, want %v", parsed, st)
			}
		})
	}
}
