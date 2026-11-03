package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/stretchr/testify/require"
)

func TestConsultClientTransport(t *testing.T) {
	for _, body := range []string{`{"message":"Hello","quick_replies":[]}`, `not json`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/internal/v1/consult", r.URL.Path)
				require.Equal(t, "Bearer fixture-token", r.Header.Get("Authorization"))
				var input types.ConsultRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&input))
				require.Equal(t, "tenant", input.TenantID)
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()
			cfg := &config.Config{}
			cfg.AI.ConsultURL = server.URL
			cfg.AI.InternalToken = "fixture-token"
			client, err := NewConsultClient(cfg)
			require.NoError(t, err)
			result, err := client.Consult(context.Background(), &types.ConsultRequest{TenantID: "tenant"})
			if body == "not json" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, "Hello", result.Message)
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err = client.Consult(ctx, &types.ConsultRequest{})
			require.Error(t, err)
		})
	}
}
func TestConsultClientRejectsStatusAndRedirect(t *testing.T) {
	for _, status := range []int{401, 500, 302} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "http://invalid.example")
				w.WriteHeader(status)
				_, _ = w.Write([]byte("sensitive-provider-detail"))
			}))
			defer server.Close()
			cfg := &config.Config{}
			cfg.AI.ConsultURL = server.URL
			cfg.AI.InternalToken = "fixture-token"
			client, err := NewConsultClient(cfg)
			require.NoError(t, err)
			_, err = client.Consult(context.Background(), &types.ConsultRequest{})
			require.Error(t, err)
			require.NotContains(t, err.Error(), "sensitive-provider-detail")
		})
	}
}

func TestConsultClientAcceptsBaseOrFullEndpoint(t *testing.T) {
	for _, path := range []string{"", "/", "/internal/v1/consult", "/internal/v1/consult/", "/prefix", "/prefix/internal/v1/consult/"} {
		t.Run(path, func(t *testing.T) {
			expected := "/internal/v1/consult"
			if strings.HasPrefix(path, "/prefix") {
				expected = "/prefix/internal/v1/consult"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, expected, r.URL.Path)
				require.Equal(t, "Bearer fixture-token", r.Header.Get("Authorization"))
				_, _ = w.Write([]byte(`{"message":"Hello"}`))
			}))
			defer server.Close()
			cfg := &config.Config{}
			cfg.AI.ConsultURL = server.URL + path
			cfg.AI.InternalToken = "fixture-token"
			client, err := NewConsultClient(cfg)
			require.NoError(t, err)
			_, err = client.Consult(context.Background(), &types.ConsultRequest{})
			require.NoError(t, err)
		})
	}
}
