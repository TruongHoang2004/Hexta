package bootstrap

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func TestAIAgentPortsAreWired(t *testing.T) {
	// Validate the actual composition without opening DB/Redis/provider clients.
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildCache(),
		fx.Supply(&gorm.DB{}), fx.Invoke(func(service.IAIAgentService) {}), fx.NopLogger,
	); err != nil {
		t.Fatalf("AI agent dependency ports are not wired: %v", err)
	}
}
