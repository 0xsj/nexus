// Package ledger provides the bounded context for immutable audit trail and event projection.
//
// # Purpose
//
// Ledger maintains a comprehensive, immutable record of all significant events
// across the Proof platform. It subscribes to domain events from other bounded
// contexts and projects them into queryable audit logs. Ledger serves compliance
// requirements, enables user-facing activity history, and provides forensic
// capabilities for investigating credential lifecycle events.
//
// # Core Responsibilities
//
//   - Subscribe to domain events from all bounded contexts
//   - Project events into immutable audit records
//   - Provide queryable activity history per user, credential, or organization
//   - Support compliance exports (audit reports, data retention)
//   - Track credential lifecycle (issued, presented, verified, revoked)
//   - Record verification attempts (who verified what, when, outcome)
//   - Maintain tamper-evident log integrity
//
// # What Ledger Does NOT Do
//
//   - Aggregate reconstruction (that's event sourcing in each domain)
//   - Business logic or validation (read-only projection)
//   - Event storage for replay (that's pkg/eventsourcing)
//   - Real-time notifications (that's Notification context)
//
// # Key Entities
//
//   - AuditEntry: An immutable record of a single event, containing event type,
//     actor, subject, timestamp, and relevant metadata.
//
//   - ActivityStream: A chronological view of entries filtered by user,
//     credential, organization, or other criteria.
//
//   - VerificationLog: Records of external verification attempts against
//     user credentials (who checked, when, what was disclosed).
//
//   - ComplianceReport: Aggregated audit data formatted for regulatory or
//     internal compliance requirements.
//
// # Domain Concepts
//
// ## Event Categories
//
// Ledger categorizes events by their domain and significance:
//
//	Identity Events:
//	- UserRegistered, UserDeleted
//	- SessionCreated, SessionRevoked
//	- DIDLinked, DIDUnlinked
//	- APIKeyCreated, APIKeyRevoked
//
//	Wallet Events:
//	- WalletLinked, WalletUnlinked
//	- SignatureVerified
//
//	Verification Events:
//	- VerificationInitiated, VerificationCompleted, VerificationFailed
//	- OAuthTokensReceived, OAuthTokensRevoked
//	- ProviderDataFetched
//
//	Credential Events:
//	- CredentialIssued, CredentialRevoked, CredentialExpired
//	- CredentialUpdated (re-verification with new data)
//
//	Presentation Events:
//	- PresentationCreated, PresentationShared
//	- PresentationVerified (external party verified)
//	- ShareLinkCreated, ShareLinkAccessed, ShareLinkRevoked
//
//	Trust Events:
//	- VouchGiven, VouchReceived, VouchRevoked
//	- ReputationUpdated
//
//	Organization Events:
//	- MemberAdded, MemberRemoved, RoleChanged
//	- IssuerRegistered, IssuerSuspended
//
// ## Audit Entry Structure
//
//	AuditEntry {
//	    ID          EntryID
//	    Timestamp   time.Time       // When the event occurred
//	    EventType   string          // e.g., "credential.issued"
//	    ActorID     string          // Who caused the event (user, system, external)
//	    ActorType   ActorType       // User, System, ExternalVerifier, Issuer
//	    SubjectID   string          // What the event is about
//	    SubjectType SubjectType     // User, Credential, Presentation, etc.
//	    Metadata    map[string]any  // Event-specific details
//	    ContextID   string          // Correlation ID for related events
//	    Hash        string          // Integrity hash (links to previous entry)
//	}
//
// ## Tamper Evidence
//
// Ledger maintains integrity through hash chaining:
//
//	Entry N:   Hash = SHA256(Entry N data + Entry N-1 Hash)
//	Entry N+1: Hash = SHA256(Entry N+1 data + Entry N Hash)
//
// This creates a verifiable chain where any modification to historical
// entries would invalidate subsequent hashes.
//
// ## Retention Policies
//
// Different event types have different retention requirements:
//
//   - Security events (auth, sessions): 2 years minimum
//   - Credential lifecycle: Indefinite (or until user deletion)
//   - Verification attempts: 5 years (compliance)
//   - User activity: Configurable per user preference
//
// # Relationships to Other Contexts
//
//   - All Contexts: Ledger subscribes to domain events from every bounded
//     context. It is a read-only consumer, never publishing commands back.
//
//   - Identity: Provides actor context (who performed actions). User deletion
//     triggers retention policy evaluation.
//
//   - Credential: Primary source of credential lifecycle events.
//
//   - Presentation: Source of sharing and external verification events.
//
//   - Organization: Source of membership and issuer events.
//
//   - Notification: May read from Ledger to include activity summaries.
//
// # Event Subscription
//
// Ledger subscribes to domain events via the event bus:
//
//	type EventSubscriber interface {
//	    Subscribe(eventTypes []string, handler EventHandler) error
//	}
//
//	type EventHandler func(ctx context.Context, event DomainEvent) error
//
//	// Ledger's projection handler
//	func (l *Ledger) HandleEvent(ctx context.Context, event DomainEvent) error {
//	    entry := l.projectEvent(event)
//	    return l.repository.Append(ctx, entry)
//	}
//
// # Example Use Cases
//
// ## Querying User Activity
//
//	query := query.GetUserActivity{
//	    UserID:    userID,
//	    FromTime:  time.Now().AddDate(0, -1, 0), // Last month
//	    ToTime:    time.Now(),
//	    EventTypes: []string{"credential.*", "presentation.*"},
//	    Limit:     50,
//	}
//
//	activity, err := handler.Handle(ctx, query)
//	// Returns chronological list of audit entries
//
// ## Credential Lifecycle History
//
//	query := query.GetCredentialHistory{
//	    CredentialID: credentialID,
//	}
//
//	history, err := handler.Handle(ctx, query)
//	// Returns: Issued -> Presented (3 times) -> Verified (2 times) -> Revoked
//
// ## External Verification Log
//
//	query := query.GetVerificationLog{
//	    UserID:   userID,
//	    FromTime: time.Now().AddDate(-1, 0, 0), // Last year
//	}
//
//	log, err := handler.Handle(ctx, query)
//	// Returns: Who verified your credentials, when, what they saw
//
// ## Compliance Export
//
//	cmd := command.GenerateComplianceReport{
//	    OrganizationID: orgID,
//	    ReportType:     "SOC2",
//	    Period:         "2024-Q4",
//	}
//
//	report, err := handler.Handle(ctx, cmd)
//	// Returns formatted compliance report with relevant audit entries
//
// # Architecture Notes
//
// Ledger follows the standard bounded context structure:
//
//	internal/ledger/
//	├── domain/
//	│   ├── entry.go           // AuditEntry aggregate
//	│   ├── stream.go          // ActivityStream value object
//	│   ├── actor.go           // ActorType, ActorID value objects
//	│   ├── subject.go         // SubjectType, SubjectID value objects
//	│   ├── hash.go            // Hash chaining logic
//	│   ├── retention.go       // RetentionPolicy value object
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // Ledger's own events (if any)
//	│   └── repository.go      // Repository interface (append-only)
//	├── application/
//	│   ├── projection/
//	│   │   ├── handlers.go    // Event handlers that project to audit entries
//	│   │   └── mapper.go      // Maps domain events to audit entries
//	│   ├── command/
//	│   │   ├── commands.go    // GenerateReport, PurgeExpired, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetUserActivity, GetCredentialHistory, etc.
//	│       ├── handlers.go    // Query handlers
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── subscription/
//	│   │   └── event_bus.go   // Subscribes to domain events
//	│   ├── persistence/
//	│   │   └── postgres/
//	│   │       ├── repository.go  // Append-only storage
//	│   │       ├── queries.go
//	│   │       └── migrations/
//	│   └── export/
//	│       ├── csv.go         // CSV export
//	│       └── json.go        // JSON export
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           └── responses.go
//	└── provider.go            // Dependency injection setup
//
// # Storage Considerations
//
// Ledger is append-only with high write volume. Storage strategy:
//
//   - Primary: PostgreSQL with time-partitioned tables
//   - Indexes: actor_id, subject_id, event_type, timestamp
//   - Archival: Cold storage for entries beyond retention window
//   - Optional: Append to blockchain for critical events (future)
//
// # Relationship to pkg/eventsourcing
//
//	| Concern           | pkg/eventsourcing          | internal/ledger            |
//	|-------------------|----------------------------|----------------------------|
//	| Purpose           | Aggregate state rebuild    | Cross-domain audit trail   |
//	| Scope             | Single aggregate           | Entire platform            |
//	| Consumers         | Domain repositories        | Users, compliance, admins  |
//	| Retention         | Forever (event replay)     | Policy-based               |
//	| Queryability      | By aggregate ID            | By actor, subject, time    |
//
// Both can coexist:
//   - Event sourcing stores granular domain events for state reconstruction
//   - Ledger projects significant events for human-readable audit trail
//
// # Future Considerations
//
//   - Real-time activity feed via WebSocket subscription
//   - Anomaly detection (unusual verification patterns)
//   - GDPR right-to-erasure handling (anonymization vs. deletion)
//   - Multi-region replication for compliance
//   - Blockchain anchoring for critical credential events
//   - GraphQL API for flexible activity queries
package ledger
