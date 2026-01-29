// Package verification provides the bounded context for verification flow orchestration.
//
// # Purpose
//
// Verification orchestrates the process of verifying a user's external accounts
// and triggering credential issuance. It owns the OAuth state machine, coordinates
// with Integration for data fetching, and requests credential issuance from the
// Credential context. Verification is a thin orchestration layer, not a data
// processing layer.
//
// # Core Responsibilities
//
//   - Manage OAuth flow lifecycle (initiate, callback, token exchange)
//   - Maintain verification state and status progression
//   - Store and validate OAuth state parameters (CSRF protection)
//   - Coordinate with Integration to fetch normalized provider data
//   - Coordinate with Credential to issue verifiable credentials
//   - Track verification history per user and provider
//   - Handle verification failures and retry logic
//
// # What Verification Does NOT Do
//
//   - Provider-specific API calls (Integration context)
//   - Data normalization or mapping (Integration context)
//   - Credential signing or storage (Credential context)
//   - Schema validation (Schema context)
//
// # Key Entities
//
//   - Verification: The aggregate root representing a single verification attempt.
//     Tracks status progression from initiation through credential issuance.
//
//   - OAuthState: Value object for CSRF protection during OAuth flow. Contains
//     state token, expiration, and associated verification ID.
//
//   - ProviderToken: OAuth tokens (access, refresh) received after successful
//     authorization. Passed to Integration for data fetching.
//
//   - VerificationStatus: Enum representing the verification state machine
//     (Pending, OAuthStarted, TokenReceived, DataFetched, CredentialIssued, Failed).
//
// # Domain Concepts
//
// ## Verification State Machine
//
//	┌─────────┐
//	│ Pending │
//	└────┬────┘
//	     │ InitiateOAuth
//	     ▼
//	┌──────────────┐
//	│ OAuthStarted │
//	└──────┬───────┘
//	       │ OAuth callback received
//	       ▼
//	┌────────────────┐
//	│ TokenReceived  │
//	└───────┬────────┘
//	        │ Integration.Fetch()
//	        ▼
//	┌─────────────┐
//	│ DataFetched │
//	└──────┬──────┘
//	       │ Credential.Issue()
//	       ▼
//	┌──────────────────┐
//	│ CredentialIssued │  (terminal success)
//	└──────────────────┘
//
//	Any state can transition to:
//	┌────────┐
//	│ Failed │  (terminal failure)
//	└────────┘
//
// ## OAuth Flow
//
// Verification owns the OAuth dance but delegates provider specifics:
//
//  1. User requests verification for a provider (e.g., GitHub)
//  2. Verification generates OAuth state, stores it, returns auth URL
//  3. User authorizes with provider, redirected back with code
//  4. Verification validates state, exchanges code for tokens
//  5. Verification calls Integration with tokens to fetch data
//  6. Verification calls Credential with claims to issue VC
//  7. Verification marks complete, stores credential reference
//
// ## Reverification
//
// Users can reverify to refresh credentials with updated data:
//
//   - Existing verification marked as superseded
//   - New verification flow initiated
//   - Old credential optionally revoked
//   - New credential issued with fresh data
//
// # Relationships to Other Contexts
//
//   - Identity: Provides user context. Verification is always linked to a user.
//     Identity may also store provider connection metadata.
//
//   - Integration: Verification calls Integration (via port) to fetch provider
//     data after OAuth tokens are received. Integration returns normalized claims.
//
//   - Credential: Verification calls Credential (via port) to issue a VC after
//     data is fetched. Credential returns the issued credential ID.
//
//   - Schema: Verification references schema type when initiating, so Integration
//     knows what data to fetch and Credential knows what schema to use.
//
//   - Ledger: Verification events are projected to Ledger for audit trail.
//
// # Ports (Interfaces to Other Contexts)
//
//	// DataFetcher is the port to the Integration context.
//	// Verification calls this after receiving OAuth tokens.
//	type DataFetcher interface {
//		Fetch(ctx context.Context, req DataFetchRequest) (*DataFetchResult, error)
//	}
//
//	type DataFetchRequest struct {
//		UserID       identity.UserID
//		Provider     ProviderType
//		AccessToken  string
//		SchemaType   string
//	}
//
//	type DataFetchResult struct {
//		Claims    map[string]any
//		FetchedAt time.Time
//		Raw       []byte // Original response for audit
//	}
//
//	// CredentialIssuer is the port to the Credential context.
//	// Verification calls this after data is fetched and normalized.
//	type CredentialIssuer interface {
//		Issue(ctx context.Context, req CredentialIssueRequest) (*CredentialIssueResult, error)
//	}
//
//	type CredentialIssueRequest struct {
//		SubjectDID  did.DID
//		SchemaType  string
//		Claims      map[string]any
//	}
//
//	type CredentialIssueResult struct {
//		CredentialID credential.CredentialID
//		IssuedAt     time.Time
//	}
//
// # Example Use Cases
//
// ## Initiating GitHub Verification
//
//	cmd := command.InitiateVerification{
//		UserID:     userID,
//		Provider:   ProviderGitHub,
//		SchemaType: "GitHubContributor",
//		RedirectURL: "https://app.proof.com/callback",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.AuthorizationURL = "https://github.com/login/oauth/authorize?..."
//	// result.VerificationID = "ver_abc123"
//	// result.State = "random_state_token"
//
// ## Handling OAuth Callback
//
//	cmd := command.CompleteOAuthCallback{
//		State: "random_state_token",
//		Code:  "auth_code_from_github",
//	}
//
//	// Handler internally:
//	// 1. Validates state, retrieves verification
//	// 2. Exchanges code for tokens
//	// 3. Calls DataFetcher.Fetch() with tokens
//	// 4. Calls CredentialIssuer.Issue() with claims
//	// 5. Updates verification status to CredentialIssued
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.VerificationID = "ver_abc123"
//	// result.CredentialID = "cred_xyz789"
//	// result.Status = CredentialIssued
//
// ## Querying Verification History
//
//	query := query.GetUserVerifications{
//		UserID: userID,
//	}
//
//	verifications, err := handler.Handle(ctx, query)
//	// Returns all verification attempts for user with status, timestamps, credential refs
//
// # Architecture Notes
//
// Verification follows the standard bounded context structure:
//
//	internal/verification/
//	├── domain/
//	│   ├── verification.go    // Verification aggregate root
//	│   ├── status.go          // VerificationStatus enum
//	│   ├── provider.go        // ProviderType enum
//	│   ├── oauth_state.go     // OAuthState value object
//	│   ├── token.go           // ProviderToken value object
//	│   ├── ports.go           // DataFetcher, CredentialIssuer interfaces
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // VerificationStarted, Completed, Failed, etc.
//	│   ├── repository.go      // Repository interface
//	│   └── services.go        // Domain services
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // InitiateVerification, CompleteCallback, etc.
//	│   │   └── handlers.go    // Command handlers (orchestration logic)
//	│   └── query/
//	│       ├── queries.go     // GetVerification, ListUserVerifications
//	│       ├── handlers.go    // Query handlers
//	│       ├── repository.go  // Read repository interface
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── oauth/
//	│   │   ├── config.go      // OAuth client configuration per provider
//	│   │   └── client.go      // OAuth URL generation, token exchange
//	│   ├── adapters/
//	│   │   ├── integration.go // DataFetcher adapter (calls Integration context)
//	│   │   └── credential.go  // CredentialIssuer adapter (calls Credential context)
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── oauth_state_repository.go
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
// # Migration Notes
//
// The current verification implementation includes provider-specific code that
// will be extracted to the Integration context:
//
//	| File                                    | Action                              |
//	|-----------------------------------------|-------------------------------------|
//	| infrastructure/oauth/github.go          | Move to integration/adapters/github |
//	| infrastructure/oauth/linkedin.go        | Move to integration/adapters/linkedin |
//	| infrastructure/oauth/config.go          | Keep (OAuth client config)          |
//	| infrastructure/oauth/provider.go        | Keep (OAuth URL/token exchange)     |
//	| infrastructure/credential/issuer.go     | Replace with port adapter           |
//	| infrastructure/credential/adapter.go    | Replace with port adapter           |
//
// After migration, Verification will depend on ports (interfaces) rather than
// concrete implementations, enabling clean separation and testability.
//
// # Events
//
// Verification emits domain events for each state transition:
//
//   - VerificationInitiated: User started verification for a provider
//   - OAuthCompleted: OAuth tokens received successfully
//   - DataFetchCompleted: Integration returned normalized claims
//   - CredentialIssued: Credential context issued the VC
//   - VerificationFailed: Any step failed (includes reason)
//   - VerificationSuperseded: Replaced by a newer verification
//
// These events are consumed by Ledger for audit trail and potentially
// by Notification for user alerts.
//
// # Future Considerations
//
//   - Scheduled reverification (refresh credentials periodically)
//   - Batch verification (verify multiple providers in one flow)
//   - Verification delegation (org admin verifies on behalf of member)
//   - Webhook-triggered verification (provider pushes updates)
//   - Partial verification (some claims verified, others pending)
package verification
