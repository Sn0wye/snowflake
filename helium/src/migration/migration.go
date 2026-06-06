package migration

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Runner struct {
	sqlDB         *sql.DB
	migrationsDir string
}

func NewRunner(sqlDB *sql.DB, migrationsDir string) *Runner {
	return &Runner{
		sqlDB:         sqlDB,
		migrationsDir: migrationsDir,
	}
}

func (r *Runner) Up() error {
	m, err := r.newMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	log.Println("Migrations applied successfully")
	return nil
}

func (r *Runner) Down() error {
	m, err := r.newMigrate()
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down failed: %w", err)
	}

	log.Println("Migration rolled back successfully")
	return nil
}

func (r *Runner) newMigrate() (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(r.sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", r.migrationsDir),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}
