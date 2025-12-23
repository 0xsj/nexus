package did

import "fmt"

// Method represents a DID method.
// Each method defines how DIDs are created, resolved, and managed.
type Method string

const (
	// MethodKey is the did:key method.
	// Self-contained DIDs derived from public keys.
	// No resolution infrastructure required.
	// Example: did:key:z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK
	MethodKey Method = "key"

	// MethodWeb is the did:web method.
	// DIDs resolved via HTTPS from a web domain.
	// Example: did:web:example.com
	// Example: did:web:example.com:users:alice
	MethodWeb Method = "web"

	// MethodPKH is the did:pkh method.
	// DIDs derived from blockchain account addresses.
	// Used for Ethereum wallet integration.
	// Example: did:pkh:eip155:1:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb
	MethodPKH Method = "pkh"

	// MethodEthr is the did:ethr method.
	// Ethereum-based DIDs with on-chain registry.
	// Phase 2: Not yet implemented.
	// Example: did:ethr:0xab16a96d359ec26a11e2c2b3d8f8b8942d5bfcdb
	MethodEthr Method = "ethr"
)

// String returns the string representation.
func (m Method) String() string {
	return string(m)
}

// Validate checks if the method is recognized.
func (m Method) Validate() error {
	switch m {
	case MethodKey, MethodWeb, MethodPKH, MethodEthr:
		return nil
	default:
		return fmt.Errorf("unknown DID method: %s", m)
	}
}

// IsSupported returns true if the method is currently implemented.
func (m Method) IsSupported() bool {
	switch m {
	case MethodKey:
		return true
	case MethodWeb, MethodPKH, MethodEthr:
		// Phase 2
		return false
	default:
		return false
	}
}

// RequiresNetwork returns true if the method requires network resolution.
func (m Method) RequiresNetwork() bool {
	switch m {
	case MethodKey:
		return false
	case MethodWeb, MethodPKH, MethodEthr:
		return true
	default:
		return true
	}
}

// Description returns a human-readable description of the method.
func (m Method) Description() string {
	switch m {
	case MethodKey:
		return "Self-contained DID derived from public key"
	case MethodWeb:
		return "DID resolved via HTTPS from web domain"
	case MethodPKH:
		return "DID derived from blockchain account address"
	case MethodEthr:
		return "Ethereum-based DID with on-chain registry"
	default:
		return "Unknown DID method"
	}
}

// SupportedAlgorithms returns the cryptographic algorithms supported by this method.
func (m Method) SupportedAlgorithms() []string {
	switch m {
	case MethodKey:
		return []string{"Ed25519", "secp256k1", "P-256", "P-384"}
	case MethodWeb:
		return []string{"Ed25519", "secp256k1", "RSA"}
	case MethodPKH:
		return []string{"secp256k1"}
	case MethodEthr:
		return []string{"secp256k1"}
	default:
		return nil
	}
}

// AllMethods returns all known DID methods.
func AllMethods() []Method {
	return []Method{
		MethodKey,
		MethodWeb,
		MethodPKH,
		MethodEthr,
	}
}

// SupportedMethods returns all currently implemented DID methods.
func SupportedMethods() []Method {
	var supported []Method
	for _, m := range AllMethods() {
		if m.IsSupported() {
			supported = append(supported, m)
		}
	}
	return supported
}
