// Package credential provides the bounded context for verifiable credential management.
//
// # Purpose
//
// Credential is the core of the Proof platform. It issues, stores, signs, and
// manages W3C Verifiable Credentials. When a user verifies an external account,
// Credential receives the normalized claims and produces a cryptographically
// signed credential that the user owns and controls. Credentials are the
// portable, verifiable proof of a user's claims.
//
// # Core Responsibilities
//
//   - Issue verifiable credentials from verified claims
//   - Sign credentials with Proof's issuer DID
//   - Store credentials for user access
//   - Manage credential lifecycle (active, revoked, expired)
//   - Verify credential signatures and status
//   - Support multiple credential formats (JWT-VC)
//   - Manage credential schemas (to be extracted to Schema context)
//
// # Key Entities
//
//   - Credential: The aggregate root representing a verifiable credential.
//     Contains subject, claims, proof, and status.
//
//   - CredentialSubject: The entity the credential is about (user's DID).
//
//   - Claims: The verified data points within the credential.
//
//   - Proof: The cryptographic signature proving authenticity.
//
//   - CredentialStatus: Lifecycle state (Active, Revoked, Expired).
//
//   - Schema/Registry: Credential type definitions (to be extracted).
//
// # Domain Concepts
//
// ## Verifiable Credential Structure
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    VERIFIABLE CREDENTIAL                        │
//	│                                                                 │
//	│  @context: ["https://www.w3.org/2018/credentials/v1"]          │
//	│  type: ["VerifiableCredential", "GitHubContributor"]           │
//	│  id: "urn:uuid:abc123..."                                       │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Issuer                                                  │   │
//	│  │  did:web:proof.dev                                       │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Credential Subject                                      │   │
//	│  │  id: did:pkh:eip155:1:0xABC...                          │   │
//	│  │                                                          │   │
//	│  │  Claims:                                                 │   │
//	│  │    username: "alice"                                     │   │
//	│  │    commits: 1337                                         │   │
//	│  │    repositories: 42                                      │   │
//	│  │    stars: 500                                            │   │
//	│  │    languages: ["Go", "TypeScript", "Rust"]              │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  issuanceDate: "2024-01-15T10:30:00Z"                          │
//	│  expirationDate: "2025-01-15T10:30:00Z" (optional)             │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Proof                                                   │   │
//	│  │  type: "JsonWebSignature2020"                           │   │
//	│  │  created: "2024-01-15T10:30:00Z"                        │   │
//	│  │  verificationMethod: "did:web:proof.dev#key-1"          │   │
//	│  │  proofPurpose: "assertionMethod"                        │   │
//	│  │  jws: "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQi..." │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## Credential Types (Current)
//
//	| Type                   | Source     | Key Claims                        |
//	|------------------------|------------|-----------------------------------|
//	| GitHubContributor      | GitHub     | username, commits, repos, stars   |
//	| ProfessionalExperience | LinkedIn   | employer, role, tenure, industry  |
//
// Schema definitions will be extracted to the Schema context.
//
// ## Credential Lifecycle
//
//	┌──────────┐
//	│  Issued  │
//	└────┬─────┘
//	     │
//	     ├─────────────────────────────┐
//	     │                             │
//	     ▼                             ▼
//	┌──────────┐                 ┌──────────┐
//	│  Active  │                 │  Revoked │
//	└────┬─────┘                 └──────────┘
//	     │
//	     │ (time passes)
//	     ▼
//	┌──────────┐
//	│ Expired  │
//	└──────────┘
//
// Status transitions:
//
//   - Issued → Active: Immediate upon issuance
//   - Active → Revoked: Issuer or user revokes
//   - Active → Expired: Expiration date reached
//   - Revoked/Expired: Terminal states
//
// ## Signing Process
//
//	┌─────────────┐     ┌─────────────┐     ┌─────────────┐
//	│   Claims    │────►│  Build VC   │────►│    Sign     │
//	│   (data)    │     │  Structure  │     │  (Ed25519)  │
//	└─────────────┘     └─────────────┘     └──────┬──────┘
//	                                               │
//	                                               ▼
//	                                        ┌─────────────┐
//	                                        │   JWT-VC    │
//	                                        │  (encoded)  │
//	                                        └─────────────┘
//
// Signing uses:
//   - Proof's issuer DID (did:web:proof.dev)
//   - Ed25519 key pair
//   - JWT encoding with VC claims
//
// ## Credential Aggregate
//
//	Credential {
//	    ID              CredentialID
//	    Type            CredentialType      // GitHubContributor, etc.
//	    IssuerDID       did.DID             // did:web:proof.dev
//	    SubjectDID      did.DID             // User's DID
//	    Claims          map[string]any      // Verified data
//	    IssuedAt        time.Time
//	    ExpiresAt       *time.Time
//	    Status          CredentialStatus
//	    RevokedAt       *time.Time
//	    RevocationReason *string
//	    JWT             string              // Encoded JWT-VC
//	    VerificationID  *verification.VerificationID
//	}
//
// ## Verification (of Credentials)
//
// Credential verification checks:
//
//  1. Signature validity (JWT signature)
//  2. Issuer DID resolution (did:web:proof.dev)
//  3. Expiration status (not expired)
//  4. Revocation status (not revoked)
//  5. Schema compliance (claims match schema)
//
// Verification can be performed:
//   - Internally (Proof platform)
//   - Externally (any party with the JWT)
//
// # Relationships to Other Contexts
//
//   - Verification: Triggers credential issuance with verified claims.
//     Passes subject DID and normalized claim data.
//
//   - Identity: Provides subject DID (user's primary DID).
//
//   - Schema (future): Provides credential type definitions and
//     claim validation rules.
//
//   - Presentation: Reads credentials for inclusion in VPs.
//     Does not modify credentials.
//
//   - Profile: Reads credentials to generate badges.
//
//   - Issuer (future): External issuers will call Credential to
//     issue credentials on their behalf.
//
//   - Ledger: Credential events projected for audit trail.
//
// # Architecture
//
//	internal/credential/
//	├── domain/
//	│   ├── credential.go      // Credential aggregate root
//	│   ├── values.go          // CredentialType, Status, Claims, etc.
//	│   ├── signing.go         // Signing interfaces
//	│   ├── verification.go    // Verification logic
//	│   ├── registry.go        // Schema registry (→ Schema context)
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // CredentialIssued, Revoked, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // IssueCredential, RevokeCredential, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetCredential, ListUserCredentials, etc.
//	│       ├── handlers.go    // Query handlers
//	│       ├── repository.go  // Read repository interface
//	│       └── view.go        // Read models
//	├── infrastructure/
//	│   ├── signing/
//	│   │   └── signer.go      // Ed25519 signing implementation
//	│   ├── verification/
//	│   │   └── verifier.go    // Credential verification implementation
//	│   └── persistence/
//	│       ├── memory/
//	│       │   └── repository.go  // In-memory repository (testing)
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── mapper.go
//	│           ├── queries.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── requests.go
//	│           ├── responses.go
//	│           ├── errors.go
//	│           └── router.go
//	└── provider.go            // Dependency injection
//
// # Events
//
//   - CredentialIssued: New credential created
//   - CredentialRevoked: Credential revoked by issuer or user
//   - CredentialExpired: Credential reached expiration date
//   - CredentialVerified: Credential signature verified (query event)
//   - CredentialAccessed: Credential retrieved by user (query event)
//
// # Issuer DID
//
// Proof acts as the issuer for platform-verified credentials:
//
//	Issuer DID: did:web:proof.dev
//
//	DID Document (https://proof.dev/.well-known/did.json):
//	{
//	    "@context": ["https://www.w3.org/ns/did/v1"],
//	    "id": "did:web:proof.dev",
//	    "verificationMethod": [{
//	        "id": "did:web:proof.dev#key-1",
//	        "type": "JsonWebKey2020",
//	        "controller": "did:web:proof.dev",
//	        "publicKeyJwk": { ... }
//	    }],
//	    "assertionMethod": ["did:web:proof.dev#key-1"]
//	}
//
// External verifiers can resolve this DID to verify credentials.
//
// # Planned Extraction
//
// The following will be extracted to the Schema context:
//
//   - registry.go → Schema context
//   - Credential type definitions → Schema context
//   - Claim validation rules → Schema context
//
// Credential will then depend on Schema via a port:
//
//	type SchemaResolver interface {
//	    GetSchema(ctx context.Context, credType string) (*Schema, error)
//	    ValidateClaims(ctx context.Context, credType string, claims map[string]any) error
//	}
//
// # Security Considerations
//
//   - Private signing key stored securely (HSM in production)
//   - JWTs are tamper-evident (signature verification)
//   - Revocation checked on every verification
//   - Credentials are user-owned (stored, not controlled)
//   - No sensitive data in credentials (only verified facts)
//
// # JWT-VC Format
//
// Credentials are encoded as JWTs for portability:
//
//	Header:
//	{
//	    "alg": "EdDSA",
//	    "typ": "JWT"
//	}
//
//	Payload:
//	{
//	    "iss": "did:web:proof.dev",
//	    "sub": "did:pkh:eip155:1:0xABC...",
//	    "iat": 1705312200,
//	    "exp": 1736848200,
//	    "vc": {
//	        "@context": [...],
//	        "type": ["VerifiableCredential", "GitHubContributor"],
//	        "credentialSubject": {
//	            "id": "did:pkh:eip155:1:0xABC...",
//	            "username": "alice",
//	            "commits": 1337,
//	            ...
//	        }
//	    }
//	}
//
//	Signature: <Ed25519 signature>
//
// # Future Considerations
//
//   - JSON-LD credentials (richer semantics)
//   - BBS+ signatures (selective disclosure)
//   - Status List 2021 (efficient revocation)
//   - Credential refresh (update without full re-verification)
//   - Batch issuance (multiple credentials at once)
//   - External issuer support (via Issuer context)
//   - On-chain anchoring (credential hash on blockchain)
package credential
