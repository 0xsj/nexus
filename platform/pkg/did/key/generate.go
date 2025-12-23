package key

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did"
)

// Generate creates a new did:key DID from a new Ed25519 key pair.
// Returns the DID and the generated key pair.
func Generate() (did.DID, *ed25519.KeyPair, error) {
	const op = "did.key.Generate"

	kp, err := ed25519.Generate()
	if err != nil {
		return did.DID{}, nil, did.ErrResolutionFailed(op, "", err)
	}

	d, err := FromKeyPair(kp)
	if err != nil {
		return did.DID{}, nil, err
	}

	return d, kp, nil
}

// FromKeyPair creates a did:key DID from an existing key pair.
func FromKeyPair(kp crypto.KeyPair) (did.DID, error) {
	const op = "did.key.FromKeyPair"

	if kp == nil {
		return did.DID{}, did.ErrInvalidDID(op, "key pair cannot be nil")
	}

	// Get multibase-encoded public key
	multibase := kp.PublicKeyMultibase()
	if multibase == "" {
		return did.DID{}, did.ErrInvalidDID(op, "key pair does not support multibase encoding")
	}

	return did.New(did.MethodKey, multibase)
}

// FromPublicKey creates a did:key DID from a raw Ed25519 public key.
func FromPublicKey(publicKey []byte) (did.DID, error) {
	const op = "did.key.FromPublicKey"

	kp, err := ed25519.FromPublicKey(publicKey)
	if err != nil {
		return did.DID{}, did.ErrInvalidDID(op, err.Error())
	}

	return FromKeyPair(kp)
}

// FromMultibase creates a did:key DID from a multibase-encoded public key.
// The multibase string should include the multicodec prefix.
func FromMultibase(multibase string) (did.DID, error) {
	const op = "did.key.FromMultibase"

	if multibase == "" {
		return did.DID{}, did.ErrInvalidDID(op, "multibase string cannot be empty")
	}

	return did.New(did.MethodKey, multibase)
}

// MustGenerate creates a new did:key DID and panics if generation fails.
// Only use for tests.
func MustGenerate() (did.DID, *ed25519.KeyPair) {
	d, kp, err := Generate()
	if err != nil {
		panic(err)
	}
	return d, kp
}
