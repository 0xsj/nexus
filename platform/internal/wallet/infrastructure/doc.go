// internal/wallet/infrastructure/
// ├── validation/
// │   ├── evm.go              # EVM address validation (FamilyAddressValidator)
// │   ├── solana.go           # Solana address validation (future)
// │   ├── cosmos.go           # Cosmos address validation (future)
// │   └── validator.go        # Composite AddressValidationService
// ├── signature/
// │   ├── siwe.go             # SIWE signature verification (EVM)
// │   ├── solana.go           # Solana signature verification (future)
// │   ├── cosmos.go           # Cosmos signature verification (future)
// │   └── verifier.go         # Composite SignatureVerificationService
// ├── did/
// │   └── derivation.go       # DIDDerivationService implementation
// ├── challenge/
// │   └── service.go          # ChallengeService implementation
// ├── nonce/
// │   └── service.go          # NonceService implementation
// └── persistence/
//
//	└── postgres/
//	    ├── mapper.go       # Domain <-> DB mapping
//	    ├── queries.go      # SQL queries
//	    ├── wallet.go       # WalletRepository implementation
//	    ├── challenge.go    # ChallengeRepository implementation
//	    ├── nonce.go        # NonceRepository implementation
//	    ├── reader.go       # WalletReader implementation (CQRS read side)
//	    └── migrations/
//	        └── 001_create_wallet_tables.sql
package infrastructure
