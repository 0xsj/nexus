package adapters

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/crypto/ed25519"
	"github.com/0xsj/nexus/platform/pkg/did/key"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
)

// Compile-time interface check.
var _ domain.DIDGenerator = (*DIDGenerator)(nil)

// DIDGenerator generates real cryptographic DIDs using pkg/crypto and pkg/did.
type DIDGenerator struct{}

// NewDIDGenerator creates a new DIDGenerator.
func NewDIDGenerator() *DIDGenerator {
	return &DIDGenerator{}
}

// GenerateDIDKey generates a new did:key from a fresh Ed25519 keypair.
func (g *DIDGenerator) GenerateDIDKey(_ context.Context) (string, error) {
	kp, err := ed25519.Generate()
	if err != nil {
		return "", err
	}

	d, err := key.FromKeyPair(kp)
	if err != nil {
		return "", err
	}

	return d.String(), nil
}

// DeriveDIDPKH derives a did:pkh from a wallet address and EVM chain ID.
func (g *DIDGenerator) DeriveDIDPKH(_ context.Context, address string, chainID string) (string, error) {
	d, err := pkh.GenerateEVM(address, chain.ChainID(chainID))
	if err != nil {
		return "", err
	}

	return d.String(), nil
}
