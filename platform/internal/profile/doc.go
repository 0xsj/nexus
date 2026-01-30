// Package profile provides the bounded context for public-facing user profiles.
//
// # Purpose
//
// Profile manages the public representation of a user's verified identity.
// It aggregates credentials into displayable badges, controls what information
// is publicly visible, and provides embeddable widgets for external sites.
// Profile is the "storefront" of a user's verified professional identity.
//
// # Core Responsibilities
//
//   - Manage public profile pages with customizable visibility
//   - Aggregate credentials into displayable badges
//   - Control privacy settings (what's public, what requires request)
//   - Generate embeddable widgets for external websites
//   - Provide vanity URLs (proof.dev/@username)
//   - Track profile views and engagement metrics
//   - Support profile theming and customization
//
// # What Profile Does NOT Do
//
//   - Store credentials (Credential context)
//   - Manage sharing/presentations (Presentation context)
//   - Handle authentication (Identity context)
//   - Issue or verify credentials (Credential/Verification contexts)
//
// # Key Entities
//
//   - Profile: The aggregate root representing a user's public profile.
//     Contains display settings, privacy controls, and badge configuration.
//
//   - Badge: A visual representation of a verified credential or achievement.
//     Derived from credentials but optimized for display.
//
//   - PrivacySettings: Controls what information is visible publicly,
//     to connections only, or hidden entirely.
//
//   - ProfileSection: A configurable section of the profile (skills,
//     experience, certifications, etc.) with its own visibility rules.
//
//   - VanityURL: A custom URL slug for the profile (proof.dev/@alice).
//
//   - Embed: Configuration for embeddable profile widgets.
//
// # Domain Concepts
//
// ## Profile Structure
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                      PUBLIC PROFILE                             │
//	│  ┌───────────────────────────────────────────────────────────┐ │
//	│  │  Header                                                    │ │
//	│  │  ┌──────┐                                                  │ │
//	│  │  │Avatar│  Alice Chen                                      │ │
//	│  │  └──────┘  @alice · did:key:z6Mk...                       │ │
//	│  │            "Senior Software Engineer"                      │ │
//	│  └───────────────────────────────────────────────────────────┘ │
//	│                                                                 │
//	│  ┌───────────────────────────────────────────────────────────┐ │
//	│  │  Verified Badges                                          │ │
//	│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐         │ │
//	│  │  │ GitHub  │ │LinkedIn │ │  AWS    │ │Coursera │         │ │
//	│  │  │  1337   │ │ 5 yrs   │ │Solutions│ │ ML Cert │         │ │
//	│  │  │ commits │ │  exp    │ │Architect│ │         │         │ │
//	│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘         │ │
//	│  └───────────────────────────────────────────────────────────┘ │
//	│                                                                 │
//	│  ┌───────────────────────────────────────────────────────────┐ │
//	│  │  Experience (from LinkedIn credential)                    │ │
//	│  │  • Senior Engineer at TechCorp (2020-present) ✓          │ │
//	│  │  • Engineer at StartupXYZ (2018-2020) ✓                  │ │
//	│  └───────────────────────────────────────────────────────────┘ │
//	│                                                                 │
//	│  ┌───────────────────────────────────────────────────────────┐ │
//	│  │  Vouches                                                   │ │
//	│  │  "Alice is an exceptional engineer" - Bob (verified)      │ │
//	│  └───────────────────────────────────────────────────────────┘ │
//	└─────────────────────────────────────────────────────────────────┘
//
// ## Badges
//
// Badges are the visual summary of credentials:
//
//	Badge {
//	    ID              BadgeID
//	    CredentialID    credential.CredentialID
//	    Type            BadgeType           // GitHub, LinkedIn, Certification, etc.
//	    DisplayName     string              // "GitHub Contributor"
//	    PrimaryValue    string              // "1337 commits"
//	    SecondaryValue  string              // "42 repositories" (optional)
//	    Icon            string              // Icon identifier or URL
//	    VerifiedAt      time.Time           // When credential was issued
//	    Visibility      Visibility          // Public, ConnectionsOnly, Private
//	}
//
// Badge types and their display:
//
//	| Credential Type        | Badge Display                              |
//	|------------------------|--------------------------------------------|
//	| GitHubContributor      | commits count, top languages               |
//	| ProfessionalExperience | years of experience, current role          |
//	| CourseCompletion       | course name, provider logo                 |
//	| CloudCertification     | cert name, level, expiry indicator         |
//	| FreelanceReputation    | jobs completed, success rate               |
//
// ## Privacy Levels
//
//   - Public: Visible to anyone viewing the profile
//   - ConnectionsOnly: Visible only to users you've connected with
//   - RequestRequired: Viewer must request access, you approve
//   - Private: Hidden from profile entirely (credential still exists)
//
// ## Profile Sections
//
// Users organize their profile into sections:
//
//   - Overview: Bio, headline, key badges
//   - Experience: Employment history (from LinkedIn credentials)
//   - Skills: Technical skills with verification badges
//   - Certifications: Professional certifications
//   - Education: Degrees, courses (from education credentials)
//   - Contributions: Open source work (from GitHub credentials)
//   - Vouches: Endorsements from other users
//   - Custom: User-defined sections
//
// Each section has independent visibility settings.
//
// ## Embeddable Widgets
//
// Users can embed verification badges on external sites:
//
//	<!-- Badge widget -->
//	<iframe src="https://proof.dev/embed/@alice/badge/github" />
//
//	<!-- Full profile card -->
//	<iframe src="https://proof.dev/embed/@alice/card" />
//
//	<!-- Verification button -->
//	<a href="https://proof.dev/@alice">
//	    <img src="https://proof.dev/embed/@alice/verify-button" />
//	</a>
//
// # Relationships to Other Contexts
//
//   - Identity: Provides user information (name, avatar, DID).
//     Profile is always linked to an Identity.
//
//   - Credential: Profile reads credentials to generate badges.
//     Listens for credential events to update badges.
//
//   - Presentation: Profile may link to presentations for detailed
//     verification. "View verified details" → Presentation share link.
//
//   - Trust: Profile displays vouches received from other users.
//     Vouch count contributes to profile completeness.
//
//   - Organization: Organizational profiles aggregate member badges.
//     Issuer profiles show credentials they've issued.
//
//   - Ledger: Profile view events projected to Ledger.
//
// # Ports (Interfaces to Other Contexts)
//
//	// IdentityReader retrieves user display information.
//	type IdentityReader interface {
//	    GetUser(ctx context.Context, userID identity.UserID) (*identity.User, error)
//	}
//
//	// CredentialReader retrieves credentials for badge generation.
//	type CredentialReader interface {
//	    GetUserCredentials(ctx context.Context, userID identity.UserID) ([]*credential.Credential, error)
//	}
//
//	// TrustReader retrieves vouches for display.
//	type TrustReader interface {
//	    GetVouchesForUser(ctx context.Context, userID identity.UserID) ([]*trust.Vouch, error)
//	}
//
//	// PresentationReader retrieves share links for "view details" actions.
//	type PresentationReader interface {
//	    GetUserPresentations(ctx context.Context, userID identity.UserID) ([]*presentation.Presentation, error)
//	}
//
// # Example Use Cases
//
// ## Creating a Profile
//
//	cmd := command.CreateProfile{
//	    UserID:      userID,
//	    DisplayName: "Alice Chen",
//	    Headline:    "Senior Software Engineer",
//	    Bio:         "Building the future of verified identity.",
//	    VanitySlug:  "alice", // proof.dev/@alice
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.ProfileID = "prof_abc123"
//	// result.URL = "https://proof.dev/@alice"
//
// ## Configuring Badge Visibility
//
//	cmd := command.UpdateBadgeVisibility{
//	    ProfileID:    profileID,
//	    CredentialID: linkedinCredID,
//	    Visibility:   ConnectionsOnly,
//	}
//
//	err := handler.Handle(ctx, cmd)
//	// LinkedIn badge now only visible to connections
//
// ## Generating an Embed Widget
//
//	query := query.GetEmbedCode{
//	    ProfileID:  profileID,
//	    EmbedType:  BadgeEmbed,
//	    BadgeType:  "github",
//	    Theme:      "dark",
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.HTML = "<iframe src='...' />"
//	// result.URL = "https://proof.dev/embed/@alice/badge/github?theme=dark"
//
// ## Viewing a Public Profile
//
//	query := query.GetPublicProfile{
//	    VanitySlug: "alice",
//	    ViewerID:   viewerUserID, // optional, for connections-only content
//	}
//
//	result, err := handler.Handle(ctx, query)
//	// result.Profile = { displayName, headline, badges, sections, ... }
//	// result.ViewRecorded = true
//
// ## Claiming a Vanity URL
//
//	cmd := command.ClaimVanityURL{
//	    ProfileID: profileID,
//	    Slug:      "alice-chen",
//	}
//
//	result, err := handler.Handle(ctx, cmd)
//	// result.URL = "https://proof.dev/@alice-chen"
//	// Old slug remains as redirect (optional)
//
// # Architecture Notes
//
// Profile follows the standard bounded context structure:
//
//	internal/profile/
//	├── domain/
//	│   ├── profile.go         // Profile aggregate root
//	│   ├── badge.go           // Badge entity
//	│   ├── section.go         // ProfileSection entity
//	│   ├── privacy.go         // PrivacySettings, Visibility value objects
//	│   ├── vanity.go          // VanityURL value object
//	│   ├── embed.go           // Embed configuration value object
//	│   ├── theme.go           // ProfileTheme value object
//	│   ├── ports.go           // Reader interfaces to other contexts
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // ProfileCreated, BadgeUpdated, ProfileViewed, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // CreateProfile, UpdateBadge, ClaimVanityURL, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   ├── query/
//	│   │   ├── queries.go     // GetProfile, GetPublicProfile, GetEmbedCode, etc.
//	│   │   ├── handlers.go    // Query handlers
//	│   │   └── views.go       // Read models (PublicProfileView, BadgeView, etc.)
//	│   └── projection/
//	│       └── badge_projector.go  // Updates badges when credentials change
//	├── infrastructure/
//	│   ├── adapters/
//	│   │   ├── identity.go    // IdentityReader adapter
//	│   │   ├── credential.go  // CredentialReader adapter
//	│   │   ├── trust.go       // TrustReader adapter
//	│   │   └── presentation.go // PresentationReader adapter
//	│   ├── embed/
//	│   │   ├── renderer.go    // Generates embed HTML/images
//	│   │   └── templates/     // Embed widget templates
//	│   ├── avatar/
//	│   │   └── service.go     // Avatar upload/storage
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
//	│           ├── badge_repository.go
//	│           ├── vanity_repository.go
//	│           ├── mapper.go
//	│           └── migrations/
//	├── interface/
//	│   └── http/
//	│       └── v1/
//	│           ├── handler.go
//	│           ├── routes.go
//	│           ├── requests.go
//	│           ├── responses.go
//	│           ├── public.go      // Public profile viewing endpoints
//	│           └── embed.go       // Embed widget endpoints
//	└── provider.go                // Dependency injection setup
//
// # Badge Projection
//
// Badges are projections of credentials, updated via events:
//
//	CredentialIssued event → Create/update badge
//	CredentialRevoked event → Remove or mark badge as revoked
//	CredentialExpired event → Mark badge as expired
//
// This keeps badges in sync without tight coupling to Credential context.
//
// # Events
//
//   - ProfileCreated: User created their profile
//   - ProfileUpdated: Profile metadata changed
//   - BadgeAdded: New badge from credential
//   - BadgeRemoved: Badge removed (credential revoked/expired)
//   - BadgeVisibilityChanged: Badge privacy setting changed
//   - VanityURLClaimed: User claimed a vanity URL
//   - ProfileViewed: Someone viewed the profile (for analytics)
//   - EmbedAccessed: Embed widget was loaded
//
// # Future Considerations
//
//   - Profile verification levels (basic, verified, trusted)
//   - Profile completeness score (gamification)
//   - Custom themes and branding (premium feature)
//   - Profile comparison (side-by-side credential comparison)
//   - Profile search and discovery
//   - Social features (follow, connect)
//   - Profile import/export (data portability)
//   - Organizational profile pages
//   - Profile analytics dashboard
package profile
