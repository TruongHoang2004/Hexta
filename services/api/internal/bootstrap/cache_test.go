package bootstrap

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"go.uber.org/fx"
)

func TestCacheProvidesDraftCache(t *testing.T) {
	// Validate the actual provider graph without opening a Redis connection.
	if err := fx.ValidateApp(BuildCache(), fx.Invoke(func(service.IDraftCache) {})); err != nil {
		t.Fatalf("draft cache dependency is not wired: %v", err)
	}
}
