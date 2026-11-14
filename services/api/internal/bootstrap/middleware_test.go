package bootstrap

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func TestTenantMiddlewareComposition(t *testing.T) {
	// Validate the actual provider graph without opening database or Redis connections.
	if err := fx.ValidateApp(BuildMiddleware(), BuildService(), BuildRepository(), BuildCache(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*middleware.TenantMiddleware) {}), fx.NopLogger); err != nil {
		t.Fatalf("tenant middleware graph is not wired: %v", err)
	}
}
