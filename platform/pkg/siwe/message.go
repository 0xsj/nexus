package siwe

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// Message
// ============================================================================

// Message represents a SIWE (EIP-4361) message.
type Message struct {
	// Domain is the RFC 3986 authority requesting the signing.
	Domain string

	// Address is the Ethereum address performing the signing.
	Address string

	// Statement is a human-readable message (optional).
	Statement string

	// URI is the RFC 3986 URI referring to the resource.
	URI string

	// Version is the SIWE version (currently "1").
	Version string

	// ChainID is the EIP-155 chain ID.
	ChainID int

	// Nonce is a random string for replay protection.
	Nonce string

	// IssuedAt is the time the message was created.
	IssuedAt time.Time

	// ExpirationTime is when the message expires (optional).
	ExpirationTime *time.Time

	// NotBefore is when the message becomes valid (optional).
	NotBefore *time.Time

	// RequestID is an application-specific identifier (optional).
	RequestID string

	// Resources is a list of URIs the user wishes to access (optional).
	Resources []string
}

// ============================================================================
// Constructor
// ============================================================================

// NewMessage creates a new SIWE message with required fields.
func NewMessage(domain, address, uri string, chainID int, nonce string) *Message {
	return &Message{
		Domain:   domain,
		Address:  address,
		URI:      uri,
		Version:  "1",
		ChainID:  chainID,
		Nonce:    nonce,
		IssuedAt: time.Now().UTC(),
	}
}

// ============================================================================
// Builder Methods
// ============================================================================

// WithStatement sets the statement.
func (m *Message) WithStatement(statement string) *Message {
	m.Statement = statement
	return m
}

// WithExpirationTime sets the expiration time.
func (m *Message) WithExpirationTime(exp time.Time) *Message {
	m.ExpirationTime = &exp
	return m
}

// WithExpiresIn sets expiration relative to issued time.
func (m *Message) WithExpiresIn(d time.Duration) *Message {
	exp := m.IssuedAt.Add(d)
	m.ExpirationTime = &exp
	return m
}

// WithNotBefore sets the not-before time.
func (m *Message) WithNotBefore(nbf time.Time) *Message {
	m.NotBefore = &nbf
	return m
}

// WithRequestID sets the request ID.
func (m *Message) WithRequestID(requestID string) *Message {
	m.RequestID = requestID
	return m
}

// WithResources sets the resources.
func (m *Message) WithResources(resources ...string) *Message {
	m.Resources = resources
	return m
}

// ============================================================================
// Serialization
// ============================================================================

// String returns the EIP-4361 formatted message string.
// This is the message that gets signed by the wallet.
func (m *Message) String() string {
	var sb strings.Builder

	// Line 1: {domain} wants you to sign in with your Ethereum account:
	sb.WriteString(m.Domain)
	sb.WriteString(" wants you to sign in with your Ethereum account:\n")

	// Line 2: {address}
	sb.WriteString(m.Address)
	sb.WriteString("\n")

	// Line 3: (empty line) or statement
	if m.Statement != "" {
		sb.WriteString("\n")
		sb.WriteString(m.Statement)
		sb.WriteString("\n")
	}

	// Required fields
	sb.WriteString("\n")
	sb.WriteString("URI: ")
	sb.WriteString(m.URI)
	sb.WriteString("\n")

	sb.WriteString("Version: ")
	sb.WriteString(m.Version)
	sb.WriteString("\n")

	sb.WriteString("Chain ID: ")
	sb.WriteString(strconv.Itoa(m.ChainID))
	sb.WriteString("\n")

	sb.WriteString("Nonce: ")
	sb.WriteString(m.Nonce)
	sb.WriteString("\n")

	sb.WriteString("Issued At: ")
	sb.WriteString(m.IssuedAt.Format(time.RFC3339))

	// Optional fields
	if m.ExpirationTime != nil {
		sb.WriteString("\n")
		sb.WriteString("Expiration Time: ")
		sb.WriteString(m.ExpirationTime.Format(time.RFC3339))
	}

	if m.NotBefore != nil {
		sb.WriteString("\n")
		sb.WriteString("Not Before: ")
		sb.WriteString(m.NotBefore.Format(time.RFC3339))
	}

	if m.RequestID != "" {
		sb.WriteString("\n")
		sb.WriteString("Request ID: ")
		sb.WriteString(m.RequestID)
	}

	if len(m.Resources) > 0 {
		sb.WriteString("\n")
		sb.WriteString("Resources:")
		for _, resource := range m.Resources {
			sb.WriteString("\n- ")
			sb.WriteString(resource)
		}
	}

	return sb.String()
}

// ============================================================================
// Parsing
// ============================================================================

// Parse parses a SIWE message string into a Message struct.
func Parse(message string) (*Message, error) {
	const op = "siwe.Parse"

	lines := strings.Split(message, "\n")
	if len(lines) < 6 {
		return nil, ErrInvalidMessage(op, "message too short")
	}

	m := &Message{}

	// Line 1: {domain} wants you to sign in with your Ethereum account:
	domainLine := lines[0]
	domainMatch := regexp.MustCompile(`^(.+) wants you to sign in with your Ethereum account:$`).FindStringSubmatch(domainLine)
	if domainMatch == nil {
		return nil, ErrInvalidMessage(op, "invalid first line")
	}
	m.Domain = domainMatch[1]

	// Line 2: {address}
	m.Address = strings.TrimSpace(lines[1])
	if !IsValidAddress(m.Address) {
		return nil, ErrInvalidAddress(op, m.Address)
	}

	// Parse remaining lines
	lineIndex := 2

	// Check for statement (non-empty line after address, before URI)
	// Skip empty lines first
	for lineIndex < len(lines) && strings.TrimSpace(lines[lineIndex]) == "" {
		lineIndex++
	}

	// Collect statement lines until we hit a field
	var statementLines []string
	for lineIndex < len(lines) {
		line := lines[lineIndex]
		if strings.HasPrefix(line, "URI:") ||
			strings.HasPrefix(line, "Version:") ||
			strings.HasPrefix(line, "Chain ID:") ||
			strings.HasPrefix(line, "Nonce:") ||
			strings.HasPrefix(line, "Issued At:") {
			break
		}
		if strings.TrimSpace(line) != "" {
			statementLines = append(statementLines, line)
		}
		lineIndex++
	}
	m.Statement = strings.Join(statementLines, "\n")

	// Parse key-value fields
	for lineIndex < len(lines) {
		line := strings.TrimSpace(lines[lineIndex])
		lineIndex++

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "URI: ") {
			m.URI = strings.TrimPrefix(line, "URI: ")
		} else if strings.HasPrefix(line, "Version: ") {
			m.Version = strings.TrimPrefix(line, "Version: ")
		} else if strings.HasPrefix(line, "Chain ID: ") {
			chainIDStr := strings.TrimPrefix(line, "Chain ID: ")
			chainID, err := strconv.Atoi(chainIDStr)
			if err != nil {
				return nil, ErrInvalidChainID(op, chainIDStr)
			}
			m.ChainID = chainID
		} else if strings.HasPrefix(line, "Nonce: ") {
			m.Nonce = strings.TrimPrefix(line, "Nonce: ")
		} else if strings.HasPrefix(line, "Issued At: ") {
			issuedAtStr := strings.TrimPrefix(line, "Issued At: ")
			issuedAt, err := time.Parse(time.RFC3339, issuedAtStr)
			if err != nil {
				return nil, ErrInvalidMessage(op, "invalid Issued At: "+issuedAtStr)
			}
			m.IssuedAt = issuedAt
		} else if strings.HasPrefix(line, "Expiration Time: ") {
			expStr := strings.TrimPrefix(line, "Expiration Time: ")
			exp, err := time.Parse(time.RFC3339, expStr)
			if err != nil {
				return nil, ErrInvalidMessage(op, "invalid Expiration Time: "+expStr)
			}
			m.ExpirationTime = &exp
		} else if strings.HasPrefix(line, "Not Before: ") {
			nbfStr := strings.TrimPrefix(line, "Not Before: ")
			nbf, err := time.Parse(time.RFC3339, nbfStr)
			if err != nil {
				return nil, ErrInvalidMessage(op, "invalid Not Before: "+nbfStr)
			}
			m.NotBefore = &nbf
		} else if strings.HasPrefix(line, "Request ID: ") {
			m.RequestID = strings.TrimPrefix(line, "Request ID: ")
		} else if line == "Resources:" {
			// Parse resources
			for lineIndex < len(lines) {
				resourceLine := lines[lineIndex]
				if !strings.HasPrefix(resourceLine, "- ") {
					break
				}
				m.Resources = append(m.Resources, strings.TrimPrefix(resourceLine, "- "))
				lineIndex++
			}
		}
	}

	// Validate required fields
	if err := m.ValidateFormat(); err != nil {
		return nil, err
	}

	return m, nil
}

// ============================================================================
// Validation
// ============================================================================

// ValidateFormat validates the message format (not timing).
func (m *Message) ValidateFormat() error {
	const op = "siwe.Message.ValidateFormat"

	if m.Domain == "" {
		return ErrInvalidDomain(op, "domain is required")
	}

	if !IsValidAddress(m.Address) {
		return ErrInvalidAddress(op, m.Address)
	}

	if m.URI == "" {
		return ErrInvalidURI(op, "URI is required")
	}

	if _, err := url.Parse(m.URI); err != nil {
		return ErrInvalidURI(op, m.URI)
	}

	if m.Version != "1" {
		return ErrInvalidMessage(op, "unsupported version: "+m.Version)
	}

	if m.ChainID <= 0 {
		return ErrInvalidChainID(op, strconv.Itoa(m.ChainID))
	}

	if m.Nonce == "" {
		return ErrInvalidNonce(op)
	}

	if m.IssuedAt.IsZero() {
		return ErrInvalidMessage(op, "Issued At is required")
	}

	return nil
}

// Validate validates the message format and timing.
func (m *Message) Validate() error {
	return m.ValidateAt(time.Now())
}

// ValidateAt validates the message at a specific time.
func (m *Message) ValidateAt(now time.Time) error {
	const op = "siwe.Message.Validate"

	// Validate format first
	if err := m.ValidateFormat(); err != nil {
		return err
	}

	// Check expiration
	if m.ExpirationTime != nil && now.After(*m.ExpirationTime) {
		return ErrMessageExpired(op)
	}

	// Check not-before
	if m.NotBefore != nil && now.Before(*m.NotBefore) {
		return ErrMessageNotYetValid(op)
	}

	return nil
}

// ValidateDomain validates the domain matches expected.
func (m *Message) ValidateDomain(expected string) error {
	const op = "siwe.Message.ValidateDomain"

	if m.Domain != expected {
		return ErrDomainMismatch(op, expected, m.Domain)
	}
	return nil
}

// ValidateNonce validates the nonce matches expected.
func (m *Message) ValidateNonce(expected string) error {
	const op = "siwe.Message.ValidateNonce"

	if m.Nonce != expected {
		return ErrNonceMismatch(op)
	}
	return nil
}

// IsExpired returns true if the message has expired.
func (m *Message) IsExpired() bool {
	if m.ExpirationTime == nil {
		return false
	}
	return time.Now().After(*m.ExpirationTime)
}

// ============================================================================
// Address Utilities
// ============================================================================

// IsValidAddress checks if a string is a valid Ethereum address.
func IsValidAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Check hex characters
	for _, c := range address[2:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// NormalizeAddress normalizes an Ethereum address to checksum format.
// For simplicity, this just lowercases. Full EIP-55 checksum can be added.
func NormalizeAddress(address string) string {
	return strings.ToLower(address)
}

// AddressesEqual compares two addresses case-insensitively.
func AddressesEqual(a, b string) bool {
	return strings.EqualFold(a, b)
}

// ============================================================================
// Nonce Generation
// ============================================================================

// GenerateNonce generates a cryptographically secure random nonce.
func GenerateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateNonceWithLength generates a nonce of specified byte length.
func GenerateNonceWithLength(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ============================================================================
// Chain IDs
// ============================================================================

// Common Ethereum chain IDs.
const (
	ChainIDEthereum       = 1
	ChainIDGoerli         = 5
	ChainIDSepolia        = 11155111
	ChainIDPolygon        = 137
	ChainIDPolygonMumbai  = 80001
	ChainIDArbitrum       = 42161
	ChainIDArbitrumGoerli = 421613
	ChainIDOptimism       = 10
	ChainIDOptimismGoerli = 420
	ChainIDBase           = 8453
	ChainIDBaseGoerli     = 84531
)

// ChainName returns a human-readable name for a chain ID.
func ChainName(chainID int) string {
	switch chainID {
	case ChainIDEthereum:
		return "Ethereum Mainnet"
	case ChainIDGoerli:
		return "Goerli"
	case ChainIDSepolia:
		return "Sepolia"
	case ChainIDPolygon:
		return "Polygon"
	case ChainIDPolygonMumbai:
		return "Polygon Mumbai"
	case ChainIDArbitrum:
		return "Arbitrum One"
	case ChainIDArbitrumGoerli:
		return "Arbitrum Goerli"
	case ChainIDOptimism:
		return "Optimism"
	case ChainIDOptimismGoerli:
		return "Optimism Goerli"
	case ChainIDBase:
		return "Base"
	case ChainIDBaseGoerli:
		return "Base Goerli"
	default:
		return fmt.Sprintf("Chain %d", chainID)
	}
}
