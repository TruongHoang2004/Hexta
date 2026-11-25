package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"testing"
)

func TestMessagingComposition(t *testing.T) {
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildController(), BuildValidator(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*controller.MessagingQueryController, *controller.MessagingCommandController, *controller.ReplySettingsController) {
	}), fx.NopLogger); err != nil {
		t.Fatalf("messaging graph is not wired: %v", err)
	}
}

func TestMessagingCommandProcessComposition(t *testing.T) {
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildController(), BuildValidator(), BuildMiddleware(), MessagingCommandRouterModule, BuildTurnExpiry(), BuildReplyRouter(), BuildCache(), fx.Supply(&gorm.DB{}, zap.NewNop().Sugar()), fx.NopLogger); err != nil {
		t.Fatalf("command process graph is not wired: %v", err)
	}
}
