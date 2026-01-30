// Package wallet provides the bounded context for multi-chain wallet management.
//
// # Purpose
//
// Wallet bridges Web3 wallets to the Proof platform. It handles wallet address
// validation, signature verification (SIWE), challenge-response authentication,
// and DID derivation from wallet addresses. Wallet enables self-sovereign
// identity by allowing users to authenticate with cryptographic keys they
// control.
//
// # Core Responsibilities
//
//   - Validate wallet addresses across multiple chains
//   - Generate and manage authentication challenges (nonces)
//   - Verify SIWE (Sign-In with Ethereum) signatures
//   - Derive DIDs from wallet addresses (did:pkh)
//   - Link multiple wallets to a single user identity
//   - Track wallet metadata (chain, label, primary status)
//
// # Key Entities
//
//   - Wallet: The aggregate root representing a linked wallet. Contains
//     address, chain information, verification status, and metadata.
//
//   - Address: A validated blockchain address with chain context.
//     Supports EVM chains with extensibility for others.
//
//   - Chain: Represents a blockchain network (Ethereum, Polygon, etc.)
//     with chain ID and configuration.
//
//   - Challenge: A time-limited nonce for signature verification.
//     Prevents replay attacks.
//
//   - Signature: A cryptographic signature from the wallet owner.
//     Used to prove address ownership.
//
// # Domain Concepts
//
// ## SIWE Authentication Flow
//
//	┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
//	│  Client  │     │  Proof   │     │  Client  │     │  Proof   │
//	│  Request │────►│  Issue   │────►│   Sign   │────►│  Verify  │
//	│  Nonce   │     │ Challenge│     │  Message │     │   Sig    │
//	└──────────┘     └──────────┘     └──────────┘     └──────────┘
//	                                                         │
//	                                                         ▼
//	                                                   ┌──────────┐
//	                                                   │  Wallet  │
//	                                                   │  Linked  │
//	                                                   └──────────┘
//
// Detailed flow:
//
//  1. Client requests a challenge for address 0xABC...
//  2. Proof generates nonce, stores with expiration
//  3. Proof returns SIWE message template
//  4. Client signs message with wallet (MetaMask, etc.)
//  5. Client submits signature to Proof
//  6. Proof verifies signature recovers to claimed address
//  7. Proof links wallet to user (or creates new user)
//  8. Session created, DID derived
//
// ## SIWE Message Format
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│  proof.dev wants you to sign in with your Ethereum account:    │
//	│  0x1234567890abcdef1234567890abcdef12345678                     │
//	│                                                                 │
//	│  Sign in to Proof                                               │
//	│                                                                 │
//	│  URI: https://proof.dev                                         │
//	│  Version: 1                                                     │
//	│  Chain ID: 1                                                    │
//	│  Nonce: abc123xyz                                               │
//	│  Issued At: 2024-01-15T10:30:00Z                               │
//	│  Expiration Time: 2024-01-15T10:45:00Z                         │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## Multi-Chain Support
//
//	| Chain         | Chain ID | Address Format        | DID Method          |
//	|---------------|----------|-----------------------|---------------------|
//	| Ethereum      | 1        | 0x... (20 bytes)      | did:pkh:eip155:1:   |
//	| Polygon       | 137      | 0x... (20 bytes)      | did:pkh:eip155:137: |
//	| Arbitrum      | 42161    | 0x... (20 bytes)      | did:pkh:eip155:42161:|
//	| Optimism      | 10       | 0x... (20 bytes)      | did:pkh:eip155:10:  |
//	| Base          | 8453     | 0x... (20 bytes)      | did:pkh:eip155:8453:|
//
// All EVM chains use the same address format and signature scheme,
// making verification consistent across chains.
//
// ## DID Derivation
//
// Wallet addresses are converted to did:pkh identifiers:
//
//	Address: 0x1234567890abcdef1234567890abcdef12345678
//	Chain:   Ethereum (eip155:1)
//	DID:     did:pkh:eip155:1:0x1234567890abcdef1234567890abcdef12345678
//
// The did:pkh method encodes:
//   - Namespace: pkh (public key hash)
//   - Chain namespace: eip155 (EVM chains per CAIP-2)
//   - Chain reference: 1 (Ethereum mainnet)
//   - Address: checksummed address
//
// ## Wallet Aggregate
//
//	Wallet {
//	    ID              WalletID
//	    UserID          identity.UserID
//	    Address         Address
//	    Chain           Chain
//	    DID             did.DID             // Derived did:pkh
//	    Label           string              // User-assigned label
//	    IsPrimary       bool                // Primary wallet for user
//	    Status          WalletStatus        // Active, Unverified, Revoked
//	    VerifiedAt      time.Time
//	    CreatedAt       time.Time
//	}
//
// Users can link multiple wallets:
//
//	User Account
//	├── Wallet 1 (Primary)
//	│   ├── Address: 0xAAA...
//	│   ├── Chain: Ethereum
//	│   └── DID: did:pkh:eip155:1:0xAAA...
//	├── Wallet 2
//	│   ├── Address: 0xAAA... (same address)
//	│   ├── Chain: Polygon
//	│   └── DID: did:pkh:eip155:137:0xAAA...
//	└── Wallet 3
//	    ├── Address: 0xBBB...
//	    ├── Chain: Ethereum
//	    └── DID: did:pkh:eip155:1:0xBBB...
//
// # Relationships to Other Contexts
//
//   - Identity: Wallet provides wallet-based authentication. On successful
//     verification, Identity creates or updates user and session.
//
//   - Credential: Credentials are issued to DIDs derived from wallets.
//     User's primary wallet DID is typically the credential subject.
//
//   - Presentation: Presentations are signed with wallet keys (for did:pkh
//     holders). Proves holder controls the wallet.
//
// # Architecture
//
//	internal/wallet/
//	├── domain/
//	│   ├── wallet.go          // Wallet aggregate root
//	│   ├── address.go         // Address value object
//	│   ├── chain.go           // Chain value object and registry
//	│   ├── signature.go       // Signature value object
//	│   ├── status.go          // WalletStatus enum
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // WalletLinked, WalletUnlinked, etc.
//	│   ├── repository.go      // Repository interface
//	│   └── services.go        // Domain services
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // LinkWallet, UnlinkWallet, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetWallet, ListUserWallets, etc.
//	│       ├── handlers.go    // Query handlers
//	│       ├── repository.go  // Read repository interface
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── challenge/
//	│   │   └── service.go     // Challenge generation
//	│   ├── did/
//	│   │   └── derivation.go  // did:pkh derivation from address
//	│   ├── nonce/
//	│   │   └── service.go     // Nonce generation and storage
//	│   ├── signature/
//	│   │   ├── siwe.go        // SIWE message handling
//	│   │   └── verifier.go    // Signature verification
//	│   ├── validation/
//	│   │   ├── evm.go         // EVM address validation
//	│   │   └── validator.go   // Address validator interface
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── wallet.go
//	│           ├── challenge.go
//	│           ├── nonce.go
//	│           ├── mapper.go
//	│           ├── queries.go
//	│           ├── reader.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── auth.go        // SIWE auth endpoints
//	│           ├── challenge.go   // Challenge endpoints
//	│           ├── wallet.go      // Wallet management endpoints
//	│           ├── requests.go
//	│           ├── responses.go
//	│           ├── errors.go
//	│           └── routes.go
//	└── provider.go            // Dependency injection
//
// # Events
//
//   - ChallengeCreated: Auth challenge issued for address
//   - ChallengeExpired: Challenge reached expiration
//   - SignatureVerified: Valid signature received
//   - SignatureRejected: Invalid signature rejected
//   - WalletLinked: Wallet linked to user account
//   - WalletUnlinked: Wallet removed from user account
//   - WalletSetPrimary: Wallet designated as primary
//   - WalletLabelUpdated: Wallet label changed
//
// # Security Considerations
//
//   - Challenges are time-limited (typically 15 minutes)
//   - Challenges are single-use (consumed on successful verification)
//   - Nonces prevent replay attacks
//   - Address checksums validated before verification
//   - Signature recovery validates against claimed address
//   - Rate limiting on challenge generation
//
// # Chain Registry
//
// Supported chains are configured in the chain registry:
//
//	registry := chain.NewRegistry()
//	registry.Register(chain.Ethereum)   // Chain ID 1
//	registry.Register(chain.Polygon)    // Chain ID 137
//	registry.Register(chain.Arbitrum)   // Chain ID 42161
//	// ...
//
// New chains can be added by registering their configuration:
//
//	customChain := chain.Chain{
//	    ID:        12345,
//	    Name:      "CustomChain",
//	    Namespace: "eip155",
//	}
//	registry.Register(customChain)
//
// # Future Considerations
//
//   - Non-EVM chain support (Solana, Cosmos, etc.)
//   - Hardware wallet integration
//   - Multi-sig wallet support
//   - Wallet activity monitoring
//   - ENS/domain name resolution
//   - WalletConnect v2 integration
package wallet
