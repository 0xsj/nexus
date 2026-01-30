// Package verification provides the bounded context for external account verification.
//
// # Purpose
//
// Verification orchestrates the process of connecting external accounts (GitHub,
// LinkedIn, etc.) and triggering credential issuance. It manages OAuth flows,
// coordinates data fetching, and requests credential creation. In the current
// implementation, Verification contains provider-specific logic that will be
// extracted to the Integration context in a future refactor.
//
// # Core Responsibilities
//
//   - Initiate OAuth flows with external providers
//   - Manage OAuth state for CSRF protection
//   - Handle OAuth callbacks and token exchange
//   - Store provider tokens for data access
//   - Fetch data from provider APIs (to be extracted to Integration)
//   - Trigger credential issuance based on fetched data
//   - Track verification status and history
//
// # Key Entities
//
//   - Verification: The aggregate root representing a verification attempt.
//     Tracks the full lifecycle from initiation to credential issuance.
//
//   - OAuthState: CSRF protection token linking OAuth callback to
//     verification attempt. Time-limited and single-use.
//
//   - ProviderToken: OAuth access and refresh tokens for a provider.
//     Used to fetch data from provider APIs.
//
// # Domain Concepts
//
// ## Verification Lifecycle
//
//	┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
//	│  Start   │────►│  OAuth   │────►│  Fetch   │────►│  Issue   │
//	│          │     │ Callback │     │   Data   │     │Credential│
//	└──────────┘     └──────────┘     └──────────┘     └──────────┘
//	     │                │                │                │
//	     │                ▼                ▼                ▼
//	     │          ┌──────────┐     ┌──────────┐     ┌──────────┐
//	     └─────────►│  Failed  │◄────│  Failed  │◄────│  Failed  │
//	                └──────────┘     └──────────┘     └──────────┘
//
// Status progression:
//
//	Pending         → User initiated verification
//	OAuthStarted    → Redirected to provider
//	OAuthCompleted  → Callback received, tokens obtained
//	DataFetched     → Provider data retrieved
//	CredentialIssued→ VC created and stored
//	Failed          → Error at any stage
//
// ## OAuth Flow
//
//	┌────────┐     ┌────────┐     ┌────────┐     ┌────────┐
//	│  User  │     │ Proof  │     │Provider│     │ Proof  │
//	│ clicks │────►│  save  │────►│  auth  │────►│callback│
//	│"verify"│     │ state  │     │ screen │     │handler │
//	└────────┘     └────────┘     └────────┘     └────────┘
//	                                                  │
//	                   ┌──────────────────────────────┘
//	                   ▼
//	              ┌────────┐     ┌────────┐     ┌────────┐
//	              │exchange│────►│ fetch  │────►│ issue  │
//	              │ token  │     │  data  │     │  VC    │
//	              └────────┘     └────────┘     └────────┘
//
// ## Supported Providers (Current)
//
//	| Provider   | OAuth Scopes                  | Credential Type        |
//	|------------|-------------------------------|------------------------|
//	| GitHub     | read:user, repo               | GitHubContributor      |
//	| LinkedIn   | r_liteprofile, r_emailaddress | ProfessionalExperience |
//
// ## Verification Aggregate
//
//	Verification {
//	    ID              VerificationID
//	    UserID          identity.UserID
//	    Provider        ProviderType
//	    Status          VerificationStatus
//	    OAuthState      string              // CSRF token
//	    CredentialID    *credential.CredentialID
//	    Error           *VerificationError
//	    StartedAt       time.Time
//	    CompletedAt     *time.Time
//	}
//
// ## Provider Tokens
//
//	ProviderToken {
//	    ID              TokenID
//	    UserID          identity.UserID
//	    Provider        ProviderType
//	    AccessToken     string              // Encrypted at rest
//	    RefreshToken    *string             // Encrypted at rest
//	    ExpiresAt       *time.Time
//	    Scopes          []string
//	    CreatedAt       time.Time
//	    UpdatedAt       time.Time
//	}
//
// Tokens are stored encrypted and used for:
//   - Initial data fetch during verification
//   - Re-verification (refresh credentials with updated data)
//   - Continuous verification (future: periodic refresh)
//
// # Current Architecture (Pre-Refactor)
//
// The current implementation includes provider-specific logic that will
// be extracted to the Integration context:
//
//	internal/verification/
//	├── domain/
//	│   ├── verification.go    // Verification aggregate
//	│   ├── values.go          // ProviderType, Status, etc.
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // VerificationStarted, Completed, etc.
//	│   ├── repository.go      // Repository interface
//	│   └── services.go        // Domain services
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // StartVerification, CompleteCallback, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetVerification, ListUserVerifications
//	│       └── handlers.go    // Query handlers
//	├── infrastructure/
//	│   ├── credential/
//	│   │   ├── adapter.go     // Credential context adapter
//	│   │   └── issuer.go      // Credential issuance logic
//	│   ├── oauth/
//	│   │   ├── config.go      // OAuth client configuration
//	│   │   ├── provider.go    // OAuth URL generation, token exchange
//	│   │   ├── github.go      // GitHub-specific API calls (→ Integration)
//	│   │   └── linkedin.go    // LinkedIn-specific API calls (→ Integration)
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── verification_repository.go
//	│           ├── oauth_state_repository.go
//	│           ├── provider_token_repository.go
//	│           ├── mapper.go
//	│           ├── queries.go
//	│           ├── reader.go
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
// # Planned Refactor
//
// After extracting Integration context, Verification will:
//
//  1. Keep: OAuth flow orchestration (state, redirects, token exchange)
//  2. Keep: Verification aggregate and status management
//  3. Move: Provider API calls → Integration context
//  4. Move: Data normalization → Integration context
//  5. Add: Port to Integration (DataFetcher interface)
//  6. Add: Port to Credential (CredentialIssuer interface)
//
// Post-refactor structure:
//
//	Verification                     Integration
//	┌─────────────────┐             ┌─────────────────┐
//	│ OAuth Flow      │             │ Provider Adapters│
//	│ State Mgmt      │────────────►│ GitHub, LinkedIn │
//	│ Token Exchange  │  DataFetcher│ Data Fetching    │
//	└────────┬────────┘             │ Normalization    │
//	         │                      └─────────────────┘
//	         │ CredentialIssuer
//	         ▼
//	┌─────────────────┐
//	│   Credential    │
//	│   VC Issuance   │
//	└─────────────────┘
//
// # Relationships to Other Contexts
//
//   - Identity: Provides user context. Verification is always for a user.
//
//   - Credential: Verification triggers credential issuance after data
//     is fetched. Currently direct call, will become port.
//
//   - Integration (future): Will handle provider-specific API calls
//     and data normalization.
//
//   - Schema (future): Will provide schema definitions for mapping
//     provider data to credential claims.
//
//   - Ledger: Verification events projected for audit trail.
//
//   - Notification: User notified on verification completion/failure.
//
// # Events
//
//   - VerificationStarted: User initiated verification
//   - OAuthStateCreated: OAuth state token generated
//   - OAuthCallbackReceived: Provider redirected back
//   - TokensReceived: OAuth tokens obtained
//   - DataFetchStarted: Provider API call initiated
//   - DataFetchCompleted: Provider data retrieved
//   - DataFetchFailed: Provider API call failed
//   - CredentialIssueRequested: Credential issuance triggered
//   - VerificationCompleted: Full flow successful
//   - VerificationFailed: Flow failed at some stage
//
// # Security Considerations
//
//   - OAuth state tokens are cryptographically random
//   - State tokens are time-limited (15 minutes)
//   - State tokens are single-use
//   - Provider tokens are encrypted at rest
//   - Tokens are scoped to minimum required permissions
//   - Token refresh handled automatically when expired
//
// # Error Handling
//
// Verification errors include:
//
//   - OAuthError: Provider denied access or returned error
//   - StateInvalidError: CSRF token mismatch or expired
//   - TokenExchangeError: Failed to exchange code for tokens
//   - DataFetchError: Provider API call failed
//   - CredentialIssueError: Credential context rejected issuance
//   - ProviderUnavailableError: Provider API unreachable
//
// Errors are captured in the Verification aggregate with details
// for debugging and user feedback.
//
// # Re-verification
//
// Users can re-verify to update credentials:
//
//  1. User initiates re-verification for existing provider
//  2. Full OAuth flow repeated (user may need to re-authorize)
//  3. Fresh data fetched from provider
//  4. New credential issued (or existing updated)
//  5. Old credential optionally superseded
//
// Re-verification is useful when:
//   - User's data has changed (new commits, new job)
//   - Credential approaching expiration
//   - User wants to refresh stale data
//
// # Future Considerations
//
//   - Scheduled re-verification (automatic refresh)
//   - Webhook-triggered verification (provider pushes updates)
//   - Partial verification (verify subset of claims)
//   - Batch verification (multiple providers in one flow)
//   - Verification delegation (org admin verifies members)
package verification
