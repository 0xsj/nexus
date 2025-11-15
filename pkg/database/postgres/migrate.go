// Package postgres provides PostgreSQL database connection and utilities.
package postgres

import (
	"embed"
	"errors"
	"fmt"

	"github.com/0xsj/nexus/pkg/observability/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Embed migrations directory into binary
// This allows migrations to run automatically in production
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs all pending migrations.
// This is safe to call multiple times - it will only apply new migrations.
//
// Example:
//
//	db, _ := postgres.New(ctx, cfg, logger)
//	if err := db.Migrate(); err != nil {
//	    log.Fatal(err)
//	}
func (db *DB) Migrate() error {
	return db.MigrateWithOptions(MigrateOptions{
		Direction: MigrateUp,
		Steps:     0, // 0 means all pending migrations
	})
}

// MigrateDirection represents the direction of migration
type MigrateDirection string

const (
	MigrateUp   MigrateDirection = "up"
	MigrateDown MigrateDirection = "down"
)

// MigrateOptions configures migration behavior
type MigrateOptions struct {
	Direction MigrateDirection
	Steps     int  // Number of steps to migrate (0 = all)
	Force     bool // Force a specific version (dangerous!)
}

// MigrateWithOptions runs migrations with custom options.
// This provides more control over migration behavior.
//
// Example:
//
//	// Apply all pending migrations
//	db.MigrateWithOptions(MigrateOptions{Direction: MigrateUp})
//
//	// Rollback last migration
//	db.MigrateWithOptions(MigrateOptions{Direction: MigrateDown, Steps: 1})
func (db *DB) MigrateWithOptions(opts MigrateOptions) error {
	m, err := db.createMigrator()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Get current version before migration
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d, manual intervention required", version)
	}

	// Log current state
	if errors.Is(err, migrate.ErrNilVersion) {
		db.logger.Info("No migrations applied yet, starting fresh")
	} else {
		db.logger.Info("Current migration version",
			logger.Uint("version", version),
		)
	}

	// Execute migration based on direction
	switch opts.Direction {
	case MigrateUp:
		if opts.Steps > 0 {
			err = m.Steps(opts.Steps)
		} else {
			err = m.Up()
		}
	case MigrateDown:
		if opts.Steps > 0 {
			err = m.Steps(-opts.Steps)
		} else {
			return fmt.Errorf("down migration requires explicit steps count")
		}
	default:
		return fmt.Errorf("invalid migration direction: %s", opts.Direction)
	}

	// Handle migration result
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			db.logger.Info("No migrations to apply, database is up to date")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Log success
	newVersion, _, _ := m.Version()
	db.logger.Info("Migrations applied successfully",
		logger.Uint("from_version", version),
		logger.Uint("to_version", newVersion),
	)

	return nil
}

// MigrationVersion returns the current migration version.
// Returns 0 if no migrations have been applied.
func (db *DB) MigrationVersion() (uint, bool, error) {
	m, err := db.createMigrator()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, err
	}

	return version, dirty, nil
}

// createMigrator creates a new migrate instance
func (db *DB) createMigrator() (*migrate.Migrate, error) {
	// Create migration source from embedded files
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database URL
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.config.Username,
		db.config.Password,
		db.config.Host,
		db.config.Port,
		db.config.Database,
		db.config.SSLMode,
	)

	// Create migrator
	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

// ForceMigrationVersion forces the migration version to a specific value.
// ⚠️  WARNING: This is dangerous and should only be used to recover from a dirty state.
//
// Example:
//
//	// Force version to 5 (use with caution!)
//	db.ForceMigrationVersion(5)
func (db *DB) ForceMigrationVersion(version uint) error {
	m, err := db.createMigrator()
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Force(int(version)); err != nil {
		return fmt.Errorf("failed to force version %d: %w", version, err)
	}

	db.logger.Warn("Forced migration version",
		logger.Uint("version", version),
		logger.String("warning", "This bypasses normal migration checks"),
	)

	return nil
}
