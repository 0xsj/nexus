package command

import (
	"context"

	"github.com/0xsj/nexus/platform/internal/schema/domain"
	"github.com/0xsj/nexus/platform/pkg/cqrs"
	pkgerrors "github.com/0xsj/nexus/platform/pkg/errors"
	"github.com/0xsj/nexus/platform/pkg/eventsourcing"
	"github.com/0xsj/nexus/platform/pkg/observability/log"
)

// ============================================================================
// Handlers
// ============================================================================

// Handlers contains all command handlers for the Schema context.
type Handlers struct {
	schemaRepo   domain.SchemaRepository
	schemaLookup domain.SchemaLookup
	issuerReader domain.IssuerReader
	publisher    domain.EventPublisher
	logger       log.Logger
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	schemaRepo domain.SchemaRepository,
	schemaLookup domain.SchemaLookup,
	issuerReader domain.IssuerReader,
	publisher domain.EventPublisher,
	logger log.Logger,
) *Handlers {
	return &Handlers{
		schemaRepo:   schemaRepo,
		schemaLookup: schemaLookup,
		issuerReader: issuerReader,
		publisher:    publisher,
		logger:       logger,
	}
}

// ============================================================================
// RegisterSchema Handler
// ============================================================================

// HandleRegisterSchema handles the RegisterSchema command.
func (h *Handlers) HandleRegisterSchema(ctx context.Context, cmd RegisterSchema) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleRegisterSchema"

	// 1. Check if schema type already exists
	exists, err := h.schemaLookup.ExistsByType(ctx, cmd.SchemaType)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}
	if exists {
		return nil, domain.SchemaAlreadyExists(op, cmd.SchemaType)
	}

	// 2. If custom schema, validate issuer exists and is active
	if cmd.IssuerID != nil {
		issuerInfo, err := h.issuerReader.GetIssuer(ctx, *cmd.IssuerID)
		if err != nil {
			return nil, domain.IssuerNotFound(op, cmd.IssuerID.String())
		}
		if !issuerInfo.Active {
			return nil, domain.IssuerInactive(op, cmd.IssuerID.String())
		}
	}

	// 3. Create schema aggregate
	schema, err := domain.RegisterSchema(
		cmd.SchemaID,
		cmd.SchemaType,
		cmd.Name,
		cmd.Description,
		cmd.Version,
		cmd.Claims,
		cmd.IssuerID,
	)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save schema
	if err := h.schemaRepo.Save(ctx, schema); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, schema.Changes())

	return &cqrs.CommandResult{
		ID:      cmd.SchemaID.String(),
		Version: schema.Version(),
		Data: RegisterSchemaResult{
			SchemaID:   cmd.SchemaID.String(),
			SchemaType: cmd.SchemaType,
			Version:    cmd.Version.String(),
		},
	}, nil
}

// ============================================================================
// AddSchemaVersion Handler
// ============================================================================

// HandleAddSchemaVersion handles the AddSchemaVersion command.
func (h *Handlers) HandleAddSchemaVersion(ctx context.Context, cmd AddSchemaVersion) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleAddSchemaVersion"

	// 1. Load schema
	schema, err := h.schemaRepo.GetByID(ctx, cmd.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Store previous version for result
	previousVersion := schema.CurrentVersion()

	// 3. Add version
	if err := schema.AddVersion(cmd.Version, cmd.Claims, cmd.ChangeSummary); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save schema
	if err := h.schemaRepo.Save(ctx, schema); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, schema.Changes())

	return &cqrs.CommandResult{
		ID:      cmd.SchemaID.String(),
		Version: schema.Version(),
		Data: AddSchemaVersionResult{
			SchemaID:        cmd.SchemaID.String(),
			Version:         cmd.Version.String(),
			PreviousVersion: previousVersion.String(),
		},
	}, nil
}

// ============================================================================
// DeprecateSchema Handler
// ============================================================================

// HandleDeprecateSchema handles the DeprecateSchema command.
func (h *Handlers) HandleDeprecateSchema(ctx context.Context, cmd DeprecateSchema) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleDeprecateSchema"

	// 1. Load schema
	schema, err := h.schemaRepo.GetByID(ctx, cmd.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. If replacement schema is specified, validate it exists
	if cmd.ReplacementSchemaID != nil {
		exists, err := h.schemaRepo.Exists(ctx, *cmd.ReplacementSchemaID)
		if err != nil {
			return nil, pkgerrors.Wrap(err, op)
		}
		if !exists {
			return nil, domain.SchemaNotFound(op, cmd.ReplacementSchemaID.String())
		}
	}

	// 3. Deprecate schema
	if err := schema.Deprecate(cmd.Reason, cmd.ReplacementSchemaID); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Save schema
	if err := h.schemaRepo.Save(ctx, schema); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 5. Publish events
	h.publishEvents(ctx, op, schema.Changes())

	replacementID := ""
	if cmd.ReplacementSchemaID != nil {
		replacementID = cmd.ReplacementSchemaID.String()
	}

	return &cqrs.CommandResult{
		ID:      cmd.SchemaID.String(),
		Version: schema.Version(),
		Data: DeprecateSchemaResult{
			SchemaID:            cmd.SchemaID.String(),
			ReplacementSchemaID: replacementID,
		},
	}, nil
}

// ============================================================================
// ActivateSchema Handler
// ============================================================================

// HandleActivateSchema handles the ActivateSchema command.
func (h *Handlers) HandleActivateSchema(ctx context.Context, cmd ActivateSchema) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleActivateSchema"

	// 1. Load schema
	schema, err := h.schemaRepo.GetByID(ctx, cmd.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Activate schema
	if err := schema.Activate(); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save schema
	if err := h.schemaRepo.Save(ctx, schema); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, schema.Changes())

	return &cqrs.CommandResult{
		ID:      cmd.SchemaID.String(),
		Version: schema.Version(),
		Data: ActivateSchemaResult{
			SchemaID: cmd.SchemaID.String(),
		},
	}, nil
}

// ============================================================================
// UpdateSchemaMetadata Handler
// ============================================================================

// HandleUpdateSchemaMetadata handles the UpdateSchemaMetadata command.
func (h *Handlers) HandleUpdateSchemaMetadata(ctx context.Context, cmd UpdateSchemaMetadata) (*cqrs.CommandResult, error) {
	const op = "Handlers.HandleUpdateSchemaMetadata"

	// 1. Load schema
	schema, err := h.schemaRepo.GetByID(ctx, cmd.SchemaID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 2. Update metadata
	if err := schema.UpdateMetadata(cmd.Name, cmd.Description); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 3. Save schema
	if err := h.schemaRepo.Save(ctx, schema); err != nil {
		return nil, pkgerrors.Wrap(err, op)
	}

	// 4. Publish events
	h.publishEvents(ctx, op, schema.Changes())

	return &cqrs.CommandResult{
		ID:      cmd.SchemaID.String(),
		Version: schema.Version(),
		Data: UpdateSchemaMetadataResult{
			SchemaID:    cmd.SchemaID.String(),
			Name:        cmd.Name,
			Description: cmd.Description,
		},
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// publishEvents publishes domain events (fire-and-forget with logging).
func (h *Handlers) publishEvents(ctx context.Context, op string, events []eventsourcing.Event) {
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

// ============================================================================
// Command Registration
// ============================================================================

// RegisterCommands registers all Schema command handlers with the command bus.
func RegisterCommands(bus *cqrs.InMemoryCommandBus, handlers *Handlers) error {
	registrations := []struct {
		name    string
		handler any
	}{
		{CommandRegisterSchema, handlers},
		{CommandAddSchemaVersion, handlers},
		{CommandDeprecateSchema, handlers},
		{CommandActivateSchema, handlers},
		{CommandUpdateSchemaMetadata, handlers},
	}

	for _, r := range registrations {
		if err := bus.Register(r.name, r.handler); err != nil {
			return err
		}
	}

	return nil
}
