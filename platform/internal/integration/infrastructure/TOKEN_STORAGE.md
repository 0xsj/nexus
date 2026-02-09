# 🔐 Token Storage - Secure OAuth Token Management

## Overview

The token storage system provides **encrypted, secure storage** for OAuth tokens (access tokens and refresh tokens) used by provider integrations.

## Architecture

### Components

1. **Database Schema** (`postgres/schema.sql`)
   - `integration_tokens` table with encrypted token columns
   - Foreign key to `integrations` table (cascade delete)
   - Indexes for expiration tracking and encryption version

2. **Encryption** (`crypto/encryption.go`)
   - AES-256-GCM encryption
   - Authenticated encryption (prevents tampering)
   - Base64-encoded keys for easy env var storage

3. **Token Storage** (`postgres/token_storage.go`)
   - Implements `domain.TokenStorage` interface
   - Automatic encryption/decryption
   - Support for token expiration tracking
   - Encryption version support (for key rotation)

4. **SQLC Queries** (`postgres/queries.sql`)
   - Type-safe database operations
   - Generated Go code
   - Uses pgx/v5 for performance

---

## Security Features

### ✅ Encryption

- **Algorithm:** AES-256-GCM (Galois/Counter Mode)
- **Key Size:** 256 bits (32 bytes)
- **Authenticated Encryption:** Prevents tampering
- **Random Nonces:** Unique nonce for each encryption

### ✅ Key Management

- Keys stored as base64-encoded environment variables
- Support for encryption version tracking (enables key rotation)
- Never logs or exposes keys

### ✅ Database Security

- Tokens stored as bytea (binary) not text
- Foreign key constraints ensure data integrity
- Cascade delete when integration is removed
- Timestamps for audit trails

---

## Usage

### 1. Generate Encryption Key

```bash
# Generate a new encryption key
go run internal/integration/infrastructure/crypto/keygen/main.go
```

Output:
```
=================================================================
Generated AES-256 Encryption Key (Base64)
=================================================================

abcd1234...base64key...xyz9876=

Add this to your environment variables:
export INTEGRATION_TOKEN_ENCRYPTION_KEY=abcd1234...base64key...xyz9876=

⚠️  KEEP THIS KEY SECURE!
   - Never commit to version control
   - Store in a secrets manager (AWS Secrets Manager, Vault, etc.)
   - Rotate regularly
=================================================================
```

### 2. Configure Environment

```bash
# Set encryption key
export INTEGRATION_TOKEN_ENCRYPTION_KEY=your-base64-key-here

# Or in .env file (NEVER commit this!)
INTEGRATION_TOKEN_ENCRYPTION_KEY=your-base64-key-here
```

### 3. Initialize Token Storage

```go
import (
    "github.com/0xsj/nexus/platform/internal/integration/infrastructure/crypto"
    "github.com/0xsj/nexus/platform/internal/integration/infrastructure/persistence/postgres"
    "github.com/jackc/pgx/v5/pgxpool"
)

// Load encryption key from environment
keyStr := os.Getenv("INTEGRATION_TOKEN_ENCRYPTION_KEY")

// Create encryptor
encryptor, err := crypto.NewEncryptorFromString(keyStr)
if err != nil {
    log.Fatal(err)
}

// Create database connection pool
pool, err := pgxpool.New(ctx, databaseURL)
if err != nil {
    log.Fatal(err)
}

// Create token storage
tokenStorage := postgres.NewTokenStorage(pool, encryptor)
```

### 4. Store Tokens

```go
// After successful OAuth flow
tokens := &domain.OAuthTokens{
    AccessToken:  "ya29.a0AfH6SMB...",
    RefreshToken: "1//0gHJV2...",
    TokenType:    "Bearer",
    ExpiresIn:    3600, // seconds
    Scopes:       []string{"user", "repo"},
}

// Store encrypted tokens
err := tokenStorage.Store(ctx, integrationID, tokens)
if err != nil {
    log.Fatal(err)
}
```

### 5. Retrieve Tokens

```go
// Get decrypted tokens
tokens, err := tokenStorage.Get(ctx, integrationID)
if err != nil {
    log.Fatal(err)
}

// Use tokens
accessToken := tokens.AccessToken
refreshToken := tokens.RefreshToken
```

### 6. Update Tokens (After Refresh)

```go
// After refreshing tokens
newTokens := &domain.OAuthTokens{
    AccessToken:  "new_access_token",
    RefreshToken: "new_refresh_token",
    TokenType:    "Bearer",
    ExpiresIn:    3600,
    Scopes:       []string{"user", "repo"},
}

// Update stored tokens
err := tokenStorage.Update(ctx, integrationID, newTokens)
if err != nil {
    log.Fatal(err)
}
```

### 7. Delete Tokens

```go
// When integration is disconnected
err := tokenStorage.Delete(ctx, integrationID)
if err != nil {
    log.Fatal(err)
}
```

---

## Database Schema

```sql
CREATE TABLE integration_tokens (
    integration_id      UUID PRIMARY KEY REFERENCES integrations(id) ON DELETE CASCADE,

    -- Encrypted token data (AES-256-GCM)
    access_token_encrypted  BYTEA NOT NULL,
    refresh_token_encrypted BYTEA,

    -- Token metadata
    token_type          VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at          TIMESTAMPTZ,
    scopes              TEXT[],

    -- Encryption versioning (for key rotation)
    encryption_version  INTEGER NOT NULL DEFAULT 1,

    -- Timestamps
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## API Reference

### TokenStorage Interface

```go
type TokenStorage interface {
    // Store stores encrypted OAuth tokens
    Store(ctx context.Context, integrationID string, tokens *OAuthTokens) error

    // Get retrieves and decrypts OAuth tokens
    Get(ctx context.Context, integrationID string) (*OAuthTokens, error)

    // Delete deletes OAuth tokens
    Delete(ctx context.Context, integrationID string) error

    // Update updates OAuth tokens (same as Store)
    Update(ctx context.Context, integrationID string, tokens *OAuthTokens) error
}
```

### Additional Methods

```go
// Exists checks if tokens exist for an integration
Exists(ctx context.Context, integrationID string) (bool, error)

// ListExpiredTokens returns integrations with expired tokens
ListExpiredTokens(ctx context.Context, limit int32) ([]string, error)

// CountByEncryptionVersion counts tokens using a specific encryption version
CountByEncryptionVersion(ctx context.Context, version int32) (int, error)
```

---

## Advanced Features

### Token Expiration Tracking

```go
// Find expired tokens that need refreshing
expiredIntegrationIDs, err := tokenStorage.ListExpiredTokens(ctx, 100)
if err != nil {
    log.Fatal(err)
}

for _, integrationID := range expiredIntegrationIDs {
    // Refresh tokens for this integration
    refreshTokens(integrationID)
}
```

### Encryption Key Rotation

```go
// Step 1: Deploy new encryption key with version 2
newEncryptor := crypto.NewEncryptorFromString(newKey)
newTokenStorage := postgres.NewTokenStorage(pool, newEncryptor)
newTokenStorage.version = 2

// Step 2: Check how many tokens still use version 1
count, err := tokenStorage.CountByEncryptionVersion(ctx, 1)
fmt.Printf("Tokens still using version 1: %d\n", count)

// Step 3: Re-encrypt tokens with new key
// (This would be done in a background job)
for each integration with version 1 tokens:
    tokens, _ := oldTokenStorage.Get(ctx, integrationID)
    newTokenStorage.Store(ctx, integrationID, tokens)
```

---

## Security Best Practices

### ✅ DO

- **Generate strong keys** using the provided key generator
- **Store keys in a secrets manager** (AWS Secrets Manager, HashiCorp Vault)
- **Rotate encryption keys** regularly (every 90 days recommended)
- **Use different keys** for dev/staging/production
- **Monitor token access** with audit logs
- **Delete tokens** immediately when integration is disconnected
- **Set expiration times** for all tokens when possible

### ❌ DON'T

- **Never commit encryption keys** to version control
- **Never log decrypted tokens**
- **Never expose tokens** in API responses
- **Never share encryption keys** between environments
- **Never use weak keys** (must be 32 bytes random)
- **Never store tokens in plaintext**
- **Never skip encryption** even for dev environments

---

## Troubleshooting

### "Failed to decrypt: message authentication failed"

**Cause:** Token was encrypted with a different key

**Solution:**
- Verify the encryption key in your environment matches the one used to encrypt
- Check if key was rotated without re-encrypting tokens
- Regenerate OAuth tokens if key is lost

### "Encryption key must be 32 bytes"

**Cause:** Invalid encryption key format

**Solution:**
- Use the key generator: `go run crypto/keygen/main.go`
- Ensure the full base64 string is in the environment variable
- Don't truncate or modify the generated key

### "Tokens not found for integration"

**Cause:** No tokens stored for this integration

**Solution:**
- Complete OAuth flow first to obtain tokens
- Check if tokens were deleted (integration disconnected)
- Verify integration ID is correct

---

## Performance Considerations

### Database Indexes

- ✅ Primary key on `integration_id` (fast lookups)
- ✅ Index on `expires_at` (fast expiration queries)
- ✅ Index on `encryption_version` (fast key rotation queries)

### Encryption Overhead

- **Encryption:** ~0.1ms per token
- **Decryption:** ~0.1ms per token
- **Minimal impact** on API latency

### Connection Pooling

Uses `pgxpool` for efficient database connections:
- Automatic connection pooling
- Connection reuse
- Health checks

---

## Testing

### Unit Tests

```go
func TestTokenStorage_StoreAndGet(t *testing.T) {
    // Setup
    pool := setupTestDB(t)
    encryptor, _ := crypto.NewEncryptor(testKey)
    storage := postgres.NewTokenStorage(pool, encryptor)

    // Store tokens
    tokens := &domain.OAuthTokens{
        AccessToken: "test_access",
        RefreshToken: "test_refresh",
    }
    err := storage.Store(ctx, integrationID, tokens)
    assert.NoError(t, err)

    // Retrieve tokens
    retrieved, err := storage.Get(ctx, integrationID)
    assert.NoError(t, err)
    assert.Equal(t, tokens.AccessToken, retrieved.AccessToken)
}
```

### Integration Tests

Test with real database:
```bash
go test -tags=integration ./internal/integration/infrastructure/persistence/postgres
```

---

## Migration

### From plaintext to encrypted storage

```go
// Step 1: Add encrypted columns
// (Already done in schema.sql)

// Step 2: Migrate existing tokens
migrateTokens := func() {
    // For each integration with plaintext tokens
    for each integration:
        plainTokens := getPlaintextTokens(integrationID)
        encryptedTokens := &domain.OAuthTokens{
            AccessToken: plainTokens.AccessToken,
            RefreshToken: plainTokens.RefreshToken,
        }
        tokenStorage.Store(ctx, integrationID, encryptedTokens)
}

// Step 3: Drop plaintext columns
// ALTER TABLE integrations DROP COLUMN access_token;
// ALTER TABLE integrations DROP COLUMN refresh_token;
```

---

## Monitoring

### Metrics to Track

- **Token refresh failures** (may indicate expired refresh tokens)
- **Decryption failures** (may indicate key mismatch)
- **Expired token count** (should trend towards zero)
- **Encryption version distribution** (for key rotation progress)

### Alerts

```
Alert if:
- Decryption failure rate > 1%
- Expired tokens > 100
- All tokens still on old encryption version after 30 days
```

---

## Production Checklist

- [ ] Encryption key generated and stored in secrets manager
- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] Token storage initialized in application
- [ ] Monitoring and alerting configured
- [ ] Key rotation procedure documented
- [ ] Incident response plan created
- [ ] Security audit completed

---

## Summary

Token storage is **production-ready** with:
- ✅ AES-256-GCM encryption
- ✅ Type-safe database operations (SQLC)
- ✅ Expiration tracking
- ✅ Key rotation support
- ✅ Comprehensive documentation
- ✅ Security best practices

**Next Steps:**
1. Generate encryption key
2. Configure environment
3. Apply database migrations
4. Wire up to HTTP endpoints
