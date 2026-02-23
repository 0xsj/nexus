package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xsj/nexus/platform/internal/identity/domain"
	"github.com/0xsj/nexus/platform/pkg/chain"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
	"github.com/0xsj/nexus/platform/pkg/siwe"
)

// Compile-time interface check.
var _ domain.WalletReader = (*WalletReader)(nil)

// WalletReader verifies SIWE signatures and derives did:pkh identifiers.
type WalletReader struct{}

// NewWalletReader creates a new WalletReader.
func NewWalletReader() *WalletReader {
	return &WalletReader{}
}

// VerifySignature verifies a SIWE signature and returns the derived DID.
func (r *WalletReader) VerifySignature(_ context.Context, req domain.WalletVerificationRequest) (*domain.WalletVerificationResult, error) {
	msg, err := siwe.Verify(req.Message, req.Signature)
	if err != nil {
		return &domain.WalletVerificationResult{Valid: false}, nil
	}

	// Validate that the recovered address matches the expected address.
	if !siwe.AddressesEqual(msg.Address, req.Address) {
		return &domain.WalletVerificationResult{Valid: false}, nil
	}

	// Derive did:pkh from the verified address.
	d, err := pkh.GenerateEVM(msg.Address, chain.ChainID(req.ChainID))
	if err != nil {
		return nil, fmt.Errorf("deriving did:pkh: %w", err)
	}

	return &domain.WalletVerificationResult{
		Valid:   true,
		Address: strings.ToLower(msg.Address),
		ChainID: req.ChainID,
		DID:     d.String(),
	}, nil
}

// GetDIDForWallet returns the did:pkh for a wallet address.
func (r *WalletReader) GetDIDForWallet(_ context.Context, address domain.WalletAddress) (string, error) {
	d, err := pkh.GenerateEVM(address.Address, chain.ChainID(address.ChainID))
	if err != nil {
		return "", err
	}

	return d.String(), nil
}
