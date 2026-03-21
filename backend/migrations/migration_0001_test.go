package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestMigration0001_ContainsRequiredAuthAndLedgerDDL(t *testing.T) {
	path := filepath.Join("0001_init_users_auth.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	sql := string(body)
	if regexp.MustCompile(`(?is)if\s+not\s+exists`).MatchString(sql) {
		t.Fatal("versioned migration must not use IF NOT EXISTS")
	}

	required := []string{
		`(?is)create\s+table\s+users\s*\(`,
		`(?is)users\s*\([^)]*\bid\s+uuid\s+primary\s+key`,
		`(?is)users\s*\([^)]*\bemail\s+text\s+unique\s+not\s+null`,
		`(?is)create\s+table\s+refresh_tokens\s*\(`,
		`(?is)refresh_tokens\s*\([^)]*\buser_id\s+uuid\s+not\s+null\s+references\s+users\s*\(\s*id\s*\)`,
		`(?is)create\s+table\s+ledgers\s*\(`,
		`(?is)ledgers\s*\([^)]*\buser_id\s+uuid\s+not\s+null\s+references\s+users\s*\(\s*id\s*\)`,
		`(?is)\bis_default\s+boolean\s+not\s+null\s+default\s+false`,
	}

	for _, pattern := range required {
		re := regexp.MustCompile(pattern)
		if !re.MatchString(sql) {
			t.Fatalf("missing required DDL pattern: %s", pattern)
		}
	}
}

func TestMigration0001Down_DropsTablesInDependencyOrder(t *testing.T) {
	path := filepath.Join("0001_init_users_auth.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	normalized := strings.ToLower(strings.Join(strings.Fields(string(body)), " "))
	required := []string{
		"drop table if exists ledgers;",
		"drop table if exists refresh_tokens;",
		"drop table if exists users;",
	}

	lastIndex := -1
	for _, stmt := range required {
		idx := strings.Index(normalized, stmt)
		if idx == -1 {
			t.Fatalf("missing required down migration statement: %s", stmt)
		}
		if idx <= lastIndex {
			t.Fatalf("down migration statement out of order: %s", stmt)
		}
		lastIndex = idx
	}
}

func TestMigration0001_UpDownCycle(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("get dedicated connection: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	tableNames := []string{"users", "refresh_tokens", "ledgers"}
	schemaName := fmt.Sprintf("task1_mig_%d", time.Now().UnixNano())

	if _, err := conn.ExecContext(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DROP SCHEMA IF EXISTS " + schemaName + " CASCADE") })

	if _, err := conn.ExecContext(ctx, "SET search_path TO "+schemaName); err != nil {
		t.Fatalf("set search_path: %v", err)
	}

	upSQL, err := os.ReadFile(filepath.Join("0001_init_users_auth.up.sql"))
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	if _, err := conn.ExecContext(ctx, string(upSQL)); err != nil {
		t.Fatalf("apply up migration: %v", err)
	}

	for _, tableName := range tableNames {
		if !tableExistsInSchema(t, ctx, conn, schemaName, tableName) {
			t.Fatalf("expected table %s to exist after up migration", tableName)
		}
	}

	downSQL, err := os.ReadFile(filepath.Join("0001_init_users_auth.down.sql"))
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	if _, err := conn.ExecContext(ctx, string(downSQL)); err != nil {
		t.Fatalf("apply down migration: %v", err)
	}

	if _, err := conn.ExecContext(ctx, string(downSQL)); err != nil {
		t.Fatalf("re-apply down migration: %v", err)
	}

	for _, tableName := range tableNames {
		if tableExistsInSchema(t, ctx, conn, schemaName, tableName) {
			t.Fatalf("expected table %s to be removed after down migration", tableName)
		}
	}
}

func tableExistsInSchema(t *testing.T, ctx context.Context, conn *sql.Conn, schemaName string, tableName string) bool {
	t.Helper()

	var exists bool
	err := conn.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2)`,
		schemaName,
		tableName,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query table existence for %s: %v", tableName, err)
	}

	return exists
}
