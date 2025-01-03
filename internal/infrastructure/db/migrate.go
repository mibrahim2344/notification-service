package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// MigrationConfig holds configuration for database migrations
type MigrationConfig struct {
	MigrationsPath string // Path to migration files
	DBConfig       PostgresConfig
}

// MigrationManager handles database migrations
type MigrationManager struct {
	migrate *migrate.Migrate
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(config MigrationConfig) (*MigrationManager, error) {
	// Create database instance
	db, err := NewPostgresDB(config.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Initialize migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.MigrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	return &MigrationManager{
		migrate: m,
	}, nil
}

// Up runs all available migrations
func (m *MigrationManager) Up() error {
	err := m.migrate.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (m *MigrationManager) Down() error {
	err := m.migrate.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	return nil
}

// Version returns the current migration version
func (m *MigrationManager) Version() (uint, bool, error) {
	return m.migrate.Version()
}

// Force forces the migration version
func (m *MigrationManager) Force(version int) error {
	err := m.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}
	return nil
}

// Steps runs n migrations up or down
func (m *MigrationManager) Steps(n int) error {
	err := m.migrate.Steps(n)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}
