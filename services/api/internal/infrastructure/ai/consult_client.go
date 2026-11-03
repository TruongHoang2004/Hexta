package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

type ConsultClient struct {
	endpoint, token string
	client          *http.Client
}

func NewConsultClient(cfg *config.Config) (*ConsultClient, error) {
	if cfg == nil {
		return nil, errors.New("AI configuration required")
	}
	endpoint := strings.TrimRight(cfg.AI.ConsultURL, "/")
	// An unconfigured optional consultant fails safely when used, without disabling unrelated API use cases.
	if endpoint == "" && cfg.AI.InternalToken == "" {
		return &ConsultClient{}, nil
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.TrimSpace(cfg.AI.InternalToken) == "" {
		return nil, errors.New("invalid internal AI configuration")
	}
	timeout := cfg.AI.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	if timeout > 120 {
		timeout = 120
	}
	return &ConsultClient{endpoint: endpoint + "/internal/v1/consult", token: cfg.AI.InternalToken, client: &http.Client{Timeout: time.Duration(timeout) * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *ConsultClient) Consult(ctx context.Context, input *types.ConsultRequest) (*types.ConsultResponse, error) {
	if c.client == nil {
		return nil, errors.New("AI consultant unavailable")
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, errors.New("invalid AI request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("invalid AI request")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, errors.New("AI consultation failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, errors.New("AI consultation failed")
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return nil, errors.New("invalid AI response")
	}
	var output types.ConsultResponse
	if json.Unmarshal(data, &output) != nil || len(output.Message) > 32000 || len(output.RecommendedProducts) > 20 || len(output.QuickReplies) > 12 {
		return nil, errors.New("invalid AI response")
	}
	return &output, nil
}
