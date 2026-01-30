// Package trust provides the bounded context for peer endorsements and reputation.
//
// # Purpose
//
// Trust adds a social layer on top of cryptographic verification. While credentials
// prove facts ("Alice has 1337 GitHub commits"), vouches express human judgment
// ("I've worked with Alice and she's an excellent engineer"). Trust aggregates
// these signals into reputation scores that help verifiers assess credibility
// beyond raw credential data.
//
// # Core Responsibilities
//
//   - Enable users to vouch for specific claims on other users' profiles
//   - Manage vouch lifecycle (request, give, accept, revoke)
//   - Calculate reputation scores from credentials and vouches
//   - Build trust graphs (who vouches for whom)
//   - Provide trust signals to verifiers
//   - Prevent gaming and abuse (rate limiting, verification requirements)
//
// # What Trust Does NOT Do
//
//   - Issue credentials (Credential context)
//   - Verify external data (Verification/Integration contexts)
//   - Manage user identity (Identity context)
//   - Display trust on profiles (Profile context consumes Trust data)
//
// # Key Entities
//
//   - Vouch: The aggregate root representing an endorsement from one user
//     to another, optionally tied to a specific credential or claim.
//
//   - VouchRequest: A request from one user asking another for a vouch.
//     Can specify what claims they'd like endorsed.
//
//   - Reputation: A calculated score for a user based on credentials,
//     vouches received, and voucher credibility.
//
//   - TrustGraph: The network of vouch relationships between users.
//     Used for transitive trust calculations.
//
//   - TrustAnchor: A highly trusted entity (organization, institution)
//     whose vouches carry additional weight.
//
// # Domain Concepts
//
// ## Vouch Structure
//
//	Vouch {
//	    ID              VouchID
//	    VoucherID       identity.UserID     // Who is vouching
//	    SubjectID       identity.UserID     // Who is being vouched for
//	    CredentialID    *credential.CredentialID  // Optional: specific credential
//	    ClaimKey        *string             // Optional: specific claim
//	    Statement       string              // "Alice is an excellent engineer"
//	    Relationship    RelationshipType    // Colleague, Manager, Client, etc.
//	    Context         string              // "Worked together at TechCorp"
//	    Strength        VouchStrength       // Endorsement, StrongEndorsement
//	    Status          VouchStatus         // Pending, Active, Revoked
//	    CreatedAt       time.Time
//	    ExpiresAt       *time.Time          // Optional expiration
//	}
//
// ## Vouch Types
//
//	General Vouch:
//	"I vouch for Alice as a professional."
//	- No specific credential or claim
//	- Broad endorsement of the person
//
//	Credential Vouch:
//	"I can confirm Alice's GitHub contributions are legitimate."
//	- Tied to a specific credential
//	- Adds human verification to cryptographic proof
//
//	Claim Vouch:
//	"I can confirm Alice led our team of 10 engineers."
//	- Tied to a specific claim within a credential
//	- Most specific and valuable endorsement
//
// ## Relationship Types
//
//   - Colleague: Worked together at same level
//   - Manager: Voucher managed the subject
//   - DirectReport: Subject managed the voucher
//   - Client: Professional client relationship
//   - Collaborator: Worked on project together
//   - Mentor: Mentorship relationship
//   - Academic: Professor, advisor, classmate
//   - Personal: Personal acquaintance (lower weight)
//
// ## Reputation Calculation
//
// Reputation is a composite score:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    REPUTATION SCORE                             │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Credential Score (40%)                                  │   │
//	│  │  - Number of verified credentials                        │   │
//	│  │  - Credential diversity (multiple providers)             │   │
//	│  │  - Credential freshness (recently verified)              │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Vouch Score (40%)                                       │   │
//	│  │  - Number of vouches received                            │   │
//	│  │  - Voucher credibility (their reputation)                │   │
//	│  │  - Vouch specificity (claim > credential > general)      │   │
//	│  │  - Relationship strength                                 │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Network Score (20%)                                     │   │
//	│  │  - Trust anchor connections                              │   │
//	│  │  - Graph centrality (well-connected)                     │   │
//	│  │  - Reciprocal vouches                                    │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  Final Score: 0-100 with confidence interval                   │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## Trust Anchors
//
// Trust anchors are verified organizations whose vouches carry extra weight:
//
//   - Employers (verified via LinkedIn integration)
//   - Educational institutions (verified via edu credentials)
//   - Professional bodies (verified via certification)
//   - Platform-verified organizations (KYB verified)
//
// A vouch from a trust anchor is worth more than a vouch from an
// unverified individual.
//
// ## Anti-Gaming Measures
//
//   - Vouchers must have verified identity (at least one credential)
//   - Rate limiting on vouch creation
//   - Reciprocal vouch detection (A vouches B, B vouches A)
//   - Cluster detection (groups vouching only for each other)
//   - Vouch age decay (older vouches worth less)
//   - Suspicious pattern flagging
//
// # Relationships to Other Contexts
//
//   - Identity: Provides user information for voucher/subject.
//     Vouchers must have verified identity.
//
//   - Credential: Trust can reference specific credentials.
//     Credential count contributes to reputation.
//
//   - Profile: Profile displays vouches and reputation score.
//     Profile context consumes Trust data.
//
//   - Organization: Organizational vouches (employer endorsement).
//     Trust anchors may be Organizations.
//
//   - Ledger: Vouch events projected for audit trail.
//
//   - Notification: Vouch requests and receipts trigger notifications.
//
// # Ports (Interfaces to Other Contexts)
//
//	// IdentityReader retrieves user information.
//	type IdentityReader interface {
//	    GetUser(ctx context.Context, userID identity.UserID) (*identity.User, error)
//	    HasVerifiedCredential(ctx context.Context, userID identity.UserID) (bool, error)
//	}
//
//	// CredentialReader retrieves credentials for vouch context.
//	type CredentialReader interface {
//	    GetCredential(ctx context.Context, id credential.CredentialID) (*credential.Credential, error)
//	    CountUserCredentials(ctx context.Context, userID identity.UserID) (int, error)
//	}
//
//	// OrganizationReader checks trust anchor status.
//	type OrganizationReader interface {
//	    IsTrustAnchor(ctx context.Context, orgID organization.OrganizationID) (bool, error)
//	}
//
// # Example Use Cases
//
// ## Requesting a Vouch
//
//	cmd := command.RequestVouch{
//	    RequesterID:  aliceUserID,
//	    TargetID:     bobUserID,
//	    CredentialID: &linkedinCredID,  // Optional: specific credential
//	    Message:      "Hi Bob, could you vouch for my work at TechCorp?",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.RequestID = "req_abc123"
//	// Notification sent to Bob
//
// ## Giving a Vouch
//
//	cmd := command.GiveVouch{
//	    VoucherID:    bobUserID,
//	    SubjectID:    aliceUserID,
//	    CredentialID: &linkedinCredID,
//	    ClaimKey:     stringPtr("role"),
//	    Statement:    "Alice was an exceptional tech lead on our team.",
//	    Relationship: Colleague,
//	    Context:      "Worked together at TechCorp 2020-2023",
//	    Strength:     StrongEndorsement,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.VouchID = "vouch_xyz789"
//	// Alice's reputation recalculated
//
// ## Querying Reputation
//
//	query := query.GetReputation{
//	    UserID: aliceUserID,
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.Score = 78
//	// result.Confidence = 0.85
//	// result.Breakdown = { credentials: 32, vouches: 30, network: 16 }
//	// result.VouchCount = 12
//	// result.TrustAnchorVouches = 2
//
// ## Querying User's Vouches
//
//	query := query.GetVouchesForUser{
//	    UserID: aliceUserID,
//	    Status: Active,
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// Returns all active vouches for Alice with voucher details
//
// ## Revoking a Vouch
//
//	cmd := command.RevokeVouch{
//	    VouchID:   vouchID,
//	    VoucherID: bobUserID,  // Only voucher can revoke
//	    Reason:    "Relationship ended",
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// Vouch marked as revoked, subject's reputation recalculated
//
// ## Checking Trust Path
//
//	query := query.GetTrustPath{
//	    FromUserID: verifierUserID,
//	    ToUserID:   aliceUserID,
//	    MaxDepth:   3,
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.Paths = [
//	//   [verifier] -> [bob] -> [alice]  (2 hops)
//	//   [verifier] -> [carol] -> [dave] -> [alice]  (3 hops)
//	// ]
//	// result.ShortestPath = 2
//
// # Architecture Notes
//
// Trust follows the standard bounded context structure:
//
//	internal/trust/
//	├── domain/
//	│   ├── vouch.go           // Vouch aggregate root
//	│   ├── request.go         // VouchRequest entity
//	│   ├── reputation.go      // Reputation value object
//	│   ├── graph.go           // TrustGraph, TrustPath
//	│   ├── anchor.go          // TrustAnchor entity
//	│   ├── relationship.go    // RelationshipType enum
//	│   ├── strength.go        // VouchStrength enum
//	│   ├── status.go          // VouchStatus enum
//	│   ├── scoring.go         // Reputation calculation logic
//	│   ├── ports.go           // Reader interfaces to other contexts
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // VouchGiven, VouchRevoked, ReputationChanged, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // RequestVouch, GiveVouch, RevokeVouch, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   ├── query/
//	│   │   ├── queries.go     // GetVouches, GetReputation, GetTrustPath, etc.
//	│   │   ├── handlers.go    // Query handlers
//	│   │   └── views.go       // Read models
//	│   └── scoring/
//	│       └── calculator.go  // Reputation score calculation service
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── identity.go    // IdentityReader adapter
//	│   │   ├── credential.go  // CredentialReader adapter
//	│   │   └── organization.go // OrganizationReader adapter
//	│   ├── graph/
//	│   │   └── service.go     // Trust graph queries (potentially graph DB)
//	│   ├── antispam/
//	│   │   └── detector.go    // Gaming/abuse detection
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── request_repository.go
//	│           ├── graph_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           ├── requests.go
//	│           └── responses.go
//	└── provider.go            // Dependency injection setup
//
// # Events
//
//   - VouchRequested: User requested a vouch from another user
//   - VouchGiven: User gave a vouch
//   - VouchAccepted: Subject accepted a vouch (if approval required)
//   - VouchRevoked: Voucher revoked their vouch
//   - VouchExpired: Vouch reached expiration date
//   - ReputationChanged: User's reputation score changed
//   - TrustAnchorAdded: Organization became a trust anchor
//   - SuspiciousActivityDetected: Gaming pattern detected
//
// # Future Considerations
//
//   - Verifiable vouch credentials (vouches as VCs)
//   - Cross-platform reputation import (bring reputation from other platforms)
//   - Reputation staking (stake tokens on vouches)
//   - Decentralized reputation (on-chain reputation registry)
//   - Industry-specific trust networks
//   - Anonymous vouching (ZK proof of relationship without revealing identity)
//   - Vouch decay curves (configurable time-based decay)
//   - Machine learning for fraud detection
//   - Reputation APIs for third-party consumption
package trust
