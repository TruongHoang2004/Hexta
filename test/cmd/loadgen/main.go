// Command loadgen drives provider-shaped webhook load into the connector and checks what reached the API.
//
// It posts signed Facebook batches or Telegram updates straight to the connector at a constant open-loop rate, each
// text carrying a run marker. With -db it polls the API database (read-only) during and after the attack until every
// accepted message is stored or -drain expires, and reports lost and duplicate messages and webhook → visible latency
// (first poll that saw the row; the API stores the provider timestamp, not its commit time). Each poll rereads the
// whole run, so very long runs need an API-side commit timestamp instead. It is meant to run
// inside the local Compose network (see infrastructure/README.demo.md); it never prints secrets or message bodies.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TruongHoang2004/Hexta/test/mockplatform"
	"github.com/jackc/pgx/v5"
)

type options struct {
	platform, account, connector, mock, secretFile, db string
	rate, customers                                    int
	duration, drain, poll                              time.Duration
}

func main() {
	var o options
	flag.StringVar(&o.platform, "platform", "telegram", "facebook or telegram")
	flag.StringVar(&o.account, "account", "", "connected mock Page ID or bot ID (required)")
	flag.StringVar(&o.connector, "connector", "http://connector:8091", "connector base URL (Facebook)")
	flag.StringVar(&o.mock, "mock", "http://mockplatform:8090", "mock platform URL (Telegram webhook lookup)")
	flag.StringVar(&o.secretFile, "app-secret-file", "/run/secrets/meta_app_secret", "Meta app secret file (Facebook)")
	flag.StringVar(&o.db, "db", "", "read-only API database URL; empty skips delivery verification")
	flag.IntVar(&o.rate, "rate", 20, "messages per second")
	flag.DurationVar(&o.duration, "duration", 30*time.Second, "attack duration")
	flag.IntVar(&o.customers, "customers", 50, "distinct simulated customers (threads)")
	flag.DurationVar(&o.drain, "drain", 5*time.Minute, "how long to wait for accepted messages to be stored")
	flag.DurationVar(&o.poll, "poll", 200*time.Millisecond, "database poll interval; bounds the stored-latency precision")
	flag.Parse()
	if (o.platform != "facebook" && o.platform != "telegram") || o.account == "" || o.rate <= 0 || o.customers <= 0 || o.duration <= 0 {
		flag.Usage()
		os.Exit(2)
	}
	run, err := newRun(o)
	if err != nil {
		log.Fatal(err)
	}
	if err := run.execute(context.Background()); err != nil {
		log.Fatal(err)
	}
}

type runState struct {
	o       options
	id      string
	target  string
	headers func(body []byte) http.Header
	idBase  int64       // run-unique offset for provider message IDs
	sent    []time.Time // send time per seq (1-based)
	status  []int       // webhook status per seq; 0 = transport error
	latency []time.Duration

	mu    sync.Mutex
	seen  map[int64]time.Time // first poll that observed a stored row, per seq
	rows  map[int64]int       // stored rows per seq at the latest poll
	pollE error
}

func newRun(o options) (*runState, error) {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	total := int(float64(o.rate) * o.duration.Seconds())
	r := &runState{o: o, id: hex.EncodeToString(b), sent: make([]time.Time, total+1), status: make([]int, total+1), latency: make([]time.Duration, total+1), seen: map[int64]time.Time{}, rows: map[int64]int{}, idBase: time.Now().UnixMilli() * 1000}
	switch o.platform {
	case "facebook":
		secret, err := os.ReadFile(o.secretFile)
		if err != nil || len(strings.TrimSpace(string(secret))) < 16 {
			return nil, errors.New("cannot read the Meta app secret file")
		}
		key := []byte(strings.TrimSpace(string(secret)))
		r.target = strings.TrimRight(o.connector, "/") + "/webhooks/facebook"
		r.headers = func(body []byte) http.Header {
			return http.Header{"Content-Type": {"application/json"}, "X-Hub-Signature-256": {mockplatform.SignFacebook(key, body)}}
		}
	case "telegram":
		res, err := http.Get(strings.TrimRight(o.mock, "/") + "/mock/telegram/webhook?bot=" + o.account)
		if err != nil {
			return nil, errors.New("mock platform unreachable")
		}
		defer res.Body.Close()
		var hook struct {
			URL    string `json:"url"`
			Secret string `json:"secret_token"`
		}
		if res.StatusCode != http.StatusOK || json.NewDecoder(res.Body).Decode(&hook) != nil {
			return nil, errors.New("bot has no webhook in the mock platform; connect it in Hexta first")
		}
		r.target = hook.URL
		r.headers = func([]byte) http.Header {
			return http.Header{"Content-Type": {"application/json"}, mockplatform.TelegramSecretHeader: {hook.Secret}}
		}
	}
	return r, nil
}

// body builds the provider webhook for one sequence number; its text starts with the run marker.
func (r *runState) body(seq int, at time.Time) []byte {
	customer := int64(seq % r.o.customers)
	text := "lg:" + r.id + ":" + strconv.Itoa(seq) + " load test message"
	if r.o.platform == "facebook" {
		return mockplatform.FacebookMessage(r.o.account, fmt.Sprintf("8%09d", customer), "m_lg_"+r.id+"_"+strconv.Itoa(seq), text, "", at)
	}
	// Telegram deduplicates by chat + message_id, so IDs must not repeat across runs for the same chats.
	id := r.idBase + int64(seq)
	return mockplatform.TelegramUpdate(id, 900000000+customer, id, text, "", at)
}

// maxInFlight bounds concurrent webhooks; when reached, sending falls behind schedule and the report says so.
const maxInFlight = 512

func (r *runState) execute(ctx context.Context) error {
	total := len(r.sent) - 1
	fmt.Printf("run %s: %s account %s, %d msg/s for %s (%d messages) over %d customers\n", r.id, r.o.platform, r.o.account, r.o.rate, r.o.duration, total, r.o.customers)
	var conn *pgx.Conn
	if r.o.db != "" {
		var err error
		if conn, err = pgx.Connect(ctx, r.o.db); err != nil {
			return errors.New("cannot connect to the API database")
		}
		defer conn.Close(context.Background())
		if _, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY"); err != nil {
			return errors.New("cannot make the database session read-only")
		}
	}
	pollCtx, stopPolling := context.WithCancel(ctx)
	defer stopPolling()
	polled := make(chan struct{})
	if conn != nil {
		go func() { defer close(polled); r.watch(pollCtx, conn) }()
	} else {
		close(polled)
	}
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{MaxIdleConnsPerHost: maxInFlight, MaxConnsPerHost: maxInFlight}}
	slots := make(chan struct{}, maxInFlight)
	var wg sync.WaitGroup
	start := time.Now()
	var behind time.Duration
	for seq := 1; seq <= total; seq++ {
		due := start.Add(time.Duration(seq-1) * time.Second / time.Duration(r.o.rate))
		if wait := time.Until(due); wait > 0 {
			time.Sleep(wait)
		}
		slots <- struct{}{}
		if lag := time.Since(due); lag > behind {
			behind = lag
		}
		wg.Add(1)
		go func(seq int) {
			defer func() { <-slots; wg.Done() }()
			at := time.Now()
			body := r.body(seq, at)
			req, _ := http.NewRequestWithContext(ctx, http.MethodPost, r.target, bytes.NewReader(body))
			req.Header = r.headers(body)
			r.sent[seq] = at
			res, err := client.Do(req)
			r.latency[seq] = time.Since(at)
			if err != nil {
				return
			}
			res.Body.Close()
			r.status[seq] = res.StatusCode
		}(seq)
	}
	wg.Wait()
	attackEnd := time.Now()
	statuses := map[int]int{}
	latencies := make([]time.Duration, 0, total)
	for seq := 1; seq <= total; seq++ {
		statuses[r.status[seq]]++
		latencies = append(latencies, r.latency[seq])
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	achieved := float64(total) / attackEnd.Sub(start).Seconds()
	fmt.Printf("webhooks: %d sent in %s (%.1f msg/s, max schedule lag %s), status %v (0 = transport error)\n", total, attackEnd.Sub(start).Round(time.Millisecond), achieved, behind.Round(time.Millisecond), statuses)
	fmt.Printf("webhook latency: p50 %s p95 %s p99 %s max %s\n", pct(latencies, 50), pct(latencies, 95), pct(latencies, 99), latencies[len(latencies)-1])
	if conn == nil {
		fmt.Println("delivery verification: NOT_RUN (no -db)")
		return nil
	}
	r.verify(attackEnd, stopPolling, polled)
	return r.pollE
}

// watch polls the API database until ctx ends, recording when each run message was first visible.
func (r *runState) watch(ctx context.Context, conn *pgx.Conn) {
	ticker := time.NewTicker(r.o.poll)
	defer ticker.Stop()
	for {
		stored, err := r.stored(ctx, conn)
		now := time.Now()
		r.mu.Lock()
		if err != nil && ctx.Err() == nil {
			r.pollE = err
		}
		for seq, n := range stored {
			if _, ok := r.seen[seq]; !ok {
				r.seen[seq] = now
			}
			r.rows[seq] = n
		}
		r.mu.Unlock()
		if err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *runState) verify(attackEnd time.Time, stopPolling context.CancelFunc, polled <-chan struct{}) {
	accepted := 0
	for _, code := range r.status[1:] {
		if code == http.StatusOK {
			accepted++
		}
	}
	deadline := attackEnd.Add(r.o.drain)
	for {
		r.mu.Lock()
		done, failed := len(r.seen) >= accepted, r.pollE != nil
		r.mu.Unlock()
		if done || failed || time.Now().After(deadline) {
			break
		}
		time.Sleep(r.o.poll)
	}
	drained := time.Since(attackEnd)
	stopPolling()
	<-polled
	var latencies []time.Duration
	lost, duplicates, unexpected := 0, 0, 0
	for seq := 1; seq < len(r.status); seq++ {
		code, rows, at := r.status[seq], r.rows[int64(seq)], r.seen[int64(seq)]
		if code != http.StatusOK {
			if rows > 0 {
				unexpected++
			}
			continue
		}
		if rows == 0 {
			lost++
			continue
		}
		duplicates += rows - 1
		latencies = append(latencies, at.Sub(r.sent[seq]))
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	fmt.Printf("delivery: %d accepted, %d stored, %d lost, %d duplicate rows, %d stored despite non-200; drained %s after the attack\n", accepted, len(latencies), lost, duplicates, unexpected, drained.Round(time.Millisecond))
	if len(latencies) > 0 {
		fmt.Printf("webhook → visible (poll %s): p50 %s p95 %s p99 %s max %s\n", r.o.poll, pct(latencies, 50), pct(latencies, 95), pct(latencies, 99), latencies[len(latencies)-1])
	}
}

// stored counts API inbound rows of this run by sequence number.
func (r *runState) stored(ctx context.Context, conn *pgx.Conn) (map[int64]int, error) {
	rows, err := conn.Query(ctx, "SELECT split_part(body, ' ', 1), count(*) FROM messaging_messages WHERE direction = 'inbound' AND body LIKE $1 GROUP BY 1", "lg:"+r.id+":%")
	if err != nil {
		return nil, errors.New("cannot query messaging_messages")
	}
	defer rows.Close()
	out := map[int64]int{}
	prefix := "lg:" + r.id + ":"
	for rows.Next() {
		var marker string
		var n int
		if rows.Scan(&marker, &n) != nil {
			return nil, errors.New("cannot read messaging_messages")
		}
		seq, err := strconv.ParseInt(strings.TrimPrefix(marker, prefix), 10, 64)
		if err == nil {
			out[seq] = n
		}
	}
	return out, rows.Err()
}

func pct(sorted []time.Duration, p int) time.Duration {
	i := (len(sorted)*p+99)/100 - 1
	return sorted[max(0, min(i, len(sorted)-1))]
}
