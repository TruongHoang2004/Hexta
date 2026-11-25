package bootstrap

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func TestOverviewComposition(t *testing.T) {
	// The overview controller resolves through the real graph with the repository bound to OverviewStore.
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildController(), BuildValidator(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*controller.OverviewController) {}), fx.NopLogger); err != nil {
		t.Fatalf("overview graph is not wired: %v", err)
	}
}
