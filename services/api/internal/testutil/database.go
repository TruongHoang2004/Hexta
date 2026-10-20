// Package testutil provisions disposable PostgreSQL databases using real Atlas history.
package testutil

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database never migrates an existing application database. The configured URL
// is used only to create/drop a randomly named test database on an explicit server.
func Database(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("API_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("SALES_TEST_DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("API_TEST_DATABASE_URL is required for Atlas/PostgreSQL database tests")
	}
	parsed, err := url.Parse(dsn)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		t.Fatal("test database URL must be a PostgreSQL URL")
	}
	if _, err := exec.LookPath("atlas"); err != nil {
		t.Fatal("Atlas CLI is required for configured PostgreSQL tests")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to the explicit test PostgreSQL server")
	}
	adminSQL, _ := admin.DB()
	t.Cleanup(func() { _ = adminSQL.Close() })
	name := "hexta_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := admin.Exec(`CREATE DATABASE "` + name + `"`).Error; err != nil {
		t.Fatal("cannot create disposable Atlas test database")
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP DATABASE IF EXISTS "` + name + `" WITH (FORCE)`).Error; err != nil {
			t.Error("could not drop disposable Atlas test database")
		}
	})
	parsed.Path = "/" + name
	query := parsed.Query()
	query.Set("search_path", "public")
	parsed.RawQuery = query.Encode()
	_, source, _, _ := runtime.Caller(0)
	dir := filepath.Clean(filepath.Join(filepath.Dir(source), "../../../../migrations/api"))
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "atlas", "migrate", "apply", "--dir", "file://"+dir, "--url", parsed.String(), "--revisions-schema", "atlas_schema_revisions")
	// Do not print migration command arguments, database URLs or driver errors.
	if err := command.Run(); err != nil {
		t.Fatal("Atlas failed to apply the API history to the disposable test database")
	}
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot connect to the migrated disposable test database")
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}
