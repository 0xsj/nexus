# Integration Context

The Integration context manages external provider connections and data fetching for the Nexus platform. It implements a provider adapter pattern to support multiple OAuth providers (GitHub, LinkedIn, Coursera, Twitter, Google, AWS).

## Architecture

### Domain Layer

#### Aggregates
- **Integration** — Manages the lifecycle of a provider connection (connected, disconnected, suspended)

#### Value Objects
- **IntegrationID** — Unique identifier for integrations
- **ProviderType** — Enum of supported providers
- **IntegrationStatus** — Connection status (connected, disconnected, suspended)
- **ProviderData** — Normalized data structure returned by all adapters
- **OAuthTokens** — OAuth token container

#### Ports (Interfaces)
- **ProviderAdapter** — Interface all provider adapters must implement
- **ProviderAdapterRegistry** — Registry for managing adapters
- **TokenStorage** — Port for storing OAuth tokens securely
- **EventPublisher** — Publishes domain events to the event bus

#### Events
- `ProviderConnectedEvent` — Provider connected to user account
- `ProviderDisconnectedEvent` — Provider disconnected
- `CredentialsRefreshedEvent` — OAuth credentials refreshed
- `DataFetchedEvent` — User data fetched from provider
- `IntegrationSuspendedEvent` — Integration suspended (e.g., expired tokens)

### Application Layer

#### Commands
- **ConnectProvider** — Connect a new provider
- **DisconnectProvider** — Disconnect a provider
- **RefreshCredentials** — Refresh OAuth credentials
- **SuspendIntegration** — Suspend an integration

#### Queries
- Get integration by ID
- List integrations by user
- List integrations by provider type

### Infrastructure Layer

#### Provider Adapters

Each adapter implements the `ProviderAdapter` interface and provides:
1. OAuth flow (authorization URL generation, code exchange, token refresh)
2. API client for fetching user data
3. Data normalization into `ProviderData` format
4. Scope validation

##### Implemented Adapters

| Provider | Status | OAuth | Data Fetching | Notes |
|----------|--------|-------|---------------|-------|
| **GitHub** | ✅ Complete | ✅ | ✅ | Fetches profile, repos, orgs, stars |
| **LinkedIn** | ✅ Complete | ✅ | ✅ | Fetches profile, positions, education |
| **Coursera** | ✅ Complete | ✅ | ✅ | Fetches enrollments, certificates |
| **Twitter** | ✅ Complete | ✅ | ✅ | OAuth 2.0 with PKCE, fetches profile & metrics |
| **Google** | ✅ Complete | ✅ | ✅ | Fetches user info via OAuth 2.0 |
| **AWS** | ✅ Complete | ✅ | ✅ | Cognito OAuth, user profile |

#### Adapter Registry

The `Registry` manages all registered adapters and provides:
- Thread-safe adapter registration
- Adapter lookup by provider type
- List of supported providers

```go
registry := adapters.NewRegistry()

config := &adapters.AdapterConfig{
    GitHub: &adapters.GitHubConfig{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURI:  "https://your-app.com/oauth/callback",
    },
}

registry, err := adapters.RegisterDefaultAdapters(config)
if err != nil {
    log.Fatal(err)
}

// Get adapter
githubAdapter, err := registry.Get(domain.ProviderTypeGitHub)
```

## Usage Examples

### 1. Initiating OAuth Flow

```go
// Get GitHub adapter
adapter, err := registry.Get(domain.ProviderTypeGitHub)
if err != nil {
    return err
}

// Generate authorization URL
state := "random-csrf-token"
redirectURI := "https://your-app.com/oauth/callback"
scopes := adapter.GetDefaultScopes()

authURL, err := adapter.GetAuthorizationURL(ctx, state, redirectURI, scopes)
if err != nil {
    return err
}

// Redirect user to authURL
```

### 2. Handling OAuth Callback

```go
// Exchange code for tokens
tokens, err := adapter.ExchangeCode(ctx, code, redirectURI)
if err != nil {
    return err
}

// Fetch user data
providerData, err := adapter.FetchUserData(ctx, tokens.AccessToken)
if err != nil {
    return err
}

// Create Integration aggregate
integration, err := domain.ConnectProvider(
    domain.NewIntegrationID(),
    userID,
    providerData.ProviderType,
    providerData.ProviderUserID,
    providerData.ProviderUsername,
    tokens.Scopes,
)

// Save integration and tokens
```

### 3. Fetching Fresh Data

```go
// Load integration
integration, err := integrationRepo.FindByID(ctx, integrationID)

// Load tokens
tokens, err := tokenStorage.Get(ctx, integration.ID().String())

// Fetch fresh data
providerData, err := adapter.FetchUserData(ctx, tokens.AccessToken)

// Record the fetch
integration.RecordFetch()
```

## Provider Data Models

Each adapter returns normalized data in the `ProviderData` structure:

### GitHub Data
```go
type GitHubData struct {
    Login           string
    Name            string
    Bio             string
    PublicRepos     int
    Followers       int
    Repositories    []GitHubRepository
    Organizations   []GitHubOrganization
    // ... more fields
}
```

### LinkedIn Data
```go
type LinkedInData struct {
    FirstName       string
    LastName        string
    Headline        string
    Positions       []LinkedInPosition
    Education       []LinkedInEducation
    Certifications  []LinkedInCertification
    // ... more fields
}
```

### Coursera Data
```go
type CourseraData struct {
    Enrollments     []CourseraEnrollment
    Certifications  []CourseraCertification
    Specializations []CourseraSpecialization
    // ... more fields
}
```

See `domain/provider_data.go` for complete data models.

## Security Considerations

### Token Storage
OAuth tokens MUST be stored securely:
- Encrypt access tokens and refresh tokens at rest
- Use the `TokenStorage` port for all token operations
- Never log tokens or include them in error messages
- Implement token rotation where supported by providers

### OAuth State
- Always validate the `state` parameter to prevent CSRF attacks
- Generate cryptographically secure random states
- Store state in session with short TTL
- Validate state matches before exchanging code

### Scope Minimization
- Request only the scopes needed for the credential being issued
- Use `adapter.GetDefaultScopes()` as a baseline
- Allow users to review requested scopes before authorization
- Document why each scope is needed

### Data Privacy
- Only fetch data required for credential issuance
- Don't store raw API responses long-term
- Provide data deletion on integration disconnect
- Comply with provider API terms of service

## Configuration

### Environment Variables
```bash
# GitHub
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
GITHUB_REDIRECT_URI=https://your-app.com/oauth/github/callback

# LinkedIn
LINKEDIN_CLIENT_ID=your-client-id
LINKEDIN_CLIENT_SECRET=your-client-secret
LINKEDIN_REDIRECT_URI=https://your-app.com/oauth/linkedin/callback

# Coursera
COURSERA_CLIENT_ID=your-client-id
COURSERA_CLIENT_SECRET=your-client-secret
COURSERA_REDIRECT_URI=https://your-app.com/oauth/coursera/callback

# ... repeat for other providers
```

### Adapter Configuration
```go
config := &adapters.AdapterConfig{
    GitHub: &adapters.GitHubConfig{
        ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
        ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
        RedirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
    },
    // ... configure other providers
}
```

## Testing

### Unit Tests
Test each adapter independently:
```bash
go test ./internal/integration/infrastructure/adapters/...
```

### Integration Tests
Test against real provider APIs (use test accounts):
```bash
go test -tags=integration ./internal/integration/...
```

### Mocking
Use the `ProviderAdapter` interface for mocking in tests:
```go
type MockAdapter struct {
    mock.Mock
}

func (m *MockAdapter) FetchUserData(ctx context.Context, token string) (*domain.ProviderData, error) {
    args := m.Called(ctx, token)
    return args.Get(0).(*domain.ProviderData), args.Error(1)
}
```

## Extending with New Providers

To add a new provider:

1. **Add Provider Type** to `domain/provider_type.go`
2. **Add Data Model** to `domain/provider_data.go`
3. **Create Adapter** in `infrastructure/adapters/{provider}.go`
4. **Implement Interface**:
   - `GetAuthorizationURL`
   - `ExchangeCode`
   - `RefreshAccessToken`
   - `FetchUserData`
   - `ValidateScopes`
   - `GetDefaultScopes`
   - `RevokeAccess`
5. **Add Config** to `infrastructure/adapters/registry.go`
6. **Register** in `RegisterDefaultAdapters`
7. **Test** thoroughly

See `github.go` for a complete reference implementation.

## Next Steps

1. **Implement remaining adapters** (LinkedIn, Coursera, Twitter, Google, AWS)
2. **Add token storage** implementation (encrypted database storage)
3. **Wire up command handlers** to use adapters
4. **Add HTTP routes** for OAuth flows
5. **Integrate with Verification context** for orchestration
6. **Add rate limiting** for API calls
7. **Implement webhook handlers** for token revocation events
