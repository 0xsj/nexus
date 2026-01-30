// Package internal contains the bounded contexts for the Proof platform.
//
// # Overview
//
// Proof is a verified identity platform that transforms self-reported claims
// into cryptographically verifiable credentials. It bridges Web2 verification
// sources (GitHub, LinkedIn, Coursera) with Web3 identity standards (DIDs,
// Verifiable Credentials), giving users portable, self-sovereign credentials
// they own and control.
//
// # Architecture
//
// The platform follows Domain-Driven Design (DDD) with CQRS and Event Sourcing
// patterns. Each bounded context is a self-contained module with:
//
//	context/
//	├── domain/          // Aggregates, entities, value objects, domain events
//	├── application/     // Commands, queries, handlers
//	├── infrastructure/  // Persistence, external services, adapters
//	├── interface/       // HTTP handlers, routes
//	└── provider.go      // Dependency injection
//
// Contexts communicate via:
//   - Ports (interfaces) for synchronous queries
//   - Domain events for asynchronous reactions
//   - Shared kernel (pkg/) for common primitives
//
// # Bounded Context Map
//
//	┌─────────────────────────────────────────────────────────────────────────────┐
//	│                              PROOF PLATFORM                                  │
//	├─────────────────────────────────────────────────────────────────────────────┤
//	│                                                                             │
//	│  ┌─────────────────────────────────────────────────────────────────────┐   │
//	│  │                     CORE IDENTITY & AUTH                             │   │
//	│  │  ┌──────────────┐    ┌──────────────┐                               │   │
//	│  │  │   Identity   │◄──►│    Wallet    │                               │   │
//	│  │  │              │    │              │                               │   │
//	│  │  │ Users, Auth  │    │ SIWE, Chain  │                               │   │
//	│  │  │ Sessions     │    │ DID Derivation│                              │   │
//	│  │  └──────────────┘    └──────────────┘                               │   │
//	│  └─────────────────────────────────────────────────────────────────────┘   │
//	│                                    │                                        │
//	│                                    ▼                                        │
//	│  ┌─────────────────────────────────────────────────────────────────────┐   │
//	│  │                      CREDENTIAL LIFECYCLE                            │   │
//	│  │                                                                      │   │
//	│  │  ┌──────────┐    ┌─────────────┐    ┌────────────┐    ┌──────────┐ │   │
//	│  │  │  Schema  │◄───│ Integration │◄───│Verification│───►│Credential│ │   │
//	│  │  │          │    │             │    │            │    │          │ │   │
//	│  │  │  Types   │    │  Providers  │    │   OAuth    │    │   VCs    │ │   │
//	│  │  │  Claims  │    │  Adapters   │    │   Flow     │    │  Issue   │ │   │
//	│  │  └──────────┘    └─────────────┘    └────────────┘    └────┬─────┘ │   │
//	│  │                                                             │      │   │
//	│  │                                                             ▼      │   │
//	│  │                                                      ┌────────────┐│   │
//	│  │                                                      │Presentation││   │
//	│  │                                                      │            ││   │
//	│  │                                                      │ Share, VP  ││   │
//	│  │                                                      │ QR, Links  ││   │
//	│  │                                                      └────────────┘│   │
//	│  └─────────────────────────────────────────────────────────────────────┘   │
//	│                                    │                                        │
//	│                                    ▼                                        │
//	│  ┌─────────────────────────────────────────────────────────────────────┐   │
//	│  │                        PUBLIC & SOCIAL                               │   │
//	│  │  ┌──────────────┐    ┌──────────────┐                               │   │
//	│  │  │   Profile    │◄──►│    Trust     │                               │   │
//	│  │  │              │    │              │                               │   │
//	│  │  │ Public Page  │    │   Vouches    │                               │   │
//	│  │  │ Badges       │    │  Reputation  │                               │   │
//	│  │  └──────────────┘    └──────────────┘                               │   │
//	│  └─────────────────────────────────────────────────────────────────────┘   │
//	│                                    │                                        │
//	│                                    ▼                                        │
//	│  ┌─────────────────────────────────────────────────────────────────────┐   │
//	│  │                       B2B & MULTI-TENANCY                            │   │
//	│  │  ┌──────────────┐    ┌──────────────┐                               │   │
//	│  │  │ Organization │◄───│    Issuer    │                               │   │
//	│  │  │              │    │              │                               │   │
//	│  │  │ Teams, Roles │    │  B2B Issue   │                               │   │
//	│  │  │ Membership   │    │  Templates   │                               │   │
//	│  │  └──────────────┘    └──────────────┘                               │   │
//	│  └─────────────────────────────────────────────────────────────────────┘   │
//	│                                    │                                        │
//	│                                    ▼                                        │
//	│  ┌─────────────────────────────────────────────────────────────────────┐   │
//	│  │                     PLATFORM INFRASTRUCTURE                          │   │
//	│  │  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐          │   │
//	│  │  │    Ledger    │    │ Notification │    │   Billing    │          │   │
//	│  │  │              │    │              │    │              │          │   │
//	│  │  │ Audit Trail  │    │ Email, Push  │    │   (Future)   │          │   │
//	│  │  │ Event Log    │    │   In-App     │    │              │          │   │
//	│  │  └──────────────┘    └──────────────┘    └──────────────┘          │   │
//	│  └─────────────────────────────────────────────────────────────────────┘   │
//	│                                                                             │
//	└─────────────────────────────────────────────────────────────────────────────┘
//
// # Bounded Contexts
//
// ## Core Identity & Auth
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ Identity    │ User accounts, authentication (magic link, OAuth),         │
//	│             │ sessions, API keys, DID management. The foundation of      │
//	│             │ user identity within the platform.                         │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Wallet      │ Multi-chain wallet linking, SIWE signature verification,   │
//	│             │ DID derivation (did:pkh). Bridges Web3 wallets to Proof.   │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// ## Credential Lifecycle
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ Schema      │ Credential type definitions, claim registry, versioning.   │
//	│             │ Foundational context that defines what credentials exist.  │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Integration │ Provider adapters (GitHub, LinkedIn, etc.), data fetching, │
//	│             │ normalization. Bridges external platforms to Proof.        │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Verification│ OAuth flow orchestration, coordinates Integration and      │
//	│             │ Credential contexts. Thin orchestration layer.             │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Credential  │ Verifiable Credential issuance, storage, signing,          │
//	│             │ revocation. Core VC lifecycle management.                  │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Presentation│ Verifiable Presentations, share links, selective           │
//	│             │ disclosure, QR codes. Controls how credentials are shared. │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// ## Public & Social
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ Profile     │ Public profile pages, badge display, privacy controls,     │
//	│             │ embeddable widgets. The user's public identity storefront. │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Trust       │ Peer vouches, endorsements, reputation scoring, trust      │
//	│             │ graph. Adds social proof layer to cryptographic proofs.    │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// ## B2B & Multi-tenancy
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ Organization│ Teams, membership, roles, KYB verification. Enables        │
//	│             │ multi-user collaboration and organizational identity.      │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Issuer      │ B2B credential issuance, templates, branding, batch        │
//	│             │ operations. Transforms Proof into an issuance platform.    │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// ## Platform Infrastructure
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ Ledger      │ Immutable audit trail, event projection, compliance        │
//	│             │ logging. Cross-domain activity history.                    │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Notification│ Email, push, in-app notifications. Cross-cutting           │
//	│             │ communications triggered by domain events.                 │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ Billing     │ Subscription tiers, usage metering, payments.              │
//	│             │ (Future phase)                                             │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// # Context Dependencies
//
//	                          Schema
//	                             │
//	                             ▼
//	                       Integration
//	                             │
//	                             ▼
//	Identity ◄─────────► Verification ─────────► Credential
//	    │                                            │
//	    │                                            ▼
//	    │                                      Presentation
//	    │                                            │
//	    ▼                                            ▼
//	 Wallet                                      Profile ◄──► Trust
//	                                                │
//	                                                ▼
//	                         Organization ◄───── Issuer
//
//	                   ┌─────────────────────────────┐
//	                   │  Ledger     Notification    │  (Cross-cutting)
//	                   │  (subscribes to all events) │
//	                   └─────────────────────────────┘
//
// # Shared Kernel (pkg/)
//
// Common primitives shared across bounded contexts:
//
//	┌─────────────┬────────────────────────────────────────────────────────────┐
//	│ vc          │ W3C Verifiable Credentials data model, JWT encoding        │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ presentation│ W3C Verifiable Presentations data model, JWT encoding      │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ did         │ DID core, did:key and did:pkh methods, resolution          │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ crypto      │ Ed25519, secp256k1, BBS+ signing algorithms                │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ siwe        │ Sign-In with Ethereum message parsing and verification     │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ jwt         │ JWT primitives (claims, headers, encoding)                 │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ eventsourcing│ Aggregate, event, store interfaces                        │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ cqrs        │ Command and query bus                                      │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ errors      │ Domain error types with codes and severity                 │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ types       │ Common value types (Email, ID, Timestamp, URL)             │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ database    │ PostgreSQL adapter, transactions                           │
//	├─────────────┼────────────────────────────────────────────────────────────┤
//	│ http        │ Server, middleware, request/response helpers               │
//	└─────────────┴────────────────────────────────────────────────────────────┘
//
// # Implementation Status
//
//	┌─────────────┬──────────┬─────────────────────────────────────────────────┐
//	│ Context     │ Status   │ Notes                                           │
//	├─────────────┼──────────┼─────────────────────────────────────────────────┤
//	│ Identity    │ ✅ Done  │ Full implementation with tests                  │
//	│ Wallet      │ ✅ Done  │ Full implementation with tests                  │
//	│ Verification│ ✅ Done  │ Needs refactoring (extract Integration)         │
//	│ Credential  │ ✅ Done  │ Full implementation with tests                  │
//	├─────────────┼──────────┼─────────────────────────────────────────────────┤
//	│ Schema      │ 🔲 New   │ Extract from Credential, add versioning         │
//	│ Integration │ 🔲 New   │ Extract from Verification                       │
//	│ Presentation│ 🔲 New   │ pkg/presentation exists, need domain layer      │
//	│ Profile     │ 🔲 New   │ Public profiles, badges, embeds                 │
//	│ Trust       │ 🔲 New   │ Vouches, reputation                             │
//	│ Organization│ 🔲 New   │ Teams, membership                               │
//	│ Issuer      │ 🔲 New   │ B2B credential issuance                         │
//	│ Ledger      │ 🔲 New   │ Audit trail, event projection                   │
//	│ Notification│ 🔲 New   │ Email, push, in-app                             │
//	│ Billing     │ 🔲 Future│ Subscriptions, payments                         │
//	└─────────────┴──────────┴─────────────────────────────────────────────────┘
//
// # Event Flow
//
// Domain events flow through the system:
//
//	┌──────────────┐     ┌──────────────┐     ┌──────────────┐
//	│ Verification │────►│  Credential  │────►│   Ledger     │
//	│   Context    │     │   Context    │     │  (audit)     │
//	└──────────────┘     └──────────────┘     └──────────────┘
//	                            │                    │
//	                            ▼                    │
//	                     ┌──────────────┐            │
//	                     │   Profile    │◄───────────┘
//	                     │  (badges)    │
//	                     └──────────────┘
//	                            │
//	                            ▼
//	                     ┌──────────────┐
//	                     │ Notification │
//	                     │  (alerts)    │
//	                     └──────────────┘
//
// Example flow for credential issuance:
//
//  1. User initiates GitHub verification
//  2. Verification orchestrates OAuth flow
//  3. Verification calls Integration to fetch GitHub data
//  4. Verification calls Credential to issue VC
//  5. Credential emits CredentialIssued event
//  6. Ledger projects event to audit trail
//  7. Profile updates badges
//  8. Notification sends email to user
//
// # Key Differentiators
//
//	┌─────────────────────┬─────────────────────────────────────────────────────┐
//	│ Traditional Profiles│ Proof                                               │
//	├─────────────────────┼─────────────────────────────────────────────────────┤
//	│ Self-reported claims│ Verified via source systems                         │
//	│ Platform-locked     │ Portable credentials (W3C VCs)                      │
//	│ All-or-nothing      │ Selective disclosure                                │
//	│ Trust the platform  │ Trust the cryptography                              │
//	│ Profile dies        │ Credentials travel with you                         │
//	└─────────────────────┴─────────────────────────────────────────────────────┘
//
// # Getting Started
//
// To work on a specific bounded context:
//
//  1. Read the context's doc.go for full documentation
//  2. Understand the domain model (domain/ package)
//  3. Review existing patterns in Identity or Credential contexts
//  4. Follow leaf-first development (errors, types, values → entities → aggregates)
//  5. Write tests alongside implementation
//
// Each context follows identical structure and patterns. See the Identity
// context as the reference implementation.
package internal
