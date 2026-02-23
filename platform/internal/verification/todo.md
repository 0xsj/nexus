# Verification — Core Product Loop TODO

## What's Done

- Verification, Integration, and Credential contexts wired end-to-end
- `IntegrationBridge` adapter bridges Verification ports to Integration's provider registry
- `HandleReceiveOAuthCallback` orchestrates: validate state → exchange code → fetch data → issue credential
- GitHub and Google adapter configs wired from env vars in main.go
- `go build` and `go vet` pass clean

## Next Steps

### 1. OAuth State → Verification ID Lookup

**Blocker for working demo.** After OAuth redirect, the frontend only has `code` and `state` — it doesn't know the `verification_id`. Options:

- [ ] Wire the existing `OAuthStateRepository` port (currently null) with a real implementation (postgres or in-memory) to map `state → verification_id`
- [ ] Add a state-based callback endpoint: `POST /api/v1/verifications/callback` that accepts `{code, state}` and resolves the verification internally
- [ ] Or: encode verification ID in the OAuth state param (e.g. `state = verificationID + ":" + random`)

### 2. Frontend Callback Page

- [ ] Create route at `/auth/callback/:provider` (e.g. `/auth/callback/github`)
- [ ] Extract `code` and `state` from URL query params (provider redirects with these)
- [ ] Call `POST /api/v1/verifications/{id}/callback` with `{code, state, redirect_uri}`
- [ ] Handle success (redirect to credential view) and error states
- [ ] Store verification ID before redirect (localStorage or encode in state — depends on #1)

### 3. End-to-End Smoke Test

- [ ] Set real `OAUTH_GITHUB_CLIENT_ID` and `OAUTH_GITHUB_CLIENT_SECRET`
- [ ] `POST /api/v1/verifications` with `{provider_type: "github", redirect_uri: "..."}` → get `auth_url`
- [ ] Complete OAuth flow in browser
- [ ] Verify callback returns `credential_id`
- [ ] `GET /api/v1/credentials/{credentialId}` → confirm issued credential

### 4. Remaining Provider Configs

- [ ] Wire LinkedIn env vars (`OAUTH_LINKEDIN_CLIENT_ID`, `OAUTH_LINKEDIN_CLIENT_SECRET`)
- [ ] Wire Twitter env vars
- [ ] Wire Coursera env vars
- [ ] Wire AWS env vars
- [ ] Add redirect URIs for each in main.go (`appBaseURL + "/auth/callback/{provider}"`)

### 5. Token Persistence (Optional)

Access tokens are currently ephemeral (used once in the handler, then discarded). If re-fetching or refreshing provider data is needed later:

- [ ] Wire `ProviderTokenRepository` port with a real implementation
- [ ] Store access/refresh tokens after code exchange
- [ ] Add refresh logic for expired tokens

### 6. Tests

- [ ] Unit tests for `HandleReceiveOAuthCallback` error paths:
  - State mismatch → aggregate fails
  - Code exchange failure → `RecordDataFetchFailed("CODE_EXCHANGE_FAILED")`
  - Data fetch failure → `RecordDataFetchFailed("DATA_FETCH_FAILED")`
  - Credential issuance failure → `Fail("CREDENTIAL_ISSUANCE_FAILED")`
- [ ] Unit tests for `HandleStartVerification` auth URL generation
- [ ] Integration test for `IntegrationBridge` with mock provider adapter
