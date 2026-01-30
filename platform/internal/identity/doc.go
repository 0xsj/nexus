// Package identity provides the bounded context for user accounts and authentication.
//
// # Purpose
//
// Identity is the foundation of the Proof platform. It manages user accounts,
// authentication methods, sessions, API keys, and the association of decentralized
// identifiers (DIDs) with user accounts. Every other bounded context depends on
// Identity for user context.
//
// # Core Responsibilities
//
//   - User registration and account management
//   - Authentication via multiple methods (magic link, OAuth, wallet)
//   - Session lifecycle (create, validate, revoke)
//   - API key management for programmatic access
//   - DID management (primary and linked DIDs)
//   - Connection tracking for external OAuth providers
//
// # Key Entities
//
//   - User: The aggregate root representing a registered user. Contains
//     profile information, primary DID, and account status.
//
//   - Session: An authenticated session with expiration and revocation
//     support. Tracks device/client information.
//
//   - APIKey: A long-lived credential for programmatic API access.
//     Supports scoping and revocation.
//
//   - MagicLink: A one-time authentication token sent via email.
//     Time-limited with single-use enforcement.
//
//   - LinkedDID: Additional DIDs associated with a user account.
//     Supports multiple identity methods (wallet-derived, custodial).
//
//   - Connection: Tracks OAuth connections to external providers.
//     Stores token metadata (not tokens themselves — see Verification).
//
// # Domain Concepts
//
// ## Authentication Methods
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    AUTHENTICATION METHODS                       │
//	│                                                                 │
//	│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
//	│  │ Magic Link  │  │   Wallet    │  │    OAuth    │            │
//	│  │             │  │   (SIWE)    │  │  (Social)   │            │
//	│  │  Email +    │  │             │  │             │            │
//	│  │  One-time   │  │  Sign msg   │  │  GitHub/    │            │
//	│  │  Token      │  │  with key   │  │  Google     │            │
//	│  └─────────────┘  └─────────────┘  └─────────────┘            │
//	│        │                │                │                     │
//	│        └────────────────┴────────────────┘                     │
//	│                         │                                      │
//	│                         ▼                                      │
//	│                  ┌─────────────┐                               │
//	│                  │   Session   │                               │
//	│                  │   Created   │                               │
//	│                  └─────────────┘                               │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## DID Architecture
//
// Users have a primary DID and can link additional DIDs:
//
//	User Account
//	├── Primary DID (did:key:... or did:pkh:...)
//	│   └── Created at registration
//	└── Linked DIDs
//	    ├── did:pkh:eip155:1:0xABC... (Ethereum wallet)
//	    ├── did:pkh:eip155:137:0xDEF... (Polygon wallet)
//	    └── did:key:z6Mk... (additional custodial)
//
// DID derivation strategy:
//   - Email registration → did:key (custodial, Proof holds keys)
//   - Wallet registration → did:pkh (self-sovereign, user holds keys)
//
// ## Session Management
//
//	Session {
//	    ID              SessionID
//	    UserID          UserID
//	    Token           string          // JWT or opaque token
//	    ExpiresAt       time.Time
//	    CreatedAt       time.Time
//	    LastActiveAt    time.Time
//	    IPAddress       string
//	    UserAgent       string
//	    Revoked         bool
//	}
//
// Sessions support:
//   - Sliding expiration (extend on activity)
//   - Explicit revocation (logout, security)
//   - Device tracking (IP, user agent)
//   - Concurrent session limits (optional)
//
// # Relationships to Other Contexts
//
//   - Wallet: Identity delegates wallet-based authentication to Wallet
//     context. Wallet returns verified address, Identity creates session.
//
//   - Verification: Identity stores connection metadata. Verification
//     handles OAuth tokens and data fetching.
//
//   - All Contexts: Every context references Identity for user context.
//     UserID is the universal identifier across the platform.
//
// # Architecture
//
//	internal/identity/
//	├── domain/
//	│   ├── user.go            // User aggregate root
//	│   ├── session.go         // Session entity
//	│   ├── apikey.go          // APIKey entity
//	│   ├── magiclink.go       // MagicLink entity
//	│   ├── linked_did.go      // LinkedDID entity
//	│   ├── connection.go      // Connection entity (OAuth metadata)
//	│   ├── values.go          // Value objects (Email, UserID, etc.)
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // UserRegistered, SessionCreated, etc.
//	│   ├── events_did.go      // DID-specific events
//	│   ├── repository.go      // Repository interface
//	│   ├── services.go        // Domain services
//	│   └── did_service.go     // DID generation/resolution service
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // RegisterUser, CreateSession, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetUser, GetSession, etc.
//	│       ├── handlers.go    // Query handlers
//	│       ├── repository.go  // Read repository interface
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── challenge/         // Auth challenge generation
//	│   ├── did/               // DID generation service impl
//	│   ├── email/             // Email sending service
//	│   ├── magiclink/         // Magic link service impl
//	│   ├── oauth/             // OAuth provider configuration
//	│   ├── signature/         // Signature verification
//	│   ├── token/             // JWT/token service
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── user.go
//	│           ├── session.go
//	│           ├── apikey.go
//	│           ├── connection.go
//	│           ├── mapper.go
//	│           ├── queries.go
//	│           ├── reader.go
//	│           ├── token.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── auth.go        // Auth endpoints
//	│           ├── user.go        // User endpoints
//	│           ├── session.go     // Session endpoints
//	│           ├── apikey.go      // API key endpoints
//	│           ├── middleware.go  // Auth middleware
//	│           ├── request.go
//	│           ├── response.go
//	│           ├── errors.go
//	│           └── routes.go
//	└── provider.go            // Dependency injection
//
// # Events
//
//   - UserRegistered: New user account created
//   - UserUpdated: Profile information changed
//   - UserDeleted: Account deleted
//   - SessionCreated: New session started
//   - SessionRevoked: Session explicitly revoked
//   - SessionExpired: Session reached expiration
//   - APIKeyCreated: New API key generated
//   - APIKeyRevoked: API key revoked
//   - MagicLinkSent: Magic link email sent
//   - MagicLinkUsed: Magic link consumed
//   - DIDLinked: Additional DID linked to account
//   - DIDUnlinked: DID removed from account
//   - ConnectionAdded: OAuth provider connected
//   - ConnectionRemoved: OAuth provider disconnected
//
// # Security Considerations
//
//   - Passwords are never stored (passwordless auth only)
//   - Magic links are single-use and time-limited
//   - Sessions support revocation and concurrent limits
//   - API keys are hashed, only shown once at creation
//   - All auth events logged for audit trail
package identity
