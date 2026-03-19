package migrations_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
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
