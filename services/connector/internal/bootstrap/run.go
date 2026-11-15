package bootstrap

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/packages/shared/outbound"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin/facebook"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/secret"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/transport"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"os"
	"strings"
	"time"
)

func Run(ctx context.Context) error {
	for _, key := range []string{"CONNECTOR_DATABASE_URL", "KAFKA_BROKERS", "META_APP_ID", "META_APP_SECRET_FILE", "META_VERIFY_TOKEN_FILE", "CONNECTOR_MANAGEMENT_TOKEN_FILE", "CONNECTOR_CREDENTIAL_KEY_FILE"} {
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
	appSecret, err := read("META_APP_SECRET_FILE")
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
	// 32-byte AES-256-GCM key as 64 hex characters; rotating it requires owners to reconnect Pages.
	keyFile, err := read("CONNECTOR_CREDENTIAL_KEY_FILE")
	if err != nil {
		return err
	}
	sealer, err := secret.NewSealer(string(keyFile), 1)
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
	graph := facebook.NewGraph(os.Getenv("META_GRAPH_URL"), os.Getenv("META_GRAPH_VERSION"), os.Getenv("META_APP_ID"), appSecret)
	channels := &service.ChannelService{Store: store, Provider: graph, Sealer: sealer, KeyVer: sealer.Version, AppID: os.Getenv("META_APP_ID")}
	sender := &service.OutboundService{Store: store, Sender: graph, Sealer: sealer}
	h := &transport.Handler{Inbound: inbound, Plugin: p, AppID: os.Getenv("META_APP_ID"), Secret: appSecret, VerifyToken: string(verify), ManagementToken: string(management), Channels: channels}
	addr := os.Getenv("CONNECTOR_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8091"
	}
	server := &http.Server{Addr: addr, Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	pub := kafkatransport.NewPublisher(brokers)
	defer pub.Writer.Close()
	publishers := map[string]*kafkatransport.Publisher{outbound.ResultTopic: kafkatransport.NewTopicPublisher(brokers, outbound.ResultTopic), outbound.HealthTopic: kafkatransport.NewTopicPublisher(brokers, outbound.HealthTopic)}
	for _, p := range publishers {
		defer p.Writer.Close()
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{}, 3)
	go func() {
		defer func() { done <- struct{}{} }()
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
	go func() {
		defer func() { done <- struct{}{} }()
		relayOutbound(workerCtx, store, publishers)
	}()
	go func() {
		defer func() { done <- struct{}{} }()
		consumeIntents(workerCtx, brokers, sender)
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
	for i := 0; i < 3; i++ {
		<-done
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// relayOutbound publishes committed result/health events at least once; API consumers deduplicate by identity.
func relayOutbound(ctx context.Context, store *repository.Store, publishers map[string]*kafkatransport.Publisher) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				v, err := store.ClaimOutbound(ctx)
				if err != nil {
					break
				}
				publisher := publishers[v.Topic]
				call, stop := context.WithTimeout(ctx, 10*time.Second)
				err = publisher.Publish(call, v.Key, v.Payload)
				stop()
				if store.FinishOutbound(ctx, v, err == nil) != nil || err != nil {
					break
				}
			}
		}
	}
}

// consumeIntents processes one intent at a time and commits its offset only after the attempt outcome is durable.
func consumeIntents(ctx context.Context, brokers []string, sender *service.OutboundService) {
	reader := kafkatransport.NewTopicReader(brokers, "hexta-connector-outbound-v1", outbound.IntentTopic)
	defer reader.Close()
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		for {
			err = sender.Handle(ctx, m.Value)
			if err == nil || errors.Is(err, service.ErrInvalidIntent) {
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
		if reader.CommitMessages(ctx, m) != nil && ctx.Err() != nil {
			return
		}
	}
}
