// Package schema provides the bounded context for credential schema management.
//
// # Purpose
//
// Schema is the authoritative source for credential type definitions within Proof.
// It defines what kinds of credentials can be issued, what claims they contain,
// and the validation rules for those claims. This context acts as the foundation
// that other contexts (Integration, Credential, Issuer) reference when working
// with verifiable credentials.
//
// # Core Responsibilities
//
//   - Define credential types (e.g., GitHubContributor, ProfessionalExperience)
//   - Manage claim type definitions with data types and validation rules
//   - Version schemas to support evolution without breaking existing credentials
//   - Provide schema resolution for credential issuance and verification
//   - Support custom schemas created by external issuers
//
// # Key Entities
//
//   - Schema: A versioned definition of a credential type, containing metadata
//     and a collection of claim definitions.
//
//   - ClaimDefinition: Specifies a single claim within a schema, including its
//     key, data type, validation rules, and whether it's required or optional.
//
//   - SchemaVersion: Represents a specific version of a schema, enabling
//     backwards compatibility and migration paths.
//
//   - SchemaRegistry: The aggregate root that manages all registered schemas
//     and handles schema resolution by type and version.
//
// # Domain Concepts
//
// ## Credential Types
//
// Each credential type represents a category of verifiable claim:
//
//   - GitHubContributor: commits, repositories, stars, languages, contributions
//   - ProfessionalExperience: employer, role, tenure, verified via LinkedIn
//   - CourseCompletion: course name, provider, completion date, grade
//   - Certification: issuer, cert type, issue date, expiry date
//
// ## Claim Types
//
// Claims are the individual data points within a credential:
//
//   - Primitive types: string, integer, boolean, date, datetime, duration
//   - Complex types: string[], object, enum
//   - Semantic types: email, url, did, address (with validation)
//
// ## Schema Versioning
//
// Schemas evolve over time. Versioning strategy:
//
//   - MAJOR: Breaking changes (removed required claims, type changes)
//   - MINOR: Backwards-compatible additions (new optional claims)
//   - Credentials reference the schema version they were issued against
//   - Verifiers can specify acceptable schema versions
//
// # Relationships to Other Contexts
//
//   - Integration: Reads schema definitions to know what data to fetch from
//     external providers and how to map provider-specific fields to claims.
//
//   - Credential: References schemas when issuing credentials. Validates that
//     credential claims conform to the schema definition.
//
//   - Issuer: External issuers can define custom schemas for their credential
//     types, registered and versioned through this context.
//
//   - Presentation: Uses schema metadata for selective disclosure hints
//     (which claims can be disclosed independently).
//
//   - Verification: Schema defines what claims are expected for each
//     credential type, guiding the verification flow.
//
// # Example Use Cases
//
// ## Registering a Built-in Schema
//
// Proof ships with predefined schemas for supported integrations:
//
//	schema := NewSchema(
//		"GitHubContributor",
//		"Verified GitHub contribution metrics",
//		WithClaim("username", StringType, Required),
//		WithClaim("commits", IntegerType, Required),
//		WithClaim("repositories", IntegerType, Required),
//		WithClaim("stars", IntegerType, Optional),
//		WithClaim("languages", StringArrayType, Optional),
//		WithClaim("contributionYears", IntegerType, Optional),
//	)
//	registry.Register(schema)
//
// ## Resolving a Schema for Credential Issuance
//
// When issuing a credential, the Credential context resolves the schema:
//
//	schema, err := registry.Resolve("GitHubContributor", "1.0")
//	if err != nil {
//		return err
//	}
//	// Validate claims against schema before signing
//	if err := schema.ValidateClaims(claims); err != nil {
//		return err
//	}
//
// ## Custom Issuer Schema
//
// External issuers can define their own credential types:
//
//	customSchema := NewSchema(
//		"AcmeCertification",
//		"Acme Corp professional certification",
//		WithIssuer(acmeOrgID),
//		WithClaim("certificationLevel", EnumType("bronze", "silver", "gold"), Required),
//		WithClaim("examScore", IntegerType, Optional),
//		WithClaim("validUntil", DateType, Required),
//	)
//	registry.RegisterCustom(customSchema)
//
// # Architecture Notes
//
// Schema follows the standard bounded context structure:
//
//	internal/schema/
//	├── domain/
//	│   ├── schema.go          // Schema aggregate root
//	│   ├── claim.go           // ClaimDefinition entity
//	│   ├── registry.go        // SchemaRegistry aggregate
//	│   ├── types.go           // ClaimType, DataType value objects
//	│   ├── version.go         // SchemaVersion value object
//	│   ├── errors.go          // Domain errors
//	│   ├── events.go          // SchemaRegistered, SchemaDeprecated, etc.
//	│   └── repository.go      // Repository interface
//	├── application/
//	│   ├── command/
//	│   │   ├── commands.go    // RegisterSchema, DeprecateSchema, etc.
//	│   │   └── handlers.go    // Command handlers
//	│   └── query/
//	│       ├── queries.go     // GetSchema, ListSchemas, etc.
//	│       ├── handlers.go    // Query handlers
//	│       └── views.go       // Read models
//	├── infrastructure/
//	│   └── persistence/
//	│       └── postgres/
//	│           ├── repository.go
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
// # Built-in Schema Types
//
// Proof provides schemas for all supported integrations:
//
//	| Schema Type            | Provider   | Key Claims                              |
//	|------------------------|------------|-----------------------------------------|
//	| GitHubContributor      | GitHub     | commits, repos, stars, languages        |
//	| ProfessionalExperience | LinkedIn   | employer, role, tenure, industry        |
//	| CourseCompletion       | Coursera   | course, provider, date, grade           |
//	| CloudCertification     | AWS/GCP    | cert type, level, issue date, expiry    |
//	| FreelanceReputation    | Upwork     | jobs completed, success rate, earnings  |
//
// # Future Considerations
//
//   - JSON Schema / JSON-LD compatibility for interoperability
//   - Schema inheritance (base schemas extended by specific types)
//   - Claim-level selective disclosure hints for BBS+ signatures
//   - Schema discovery endpoint for external verifiers
//   - Deprecation workflow with migration guidance
package schema
