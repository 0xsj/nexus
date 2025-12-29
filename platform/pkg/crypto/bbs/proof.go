package bbs

import (
	"github.com/0xsj/nexus/platform/pkg/crypto"
)

// Proof represents a BBS+ zero-knowledge proof.
// Enables selective disclosure of signed messages.
type Proof struct {
	// ProofBytes is the raw proof data.
	ProofBytes []byte

	// RevealedMessages contains the disclosed messages.
	// Key is the message index, value is the message content.
	RevealedMessages map[int][]byte

	// RevealedIndexes indicates which message indexes are disclosed.
	RevealedIndexes []int

	// Nonce is the verifier-provided nonce bound to the proof.
	Nonce []byte
}

// ProofRequest specifies which messages to reveal in a proof.
type ProofRequest struct {
	// RevealIndexes specifies which message indexes to reveal.
	// Messages not in this list remain hidden.
	RevealIndexes []int

	// Nonce is a verifier-provided value to prevent replay.
	Nonce []byte
}

// NewProofRequest creates a new proof request.
func NewProofRequest(revealIndexes []int, nonce []byte) *ProofRequest {
	return &ProofRequest{
		RevealIndexes: revealIndexes,
		Nonce:         nonce,
	}
}

// DeriveProof creates a zero-knowledge proof from a signature.
// Reveals only the messages at the specified indexes.
// Phase 2: Not yet implemented.
func DeriveProof(
	publicKey []byte,
	signature []byte,
	messages [][]byte,
	request *ProofRequest,
) (*Proof, error) {
	const op = "bbs.DeriveProof"
	return nil, crypto.ErrUnsupportedAlgorithm(op, crypto.AlgorithmBLS12381)
}

// IsRevealed returns true if the message at index is revealed.
func (p *Proof) IsRevealed(index int) bool {
	for _, i := range p.RevealedIndexes {
		if i == index {
			return true
		}
	}
	return false
}

// GetRevealedMessage returns the revealed message at index.
// Returns nil if the message is not revealed.
func (p *Proof) GetRevealedMessage(index int) []byte {
	return p.RevealedMessages[index]
}

// RevealedCount returns the number of revealed messages.
func (p *Proof) RevealedCount() int {
	return len(p.RevealedIndexes)
}

// TotalMessages returns the total number of messages in the original signature.
// Phase 2: This will be encoded in the proof.
func (p *Proof) TotalMessages() int {
	// Phase 2: extract from proof structure
	return 0
}

// HiddenCount returns the number of hidden messages.
func (p *Proof) HiddenCount() int {
	return p.TotalMessages() - p.RevealedCount()
}

// Bytes returns the raw proof bytes.
func (p *Proof) Bytes() []byte {
	return p.ProofBytes
}

// ============================================================================
// Credential Selective Disclosure
// ============================================================================

// CredentialProofRequest represents a request to prove credential attributes.
// Maps credential field names to reveal/hide decisions.
type CredentialProofRequest struct {
	// RevealFields lists the credential fields to reveal.
	// Example: ["degree", "institution"] reveals these, hides others.
	RevealFields []string

	// Predicates lists conditions to prove without revealing values.
	// Phase 2: Example: "age >= 18" proves condition without revealing age.
	Predicates []Predicate

	// Nonce is verifier-provided randomness.
	Nonce []byte
}

// Predicate represents a condition to prove without revealing the value.
// Phase 2: Not yet implemented.
type Predicate struct {
	// Field is the credential field name.
	Field string

	// Operator is the comparison operator (>=, <=, ==, !=, >, <).
	Operator string

	// Value is the threshold/comparison value.
	Value interface{}
}

// NewCredentialProofRequest creates a new credential proof request.
func NewCredentialProofRequest(revealFields []string, nonce []byte) *CredentialProofRequest {
	return &CredentialProofRequest{
		RevealFields: revealFields,
		Predicates:   nil,
		Nonce:        nonce,
	}
}

// WithPredicate adds a predicate to the proof request.
func (r *CredentialProofRequest) WithPredicate(field, operator string, value interface{}) *CredentialProofRequest {
	r.Predicates = append(r.Predicates, Predicate{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return r
}
