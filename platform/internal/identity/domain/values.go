package domain

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

// ============================================================================
// Identity Type
// ============================================================================

// IdentityType represents how a user identifies themselves.
// DID is the primary identity; others are linked identities for convenience.
type IdentityType string

const (
	// IdentityTypeDID is a decentralized identifier (primary, self-sovereign).
	IdentityTypeDID IdentityType = "did"

	// IdentityTypeWallet is a blockchain wallet address.
	IdentityTypeWallet IdentityType = "wallet"

	// IdentityTypeEmail is a traditional email (bridge for Web2 onboarding).
	IdentityTypeEmail IdentityType = "email"

	// IdentityTypeENS is an Ethereum Name Service domain.
	IdentityTypeENS IdentityType = "ens"

	// IdentityTypeLens is a Lens Protocol handle.
	IdentityTypeLens IdentityType = "lens"
)

// String returns the string representation.
func (t IdentityType) String() string {
	return string(t)
}

// IsValid checks if the identity type is supported.
func (t IdentityType) IsValid() bool {
	switch t {
	case IdentityTypeDID, IdentityTypeWallet, IdentityTypeEmail, IdentityTypeENS, IdentityTypeLens:
		return true
	default:
		return false
	}
}

// IsDecentralized returns true if the identity type is self-sovereign.
func (t IdentityType) IsDecentralized() bool {
	switch t {
	case IdentityTypeDID, IdentityTypeWallet, IdentityTypeENS, IdentityTypeLens:
		return true
	default:
		return false
	}
}

// IsBridge returns true if the identity type is for Web2 compatibility.
func (t IdentityType) IsBridge() bool {
	return t == IdentityTypeEmail
}

// ============================================================================
// Auth Method
// ============================================================================

// AuthMethod represents how a user authenticates.
// Wallet signatures and DID auth are primary; OAuth is a bridge.
type AuthMethod string

const (
	// AuthMethodWallet is wallet signature auth (EIP-4361 SIWE, etc.).
	AuthMethodWallet AuthMethod = "wallet"

	// AuthMethodDIDAuth is DID-based authentication (DID AuthN).
	AuthMethodDIDAuth AuthMethod = "did_auth"

	// AuthMethodPasskey is WebAuthn/FIDO2 passkey auth.
	AuthMethodPasskey AuthMethod = "passkey"

	// AuthMethodOAuth is OAuth provider auth (bridge for Web2).
	AuthMethodOAuth AuthMethod = "oauth"

	// AuthMethodVerifiablePresentation is VP-based auth.
	AuthMethodVerifiablePresentation AuthMethod = "vp"
)

// String returns the string representation.
func (m AuthMethod) String() string {
	return string(m)
}

// IsValid checks if the auth method is supported.
func (m AuthMethod) IsValid() bool {
	switch m {
	case AuthMethodWallet, AuthMethodDIDAuth, AuthMethodPasskey, AuthMethodOAuth, AuthMethodVerifiablePresentation:
		return true
	default:
		return false
	}
}

// IsDecentralized returns true if the auth method is self-sovereign.
func (m AuthMethod) IsDecentralized() bool {
	switch m {
	case AuthMethodWallet, AuthMethodDIDAuth, AuthMethodPasskey, AuthMethodVerifiablePresentation:
		return true
	default:
		return false
	}
}

// IsBridge returns true if the auth method is for Web2 compatibility.
func (m AuthMethod) IsBridge() bool {
	return m == AuthMethodOAuth
}

// ============================================================================
// Wallet Provider
// ============================================================================

// WalletProvider represents supported wallet types.
type WalletProvider string

const (
	WalletProviderMetaMask      WalletProvider = "metamask"
	WalletProviderWalletConnect WalletProvider = "walletconnect"
	WalletProviderCoinbase      WalletProvider = "coinbase"
	WalletProviderRainbow       WalletProvider = "rainbow"
	WalletProviderPhantom       WalletProvider = "phantom"
	WalletProviderKeplr         WalletProvider = "keplr"
	WalletProviderLeap          WalletProvider = "leap"
)

// String returns the string representation.
func (p WalletProvider) String() string {
	return string(p)
}

// IsValid checks if the wallet provider is supported.
func (p WalletProvider) IsValid() bool {
	switch p {
	case WalletProviderMetaMask, WalletProviderWalletConnect, WalletProviderCoinbase,
		WalletProviderRainbow, WalletProviderPhantom, WalletProviderKeplr, WalletProviderLeap:
		return true
	default:
		return false
	}
}

// DisplayName returns the display name.
func (p WalletProvider) DisplayName() string {
	switch p {
	case WalletProviderMetaMask:
		return "MetaMask"
	case WalletProviderWalletConnect:
		return "WalletConnect"
	case WalletProviderCoinbase:
		return "Coinbase Wallet"
	case WalletProviderRainbow:
		return "Rainbow"
	case WalletProviderPhantom:
		return "Phantom"
	case WalletProviderKeplr:
		return "Keplr"
	case WalletProviderLeap:
		return "Leap"
	default:
		return string(p)
	}
}

// DefaultChain returns the default chain for the wallet provider.
func (p WalletProvider) DefaultChain() Chain {
	switch p {
	case WalletProviderPhantom:
		return ChainSolana
	case WalletProviderKeplr, WalletProviderLeap:
		return ChainCosmos
	default:
		return ChainEthereum
	}
}

// ============================================================================
// Chain
// ============================================================================

// Chain represents a blockchain network.
type Chain string

const (
	ChainEthereum Chain = "ethereum"
	ChainPolygon  Chain = "polygon"
	ChainArbitrum Chain = "arbitrum"
	ChainOptimism Chain = "optimism"
	ChainBase     Chain = "base"
	ChainSolana   Chain = "solana"
	ChainCosmos   Chain = "cosmos"
)

// String returns the string representation.
func (c Chain) String() string {
	return string(c)
}

// IsValid checks if the chain is supported.
func (c Chain) IsValid() bool {
	switch c {
	case ChainEthereum, ChainPolygon, ChainArbitrum, ChainOptimism, ChainBase, ChainSolana, ChainCosmos:
		return true
	default:
		return false
	}
}

// IsEVM returns true if the chain is EVM-compatible.
func (c Chain) IsEVM() bool {
	switch c {
	case ChainEthereum, ChainPolygon, ChainArbitrum, ChainOptimism, ChainBase:
		return true
	default:
		return false
	}
}

// ChainID returns the chain ID for EVM chains.
func (c Chain) ChainID() uint64 {
	switch c {
	case ChainEthereum:
		return 1
	case ChainPolygon:
		return 137
	case ChainArbitrum:
		return 42161
	case ChainOptimism:
		return 10
	case ChainBase:
		return 8453
	default:
		return 0
	}
}

// ChainIDString returns the chain ID as a string.
func (c Chain) ChainIDString() string {
	return fmt.Sprintf("%d", c.ChainID())
}

// DisplayName returns the display name.
func (c Chain) DisplayName() string {
	switch c {
	case ChainEthereum:
		return "Ethereum"
	case ChainPolygon:
		return "Polygon"
	case ChainArbitrum:
		return "Arbitrum"
	case ChainOptimism:
		return "Optimism"
	case ChainBase:
		return "Base"
	case ChainSolana:
		return "Solana"
	case ChainCosmos:
		return "Cosmos"
	default:
		return string(c)
	}
}

// ============================================================================
// User Status (Simplified)
// ============================================================================

// UserStatus represents the status of a user identity.
type UserStatus string

const (
	// UserStatusActive indicates the identity is active.
	UserStatusActive UserStatus = "active"

	// UserStatusSuspended indicates the identity is temporarily suspended.
	UserStatusSuspended UserStatus = "suspended"
)

// String returns the string representation.
func (s UserStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

// CanAuthenticate returns true if the user can authenticate.
func (s UserStatus) CanAuthenticate() bool {
	return s == UserStatusActive
}

// ============================================================================
// Session Status
// ============================================================================

// SessionStatus represents the status of a session.
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

// String returns the string representation.
func (s SessionStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid.
func (s SessionStatus) IsValid() bool {
	switch s {
	case SessionStatusActive, SessionStatusExpired, SessionStatusRevoked:
		return true
	default:
		return false
	}
}

// IsActive returns true if the session is active.
func (s SessionStatus) IsActive() bool {
	return s == SessionStatusActive
}

// ============================================================================
// OAuth Provider (Web2 Bridge)
// ============================================================================

// OAuthProvider represents OAuth providers for Web2 bridging.
// Used to link Web2 accounts and issue credentials from Web2 sources.
type OAuthProvider string

const (
	OAuthProviderGitHub   OAuthProvider = "github"
	OAuthProviderLinkedIn OAuthProvider = "linkedin"
	OAuthProviderGoogle   OAuthProvider = "google"
	OAuthProviderTwitter  OAuthProvider = "twitter"
	OAuthProviderDiscord  OAuthProvider = "discord"
)

// String returns the string representation.
func (p OAuthProvider) String() string {
	return string(p)
}

// IsValid checks if the provider is supported.
func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGitHub, OAuthProviderLinkedIn, OAuthProviderGoogle, OAuthProviderTwitter, OAuthProviderDiscord:
		return true
	default:
		return false
	}
}

// DisplayName returns the display name.
func (p OAuthProvider) DisplayName() string {
	switch p {
	case OAuthProviderGitHub:
		return "GitHub"
	case OAuthProviderLinkedIn:
		return "LinkedIn"
	case OAuthProviderGoogle:
		return "Google"
	case OAuthProviderTwitter:
		return "Twitter/X"
	case OAuthProviderDiscord:
		return "Discord"
	default:
		return string(p)
	}
}

// ParseOAuthProvider parses a string into an OAuthProvider.
func ParseOAuthProvider(s string) OAuthProvider {
	switch strings.ToLower(s) {
	case "github":
		return OAuthProviderGitHub
	case "linkedin":
		return OAuthProviderLinkedIn
	case "google":
		return OAuthProviderGoogle
	case "twitter", "x":
		return OAuthProviderTwitter
	case "discord":
		return OAuthProviderDiscord
	default:
		return OAuthProvider(s)
	}
}

// ============================================================================
// Disclosure Level (ZKP-Ready)
// ============================================================================

// DisclosureLevel represents how much information is revealed in a presentation.
type DisclosureLevel string

const (
	// DisclosureFull reveals the actual value.
	// Example: "age: 25"
	DisclosureFull DisclosureLevel = "full"

	// DisclosureRange reveals a range predicate.
	// Example: "age >= 18" without revealing actual age
	DisclosureRange DisclosureLevel = "range"

	// DisclosureSetMembership reveals membership in a set.
	// Example: "country IN [US, CA, UK]" without revealing which
	DisclosureSetMembership DisclosureLevel = "set_membership"

	// DisclosureExistence reveals only that the claim exists.
	// Example: "has email credential" without revealing email
	DisclosureExistence DisclosureLevel = "existence"

	// DisclosureHash reveals a hash/commitment.
	// Example: hash(email) for later verification
	DisclosureHash DisclosureLevel = "hash"

	// DisclosureZKP reveals nothing, only a zero-knowledge proof.
	// Example: ZK proof that age >= 18
	DisclosureZKP DisclosureLevel = "zkp"
)

// String returns the string representation.
func (d DisclosureLevel) String() string {
	return string(d)
}

// IsValid checks if the disclosure level is valid.
func (d DisclosureLevel) IsValid() bool {
	switch d {
	case DisclosureFull, DisclosureRange, DisclosureSetMembership,
		DisclosureExistence, DisclosureHash, DisclosureZKP:
		return true
	default:
		return false
	}
}

// RequiresZKP returns true if the disclosure level requires ZKP.
func (d DisclosureLevel) RequiresZKP() bool {
	switch d {
	case DisclosureRange, DisclosureSetMembership, DisclosureZKP:
		return true
	default:
		return false
	}
}

// PrivacyScore returns a score indicating privacy level (higher = more private).
func (d DisclosureLevel) PrivacyScore() int {
	switch d {
	case DisclosureFull:
		return 0
	case DisclosureHash:
		return 1
	case DisclosureExistence:
		return 2
	case DisclosureSetMembership:
		return 3
	case DisclosureRange:
		return 4
	case DisclosureZKP:
		return 5
	default:
		return 0
	}
}

// ============================================================================
// API Key Scope
// ============================================================================

// APIKeyScope represents permissions granted to an API key.
type APIKeyScope string

const (
	APIScopeRead              APIKeyScope = "read"
	APIScopeWrite             APIKeyScope = "write"
	APIScopeAdmin             APIKeyScope = "admin"
	APIScopeCredentialsIssue  APIKeyScope = "credentials:issue"
	APIScopeCredentialsVerify APIKeyScope = "credentials:verify"
	APIScopeCredentialsRevoke APIKeyScope = "credentials:revoke"
	APIScopePresentations     APIKeyScope = "presentations"
)

// String returns the string representation.
func (s APIKeyScope) String() string {
	return string(s)
}

// IsValid checks if the scope is valid.
func (s APIKeyScope) IsValid() bool {
	switch s {
	case APIScopeRead, APIScopeWrite, APIScopeAdmin,
		APIScopeCredentialsIssue, APIScopeCredentialsVerify,
		APIScopeCredentialsRevoke, APIScopePresentations:
		return true
	default:
		return false
	}
}

// APIKeyScopes represents a set of scopes.
type APIKeyScopes []APIKeyScope

// Contains checks if the scopes contain a specific scope.
func (s APIKeyScopes) Contains(scope APIKeyScope) bool {
	for _, sc := range s {
		if sc == scope {
			return true
		}
	}
	return false
}

// HasAny checks if the scopes contain any of the given scopes.
func (s APIKeyScopes) HasAny(scopes ...APIKeyScope) bool {
	for _, scope := range scopes {
		if s.Contains(scope) {
			return true
		}
	}
	return false
}

// HasAll checks if the scopes contain all of the given scopes.
func (s APIKeyScopes) HasAll(scopes ...APIKeyScope) bool {
	for _, scope := range scopes {
		if !s.Contains(scope) {
			return false
		}
	}
	return true
}

// Strings returns the scopes as a string slice.
func (s APIKeyScopes) Strings() []string {
	result := make([]string, len(s))
	for i, scope := range s {
		result[i] = scope.String()
	}
	return result
}

// ParseAPIKeyScopes parses a string slice into APIKeyScopes.
func ParseAPIKeyScopes(scopes []string) APIKeyScopes {
	result := make(APIKeyScopes, 0, len(scopes))
	for _, s := range scopes {
		scope := APIKeyScope(s)
		if scope.IsValid() {
			result = append(result, scope)
		}
	}
	return result
}

// ============================================================================
// Token Type
// ============================================================================

// TokenType represents the type of authentication token.
// Used for API compatibility and session management.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// String returns the string representation.
func (t TokenType) String() string {
	return string(t)
}

// IsValid checks if the token type is valid.
func (t TokenType) IsValid() bool {
	switch t {
	case TokenTypeAccess, TokenTypeRefresh:
		return true
	default:
		return false
	}
}

// DefaultExpiration returns the default expiration duration.
func (t TokenType) DefaultExpiration() time.Duration {
	switch t {
	case TokenTypeAccess:
		return 15 * time.Minute
	case TokenTypeRefresh:
		return 7 * 24 * time.Hour
	default:
		return 15 * time.Minute
	}
}

// ============================================================================
// Linked Identity (Value Object)
// ============================================================================

// LinkedIdentity represents a linked external identity (email, OAuth, etc.).
// These are bridges from Web2 to the user's primary DID identity.
type LinkedIdentity struct {
	Type       IdentityType
	Value      string    // email address, OAuth provider user ID, etc.
	Provider   string    // OAuth provider name, chain name, empty for email
	Verified   bool      // Whether the identity has been verified
	VerifiedAt time.Time // When verification occurred
	LinkedAt   time.Time // When the identity was linked
}

// NewEmailIdentity creates a linked email identity.
func NewEmailIdentity(email string, verified bool) (LinkedIdentity, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))

	if normalized == "" {
		return LinkedIdentity{}, ErrInvalidEmail("LinkedIdentity.NewEmail", email)
	}

	// Parse email address
	addr, err := mail.ParseAddress(normalized)
	if err != nil {
		return LinkedIdentity{}, ErrInvalidEmail("LinkedIdentity.NewEmail", email)
	}

	identity := LinkedIdentity{
		Type:     IdentityTypeEmail,
		Value:    addr.Address,
		Verified: verified,
		LinkedAt: time.Now(),
	}

	if verified {
		identity.VerifiedAt = time.Now()
	}

	return identity, nil
}

// NewOAuthIdentity creates a linked OAuth identity.
func NewOAuthIdentity(provider OAuthProvider, providerUserID string) LinkedIdentity {
	return LinkedIdentity{
		Type:       IdentityTypeEmail, // OAuth links as email-like identities
		Value:      providerUserID,
		Provider:   provider.String(),
		Verified:   true, // OAuth is inherently verified
		VerifiedAt: time.Now(),
		LinkedAt:   time.Now(),
	}
}

// NewWalletIdentity creates a linked wallet identity.
func NewWalletIdentity(address string, chain Chain) LinkedIdentity {
	return LinkedIdentity{
		Type:       IdentityTypeWallet,
		Value:      strings.ToLower(address),
		Provider:   chain.String(),
		Verified:   true, // Wallet signatures verify ownership
		VerifiedAt: time.Now(),
		LinkedAt:   time.Now(),
	}
}

// NewENSIdentity creates a linked ENS identity.
func NewENSIdentity(ensName string) LinkedIdentity {
	return LinkedIdentity{
		Type:     IdentityTypeENS,
		Value:    strings.ToLower(ensName),
		Verified: false, // Needs on-chain verification
		LinkedAt: time.Now(),
	}
}

// IsVerified returns true if the identity is verified.
func (l LinkedIdentity) IsVerified() bool {
	return l.Verified
}

// IsBridge returns true if the identity is a Web2 bridge.
func (l LinkedIdentity) IsBridge() bool {
	return l.Type.IsBridge()
}

// Domain returns the domain for email identities.
func (l LinkedIdentity) Domain() string {
	if l.Type != IdentityTypeEmail {
		return ""
	}
	parts := strings.Split(l.Value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// ============================================================================
// Wallet Address (Value Object)
// ============================================================================

// WalletAddress represents a blockchain wallet address.
type WalletAddress struct {
	Address string
	Chain   Chain
}

// NewWalletAddress creates a new wallet address.
func NewWalletAddress(address string, chain Chain) WalletAddress {
	// Normalize: lowercase for EVM, keep original for others
	normalized := address
	if chain.IsEVM() {
		normalized = strings.ToLower(address)
	}

	return WalletAddress{
		Address: normalized,
		Chain:   chain,
	}
}

// String returns the address string.
func (w WalletAddress) String() string {
	return w.Address
}

// IsEVM returns true if the address is on an EVM chain.
func (w WalletAddress) IsEVM() bool {
	return w.Chain.IsEVM()
}

// Equals checks if two addresses are equal.
func (w WalletAddress) Equals(other WalletAddress) bool {
	return w.Address == other.Address && w.Chain == other.Chain
}

// ToDID converts the wallet address to a did:pkh DID.
func (w WalletAddress) ToDID() string {
	if w.Chain.IsEVM() {
		return fmt.Sprintf("did:pkh:eip155:%d:%s", w.Chain.ChainID(), w.Address)
	}
	// Simplified for non-EVM
	return fmt.Sprintf("did:pkh:%s:%s", w.Chain.String(), w.Address)
}

// ============================================================================
// Challenge (for wallet auth)
// ============================================================================

// Challenge represents an authentication challenge for wallet signatures.
type Challenge struct {
	Nonce     string
	Message   string
	Domain    string
	URI       string
	IssuedAt  time.Time
	ExpiresAt time.Time
	Chain     Chain
	Address   string
}

// NewChallenge creates a new authentication challenge.
func NewChallenge(nonce, domain, uri, address string, chain Chain, ttl time.Duration) Challenge {
	now := time.Now()
	return Challenge{
		Nonce:     nonce,
		Domain:    domain,
		URI:       uri,
		Address:   address,
		Chain:     chain,
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
	}
}

// IsExpired returns true if the challenge has expired.
func (c Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsValid returns true if the challenge is valid (not expired).
func (c Challenge) IsValid() bool {
	return !c.IsExpired()
}

// SIWE returns the Sign-In With Ethereum (EIP-4361) message format.
func (c Challenge) SIWE() string {
	return fmt.Sprintf(`%s wants you to sign in with your Ethereum account:
%s

Sign in to Proof

URI: %s
Version: 1
Chain ID: %d
Nonce: %s
Issued At: %s
Expiration Time: %s`,
		c.Domain,
		c.Address,
		c.URI,
		c.Chain.ChainID(),
		c.Nonce,
		c.IssuedAt.UTC().Format(time.RFC3339),
		c.ExpiresAt.UTC().Format(time.RFC3339),
	)
}

// ============================================================================
// Signature (for verification)
// ============================================================================

// Signature represents a cryptographic signature.
type Signature struct {
	Value     string // Hex-encoded signature
	Algorithm string // e.g., "ES256K", "Ed25519", "secp256k1"
	KeyID     string // DID key ID or wallet address
	Created   time.Time
}

// NewSignature creates a new signature.
func NewSignature(value, algorithm, keyID string) Signature {
	return Signature{
		Value:     value,
		Algorithm: algorithm,
		KeyID:     keyID,
		Created:   time.Now(),
	}
}

// IsEmpty returns true if the signature is empty.
func (s Signature) IsEmpty() bool {
	return s.Value == ""
}
