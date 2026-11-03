package bootstrap

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/facebook"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"strings"
	"time"
)

func Run(ctx context.Context) error {
	for _, key := range []string{"CONNECTOR_DATABASE_URL", "KAFKA_BROKERS", "META_APP_ID", "META_APP_SECRET_FILE", "META_VERIFY_TOKEN_FILE", "CONNECTOR_MANAGEMENT_TOKEN_FILE"} {
		if os.Getenv(key) == "" {
			return errors.New("missing connector configuration")
		}
	}
	read := func(key string) ([]byte, error) {
		v, err := os.ReadFile(os.Getenv(key))
		if err != nil || len(strings.TrimSpace(string(v))) < 16 {
			return nil, errors.New("invalid secret file")
		}
		return []byte(strings.TrimSpace(string(v))), nil
	}
	secret, err := read("META_APP_SECRET_FILE")
	if err != nil {
		return err
	}
	verify, err := read("META_VERIFY_TOKEN_FILE")
	if err != nil {
		return err
	}
	management, err := read("CONNECTOR_MANAGEMENT_TOKEN_FILE")
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, os.Getenv("CONNECTOR_DATABASE_URL"))
	if err != nil {
		return errors.New("database unavailable")
	}
	defer pool.Close()
	if pool.Ping(ctx) != nil {
		return errors.New("database unavailable")
	}
	store := &repository.Store{Pool: pool}
	p := facebook.FacebookPlugin{}
	registry, err := plugin.NewRegistry(p)
	if err != nil {
		return err
	}
	inbound := &service.InboundService{Registry: registry, Store: store}
	h := &transport.Handler{Inbound: inbound, Plugin: p, AppID: os.Getenv("META_APP_ID"), Secret: secret, VerifyToken: string(verify), ManagementToken: string(management), Put: func(r *http.Request, c service.Connection) error { return store.PutConnection(r.Context(), c) }}
	addr := os.Getenv("CONNECTOR_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8091"
	}
	server := &http.Server{Addr: addr, Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	pub := kafkatransport.NewPublisher(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	defer pub.Writer.Close()
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				v, err := store.Claim(workerCtx)
				if err != nil {
					continue
				}
				call, stop := context.WithTimeout(workerCtx, 10*time.Second)
				err = service.Publish(call, pub, v.Payload)
				stop()
				_ = store.Finish(workerCtx, v, err == nil)
			}
		}
	}()
	errch := make(chan error, 1)
	go func() { errch <- server.ListenAndServe() }()
	select {
	case <-ctx.Done():
	case err = <-errch:
	}
	cancel()
	shutdown, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	_ = server.Shutdown(shutdown)
	<-done
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
