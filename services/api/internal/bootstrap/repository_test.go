package bootstrap

import (
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func TestRepositoryProvidesDBAuthoritativeSessionRepository(t *testing.T) {
	// Build the real repository graph so session validity cannot silently move to a cache path.
	var sessionRepo repository.ISessionRepository
	app := fx.New(
		BuildRepository(),
		fx.Supply(&gorm.DB{}),
		fx.Populate(&sessionRepo),
		fx.NopLogger,
	)
	if err := app.Err(); err != nil {
		t.Fatalf("session repository graph is not wired: %v", err)
	}
	if _, ok := sessionRepo.(*repository.SessionRepository); !ok {
		t.Fatalf("expected ISessionRepository to be *repository.SessionRepository, got %T", sessionRepo)
	}
}

func TestAuthServiceComposition(t *testing.T) {
	// Validate the actual provider graph without opening database or Redis connections.
	if err := fx.ValidateApp(BuildService(), BuildRepository(), BuildCache(), fx.Supply(&gorm.DB{}), fx.Invoke(func(*service.AuthService) {}), fx.NopLogger); err != nil {
		t.Fatalf("auth service graph is not wired: %v", err)
	}
}
