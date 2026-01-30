// Package notification provides the bounded context for user communications.
//
// # Purpose
//
// Notification handles all outbound communications to users across the platform.
// It subscribes to domain events from other contexts and delivers timely,
// relevant notifications via email, push, and in-app channels. Notification
// is a cross-cutting concern that keeps users informed without coupling other
// contexts to delivery mechanisms.
//
// # Core Responsibilities
//
//   - Subscribe to domain events that require user notification
//   - Manage notification preferences per user
//   - Render notifications from templates
//   - Deliver via multiple channels (email, push, in-app)
//   - Track delivery status and engagement
//   - Handle notification batching and digests
//   - Manage unsubscribe and suppression lists
//   - Provide notification history and inbox
//
// # What Notification Does NOT Do
//
//   - Generate events (consumes events from other contexts)
//   - Make business decisions (only delivery)
//   - Store domain data (only notification metadata)
//   - Handle real-time messaging (WebSocket is separate infrastructure)
//
// # Key Entities
//
//   - Notification: The aggregate root representing a single notification
//     to be delivered. Contains content, recipient, channel, and status.
//
//   - NotificationPreferences: User's preferences for what notifications
//     they want and via which channels.
//
//   - Template: A reusable notification template with placeholders for
//     dynamic content. Supports multiple channels (email, push).
//
//   - DeliveryAttempt: Records each attempt to deliver a notification,
//     including status, timestamp, and any errors.
//
//   - Digest: A batched summary of multiple notifications sent periodically
//     instead of individually.
//
//   - Channel: A delivery mechanism (email, push, in-app, SMS).
//
// # Domain Concepts
//
// ## Notification Lifecycle
//
//	┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
//	│  Event   │────▶│  Queued  │────▶│ Sending  │────▶│Delivered │
//	│ Received │     │          │     │          │     │          │
//	└──────────┘     └──────────┘     └──────────┘     └──────────┘
//	                       │                │                │
//	                       ▼                ▼                ▼
//	                 ┌──────────┐     ┌──────────┐     ┌──────────┐
//	                 │Suppressed│     │  Failed  │     │   Read   │
//	                 └──────────┘     └──────────┘     └──────────┘
//
//	Event Received: Domain event triggers notification creation
//	Queued:         Notification created, waiting for delivery
//	Suppressed:     Skipped due to user preferences or rate limiting
//	Sending:        Delivery in progress
//	Failed:         Delivery failed (will retry or abandon)
//	Delivered:      Successfully delivered to channel
//	Read:           User viewed/opened the notification
//
// ## Notification Categories
//
// Notifications are categorized for preference management:
//
//	| Category       | Examples                                           |
//	|----------------|----------------------------------------------------|
//	| Security       | Login alerts, API key created, session revoked     |
//	| Credentials    | Credential issued, verified, expiring, revoked     |
//	| Verification   | Verification started, completed, failed            |
//	| Social         | Vouch received, vouch requested, profile viewed    |
//	| Organization   | Invitation received, role changed, member added    |
//	| Issuer         | Credential request, batch completed                |
//	| System         | Maintenance, policy updates, new features          |
//
// ## Delivery Channels
//
//	Email:
//	- Primary channel for important notifications
//	- Rich HTML templates with branding
//	- Supports digests and batching
//	- Tracks opens and clicks
//
//	Push:
//	- Mobile and web push notifications
//	- Real-time alerts for time-sensitive events
//	- Short-form content
//	- Requires device registration
//
//	In-App:
//	- Notification inbox within the application
//	- Always delivered (no suppression)
//	- Supports rich content and actions
//	- Read/unread tracking
//
//	SMS (future):
//	- Critical security alerts only
//	- Opt-in required
//	- Character-limited content
//
// ## User Preferences
//
//	NotificationPreferences {
//	    UserID          identity.UserID
//	    GlobalEnabled   bool                    // Master switch
//	    Channels        map[Channel]bool        // Which channels enabled
//	    Categories      map[Category]ChannelSet // Per-category channel preferences
//	    DigestEnabled   bool                    // Batch into digest
//	    DigestFrequency DigestFrequency         // Daily, Weekly
//	    QuietHours      *QuietHours             // Don't notify during these hours
//	    Timezone        string                  // For quiet hours and digest timing
//	}
//
// Example preference:
//
//	{
//	    "globalEnabled": true,
//	    "channels": { "email": true, "push": true, "inApp": true },
//	    "categories": {
//	        "security":     { "email": true, "push": true, "inApp": true },
//	        "credentials":  { "email": true, "push": false, "inApp": true },
//	        "social":       { "email": false, "push": false, "inApp": true },
//	        "organization": { "email": true, "push": true, "inApp": true }
//	    },
//	    "digestEnabled": true,
//	    "digestFrequency": "daily",
//	    "quietHours": { "start": "22:00", "end": "08:00" },
//	    "timezone": "America/New_York"
//	}
//
// ## Templates
//
// Notification templates support multiple channels:
//
//	Template {
//	    ID              TemplateID
//	    EventType       string              // e.g., "credential.issued"
//	    Category        Category
//	    Name            string              // Human-readable name
//	    Channels        map[Channel]ChannelTemplate
//	}
//
//	ChannelTemplate {
//	    Subject         string              // Email subject / push title
//	    Body            string              // Template with placeholders
//	    Format          TemplateFormat      // HTML, Markdown, Plain
//	    ActionURL       string              // Primary action link
//	    ActionLabel     string              // Button text
//	}
//
// Placeholder syntax:
//
//	"Hi {{.RecipientName}}, your {{.CredentialType}} credential has been issued."
//
// ## Digests
//
// Digest batches multiple notifications:
//
//	┌─────────────────────────────────────────────────────────────┐
//	│                    DAILY DIGEST                             │
//	│                    January 15, 2024                         │
//	│                                                             │
//	│  Credentials                                                │
//	│  • Your GitHub credential was verified (2 hours ago)        │
//	│  • AWS certification expiring in 30 days                    │
//	│                                                             │
//	│  Social                                                     │
//	│  • Bob vouched for your engineering skills                  │
//	│  • Carol requested a vouch from you                         │
//	│  • 12 people viewed your profile                            │
//	│                                                             │
//	│  Organization                                               │
//	│  • You were added to Acme Corp                              │
//	└─────────────────────────────────────────────────────────────┘
//
// # Event Subscriptions
//
// Notification subscribes to events from all contexts:
//
//	| Source Context | Events                                          |
//	|----------------|-------------------------------------------------|
//	| Identity       | UserRegistered, SessionCreated, APIKeyCreated   |
//	| Wallet         | WalletLinked, WalletUnlinked                    |
//	| Verification   | VerificationCompleted, VerificationFailed       |
//	| Credential     | CredentialIssued, CredentialExpiring, Revoked   |
//	| Presentation   | ShareLinkAccessed, PresentationVerified         |
//	| Trust          | VouchReceived, VouchRequested                   |
//	| Organization   | InvitationReceived, MemberAdded, RoleChanged    |
//	| Issuer         | CredentialRequested, BatchCompleted             |
//	| Profile        | ProfileViewed (aggregated)                      |
//
// # Relationships to Other Contexts
//
//   - All Contexts: Notification subscribes to domain events.
//     It is a pure consumer, never publishing back.
//
//   - Identity: Provides recipient information (email, name).
//     Notification preferences stored per user.
//
//   - Ledger: Notification delivery events may be logged.
//     Audit trail of what was sent to whom.
//
// # Ports (Interfaces to Other Contexts)
//
//	// IdentityReader retrieves recipient information.
//	type IdentityReader interface {
//	    GetUser(ctx context.Context, userID identity.UserID) (*identity.User, error)
//	    GetUserEmail(ctx context.Context, userID identity.UserID) (string, error)
//	}
//
//	// EventSubscriber subscribes to domain events.
//	type EventSubscriber interface {
//	    Subscribe(eventTypes []string, handler EventHandler) error
//	}
//
// # Example Use Cases
//
// ## Handling a Credential Issued Event
//
//	// Event received from Credential context
//	event := CredentialIssuedEvent{
//	    CredentialID: "cred_abc123",
//	    SubjectID:    "user_xyz789",
//	    SchemaType:   "GitHubContributor",
//	    IssuedAt:     time.Now(),
//	}
//
//	// Notification handler processes event
//	func (h *Handler) HandleCredentialIssued(ctx context.Context, event CredentialIssuedEvent) error {
//	    // 1. Check user preferences
//	    prefs, _ := h.prefsRepo.Get(ctx, event.SubjectID)
//	    if !prefs.ShouldNotify(CategoryCredentials, ChannelEmail) {
//	        return nil // Suppressed
//	    }
//
//	    // 2. Create notification
//	    notification := domain.NewNotification(
//	        event.SubjectID,
//	        "credential.issued",
//	        map[string]any{
//	            "CredentialType": event.SchemaType,
//	            "IssuedAt":       event.IssuedAt,
//	        },
//	    )
//
//	    // 3. Queue for delivery
//	    return h.notificationRepo.Save(ctx, notification)
//	}
//
// ## Updating Notification Preferences
//
//	cmd := command.UpdatePreferences{
//	    UserID: userID,
//	    Updates: PreferenceUpdates{
//	        Categories: map[Category]ChannelSet{
//	            Social: {Email: false, Push: false, InApp: true},
//	        },
//	        DigestEnabled:   true,
//	        DigestFrequency: Daily,
//	    },
//	}
//
//	err := handler.Handle(ctx, cmd)
//
// ## Querying Notification Inbox
//
//	query := query.GetInbox{
//	    UserID:     userID,
//	    Unread:     true,  // Only unread
//	    Limit:      20,
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.Notifications = [...]
//	// result.UnreadCount = 5
//
// ## Marking Notifications as Read
//
//	cmd := command.MarkAsRead{
//	    UserID:          userID,
//	    NotificationIDs: []NotificationID{id1, id2, id3},
//	}
//
//	err := handler.Handle(ctx, cmd)
//
// ## Sending a Digest
//
//	// Scheduled job runs daily/weekly
//	cmd := command.SendDigests{
//	    Frequency: Daily,
//	    AsOf:      time.Now(),
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.DigestsSent = 1500
//	// result.UsersSkipped = 200 (no pending notifications)
//
// # Architecture Notes
//
// Notification follows the standard bounded context structure:
//
//	internal/notification/
//	├── domain/
//	│   ├── notification.go    // Notification aggregate root
//	│   ├── preferences.go     // NotificationPreferences entity
//	│   ├── template.go        // Template entity
//	│   ├── digest.go          // Digest value object
//	│   ├── channel.go         // Channel enum (Email, Push, InApp)
//	│   ├── category.go        // Category enum
//	│   ├── delivery.go        // DeliveryAttempt, DeliveryStatus
//	│   ├── suppression.go     // SuppressionRule, SuppressionList
//	│   ├── ports.go           // IdentityReader, EventSubscriber
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // NotificationSent, NotificationFailed, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // UpdatePreferences, MarkAsRead, SendDigests
//	│   │   └── handlers.go    // Command handlers
//	│   ├── query/
//	│   │   ├── queries.go     // GetInbox, GetPreferences, etc.
//	│   │   ├── handlers.go    // Query handlers
//	│   │   └── views.go       // Read models
//	│   ├── subscription/
//	│   │   ├── handlers.go    // Event handlers for each source context
//	│   │   └── router.go      // Routes events to appropriate handlers
//	│   └── delivery/
//	│       ├── service.go     // Orchestrates delivery across channels
//	│       └── retry.go       // Retry logic for failed deliveries
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   └── identity.go    // IdentityReader adapter
//	│   ├── channels/
//	│   │   ├── email/
//	│   │   │   ├── sender.go      // Email delivery (SendGrid, SES, etc.)
//	│   │   │   └── templates/     // HTML email templates
//	│   │   ├── push/
//	│   │   │   ├── sender.go      // Push notification (FCM, APNs)
//	│   │   │   └── registry.go    // Device token registry
//	│   │   └── inapp/
//	│   │       └── store.go       // In-app notification storage
//	│   ├── templates/
//	│   │   ├── renderer.go        // Template rendering engine
//	│   │   └── defaults/          // Default notification templates
//	│   ├── scheduling/
//	│   │   ├── digest.go          // Digest scheduling
//	│   │   └── quiethours.go      // Quiet hours enforcement
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── preferences_repository.go
//	│           ├── template_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           ├── requests.go
//	│           ├── responses.go
//	│           └── unsubscribe.go  // One-click unsubscribe handling
//	└── provider.go                 // Dependency injection setup
//
// # Delivery Infrastructure
//
//	| Channel | Provider Options                           |
//	|---------|--------------------------------------------|
//	| Email   | SendGrid, AWS SES, Postmark, Mailgun       |
//	| Push    | Firebase Cloud Messaging, Apple APNs       |
//	| In-App  | PostgreSQL + WebSocket for real-time       |
//	| SMS     | Twilio, AWS SNS (future)                   |
//
// # Events
//
//   - NotificationCreated: Notification queued for delivery
//   - NotificationSent: Successfully delivered via channel
//   - NotificationFailed: Delivery failed
//   - NotificationRead: User viewed notification
//   - NotificationSuppressed: Skipped due to preferences/rules
//   - PreferencesUpdated: User changed notification preferences
//   - DigestSent: Digest email delivered
//   - DeviceRegistered: Push notification device registered
//   - DeviceUnregistered: Push notification device removed
//
// # Rate Limiting & Suppression
//
// To prevent notification fatigue:
//
//   - Per-user rate limits (max N notifications per hour)
//   - Per-category rate limits (max N social notifications per day)
//   - Duplicate suppression (same notification within time window)
//   - Quiet hours enforcement (defer delivery)
//   - Global suppression list (bounced emails, unsubscribed)
//
// # Future Considerations
//
//   - Rich push notifications (images, action buttons)
//   - Notification webhooks (deliver to external endpoints)
//   - AI-powered notification timing (optimal send time)
//   - Notification A/B testing (template variants)
//   - Multi-language templates
//   - Notification analytics dashboard
//   - Custom notification rules (user-defined triggers)
//   - Slack/Discord integration
//   - WhatsApp Business integration
package notification
