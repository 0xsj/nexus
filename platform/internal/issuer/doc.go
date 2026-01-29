// Package issuer provides the bounded context for B2B credential issuance.
//
// # Purpose
//
// Issuer enables verified organizations to issue their own credentials through
// Proof's infrastructure. Universities can issue diplomas, employers can issue
// employment verifications, certification bodies can issue professional licenses.
// This transforms Proof from a credential aggregator into a credential issuance
// platform, opening B2B revenue streams and expanding the credential ecosystem.
//
// # Core Responsibilities
//
//   - Enable verified organizations to become credential issuers
//   - Manage credential templates (reusable credential definitions)
//   - Handle credential issuance workflows (draft, review, issue)
//   - Support batch issuance (issue to many recipients at once)
//   - Manage issuer branding on credentials
//   - Provide issuer analytics (credentials issued, verified, revoked)
//   - Handle credential revocation by issuers
//   - Expose issuance APIs for programmatic integration
//
// # What Issuer Does NOT Do
//
//   - Store issued credentials (Credential context)
//   - Manage organization membership (Organization context)
//   - Define base schema types (Schema context)
//   - Handle recipient identity (Identity context)
//
// # Key Entities
//
//   - Issuer: The aggregate root representing an organization's issuer
//     profile. Links to Organization and contains issuance settings.
//
//   - Template: A reusable credential template defining the structure,
//     claims, and branding for a type of credential the issuer issues.
//
//   - IssuanceRequest: A request to issue a credential, either individual
//     or part of a batch. Tracks workflow state.
//
//   - Batch: A group of issuance requests processed together.
//     Used for bulk operations like graduating classes.
//
//   - IssuerBranding: Visual customization for credentials (logo, colors,
//     certificate design).
//
//   - IssuancePolicy: Rules governing how credentials are issued
//     (approval workflows, expiration defaults, revocation policies).
//
// # Domain Concepts
//
// ## Issuer Lifecycle
//
//	┌─────────────────┐
//	│  Organization   │
//	│   (Verified)    │
//	└────────┬────────┘
//	         │ Register as Issuer
//	         ▼
//	┌─────────────────┐
//	│     Issuer      │
//	│    (Pending)    │
//	└────────┬────────┘
//	         │ Configure templates, branding
//	         ▼
//	┌─────────────────┐
//	│     Issuer      │
//	│    (Active)     │
//	└────────┬────────┘
//	         │ Issue credentials
//	         ▼
//	┌─────────────────┐
//	│   Credentials   │
//	│   (in Wallet)   │
//	└─────────────────┘
//
// ## Credential Templates
//
// Templates define reusable credential structures:
//
//	Template {
//	    ID              TemplateID
//	    IssuerID        IssuerID
//	    Name            string              // "Bachelor's Degree"
//	    Description     string
//	    SchemaType      string              // Base schema from Schema context
//	    ClaimOverrides  []ClaimOverride     // Issuer-specific claim customization
//	    Branding        TemplateBranding    // Visual design
//	    DefaultExpiry   *time.Duration      // Default credential lifetime
//	    IssuancePolicy  IssuancePolicy      // Approval requirements
//	    Status          TemplateStatus      // Draft, Active, Archived
//	    Version         int                 // Template versioning
//	}
//
// Example templates:
//
//	| Issuer Type     | Template Name              | Base Schema          |
//	|-----------------|----------------------------|----------------------|
//	| University      | Bachelor's Degree          | AcademicDegree       |
//	| University      | Course Completion          | CourseCompletion     |
//	| Employer        | Employment Verification    | EmploymentHistory    |
//	| Employer        | Performance Award          | Achievement          |
//	| Cert Body       | Professional License       | ProfessionalLicense  |
//	| Cert Body       | Continuing Education       | CourseCompletion     |
//
// ## Issuance Workflow
//
//	┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
//	│  Draft   │────▶│  Review  │────▶│ Approved │────▶│  Issued  │
//	└──────────┘     └──────────┘     └──────────┘     └──────────┘
//	                       │                                 │
//	                       ▼                                 ▼
//	                 ┌──────────┐                      ┌──────────┐
//	                 │ Rejected │                      │ Revoked  │
//	                 └──────────┘                      └──────────┘
//
//	Draft:    Credential data entered, not yet submitted
//	Review:   Pending approval (if policy requires)
//	Approved: Ready for issuance
//	Issued:   Credential signed and delivered to recipient
//	Rejected: Review denied, not issued
//	Revoked:  Previously issued credential revoked
//
// ## Batch Issuance
//
// For bulk operations (graduation ceremonies, annual certifications):
//
//	Batch {
//	    ID              BatchID
//	    IssuerID        IssuerID
//	    TemplateID      TemplateID
//	    Name            string              // "Class of 2024 Graduates"
//	    Requests        []IssuanceRequest   // Individual recipients
//	    Status          BatchStatus         // Draft, Processing, Completed, Failed
//	    Stats           BatchStats          // Issued, failed, pending counts
//	    CreatedAt       time.Time
//	    CompletedAt     *time.Time
//	}
//
// Batch workflow:
//  1. Create batch with template
//  2. Upload recipient list (CSV, API, manual entry)
//  3. Validate all recipients
//  4. Process batch (issue credentials)
//  5. Report results (success/failure per recipient)
//
// ## Issuer Branding
//
// Credentials carry issuer branding:
//
//	IssuerBranding {
//	    Logo            Image               // Issuer logo
//	    PrimaryColor    string              // Brand color
//	    SecondaryColor  string
//	    CertificateTemplate  string         // Visual certificate design
//	    SignatureImage  *Image              // Official signature
//	    WatermarkText   string              // Background watermark
//	}
//
// This branding appears on:
//   - Credential display in recipient's wallet
//   - Public verification pages
//   - Embeddable widgets
//   - PDF exports
//
// ## Issuance Policies
//
//	IssuancePolicy {
//	    RequiresApproval    bool            // Needs review before issuance
//	    ApproverRoles       []Role          // Who can approve
//	    AutoExpire          bool            // Credentials expire automatically
//	    DefaultExpiryDays   int             // Days until expiration
//	    AllowRevocation     bool            // Can credentials be revoked
//	    RevocationReasons   []string        // Valid revocation reasons
//	    NotifyRecipient     bool            // Email recipient on issuance
//	    RequireRecipientDID bool            // Recipient must have Proof account
//	}
//
// # Relationships to Other Contexts
//
//   - Organization: Issuer requires a verified Organization.
//     Organization members with appropriate roles can issue.
//
//   - Schema: Templates reference schema types from Schema context.
//     Issuer extends schemas with custom claims.
//
//   - Credential: Issued credentials are stored in Credential context.
//     Issuer calls Credential context to create and sign VCs.
//
//   - Identity: Recipients are identified by DID or email.
//     If email, recipient can claim credential when they register.
//
//   - Ledger: All issuance events projected for audit trail.
//
//   - Notification: Recipients notified when credentials issued.
//
// # Ports (Interfaces to Other Contexts)
//
//	// OrganizationReader verifies issuer's organization status.
//	type OrganizationReader interface {
//	    GetOrganization(ctx context.Context, id organization.OrganizationID) (*organization.Organization, error)
//	    IsVerified(ctx context.Context, id organization.OrganizationID) (bool, error)
//	    GetMemberRole(ctx context.Context, orgID organization.OrganizationID, userID identity.UserID) (organization.Role, error)
//	}
//
//	// SchemaReader retrieves base schemas for templates.
//	type SchemaReader interface {
//	    GetSchema(ctx context.Context, schemaType string) (*schema.Schema, error)
//	    ValidateClaims(ctx context.Context, schemaType string, claims map[string]any) error
//	}
//
//	// CredentialIssuer creates signed credentials.
//	type CredentialIssuer interface {
//	    Issue(ctx context.Context, req CredentialIssueRequest) (*credential.Credential, error)
//	    Revoke(ctx context.Context, credentialID credential.CredentialID, reason string) error
//	}
//
//	// IdentityResolver resolves recipient identity.
//	type IdentityResolver interface {
//	    ResolveByDID(ctx context.Context, did did.DID) (*identity.User, error)
//	    ResolveByEmail(ctx context.Context, email string) (*identity.User, error)
//	}
//
//	// NotificationService notifies recipients.
//	type NotificationService interface {
//	    NotifyCredentialIssued(ctx context.Context, recipient identity.UserID, credential credential.CredentialID) error
//	    NotifyCredentialRevoked(ctx context.Context, recipient identity.UserID, credential credential.CredentialID, reason string) error
//	}
//
// # Example Use Cases
//
// ## Registering as an Issuer
//
//	cmd := command.RegisterIssuer{
//	    OrganizationID: orgID,
//	    RequestedBy:    adminUserID,
//	    DisplayName:    "Acme University",
//	    Description:    "Official credentials from Acme University",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.IssuerID = "iss_abc123"
//	// result.Status = Pending (until branding configured)
//
// ## Creating a Template
//
//	cmd := command.CreateTemplate{
//	    IssuerID:    issuerID,
//	    Name:        "Bachelor of Science",
//	    Description: "Undergraduate degree in sciences",
//	    SchemaType:  "AcademicDegree",
//	    ClaimOverrides: []ClaimOverride{
//	        {Key: "degreeType", FixedValue: "Bachelor of Science"},
//	        {Key: "institution", FixedValue: "Acme University"},
//	    },
//	    DefaultExpiry: nil,  // Degrees don't expire
//	    Policy: IssuancePolicy{
//	        RequiresApproval: true,
//	        ApproverRoles:    []Role{Admin, Owner},
//	        NotifyRecipient:  true,
//	    },
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.TemplateID = "tpl_xyz789"
//
// ## Issuing a Single Credential
//
//	cmd := command.IssueCredential{
//	    IssuerID:    issuerID,
//	    TemplateID:  templateID,
//	    IssuedBy:    adminUserID,
//	    Recipient: RecipientInfo{
//	        Email: "alice@example.com",  // or DID
//	    },
//	    Claims: map[string]any{
//	        "recipientName": "Alice Chen",
//	        "major":         "Computer Science",
//	        "graduationDate": "2024-05-15",
//	        "honors":        "Magna Cum Laude",
//	    },
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.IssuanceRequestID = "req_abc123"
//	// result.Status = Review (if approval required) or Issued
//
// ## Creating a Batch
//
//	cmd := command.CreateBatch{
//	    IssuerID:   issuerID,
//	    TemplateID: templateID,
//	    Name:       "Class of 2024 Graduates",
//	    Recipients: []RecipientData{
//	        {Email: "alice@example.com", Claims: map[string]any{...}},
//	        {Email: "bob@example.com", Claims: map[string]any{...}},
//	        // ... hundreds more
//	    },
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.BatchID = "batch_abc123"
//	// result.TotalRecipients = 250
//	// result.Status = Draft
//
// ## Processing a Batch
//
//	cmd := command.ProcessBatch{
//	    BatchID:   batchID,
//	    IssuedBy:  adminUserID,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.Status = Processing
//	// Credentials issued asynchronously
//	// result.Stats = { total: 250, issued: 0, pending: 250, failed: 0 }
//
// ## Revoking a Credential
//
//	cmd := command.RevokeCredential{
//	    IssuerID:     issuerID,
//	    CredentialID: credentialID,
//	    RevokedBy:    adminUserID,
//	    Reason:       "Degree rescinded due to academic misconduct",
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// Credential marked as revoked in Credential context
//	// Recipient notified
//	// Event logged to Ledger
//
// ## Querying Issuer Analytics
//
//	query := query.GetIssuerAnalytics{
//	    IssuerID:  issuerID,
//	    FromDate:  time.Now().AddDate(-1, 0, 0),  // Last year
//	    ToDate:    time.Now(),
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.TotalIssued = 1500
//	// result.TotalRevoked = 3
//	// result.TotalVerified = 4200  // Times credentials were verified
//	// result.ByTemplate = { "Bachelor of Science": 500, "Master of Arts": 200, ... }
//	// result.ByMonth = [{ month: "2024-01", issued: 50 }, ...]
//
// # Architecture Notes
//
// Issuer follows the standard bounded context structure:
//
//	internal/issuer/
//	├── domain/
//	│   ├── issuer.go          // Issuer aggregate root
//	│   ├── template.go        // Template entity
//	│   ├── request.go         // IssuanceRequest entity
//	│   ├── batch.go           // Batch entity
//	│   ├── branding.go        // IssuerBranding, TemplateBranding value objects
//	│   ├── policy.go          // IssuancePolicy value object
//	│   ├── recipient.go       // RecipientInfo value object
//	│   ├── status.go          // IssuerStatus, RequestStatus, BatchStatus enums
//	│   ├── ports.go           // Interfaces to other contexts
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // IssuerRegistered, CredentialIssued, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // RegisterIssuer, CreateTemplate, IssueCredential, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   ├── query/
//	│   │   ├── queries.go     // GetIssuer, ListTemplates, GetAnalytics, etc.
//	│   │   ├── handlers.go    // Query handlers
//	│   │   └── views.go       // Read models
//	│   └── batch/
//	│       └── processor.go   // Batch processing service
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── organization.go // OrganizationReader adapter
//	│   │   ├── schema.go       // SchemaReader adapter
//	│   │   ├── credential.go   // CredentialIssuer adapter
//	│   │   ├── identity.go     // IdentityResolver adapter
//	│   │   └── notification.go // NotificationService adapter
//	│   ├── branding/
//	│   │   ├── storage.go      // Logo/image storage
//	│   │   └── renderer.go     // Certificate rendering
//	│   ├── import/
//	│   │   ├── csv.go          // CSV recipient import
//	│   │   └── validator.go    // Recipient data validation
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── template_repository.go
//	│           ├── request_repository.go
//	│           ├── batch_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   ├── http/
//	│   │   └── v1/
//	│   │       ├── handler.go
//	│   │       ├── routes.go
//	│   │       ├── requests.go
//	│   │       ├── responses.go
//	│   │       └── middleware.go   // Issuer permission checks
//	│   └── api/
//	│       └── v1/
//	│           ├── handler.go      // Programmatic issuance API
//	│           ├── routes.go
//	│           └── auth.go         // API key authentication
//	└── provider.go                 // Dependency injection setup
//
// # Events
//
//   - IssuerRegistered: Organization registered as issuer
//   - IssuerActivated: Issuer completed setup, now active
//   - IssuerSuspended: Issuer suspended (policy violation, etc.)
//   - TemplateCreated: New credential template created
//   - TemplateUpdated: Template modified
//   - TemplateArchived: Template no longer in use
//   - CredentialRequested: Issuance request created
//   - CredentialApproved: Request approved (if review required)
//   - CredentialRejected: Request rejected
//   - CredentialIssued: Credential issued to recipient
//   - CredentialRevoked: Credential revoked by issuer
//   - BatchCreated: Batch issuance initiated
//   - BatchCompleted: Batch processing finished
//   - BatchFailed: Batch processing encountered errors
//
// # API Access
//
// Issuers can integrate programmatically:
//
//	POST /api/v1/issuers/{id}/credentials
//	Authorization: Bearer <api_key>
//
//	{
//	    "template_id": "tpl_xyz789",
//	    "recipient": { "email": "alice@example.com" },
//	    "claims": { ... }
//	}
//
// This enables:
//   - LMS integration (auto-issue on course completion)
//   - HR system integration (employment verification on hire)
//   - CI/CD integration (issue contributor credentials)
//
// # Future Considerations
//
//   - Credential pricing (issuers charge for credentials)
//   - Credential marketplace (discover issuers)
//   - Issuer reputation (based on credential verification rates)
//   - Credential templates marketplace (share/sell templates)
//   - Automated issuance triggers (webhooks, integrations)
//   - Multi-signature issuance (require multiple approvers)
//   - Credential amendments (update without full reissuance)
//   - Issuer federation (cross-issuer credential recognition)
//   - White-label issuer portals
package issuer