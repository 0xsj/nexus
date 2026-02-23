package adapters

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/siwe"
)

// Compile-time interface check.
var _ domain.SignatureVerifier = (*SIWESignatureVerifier)(nil)

// SIWESignatureVerifier verifies SIWE (EIP-4361) signatures using pkg/siwe.
type SIWESignatureVerifier struct{}

// NewSIWESignatureVerifier creates a new SIWESignatureVerifier.
func NewSIWESignatureVerifier() *SIWESignatureVerifier {
	return &SIWESignatureVerifier{}
}

// VerifySignature verifies that the signature was produced by the given address.
func (v *SIWESignatureVerifier) VerifySignature(_ context.Context, address domain.WalletAddress, message string, signature string) (bool, error) {
	recovered, err := siwe.RecoverAddress(message, signature)
	if err != nil {
		return false, nil
	}

	return siwe.AddressesEqual(recovered, address.String()), nil
}
