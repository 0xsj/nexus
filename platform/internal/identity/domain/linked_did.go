package domain

import (
	"time"

	"github.com/0xsj/nexus/platform/pkg/did"
)

// ============================================================================
// DID Source
// ============================================================================

// DIDSource indicates how a DID was created/linked.
type DIDSource string

const (
	// DIDSourceCustodial indicates Nexus generated the DID (did:key for email users).
	DIDSourceCustodial DIDSource = "custodial"

	// DIDSourceWallet indicates the DID was derived from a user's wallet (did:pkh).
	DIDSourceWallet DIDSource = "wallet"

	// DIDSourceImported indicates the user imported an existing DID.
	DIDSourceImported DIDSource = "imported"
)

// String returns the string representation.
func (s DIDSource) String() string {
	return string(s)
}

// IsValid checks if the source is valid.
func (s DIDSource) IsValid() bool {
	switch s {
	case DIDSourceCustodial, DIDSourceWallet, DIDSourceImported:
		return true
	default:
		return false
	}
}

// IsSelfCustody returns true if the user controls the private key.
func (s DIDSource) IsSelfCustody() bool {
	return s == DIDSourceWallet || s == DIDSourceImported
}

// ============================================================================
// Linked DID
// ============================================================================

// LinkedDID represents a DID associated with a user.
// Users can have multiple DIDs from different sources.
// One DID is designated as "primary" for credential issuance.
type LinkedDID struct {
	// ID is a unique identifier for this linked DID entry.
	ID string

	// DID is the actual decentralized identifier.
	DID did.DID

	// Source indicates how this DID was created/linked.
	Source DIDSource

	// IsPrimary indicates if this is the user's primary DID.
	// Only one LinkedDID should have IsPrimary=true at a time.
	IsPrimary bool

	// Label is an optional user-friendly name for this DID.
	// Example: "Personal Wallet", "Work Identity"
	Label string

	// Metadata contains source-specific data.
	// For wallet DIDs: wallet address, chain ID
	// For custodial DIDs: key algorithm
	// For imported DIDs: import source
	Metadata map[string]string

	// LinkedAt is when this DID was linked to the user.
	LinkedAt time.Time

	// LastUsedAt is when this DID was last used (for signing, presenting credentials).
	LastUsedAt *time.Time
}

// ============================================================================
// Constructors
// ============================================================================

// NewCustodialDID creates a LinkedDID for a Nexus-generated custodial DID.
func NewCustodialDID(id string, d did.DID, isPrimary bool) LinkedDID {
	return LinkedDID{
		ID:        id,
		DID:       d,
		Source:    DIDSourceCustodial,
		IsPrimary: isPrimary,
		Label:     "Default Identity",
		Metadata: map[string]string{
			"algorithm": "Ed25519",
		},
		LinkedAt: time.Now(),
	}
}

// NewWalletDID creates a LinkedDID derived from a wallet address.
func NewWalletDID(id string, d did.DID, walletAddress string, chain Chain, isPrimary bool) LinkedDID {
	return LinkedDID{
		ID:        id,
		DID:       d,
		Source:    DIDSourceWallet,
		IsPrimary: isPrimary,
		Label:     chain.DisplayName() + " Wallet",
		Metadata: map[string]string{
			"address":  walletAddress,
			"chain":    chain.String(),
			"chain_id": chain.ChainIDString(),
		},
		LinkedAt: time.Now(),
	}
}

// NewImportedDID creates a LinkedDID for a user-imported DID.
func NewImportedDID(id string, d did.DID, label string, isPrimary bool) LinkedDID {
	return LinkedDID{
		ID:        id,
		DID:       d,
		Source:    DIDSourceImported,
		IsPrimary: isPrimary,
		Label:     label,
		Metadata:  map[string]string{},
		LinkedAt:  time.Now(),
	}
}

// ============================================================================
// Methods
// ============================================================================

// String returns the DID string.
func (l LinkedDID) String() string {
	return l.DID.String()
}

// IsZero returns true if the LinkedDID is uninitialized.
func (l LinkedDID) IsZero() bool {
	return l.ID == "" || l.DID.IsZero()
}

// IsCustodial returns true if Nexus manages the private key.
func (l LinkedDID) IsCustodial() bool {
	return l.Source == DIDSourceCustodial
}

// IsSelfCustody returns true if the user manages the private key.
func (l LinkedDID) IsSelfCustody() bool {
	return l.Source.IsSelfCustody()
}

// WalletAddress returns the wallet address if this is a wallet-derived DID.
func (l LinkedDID) WalletAddress() string {
	if l.Source != DIDSourceWallet {
		return ""
	}
	return l.Metadata["address"]
}

// ChainID returns the chain ID if this is a wallet-derived DID.
func (l LinkedDID) ChainID() string {
	if l.Source != DIDSourceWallet {
		return ""
	}
	return l.Metadata["chain_id"]
}

// MarkUsed updates the LastUsedAt timestamp.
func (l *LinkedDID) MarkUsed() {
	now := time.Now()
	l.LastUsedAt = &now
}

// SetPrimary sets/unsets the primary flag.
func (l *LinkedDID) SetPrimary(primary bool) {
	l.IsPrimary = primary
}

// SetLabel updates the label.
func (l *LinkedDID) SetLabel(label string) {
	l.Label = label
}

// ============================================================================
// LinkedDIDs Collection
// ============================================================================

// LinkedDIDs is a collection of LinkedDID with helper methods.
type LinkedDIDs []LinkedDID

// Primary returns the primary DID, if any.
func (lds LinkedDIDs) Primary() (LinkedDID, bool) {
	for _, ld := range lds {
		if ld.IsPrimary {
			return ld, true
		}
	}
	return LinkedDID{}, false
}

// FindByDID finds a LinkedDID by its DID string.
func (lds LinkedDIDs) FindByDID(d did.DID) (LinkedDID, bool) {
	for _, ld := range lds {
		if ld.DID.Equals(d) {
			return ld, true
		}
	}
	return LinkedDID{}, false
}

// FindByID finds a LinkedDID by its ID.
func (lds LinkedDIDs) FindByID(id string) (LinkedDID, bool) {
	for _, ld := range lds {
		if ld.ID == id {
			return ld, true
		}
	}
	return LinkedDID{}, false
}

// FindByWalletAddress finds a LinkedDID by wallet address.
func (lds LinkedDIDs) FindByWalletAddress(address string) (LinkedDID, bool) {
	for _, ld := range lds {
		if ld.Source == DIDSourceWallet && ld.WalletAddress() == address {
			return ld, true
		}
	}
	return LinkedDID{}, false
}

// BySource returns all LinkedDIDs from a specific source.
func (lds LinkedDIDs) BySource(source DIDSource) LinkedDIDs {
	var result LinkedDIDs
	for _, ld := range lds {
		if ld.Source == source {
			result = append(result, ld)
		}
	}
	return result
}

// WalletDIDs returns all wallet-derived DIDs.
func (lds LinkedDIDs) WalletDIDs() LinkedDIDs {
	return lds.BySource(DIDSourceWallet)
}

// CustodialDIDs returns all custodial DIDs.
func (lds LinkedDIDs) CustodialDIDs() LinkedDIDs {
	return lds.BySource(DIDSourceCustodial)
}

// DIDs returns all DID strings.
func (lds LinkedDIDs) DIDs() []string {
	result := make([]string, len(lds))
	for i, ld := range lds {
		result[i] = ld.DID.String()
	}
	return result
}

// HasDID checks if a specific DID is in the collection.
func (lds LinkedDIDs) HasDID(d did.DID) bool {
	_, found := lds.FindByDID(d)
	return found
}
