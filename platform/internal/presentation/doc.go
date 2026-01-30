// Package presentation provides the bounded context for verifiable presentations and selective sharing.
//
// # Purpose
//
// Presentation enables users to share their credentials with third parties in a
// controlled, privacy-preserving manner. It builds on the low-level primitives
// in pkg/presentation to provide share links, QR codes, selective disclosure
// policies, and access controls. Users decide exactly what to reveal, to whom,
// and for how long.
//
// # Core Responsibilities
//
//   - Create Verifiable Presentations from one or more credentials
//   - Define disclosure policies (which claims to include/exclude)
//   - Generate shareable links with configurable access controls
//   - Generate QR codes for in-person verification scenarios
//   - Track presentation access (who viewed, when)
//   - Support time-limited and single-use presentations
//   - Verify presentations on behalf of external verifiers
//
// # What Presentation Does NOT Do
//
//   - Store credentials (Credential context)
//   - Issue credentials (Credential context)
//   - Manage user identity (Identity context)
//   - VP encoding/signing primitives (pkg/presentation)
//
// # Key Entities
//
//   - Presentation: The aggregate root representing a Verifiable Presentation.
//     Contains selected credentials, disclosure policy, and metadata.
//
//   - DisclosurePolicy: Defines what claims are included or excluded from
//     each credential in the presentation. Supports claim-level granularity.
//
//   - ShareLink: A URL that provides access to a presentation. Configurable
//     with expiration, view limits, PIN protection, and audience restrictions.
//
//   - AccessGrant: Records each access to a presentation, including verifier
//     identity, timestamp, and what was disclosed.
//
//   - VerificationRequest: An inbound request from an external verifier
//     specifying what credentials/claims they need to see.
//
// # Domain Concepts
//
// ## Verifiable Presentations
//
// A Verifiable Presentation (VP) bundles one or more Verifiable Credentials
// with a proof that the holder controls them:
//
//	┌─────────────────────────────────────────────────────┐
//	│              Verifiable Presentation                │
//	│                                                     │
//	│  ┌─────────────────┐  ┌─────────────────┐          │
//	│  │  Credential #1  │  │  Credential #2  │   ...    │
//	│  │  (GitHub)       │  │  (LinkedIn)     │          │
//	│  │                 │  │                 │          │
//	│  │  [selected      │  │  [selected      │          │
//	│  │   claims only]  │  │   claims only]  │          │
//	│  └─────────────────┘  └─────────────────┘          │
//	│                                                     │
//	│  Holder: did:key:z6Mk...                           │
//	│  Proof:  (signature proving holder control)        │
//	│  Created: 2024-01-15T10:30:00Z                     │
//	└─────────────────────────────────────────────────────┘
//
// ## Selective Disclosure
//
// Users control exactly what they reveal:
//
//	Full Credential (GitHubContributor):
//	{
//	    "username": "alice",
//	    "commits": 1337,
//	    "repositories": 42,
//	    "stars": 500,
//	    "languages": ["Go", "TypeScript", "Rust"],
//	    "contributionYears": 5
//	}
//
//	Disclosed via Policy (prove experience without revealing identity):
//	{
//	    "commits": 1337,           // ✓ included
//	    "contributionYears": 5     // ✓ included
//	    // username, repositories, stars, languages excluded
//	}
//
// ## Disclosure Policy Types
//
//   - Allowlist: Only specified claims are included
//   - Blocklist: All claims except specified ones are included
//   - Threshold: Prove claim meets threshold without revealing value
//     (e.g., "commits >= 1000" without revealing exact count)
//   - Range: Prove claim falls within range
//     (e.g., "contributionYears between 3-10")
//
// Note: Threshold and range disclosures require BBS+ signatures (future phase).
// Current implementation supports allowlist/blocklist with JWT-VC credentials.
//
// ## Share Links
//
// Share links provide controlled access to presentations:
//
//	┌─────────────────────────────────────────────────────┐
//	│                    ShareLink                        │
//	│                                                     │
//	│  URL:        https://proof.dev/p/abc123xyz          │
//	│  Presentation: pres_789                             │
//	│                                                     │
//	│  Access Controls:                                   │
//	│  ├── Expires:      2024-02-01T00:00:00Z            │
//	│  ├── Max Views:    10                               │
//	│  ├── PIN:          ****                             │
//	│  ├── Audience:     did:web:acme.com (optional)     │
//	│  └── Single Use:   false                            │
//	│                                                     │
//	│  Stats:                                             │
//	│  ├── Views:        3                                │
//	│  └── Last Access:  2024-01-20T14:30:00Z            │
//	└─────────────────────────────────────────────────────┘
//
// ## QR Codes
//
// For in-person verification (conferences, interviews):
//
//   - QR encodes the share link URL
//   - Optionally includes embedded presentation (offline verification)
//   - Supports dynamic QR (link-based) and static QR (self-contained)
//
// ## Verification Requests
//
// External verifiers can request specific credentials:
//
//	Verifier Request:
//	"I need to verify:
//	 - GitHub: commits >= 500, any language
//	 - LinkedIn: 2+ years experience"
//
//	User Response:
//	"Here's a presentation with:
//	 - GitHub: commits = 1337 ✓
//	 - LinkedIn: tenure = 3 years ✓"
//
// # Relationships to Other Contexts
//
//   - Credential: Presentation reads credentials to include in VPs.
//     Does not modify credentials.
//
//   - Identity: Provides holder DID for signing presentations.
//     Access grants may reference verifier identity.
//
//   - Schema: References schema definitions for claim metadata
//     (which claims exist, display names, disclosure hints).
//
//   - Profile: Profile may embed or link to presentations.
//     Public profile shows badges derived from presentations.
//
//   - Ledger: Presentation events (created, accessed, revoked) are
//     projected to Ledger for audit trail.
//
//   - Trust: Presentations may include vouch credentials.
//
// # Ports (Interfaces to Other Contexts)
//
//	// CredentialReader reads credentials for inclusion in presentations.
//	type CredentialReader interface {
//	    GetCredential(ctx context.Context, id credential.CredentialID) (*credential.Credential, error)
//	    GetUserCredentials(ctx context.Context, userID identity.UserID) ([]*credential.Credential, error)
//	}
//
//	// IdentityReader retrieves holder DID for signing.
//	type IdentityReader interface {
//	    GetUserDID(ctx context.Context, userID identity.UserID) (did.DID, error)
//	}
//
//	// SchemaReader retrieves claim metadata for disclosure hints.
//	type SchemaReader interface {
//	    GetSchema(ctx context.Context, schemaType string) (*schema.Schema, error)
//	}
//
// # Example Use Cases
//
// ## Creating a Presentation with Selective Disclosure
//
//	cmd := command.CreatePresentation{
//	    UserID: userID,
//	    Credentials: []CredentialSelection{
//	        {
//	            CredentialID: githubCredID,
//	            Policy: DisclosurePolicy{
//	                Type: AllowList,
//	                Claims: []string{"commits", "contributionYears"},
//	            },
//	        },
//	    },
//	    Purpose: "Job application at Acme Corp",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.PresentationID = "pres_abc123"
//
// ## Generating a Share Link
//
//	cmd := command.CreateShareLink{
//	    PresentationID: presentationID,
//	    ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 1 week
//	    MaxViews:       10,
//	    PIN:            "1234", // optional
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.URL = "https://proof.dev/p/xyz789"
//	// result.ShareLinkID = "link_xyz789"
//
// ## Generating a QR Code
//
//	cmd := command.GenerateQRCode{
//	    ShareLinkID: shareLinkID,
//	    Format:      "png",
//	    Size:        256,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.Image = []byte{...}
//	// result.DataURL = "data:image/png;base64,..."
//
// ## Verifying a Presentation (External Verifier)
//
//	cmd := command.VerifyPresentation{
//	    ShareLinkID: shareLinkID,
//	    PIN:         "1234", // if required
//	    VerifierDID: verifierDID, // optional, for audience-restricted links
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.Valid = true
//	// result.Holder = "did:key:z6Mk..."
//	// result.Credentials = []{type, claims, issuer, issuedAt}
//	// result.AccessGrantID = "grant_abc123"
//
// ## Revoking a Share Link
//
//	cmd := command.RevokeShareLink{
//	    ShareLinkID: shareLinkID,
//	    Reason:      "No longer needed",
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// Link is now invalid, subsequent access attempts fail
//
// # Architecture Notes
//
// Presentation follows the standard bounded context structure:
//
//	internal/presentation/
//	├── domain/
//	│   ├── presentation.go    // Presentation aggregate root
//	│   ├── policy.go          // DisclosurePolicy value object
//	│   ├── share_link.go      // ShareLink entity
//	│   ├── access_grant.go    // AccessGrant entity
//	│   ├── qr.go              // QRCode value object
//	│   ├── request.go         // VerificationRequest value object
//	│   ├── ports.go           // CredentialReader, IdentityReader, SchemaReader
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // PresentationCreated, LinkAccessed, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // CreatePresentation, CreateShareLink, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetPresentation, ListShareLinks, etc.
//	│       ├── handlers.go    // Query handlers
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── credential.go  // CredentialReader adapter
//	│   │   ├── identity.go    // IdentityReader adapter
//	│   │   └── schema.go      // SchemaReader adapter
//	│   ├── qr/
//	│   │   └── generator.go   // QR code generation
//	│   ├── signing/
//	│   │   └── signer.go      // VP signing (uses pkg/presentation)
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── share_link_repository.go
//	│           ├── access_grant_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           ├── requests.go
//	│           ├── responses.go
//	│           └── public.go      // Public verification endpoint
//	└── provider.go                // Dependency injection setup
//
// # Relationship to pkg/presentation
//
//	| Concern                  | pkg/presentation           | internal/presentation      |
//	|--------------------------|----------------------------|----------------------------|
//	| VP data model            | ✓                          |                            |
//	| JWT-VP encoding/decoding | ✓                          |                            |
//	| VP signing/verification  | ✓                          |                            |
//	| Disclosure policies      |                            | ✓                          |
//	| Share links              |                            | ✓                          |
//	| Access control           |                            | ✓                          |
//	| QR generation            |                            | ✓                          |
//	| Persistence              |                            | ✓                          |
//
// # Events
//
//   - PresentationCreated: User created a new presentation
//   - PresentationRevoked: User revoked a presentation
//   - ShareLinkCreated: Share link generated for a presentation
//   - ShareLinkAccessed: Someone accessed a share link
//   - ShareLinkRevoked: Share link manually revoked
//   - ShareLinkExpired: Share link reached expiration or view limit
//   - VerificationCompleted: External verifier successfully verified
//
// # Future Considerations
//
//   - BBS+ selective disclosure (prove predicates without revealing values)
//   - Presentation templates (reusable disclosure configurations)
//   - Batch presentations (share with multiple verifiers at once)
//   - Encrypted presentations (only intended verifier can decrypt)
//   - Presentation requests protocol (DIDComm-based exchange)
//   - Offline verification (self-contained QR with embedded VP)
//   - Presentation analytics (aggregated stats, anonymized)
package presentation
