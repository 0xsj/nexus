package did

import (
	"encoding/json"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DID
		wantErr bool
	}{
		{
			name:  "valid did:key",
			input: "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			want: DID{
				method:           MethodKey,
				methodSpecificID: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
				raw:              "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			},
			wantErr: false,
		},
		{
			name:  "valid did:web",
			input: "did:web:example.com",
			want: DID{
				method:           MethodWeb,
				methodSpecificID: "example.com",
				raw:              "did:web:example.com",
			},
			wantErr: false,
		},
		{
			name:  "valid did:web with path",
			input: "did:web:example.com:users:alice",
			want: DID{
				method:           MethodWeb,
				methodSpecificID: "example.com:users:alice",
				raw:              "did:web:example.com:users:alice",
			},
			wantErr: false,
		},
		{
			name:  "valid did:pkh",
			input: "did:pkh:eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb",
			want: DID{
				method:           MethodPKH,
				methodSpecificID: "eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb",
				raw:              "did:pkh:eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb",
			},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "missing did prefix",
			input:   "key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr: true,
		},
		{
			name:    "missing method",
			input:   "did::z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr: true,
		},
		{
			name:    "missing method-specific-id",
			input:   "did:key:",
			wantErr: true,
		},
		{
			name:    "only did prefix",
			input:   "did:",
			wantErr: true,
		},
		{
			name:    "only did and method",
			input:   "did:key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if got.Method() != tt.want.method {
				t.Errorf("Method() = %v, want %v", got.Method(), tt.want.method)
			}

			if got.MethodSpecificID() != tt.want.methodSpecificID {
				t.Errorf("MethodSpecificID() = %v, want %v", got.MethodSpecificID(), tt.want.methodSpecificID)
			}

			if got.String() != tt.want.raw {
				t.Errorf("String() = %v, want %v", got.String(), tt.want.raw)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name             string
		method           Method
		methodSpecificID string
		wantStr          string
		wantErr          bool
	}{
		{
			name:             "valid did:key",
			method:           MethodKey,
			methodSpecificID: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantStr:          "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr:          false,
		},
		{
			name:             "valid did:web",
			method:           MethodWeb,
			methodSpecificID: "example.com",
			wantStr:          "did:web:example.com",
			wantErr:          false,
		},
		{
			name:             "empty method-specific-id",
			method:           MethodKey,
			methodSpecificID: "",
			wantErr:          true,
		},
		{
			name:             "invalid method",
			method:           Method("invalid"),
			methodSpecificID: "test",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.method, tt.methodSpecificID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if got.String() != tt.wantStr {
				t.Errorf("String() = %v, want %v", got.String(), tt.wantStr)
			}
		})
	}
}

func TestDID_IsZero(t *testing.T) {
	var zero DID
	if !zero.IsZero() {
		t.Error("zero value DID should be zero")
	}

	parsed, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	if parsed.IsZero() {
		t.Error("parsed DID should not be zero")
	}
}

func TestDID_IsValid(t *testing.T) {
	var zero DID
	if zero.IsValid() {
		t.Error("zero value DID should not be valid")
	}

	parsed, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	if !parsed.IsValid() {
		t.Error("parsed DID should be valid")
	}
}

func TestDID_Equals(t *testing.T) {
	d1, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	d2, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	d3, _ := Parse("did:key:z6MkoDgemAx8aw9xGMVCdoaHiGc1FbimMJLPgRsRtiBQRBFy")

	if !d1.Equals(d2) {
		t.Error("identical DIDs should be equal")
	}

	if d1.Equals(d3) {
		t.Error("different DIDs should not be equal")
	}
}

func TestDID_Fragment(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	tests := []struct {
		fragment string
		want     string
	}{
		{
			fragment: "key-1",
			want:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK#key-1",
		},
		{
			fragment: "",
			want:     "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
	}

	for _, tt := range tests {
		got := d.Fragment(tt.fragment)
		if got != tt.want {
			t.Errorf("Fragment(%q) = %v, want %v", tt.fragment, got, tt.want)
		}
	}
}

func TestDID_Path(t *testing.T) {
	d, _ := Parse("did:web:example.com")

	tests := []struct {
		path string
		want string
	}{
		{
			path: "/users/alice",
			want: "did:web:example.com/users/alice",
		},
		{
			path: "users/alice",
			want: "did:web:example.com/users/alice",
		},
		{
			path: "",
			want: "did:web:example.com",
		},
	}

	for _, tt := range tests {
		got := d.Path(tt.path)
		if got != tt.want {
			t.Errorf("Path(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestDID_JSONMarshal(t *testing.T) {
	d, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	want := `"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"`
	if string(data) != want {
		t.Errorf("json.Marshal() = %v, want %v", string(data), want)
	}
}

func TestDID_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid DID",
			input:   `"did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK"`,
			want:    "did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			wantErr: false,
		},
		{
			name:    "null",
			input:   `null`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   `""`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "invalid DID",
			input:   `"not-a-did"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d DID
			err := json.Unmarshal([]byte(tt.input), &d)

			if tt.wantErr {
				if err == nil {
					t.Error("json.Unmarshal() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("json.Unmarshal() unexpected error: %v", err)
				return
			}

			if d.String() != tt.want {
				t.Errorf("String() = %v, want %v", d.String(), tt.want)
			}
		})
	}
}

func TestDID_JSONRoundTrip(t *testing.T) {
	original, _ := Parse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var decoded DID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !original.Equals(decoded) {
		t.Errorf("round-trip failed: got %v, want %v", decoded, original)
	}
}

func TestMustParse(t *testing.T) {
	// Should not panic for valid DID
	d := MustParse("did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK")
	if d.IsZero() {
		t.Error("MustParse() returned zero DID for valid input")
	}

	// Should panic for invalid DID
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse() should panic for invalid DID")
		}
	}()
	MustParse("invalid")
}
