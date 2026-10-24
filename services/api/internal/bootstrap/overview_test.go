package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

func TestOverviewControllerDependenciesAreWired(t *testing.T) {
	if err := fx.ValidateApp(BuildRepository(), BuildService(), BuildController(), BuildValidator(), BuildCache(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*controller.OverviewController) {}), fx.NopLogger); err != nil {
		t.Fatalf("overview dependencies: %v", err)
	}
}
