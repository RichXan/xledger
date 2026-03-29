package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func ApplyMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}
	if strings.TrimSpace(dir) == "" {
		dir = "migrations"
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		if strings.Contains(name, ".down.sql") {
			continue
		}
		files = append(files, name)
	}
	sort.Strings(files)

	for _, name := range files {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, name).Scan(&exists); err != nil {
			return fmt.Errorf("query migration state for %s: %w", name, err)
		}
		if exists {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(filename, applied_at) VALUES($1, $2)`, name, time.Now().UTC()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
		log.Printf("Applied migration: %s", name)
	}

	return nil
}

func RollbackMigrations(ctx context.Context, db *sql.DB, dir string, steps int) (rolledBack int, err error) {
	if db == nil {
		return 0, fmt.Errorf("database is nil")
	}
	if strings.TrimSpace(dir) == "" {
		dir = "migrations"
	}

	rows, err := db.QueryContext(ctx, `
		SELECT filename FROM schema_migrations 
		ORDER BY applied_at DESC, filename DESC
		LIMIT $1
	`, steps)
	if err != nil {
		return 0, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	var appliedMigrations []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return 0, fmt.Errorf("scan migration filename: %w", err)
		}
		appliedMigrations = append(appliedMigrations, filename)
	}

	if len(appliedMigrations) == 0 {
		log.Println("No migrations to rollback")
		return 0, nil
	}

	for _, filename := range appliedMigrations {
		downFilename := strings.Replace(filename, ".sql", ".down.sql", 1)
		downPath := filepath.Join(dir, downFilename)

		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			log.Printf("Warning: No down migration file for %s, skipping rollback", filename)
			continue
		}

		body, err := os.ReadFile(downPath)
		if err != nil {
			return rolledBack, fmt.Errorf("read down migration %s: %w", downFilename, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return rolledBack, fmt.Errorf("begin tx for rollback %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return rolledBack, fmt.Errorf("rollback migration %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE filename = $1`, filename); err != nil {
			_ = tx.Rollback()
			return rolledBack, fmt.Errorf("remove migration record %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return rolledBack, fmt.Errorf("commit rollback %s: %w", filename, err)
		}

		rolledBack++
		log.Printf("Rolled back migration: %s", filename)
	}

	return rolledBack, nil
}

func GetMigrationStatus(ctx context.Context, db *sql.DB) ([]MigrationStatus, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return nil, fmt.Errorf("create schema_migrations table: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT filename, applied_at FROM schema_migrations 
		ORDER BY applied_at ASC, filename ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query migration status: %w", err)
	}
	defer rows.Close()

	var statuses []MigrationStatus
	for rows.Next() {
		var status MigrationStatus
		if err := rows.Scan(&status.Filename, &status.AppliedAt); err != nil {
			return nil, fmt.Errorf("scan migration status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

type MigrationStatus struct {
	Filename  string
	AppliedAt time.Time
}
