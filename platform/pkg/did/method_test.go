package did

import (
	"testing"
)

func TestMethod_String(t *testing.T) {
	tests := []struct {
		method Method
		want   string
	}{
		{MethodKey, "key"},
		{MethodWeb, "web"},
		{MethodPKH, "pkh"},
		{MethodEthr, "ethr"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.method.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMethod_Validate(t *testing.T) {
	tests := []struct {
		method  Method
		wantErr bool
	}{
		{MethodKey, false},
		{MethodWeb, false},
		{MethodPKH, false},
		{MethodEthr, false},
		{Method("invalid"), true},
		{Method(""), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			err := tt.method.Validate()
			if tt.wantErr && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestMethod_IsSupported(t *testing.T) {
	tests := []struct {
		method Method
		want   bool
	}{
		{MethodKey, true},
		{MethodWeb, false},  // Phase 2
		{MethodPKH, false},  // Phase 2
		{MethodEthr, false}, // Phase 2
		{Method("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := tt.method.IsSupported(); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMethod_RequiresNetwork(t *testing.T) {
	tests := []struct {
		method Method
		want   bool
	}{
		{MethodKey, false},
		{MethodWeb, true},
		{MethodPKH, true},
		{MethodEthr, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := tt.method.RequiresNetwork(); got != tt.want {
				t.Errorf("RequiresNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMethod_Description(t *testing.T) {
	tests := []struct {
		method Method
		want   string
	}{
		{MethodKey, "Self-contained DID derived from public key"},
		{MethodWeb, "DID resolved via HTTPS from web domain"},
		{MethodPKH, "DID derived from blockchain account address"},
		{MethodEthr, "Ethereum-based DID with on-chain registry"},
		{Method("invalid"), "Unknown DID method"},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			if got := tt.method.Description(); got != tt.want {
				t.Errorf("Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMethod_SupportedAlgorithms(t *testing.T) {
	tests := []struct {
		method    Method
		wantCount int
	}{
		{MethodKey, 4},  // Ed25519, secp256k1, P-256, P-384
		{MethodWeb, 3},  // Ed25519, secp256k1, RSA
		{MethodPKH, 1},  // secp256k1
		{MethodEthr, 1}, // secp256k1
		{Method("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			got := tt.method.SupportedAlgorithms()
			if len(got) != tt.wantCount {
				t.Errorf("SupportedAlgorithms() count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}

func TestAllMethods(t *testing.T) {
	methods := AllMethods()

	expected := []Method{MethodKey, MethodWeb, MethodPKH, MethodEthr}
	if len(methods) != len(expected) {
		t.Errorf("AllMethods() count = %v, want %v", len(methods), len(expected))
	}

	for _, e := range expected {
		found := false
		for _, m := range methods {
			if m == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AllMethods() missing %v", e)
		}
	}
}

func TestSupportedMethods(t *testing.T) {
	methods := SupportedMethods()

	// Currently only MethodKey is supported
	if len(methods) != 1 {
		t.Errorf("SupportedMethods() count = %v, want 1", len(methods))
	}

	if len(methods) > 0 && methods[0] != MethodKey {
		t.Errorf("SupportedMethods()[0] = %v, want %v", methods[0], MethodKey)
	}
}
