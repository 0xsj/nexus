package credential

import (
	"context"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/verification/domain"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Credential Issuer Service
// ============================================================================

// IssuerService implements domain.CredentialIssuerService.
// It bridges the verification module to the credential module for issuance.
type IssuerService struct {
	credentialService CredentialService
	issuerDID         string
	logger            log.Logger
}

// CredentialService is the interface to the credential module.
// This allows the verification module to issue credentials without
// directly depending on the credential module's implementation.
type CredentialService interface {
	// IssueCredential issues a new verifiable credential.
	IssueCredential(ctx context.Context, params IssueCredentialParams) (*IssuedCredential, error)
}

// IssueCredentialParams contains parameters for credential issuance.
type IssueCredentialParams struct {
	// IssuerDID is the DID of the credential issuer.
	IssuerDID string

	// HolderDID is the DID of the credential holder.
	HolderDID string

	// CredentialType is the type of credential to issue.
	CredentialType string

	// Claims are the credential claims/attributes.
	Claims map[string]any

	// Evidence is the verification evidence.
	Evidence []EvidenceParam

	// ExpiresIn is how long until the credential expires in seconds (optional).
	ExpiresIn *int64

	// Metadata is additional metadata for the credential.
	Metadata map[string]any
}

// EvidenceParam represents evidence for credential issuance.
type EvidenceParam struct {
	Type       string
	Source     string
	VerifiedAt string
	Data       map[string]any
}

// IssuedCredential represents a successfully issued credential.
type IssuedCredential struct {
	// ID is the credential identifier.
	ID string

	// Type is the credential type.
	Type string

	// SignedVC is the signed verifiable credential.
	SignedVC string

	// IssuedAt is when the credential was issued.
	IssuedAt string

	// ExpiresAt is when the credential expires.
	ExpiresAt *string
}

// NewIssuerService creates a new IssuerService.
func NewIssuerService(credentialService CredentialService, issuerDID string, logger log.Logger) *IssuerService {
	return &IssuerService{
		credentialService: credentialService,
		issuerDID:         issuerDID,
		logger:            logger,
	}
}

// ============================================================================
// Domain Interface Implementation
// ============================================================================

// IssueCredential issues a credential based on verification data.
func (s *IssuerService) IssueCredential(ctx context.Context, params domain.IssueCredentialParams) (*domain.IssuedCredential, error) {
	const op = "IssuerService.IssueCredential"

	// Validate required fields
	if params.HolderDID == "" {
		return nil, fmt.Errorf("%s: holder DID is required", op)
	}
	if params.CredentialType == "" {
		return nil, fmt.Errorf("%s: credential type is required", op)
	}
	if len(params.Claims) == 0 {
		return nil, fmt.Errorf("%s: claims are required", op)
	}

	// Build evidence params
	evidence := make([]EvidenceParam, len(params.Evidence))
	for i, e := range params.Evidence {
		evidence[i] = EvidenceParam{
			Type:       e.Type,
			Source:     e.Source,
			VerifiedAt: e.VerifiedAt,
			Data:       e.Data,
		}
	}

	// Issue credential through credential service
	issueParams := IssueCredentialParams{
		IssuerDID:      s.issuerDID,
		HolderDID:      params.HolderDID,
		CredentialType: string(params.CredentialType),
		Claims:         params.Claims,
		Evidence:       evidence,
		ExpiresIn:      params.ExpiresIn,
	}

	issued, err := s.credentialService.IssueCredential(ctx, issueParams)
	if err != nil {
		s.logger.Error("failed to issue credential",
			log.String("holder_did", params.HolderDID),
			log.String("credential_type", string(params.CredentialType)),
			log.Err(err),
		)
		return nil, domain.ErrCredentialIssuanceFailed(op, "credential service error", err)
	}

	s.logger.Info("credential issued successfully",
		log.String("credential_id", issued.ID),
		log.String("holder_did", params.HolderDID),
		log.String("credential_type", string(params.CredentialType)),
	)

	return &domain.IssuedCredential{
		ID:             issued.ID,
		CredentialType: domain.CredentialType(issued.Type),
		SignedVC:       issued.SignedVC,
		IssuedAt:       issued.IssuedAt,
		ExpiresAt:      issued.ExpiresAt,
	}, nil
}

// GetIssuerDID returns the DID used for issuing credentials.
func (s *IssuerService) GetIssuerDID(ctx context.Context) (string, error) {
	return s.issuerDID, nil
}

// ============================================================================
// Interface Compliance
// ============================================================================

var _ domain.CredentialIssuerService = (*IssuerService)(nil)
