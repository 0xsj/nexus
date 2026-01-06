package signature

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/0xsj/nexus/platform/internal/wallet/domain"
	"github.com/0xsj/nexus/platform/pkg/siwe"
)

// ============================================================================
// SIWE Verifier
// ============================================================================

// SIWEVerifier verifies Sign-In With Ethereum (EIP-4361) signatures.
type SIWEVerifier struct {
	config SIWEConfig
}

// SIWEConfig contains configuration for SIWE verification.
type SIWEConfig struct {
	// Domain is the expected domain for verification.
	Domain string

	// AllowedDomains is a list of allowed domains (if Domain is empty).
	AllowedDomains []string

	// SkipTimeValidation skips issued_at and expiration checks.
	SkipTimeValidation bool

	// MaxMessageAge is the maximum age of a message.
	MaxMessageAge time.Duration
}

// DefaultSIWEConfig returns default SIWE configuration.
func DefaultSIWEConfig() SIWEConfig {
	return SIWEConfig{
		SkipTimeValidation: false,
		MaxMessageAge:      10 * time.Minute,
	}
}

// NewSIWEVerifier creates a new SIWE verifier.
func NewSIWEVerifier(config SIWEConfig) *SIWEVerifier {
	return &SIWEVerifier{
		config: config,
	}
}

// ============================================================================
// Signature Verification
// ============================================================================

// VerifySignature verifies a signature against a message and address.
func (sv *SIWEVerifier) VerifySignature(ctx context.Context, params domain.VerifySignatureParams) error {
	const op = "SIWEVerifier.VerifySignature"

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate inputs
	if params.Address.IsZero() {
		return domain.ErrInvalidAddress(op, "", "address is required")
	}

	if params.Message == "" {
		return domain.ErrInvalidSignature(op, "message is required")
	}

	if params.Signature.IsZero() {
		return domain.ErrInvalidSignature(op, "signature is required")
	}

	// Recover address from signature
	recoveredAddress, err := sv.recoverAddress(params.Message, params.Signature)
	if err != nil {
		return domain.ErrVerificationFailed(op, err.Error())
	}

	// Compare addresses (case-insensitive for EVM)
	if !strings.EqualFold(recoveredAddress, params.Address.Normalized()) {
		return domain.ErrAddressMismatch(op, params.Address.Normalized(), recoveredAddress)
	}

	return nil
}

// VerifySIWE verifies a Sign-In With Ethereum (EIP-4361) message.
func (sv *SIWEVerifier) VerifySIWE(ctx context.Context, params domain.VerifySIWEParams) (*domain.SIWEVerificationResult, error) {
	const op = "SIWEVerifier.VerifySIWE"

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Parse SIWE message
	siweMessage, err := siwe.Parse(params.Message)
	if err != nil {
		return nil, domain.ErrInvalidSignature(op, "failed to parse SIWE message: "+err.Error())
	}

	// Recover address from signature
	recoveredAddress, err := sv.recoverAddress(params.Message, params.Signature)
	if err != nil {
		return nil, domain.ErrVerificationFailed(op, err.Error())
	}

	// Validate recovered address matches message
	if !strings.EqualFold(recoveredAddress, siweMessage.Address) {
		return nil, domain.ErrAddressMismatch(op, siweMessage.Address, recoveredAddress)
	}

	// Validate expected address if provided
	if params.ExpectedAddress != nil {
		if !strings.EqualFold(recoveredAddress, params.ExpectedAddress.Normalized()) {
			return nil, domain.ErrAddressMismatch(op, params.ExpectedAddress.Normalized(), recoveredAddress)
		}
	}

	// Validate domain if provided
	if params.ExpectedDomain != "" {
		if siweMessage.Domain != params.ExpectedDomain {
			return nil, domain.ErrVerificationFailed(op, "domain mismatch")
		}
	} else if sv.config.Domain != "" {
		if siweMessage.Domain != sv.config.Domain {
			return nil, domain.ErrVerificationFailed(op, "domain mismatch")
		}
	}

	// Validate nonce if provided
	if params.ExpectedNonce != "" {
		if siweMessage.Nonce != params.ExpectedNonce {
			return nil, domain.ErrInvalidNonce(op, params.ExpectedNonce)
		}
	}

	// Validate chain ID if provided
	if params.ExpectedChainID != nil {
		messageChainID := strconv.Itoa(siweMessage.ChainID)
		if messageChainID != params.ExpectedChainID.String() {
			return nil, domain.ErrChainMismatch(op, params.ExpectedChainID.String(), messageChainID)
		}
	}

	// Validate time constraints
	if !sv.config.SkipTimeValidation {
		if err := sv.validateTime(siweMessage); err != nil {
			return nil, err
		}
	}

	// Build result
	chainIDStr := strconv.Itoa(siweMessage.ChainID)
	chainID := domain.ChainID(chainIDStr)
	if chainID.IsEmpty() {
		chainID = domain.ChainIDEthereumMainnet
	}

	address := domain.NewAddressUnchecked(
		recoveredAddress,
		recoveredAddress,
		chainID,
		domain.ChainFamilyEVM,
	)

	siweMsg := sv.toSIWEMessage(siweMessage, address)

	return &domain.SIWEVerificationResult{
		Valid:   true,
		Address: address,
		Message: siweMsg,
	}, nil
}

// RecoverAddress recovers the signer address from a signature.
func (sv *SIWEVerifier) RecoverAddress(ctx context.Context, params domain.RecoverAddressParams) (domain.Address, error) {
	const op = "SIWEVerifier.RecoverAddress"

	// Check context
	select {
	case <-ctx.Done():
		return domain.Address{}, ctx.Err()
	default:
	}

	// Recover address
	recoveredAddress, err := sv.recoverAddress(params.Message, params.Signature)
	if err != nil {
		return domain.Address{}, domain.ErrVerificationFailed(op, err.Error())
	}

	// Build domain address
	chainID := params.ChainID
	if chainID.IsEmpty() {
		chainID = domain.ChainIDEthereumMainnet
	}

	return domain.NewAddressUnchecked(
		recoveredAddress,
		recoveredAddress,
		chainID,
		domain.ChainFamilyEVM,
	), nil
}

// ============================================================================
// Internal Methods
// ============================================================================

// recoverAddress recovers the Ethereum address from a message and signature.
func (sv *SIWEVerifier) recoverAddress(message string, signature domain.Signature) (string, error) {
	const op = "SIWEVerifier.recoverAddress"

	// Get signature bytes
	sigBytes := signature.Bytes()
	if len(sigBytes) != 65 {
		return "", domain.ErrInvalidSignature(op, "signature must be 65 bytes")
	}

	// Prepare message hash (Ethereum signed message)
	messageHash := hashEthereumSignedMessage(message)

	// Recover public key from signature
	// sigBytes format: [R (32 bytes)][S (32 bytes)][V (1 byte)]
	r := sigBytes[0:32]
	s := sigBytes[32:64]
	vByte := sigBytes[64]

	// Normalize V value (27/28 or 0/1)
	recoveryID := vByte
	if recoveryID >= 27 {
		recoveryID -= 27
	}
	if recoveryID != 0 && recoveryID != 1 {
		return "", domain.ErrInvalidSignature(op, "invalid recovery id")
	}

	// Recover address using ECDSA
	address, err := ecrecover(messageHash, r, s, recoveryID)
	if err != nil {
		return "", domain.ErrVerificationFailed(op, err.Error())
	}

	return address, nil
}

// validateTime validates the time constraints of a SIWE message.
func (sv *SIWEVerifier) validateTime(msg *siwe.Message) error {
	const op = "SIWEVerifier.validateTime"
	now := time.Now()

	// Check issued_at
	if !msg.IssuedAt.IsZero() {
		// Message shouldn't be from the future
		if msg.IssuedAt.After(now.Add(time.Minute)) {
			return domain.ErrVerificationFailed(op, "message issued in the future")
		}

		// Check max age
		if sv.config.MaxMessageAge > 0 {
			if now.Sub(msg.IssuedAt) > sv.config.MaxMessageAge {
				return domain.ErrSignatureExpired(op)
			}
		}
	}

	// Check expiration
	if msg.ExpirationTime != nil && now.After(*msg.ExpirationTime) {
		return domain.ErrSignatureExpired(op)
	}

	// Check not_before
	if msg.NotBefore != nil && now.Before(*msg.NotBefore) {
		return domain.ErrVerificationFailed(op, "message not yet valid")
	}

	return nil
}

// toSIWEMessage converts a parsed SIWE message to domain type.
func (sv *SIWEVerifier) toSIWEMessage(msg *siwe.Message, address domain.Address) domain.SIWEMessage {
	chainIDStr := strconv.Itoa(msg.ChainID)

	return domain.SIWEMessage{
		Domain:         msg.Domain,
		Address:        address,
		Statement:      msg.Statement,
		URI:            msg.URI,
		Version:        msg.Version,
		ChainID:        domain.ChainID(chainIDStr),
		Nonce:          msg.Nonce,
		IssuedAt:       msg.IssuedAt,
		ExpirationTime: msg.ExpirationTime,
		NotBefore:      msg.NotBefore,
		RequestID:      msg.RequestID,
		Resources:      msg.Resources,
	}
}

// ============================================================================
// Ethereum Message Hashing
// ============================================================================

// hashEthereumSignedMessage hashes a message with the Ethereum signed message prefix.
// Format: "\x19Ethereum Signed Message:\n" + len(message) + message
func hashEthereumSignedMessage(message string) []byte {
	prefix := "\x19Ethereum Signed Message:\n"
	lenStr := strconv.Itoa(len(message))
	prefixedMessage := prefix + lenStr + message

	// TODO: Use proper Keccak256 from golang.org/x/crypto/sha3
	return keccak256([]byte(prefixedMessage))
}

// keccak256 computes the Keccak-256 hash.
// TODO: Replace with proper implementation from golang.org/x/crypto/sha3
func keccak256(data []byte) []byte {
	// Placeholder - in production use:
	// hasher := sha3.NewLegacyKeccak256()
	// hasher.Write(data)
	// return hasher.Sum(nil)

	result := make([]byte, 32)
	copy(result, data)
	return result
}

// ecrecover recovers the Ethereum address from a signature.
// TODO: Replace with proper implementation using github.com/ethereum/go-ethereum/crypto
func ecrecover(hash, r, s []byte, v byte) (string, error) {
	// Placeholder - in production use:
	// sig := append(append(r, s...), v)
	// pubKey, err := crypto.SigToPub(hash, sig)
	// if err != nil {
	//     return "", err
	// }
	// return crypto.PubkeyToAddress(*pubKey).Hex(), nil

	// For development, return a placeholder
	// This MUST be replaced with proper implementation
	return "0x" + hex.EncodeToString(hash[:20]), nil
}
