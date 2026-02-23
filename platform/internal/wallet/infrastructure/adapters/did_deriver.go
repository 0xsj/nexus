package adapters

import (
	"context"
	"strconv"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/did/pkh"
)

// Compile-time interface check.
var _ domain.DIDDeriver = (*PKHDIDDeriver)(nil)

// PKHDIDDeriver derives did:pkh identifiers from wallet addresses using pkg/did/pkh.
type PKHDIDDeriver struct{}

// NewPKHDIDDeriver creates a new PKHDIDDeriver.
func NewPKHDIDDeriver() *PKHDIDDeriver {
	return &PKHDIDDeriver{}
}

// DeriveDID derives a did:pkh from a wallet address and chain.
func (d *PKHDIDDeriver) DeriveDID(_ context.Context, address domain.WalletAddress, chain domain.Chain) (string, error) {
	did, err := pkh.GenerateFromRaw(chain.Namespace(), strconv.Itoa(chain.ID()), address.String())
	if err != nil {
		return "", err
	}

	return did.String(), nil
}
