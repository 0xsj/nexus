// Package integration provides the bounded context for external provider adapters.
//
// # Purpose
//
// Integration is responsible for connecting to external data sources (GitHub,
// LinkedIn, Coursera, etc.), fetching user data, and normalizing it into a
// consistent format suitable for credential issuance. This context isolates
// provider-specific logic from the rest of the system, making it easy to add
// new providers without modifying the Verification or Credential contexts.
//
// # Core Responsibilities
//
//   - Define provider adapters with consistent interfaces
//   - Handle provider-specific API authentication and rate limiting
//   - Fetch raw data from external APIs
//   - Normalize provider data into schema-aligned claim sets
//   - Manage provider health and availability
//   - Support multiple integration methods (OAuth, API key, file upload, manual)
//
// # Key Entities
//
//   - Provider: Represents an external data source (GitHub, LinkedIn, etc.)
//     with its capabilities, supported credential types, and configuration.
//
//   - Adapter: The implementation that knows how to communicate with a specific
//     provider's API, handle authentication, and fetch data.
//
//   - FetchResult: The normalized output from a provider fetch, containing
//     raw data, mapped claims, and metadata about the fetch operation.
//
//   - ProviderCredentials: OAuth tokens, API keys, or other authentication
//     material needed to access a provider on behalf of a user.
//
//   - RateLimiter: Manages API rate limits per provider to avoid throttling.
//
// # Domain Concepts
//
// ## Provider Types
//
// Providers are categorized by their integration method:
//
//   - OAuth Providers: GitHub, LinkedIn, Google, Twitter
//     User authorizes access, we receive tokens, fetch via API
//
//   - API Key Providers: AWS (credential verification), Stripe
//     User provides API key, we verify against provider API
//
//   - Certificate Providers: Coursera, Udemy, edX
//     OAuth or file upload of completion certificates
//
//   - Manual Verification: Professional licenses, degrees
//     Document upload with manual or automated verification
//
// ## Data Normalization
//
// Each provider returns data in its own format. Integration normalizes this
// into claim sets that align with Schema definitions:
//
//	GitHub API Response:
//	{
//	  "public_repos": 42,
//	  "followers": 150,
//	  "contributions": { "total": 1337 }
//	}
//
//	Normalized Claims (GitHubContributor schema):
//	{
//	  "repositories": 42,
//	  "followers": 150,
//	  "commits": 1337
//	}
//
// ## Fetch Strategies
//
// Different credential types require different fetch strategies:
//
//   - Snapshot: Fetch current state (GitHub stats at this moment)
//   - Historical: Fetch time-range data (commits in last year)
//   - Verification: Confirm specific claim (does user own this repo?)
//   - Continuous: Periodic refresh for up-to-date credentials
//
// # Relationships to Other Contexts
//
//   - Schema: Integration reads schema definitions to understand what claims
//     to extract from provider data and how to map provider fields.
//
//   - Verification: Orchestrates the verification flow, calling Integration
//     to fetch data after OAuth is complete. Verification owns the OAuth
//     state machine; Integration owns the data fetching.
//
//   - Credential: Receives normalized claim sets from Integration (via
//     Verification) and issues signed credentials.
//
//   - Identity: Provides user context and linked provider connections.
//
// # Example Use Cases
//
// ## Fetching GitHub Data
//
//	adapter := github.NewAdapter(httpClient, rateLimiter)
//
//	result, err := adapter.Fetch(ctx, FetchRequest{
//		UserID:      userID,
//		Credentials: oauthTokens,
//		SchemaType:  "GitHubContributor",
//	})
//	if err != nil {
//		return err
//	}
//
//	// result.Claims contains normalized data ready for credential issuance
//	// result.Raw contains original API response for audit
//	// result.FetchedAt contains timestamp
//
// ## Adding a New Provider
//
// To add a new provider (e.g., Twitter):
//
//  1. Define supported schema types for Twitter credentials
//
//  2. Implement the Adapter interface for Twitter API
//
//  3. Create field mappings from Twitter API to schema claims
//
//  4. Register the adapter in the provider registry
//
//  5. Add OAuth configuration (if applicable)
//
//     type TwitterAdapter struct {
//     client      *http.Client
//     rateLimiter RateLimiter
//     }
//
//     func (a *TwitterAdapter) Fetch(ctx context.Context, req FetchRequest) (*FetchResult, error) {
//     // Fetch from Twitter API
//     // Normalize to schema claims
//     // Return FetchResult
//     }
//
//     func (a *TwitterAdapter) SupportedSchemas() []string {
//     return []string{"TwitterInfluencer", "TwitterDeveloper"}
//     }
//
// ## Provider Health Checks
//
//	health := registry.CheckHealth(ctx)
//	// Returns status of each provider's API availability
//	// {
//	//   "github": { "status": "healthy", "latency": "45ms" },
//	//   "linkedin": { "status": "degraded", "latency": "2s" },
//	// }
//
// # Architecture Notes
//
// Integration follows the standard bounded context structure:
//
//	internal/integration/
//	├── domain/
//	│   ├── provider.go        // Provider entity
//	│   ├── adapter.go         // Adapter interface
//	│   ├── fetch.go           // FetchRequest, FetchResult
//	│   ├── credentials.go     // ProviderCredentials value object
//	│   ├── mapping.go         // FieldMapping, ClaimMapper
//	│   ├── errors.go          // Domain errors (ProviderUnavailable, RateLimited, etc.)
//	│   ├── events.go          // DataFetched, FetchFailed, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // FetchProviderData, RefreshConnection, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetProviderStatus, ListProviders, etc.
//	│       ├── handlers.go    // Query handlers
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── github/
//	│   │   │   ├── adapter.go     // GitHub API client
//	│   │   │   ├── mapper.go      // GitHub -> schema mapping
//	│   │   │   └── types.go       // GitHub API response types
//	│   │   ├── linkedin/
//	│   │   │   ├── adapter.go
//	│   │   │   ├── mapper.go
//	│   │   │   └── types.go
//	│   │   └── registry.go        // Adapter registry
//	│   ├── ratelimit/
//	│   │   └── limiter.go         // Rate limiting implementation
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go     // Provider status endpoints
//	│           ├── routes.go
//	│           └── responses.go
//	└── provider.go                // Dependency injection setup
//
// # Supported Providers (Planned)
//
//	| Provider   | Method   | Credential Types                        | Status    |
//	|------------|----------|-----------------------------------------|-----------|
//	| GitHub     | OAuth    | GitHubContributor                       | ✅ Active |
//	| LinkedIn   | OAuth    | ProfessionalExperience                  | ✅ Active |
//	| Twitter/X  | OAuth    | TwitterInfluencer, TwitterDeveloper     | 🔲 Planned |
//	| Coursera   | OAuth    | CourseCompletion                        | 🔲 Planned |
//	| Udemy      | OAuth    | CourseCompletion                        | 🔲 Planned |
//	| AWS        | API Key  | CloudCertification                      | 🔲 Planned |
//	| Google     | OAuth    | CloudCertification                      | 🔲 Planned |
//	| Upwork     | OAuth    | FreelanceReputation                     | 🔲 Planned |
//	| Stripe     | API Key  | RevenueVerification                     | 🔲 Planned |
//
// # Migration from Verification Context
//
// Currently, provider-specific code lives in internal/verification/infrastructure/oauth/.
// This context extracts that logic:
//
//   - verification/infrastructure/oauth/github.go  -> integration/infrastructure/adapters/github/
//   - verification/infrastructure/oauth/linkedin.go -> integration/infrastructure/adapters/linkedin/
//
// Verification retains OAuth flow orchestration (state, redirects, token exchange).
// Integration takes over post-auth data fetching and normalization.
//
// # Future Considerations
//
//   - Webhook support for real-time updates from providers
//   - Batch fetching for efficiency (fetch multiple users' data)
//   - Caching layer for frequently accessed provider data
//   - Provider SDK integrations where available
//   - Custom provider definitions for Issuer-specific sources
//   - Retry policies with exponential backoff per provider
package integration
