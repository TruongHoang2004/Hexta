package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

// RunIngress is a dedicated API-owned consumer process, sharing only the API database.
func RunIngress(ctx context.Context, db *gorm.DB) error {
	if os.Getenv("KAFKA_BROKERS") == "" {
		return errors.New("KAFKA_BROKERS required")
	}
	reader := kafkatransport.NewReader(strings.Split(os.Getenv("KAFKA_BROKERS"), ","), "hexta-api-ingress-v1")
	defer reader.Close()
	store := repository.NewIngressRepository(db)
	consumer := service.NewIngressService(store)
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		var e ingress.Event
		if json.Unmarshal(m.Value, &e) != nil || e.Validate() != nil {
			if err := store.Quarantine(ctx, ingress.Hash(string(m.Value)), "invalid_envelope"); err != nil {
				return errors.New("quarantine unavailable")
			}
			if err := reader.CommitMessages(ctx, m); err != nil {
				return err
			}
			continue
		}
		for {
			if consumer.Consume(ctx, e) == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}
		if err = reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

func RunIngressProcess(ctx context.Context) error {
	dsn := os.Getenv("API_DATABASE_URL")
	if dsn == "" {
		return errors.New("API_DATABASE_URL required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return errors.New("database unavailable")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return errors.New("database unavailable")
	}
	defer sqlDB.Close()
	return RunIngress(ctx, db)
}
