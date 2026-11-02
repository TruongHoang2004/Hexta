package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

func TestMessagingComposition(t *testing.T) {
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildController(), BuildValidator(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*controller.MessagingController) {}), fx.NopLogger); err != nil {
		t.Fatalf("messaging graph is not wired: %v", err)
	}
}
