package command

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/0xsj/nexus/platform/internal/issuer/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Issuer context.
type Handlers struct {
	issuerRepo   domain.IssuerRepository
	templateRepo domain.TemplateRepository
	orgReader    domain.OrganizationReader
	schemaReader domain.SchemaReader
	publisher    domain.EventPublisher
	logger       log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	issuerRepo domain.IssuerRepository,
	templateRepo domain.TemplateRepository,
	orgReader domain.OrganizationReader,
	schemaReader domain.SchemaReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		issuerRepo:   issuerRepo,
		templateRepo: templateRepo,
		orgReader:    orgReader,
		schemaReader: schemaReader,
		publisher:    publisher,
		logger:       logger,
	}
}

// ============================================================================
// RegisterIssuer Handler
// ============================================================================

// HandleRegisterIssuer handles the RegisterIssuer command.
func (h *Handlers) HandleRegisterIssuer(ctx context.Context, cmd RegisterIssuer) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRegisterIssuer"

	// 1. Verify organization exists (non-fatal)
	if err := h.orgReader.GetOrganization(ctx, cmd.OrganizationID); err != nil {
		h.logger.Warn("organization verification skipped",
			log.String("op", op),
			log.String("organization_id", cmd.OrganizationID),
			log.Err(err),
		)
	}

	// 2. Generate issuer ID
	issuerID := domain.NewIssuerID()

	// 3. Create aggregate
	iss, err := domain.RegisterIssuer(issuerID, cmd.OrganizationID, cmd.Name, cmd.Description, cmd.WebhookURL)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save
	if err := h.issuerRepo.Save(ctx, iss); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, iss.Changes()...)

	return &cqrs.CommandResult{
		ID:      issuerID.String(),
		Version: iss.Version(),
		Data: RegisterIssuerResult{
			IssuerID: issuerID.String(),
			Status:   iss.Status().String(),
		},
	}, nil
}

// ============================================================================
// ActivateIssuer Handler
// ============================================================================

// HandleActivateIssuer handles the ActivateIssuer command.
func (h *Handlers) HandleActivateIssuer(ctx context.Context, cmd ActivateIssuer) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleActivateIssuer"

	// 1. Load aggregate
	issuerID := domain.IssuerIDFromTypesID(cmd.IssuerID)
	iss, err := h.issuerRepo.Get(ctx, issuerID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Hash the API key
	apiKeyHash := hashAPIKey(cmd.APIKey)

	// 3. Build branding
	branding := domain.NewIssuerBranding(cmd.LogoURL, cmd.PrimaryColor, cmd.SecondaryColor, cmd.CertificateDesign)

	// 4. Activate
	if err := iss.Activate(cmd.DID, apiKeyHash, branding); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save
	if err := h.issuerRepo.Save(ctx, iss); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, iss.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.IssuerID.String(),
		Version: iss.Version(),
		Data: ActivateIssuerResult{
			IssuerID: cmd.IssuerID.String(),
			Status:   iss.Status().String(),
		},
	}, nil
}

// ============================================================================
// SuspendIssuer Handler
// ============================================================================

// HandleSuspendIssuer handles the SuspendIssuer command.
func (h *Handlers) HandleSuspendIssuer(ctx context.Context, cmd SuspendIssuer) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleSuspendIssuer"

	// 1. Load aggregate
	issuerID := domain.IssuerIDFromTypesID(cmd.IssuerID)
	iss, err := h.issuerRepo.Get(ctx, issuerID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Suspend
	if err := iss.Suspend(cmd.Reason); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.issuerRepo.Save(ctx, iss); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, iss.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.IssuerID.String(),
		Version: iss.Version(),
		Data: SuspendIssuerResult{
			IssuerID: cmd.IssuerID.String(),
			Status:   iss.Status().String(),
		},
	}, nil
}

// ============================================================================
// CreateTemplate Handler
// ============================================================================

// HandleCreateTemplate handles the CreateTemplate command.
func (h *Handlers) HandleCreateTemplate(ctx context.Context, cmd CreateTemplate) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleCreateTemplate"

	// 1. Verify issuer exists and is active
	issuerID := domain.IssuerIDFromTypesID(cmd.IssuerID)
	iss, err := h.issuerRepo.Get(ctx, issuerID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	if !iss.Status().IsActive() {
		return nil, domain.IssuerNotActive(op, issuerID.String())
	}

	// 2. Verify schema exists (non-fatal)
	if err := h.schemaReader.GetSchema(ctx, cmd.SchemaType); err != nil {
		h.logger.Warn("schema verification skipped",
			log.String("op", op),
			log.String("schema_type", cmd.SchemaType),
			log.Err(err),
		)
	}

	// 3. Generate template ID
	templateID := domain.NewTemplateID()

	// 4. Create aggregate
	t, err := domain.CreateTemplate(
		templateID, issuerID,
		cmd.Name, cmd.Description, cmd.SchemaType,
		cmd.ClaimMappings, cmd.DefaultValues,
		cmd.ExpirationDays, cmd.AutoApprove,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Save
	if err := h.templateRepo.Save(ctx, t); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 6. Publish events
	h.publishEvents(ctx, op, t.Changes()...)

	return &cqrs.CommandResult{
		ID:      templateID.String(),
		Version: t.Version(),
		Data: CreateTemplateResult{
			TemplateID: templateID.String(),
			IssuerID:   cmd.IssuerID.String(),
			Status:     t.Status().String(),
		},
	}, nil
}

// ============================================================================
// UpdateTemplate Handler
// ============================================================================

// HandleUpdateTemplate handles the UpdateTemplate command.
func (h *Handlers) HandleUpdateTemplate(ctx context.Context, cmd UpdateTemplate) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdateTemplate"

	// 1. Load aggregate
	templateID := domain.TemplateIDFromTypesID(cmd.TemplateID)
	t, err := h.templateRepo.Get(ctx, templateID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Update
	if err := t.Update(
		cmd.Name, cmd.Description,
		cmd.ClaimMappings, cmd.DefaultValues,
		cmd.ExpirationDays, cmd.AutoApprove,
	); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.templateRepo.Save(ctx, t); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, t.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.TemplateID.String(),
		Version: t.Version(),
		Data: UpdateTemplateResult{
			TemplateID: cmd.TemplateID.String(),
			Version:    t.TemplateVersion(),
			Status:     t.Status().String(),
		},
	}, nil
}

// ============================================================================
// ArchiveTemplate Handler
// ============================================================================

// HandleArchiveTemplate handles the ArchiveTemplate command.
func (h *Handlers) HandleArchiveTemplate(ctx context.Context, cmd ArchiveTemplate) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleArchiveTemplate"

	// 1. Load aggregate
	templateID := domain.TemplateIDFromTypesID(cmd.TemplateID)
	t, err := h.templateRepo.Get(ctx, templateID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Archive
	if err := t.Archive(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save
	if err := h.templateRepo.Save(ctx, t); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, t.Changes()...)

	return &cqrs.CommandResult{
		ID:      cmd.TemplateID.String(),
		Version: t.Version(),
		Data: ArchiveTemplateResult{
			TemplateID: cmd.TemplateID.String(),
			Status:     t.Status().String(),
		},
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// publishEvents publishes domain events (fire-and-forget with logging).
func (h *Handlers) publishEvents(ctx context.Context, op string, events ...eventsourcing.Event) {
	if len(events) == 0 {
		return
	}

	if err := h.publisher.Publish(ctx, events...); err != nil {
		h.logger.Error("failed to publish events",
			log.String("op", op),
			log.Err(err),
		)
	}
}

// hashAPIKey hashes an API key using SHA-256.
func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", hash)
}
