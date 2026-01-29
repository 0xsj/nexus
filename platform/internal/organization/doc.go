// Package organization provides the bounded context for multi-user teams and entities.
//
// # Purpose
//
// Organization enables groups of users to collaborate under a shared entity.
// This includes companies verifying employees, recruiting firms checking candidates,
// and any scenario where multiple users need shared access to verification
// capabilities. Organization is also the foundation for the Issuer context,
// as only verified organizations can issue credentials.
//
// # Core Responsibilities
//
//   - Create and manage organizational entities
//   - Handle membership (invite, join, leave, remove)
//   - Define roles and permissions within organizations
//   - Support organizational verification (KYB - Know Your Business)
//   - Provide shared verification capabilities for members
//   - Manage organizational DIDs (did:web for verified orgs)
//   - Enable organizational profiles and branding
//
// # What Organization Does NOT Do
//
//   - Issue credentials (Issuer context, requires Organization)
//   - Manage individual user identity (Identity context)
//   - Handle billing/subscriptions (Billing context, future)
//   - Display public profiles (Profile context)
//
// # Key Entities
//
//   - Organization: The aggregate root representing a team or company.
//     Contains metadata, verification status, and settings.
//
//   - Member: A user's membership in an organization with assigned role.
//     Users can belong to multiple organizations.
//
//   - Role: Defines permissions within an organization
//     (Owner, Admin, Member, Viewer).
//
//   - Invitation: A pending invite for a user to join an organization.
//     Can be email-based or link-based.
//
//   - OrganizationDID: The decentralized identifier for verified organizations.
//     Used when the organization acts as a credential issuer.
//
//   - VerificationStatus: KYB verification state for the organization.
//
// # Domain Concepts
//
// ## Organization Structure
//
//	Organization {
//	    ID                  OrganizationID
//	    Name                string
//	    Slug                string              // proof.dev/org/acme
//	    Type                OrganizationType    // Company, Agency, Institution, Community
//	    DID                 *did.DID            // did:web:acme.proof.dev (if verified)
//	    VerificationStatus  VerificationStatus  // Unverified, Pending, Verified
//	    Settings            OrganizationSettings
//	    Branding            Branding
//	    CreatedAt           time.Time
//	}
//
// ## Organization Types
//
//   - Company: Standard business entity (startups, enterprises)
//   - Agency: Recruiting firms, staffing agencies
//   - Institution: Universities, certification bodies
//   - Community: DAOs, open source projects, professional groups
//
// Each type has different capabilities and verification requirements.
//
// ## Membership Model
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                      ORGANIZATION                               │
//	│                        (Acme Corp)                              │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Members                                                 │   │
//	│  │                                                          │   │
//	│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │   │
//	│  │  │  Alice  │  │   Bob   │  │  Carol  │  │  Dave   │    │   │
//	│  │  │  Owner  │  │  Admin  │  │ Member  │  │ Viewer  │    │   │
//	│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │   │
//	│  │                                                          │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                                                                 │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  Pending Invitations                                     │   │
//	│  │  • eve@example.com (Admin) - expires in 7 days          │   │
//	│  │  • Link invite (Member) - 3 uses remaining              │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## Roles and Permissions
//
//	| Permission              | Owner | Admin | Member | Viewer |
//	|-------------------------|-------|-------|--------|--------|
//	| View organization       |   ✓   |   ✓   |   ✓    |   ✓    |
//	| View member list        |   ✓   |   ✓   |   ✓    |   ✓    |
//	| Verify credentials      |   ✓   |   ✓   |   ✓    |        |
//	| Invite members          |   ✓   |   ✓   |        |        |
//	| Remove members          |   ✓   |   ✓   |        |        |
//	| Change member roles     |   ✓   |   ✓   |        |        |
//	| Update settings         |   ✓   |   ✓   |        |        |
//	| Issue credentials       |   ✓   |   ✓   |        |        |
//	| Manage billing          |   ✓   |       |        |        |
//	| Delete organization     |   ✓   |       |        |        |
//	| Transfer ownership      |   ✓   |       |        |        |
//
// ## Organization Verification (KYB)
//
// Organizations can be verified to unlock additional capabilities:
//
//	Unverified:
//	- Basic team features
//	- Cannot issue credentials
//	- No organizational DID
//	- Limited member count
//
//	Pending:
//	- Verification in progress
//	- Documents submitted
//	- Awaiting review
//
//	Verified:
//	- Full capabilities
//	- Can issue credentials (via Issuer context)
//	- Organizational DID (did:web)
//	- Trust anchor status eligible
//	- Unlimited members (per plan)
//
// Verification methods:
//   - Domain verification (prove ownership of company domain)
//   - Document verification (business registration, incorporation docs)
//   - Manual review (for complex cases)
//
// ## Organizational DID
//
// Verified organizations receive a did:web identifier:
//
//	did:web:acme.proof.dev
//
// This DID:
//   - Is controlled by the organization
//   - Used to sign credentials issued by the organization
//   - Resolvable via well-known DID document
//   - Links to organization's verification status
//
// # Relationships to Other Contexts
//
//   - Identity: Members are users from Identity context.
//     Organization stores membership, not user data.
//
//   - Issuer: Verified organizations can become issuers.
//     Issuer context requires an Organization.
//
//   - Trust: Organizations can be trust anchors.
//     Organizational vouches carry extra weight.
//
//   - Profile: Organizations have public profile pages.
//     Profile context renders organizational profiles.
//
//   - Credential: Organizations can verify member credentials.
//     Bulk verification for recruiting/HR use cases.
//
//   - Ledger: Organizational events projected for audit.
//
//   - Notification: Invitations and membership changes trigger notifications.
//
//   - Billing: Organizations are the billing entity (future).
//
// # Ports (Interfaces to Other Contexts)
//
//	// IdentityReader retrieves user information for membership.
//	type IdentityReader interface {
//	    GetUser(ctx context.Context, userID identity.UserID) (*identity.User, error)
//	    GetUserByEmail(ctx context.Context, email string) (*identity.User, error)
//	}
//
//	// DIDService manages organizational DIDs.
//	type DIDService interface {
//	    CreateOrganizationDID(ctx context.Context, orgSlug string) (did.DID, error)
//	    ResolveOrganizationDID(ctx context.Context, did did.DID) (*did.Document, error)
//	}
//
//	// NotificationService sends membership notifications.
//	type NotificationService interface {
//	    SendInvitation(ctx context.Context, invitation Invitation) error
//	    NotifyMembershipChange(ctx context.Context, change MembershipChange) error
//	}
//
// # Example Use Cases
//
// ## Creating an Organization
//
//	cmd := command.CreateOrganization{
//	    OwnerID:   aliceUserID,
//	    Name:      "Acme Corporation",
//	    Slug:      "acme",
//	    Type:      Company,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.OrganizationID = "org_abc123"
//	// result.URL = "https://proof.dev/org/acme"
//	// Alice automatically added as Owner
//
// ## Inviting a Member
//
//	cmd := command.InviteMember{
//	    OrganizationID: orgID,
//	    InviterID:      aliceUserID,
//	    Email:          "bob@example.com",
//	    Role:           Admin,
//	    Message:        "Join our team on Proof!",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.InvitationID = "inv_xyz789"
//	// Email sent to bob@example.com
//
// ## Accepting an Invitation
//
//	cmd := command.AcceptInvitation{
//	    InvitationID: invitationID,
//	    UserID:       bobUserID,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.MemberID = "mem_abc123"
//	// Bob is now Admin of Acme Corp
//
// ## Verifying the Organization
//
//	cmd := command.InitiateVerification{
//	    OrganizationID: orgID,
//	    Method:         DomainVerification,
//	    Domain:         "acme.com",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.VerificationToken = "proof-verify=abc123xyz"
//	// Instructions: Add TXT record to DNS
//
//	// After DNS propagation...
//	cmd := command.CompleteVerification{
//	    OrganizationID: orgID,
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.Status = Verified
//	// result.DID = "did:web:acme.proof.dev"
//
// ## Changing Member Role
//
//	cmd := command.ChangeMemberRole{
//	    OrganizationID: orgID,
//	    ChangerID:      aliceUserID,
//	    MemberID:       carolMemberID,
//	    NewRole:        Admin,
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// Carol is now Admin
//
// ## Removing a Member
//
//	cmd := command.RemoveMember{
//	    OrganizationID: orgID,
//	    RemoverID:      aliceUserID,
//	    MemberID:       daveMemberID,
//	    Reason:         "No longer with company",
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// Dave removed from organization
//
// # Architecture Notes
//
// Organization follows the standard bounded context structure:
//
//	internal/organization/
//	├── domain/
//	│   ├── organization.go    // Organization aggregate root
//	│   ├── member.go          // Member entity
//	│   ├── role.go            // Role enum and permissions
//	│   ├── invitation.go      // Invitation entity
//	│   ├── verification.go    // VerificationStatus, VerificationMethod
//	│   ├── did.go             // OrganizationDID value object
//	│   ├── settings.go        // OrganizationSettings value object
//	│   ├── branding.go        // Branding value object
//	│   ├── type.go            // OrganizationType enum
//	│   ├── ports.go           // Reader/service interfaces
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // OrganizationCreated, MemberAdded, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // CreateOrganization, InviteMember, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetOrganization, ListMembers, etc.
//	│       ├── handlers.go    // Query handlers
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── identity.go    // IdentityReader adapter
//	│   │   └── notification.go // NotificationService adapter
//	│   ├── did/
//	│   │   └── service.go     // Organizational DID management
//	│   ├── verification/
//	│   │   ├── domain.go      // Domain verification (DNS TXT)
//	│   │   └── document.go    // Document verification
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── member_repository.go
//	│           ├── invitation_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           ├── requests.go
//	│           ├── responses.go
//	│           └── middleware.go  // Org membership/permission checks
//	└── provider.go                // Dependency injection setup
//
// # Events
//
//   - OrganizationCreated: New organization created
//   - OrganizationUpdated: Settings or metadata changed
//   - OrganizationVerified: KYB verification completed
//   - OrganizationDeleted: Organization deleted
//   - MemberInvited: Invitation sent
//   - MemberAdded: User joined organization
//   - MemberRemoved: User removed from organization
//   - MemberRoleChanged: User's role changed
//   - InvitationAccepted: User accepted invitation
//   - InvitationDeclined: User declined invitation
//   - InvitationExpired: Invitation expired
//   - OwnershipTransferred: Organization ownership changed
//
// # Future Considerations
//
//   - Hierarchical organizations (parent/child orgs)
//   - Custom roles with granular permissions
//   - SSO/SAML integration for enterprise
//   - Organizational credential requirements (all members must verify X)
//   - Department/team subdivisions
//   - Organizational API keys (separate from user API keys)
//   - Audit logs scoped to organization
//   - Multi-org dashboards for agencies
//   - Organizational templates (onboarding flows)
package organization