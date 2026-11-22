package docs_test

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/docs"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/router"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type schema struct {
	Ref        string            `json:"$ref"`
	Properties map[string]schema `json:"properties"`
	Items      *schema           `json:"items"`
}
type operation struct {
	Parameters []struct {
		In string `json:"in"`
	} `json:"parameters"`
	Responses map[string]struct {
		Schema schema `json:"schema"`
	} `json:"responses"`
}
type specification struct {
	Host        string                          `json:"host"`
	BasePath    string                          `json:"basePath"`
	Paths       map[string]map[string]operation `json:"paths"`
	Definitions map[string]schema               `json:"definitions"`
}

func TestGeneratedSwaggerMatchesRegisteredAuthAndHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	auth := middleware.NewAuthMiddleware(nil)
	// Register the real route tree; no handlers or dependencies need execution.
	router.RegisterRoutes(router.Params{Public: router.CreatePublicRouterGroup(engine), Private: router.CreatePrivateRouterGroup(engine, auth)},
		&controller.HealthController{}, &controller.AuthController{}, &controller.TenantController{}, &controller.SalesController{}, &controller.WorkflowController{}, &controller.MessagingQueryController{}, &router.MessagingCommandGateway{}, &controller.ChannelController{}, &controller.TenantOnboardingController{}, &controller.AIController{}, auth, &middleware.TenantMiddleware{})
	registered := map[string]bool{}
	for _, route := range engine.Routes() {
		registered[route.Method+" "+route.Path] = true
	}
	staticJSON, err := os.ReadFile("swagger.json")
	require.NoError(t, err)
	for _, artifact := range []struct{ name, source string }{{"runtime Go document", docs.SwaggerInfo.ReadDoc()}, {"JSON document", string(staticJSON)}} {
		t.Run(artifact.name, func(t *testing.T) {
			var spec specification
			require.NoError(t, json.Unmarshal([]byte(artifact.source), &spec))
			require.Empty(t, spec.Host, "Swagger must use the host serving the document")
			for _, endpoint := range []struct{ method, route string }{{"get", "/api/v1/health"}, {"post", "/api/v1/auth/register"}, {"post", "/api/v1/auth/login"}, {"post", "/api/v1/auth/refresh"}} {
				operations, exists := spec.Paths[endpoint.route]
				require.True(t, exists, "missing documented endpoint %s", endpoint.route)
				op, exists := operations[endpoint.method]
				require.True(t, exists)
				resolved := path.Join("/", spec.BasePath, endpoint.route)
				require.True(t, registered[strings.ToUpper(endpoint.method)+" "+resolved], "generated Swagger resolves to an unregistered route: %s", resolved)
				success, exists := op.Responses["200"]
				require.True(t, exists)
				envelope := resolveSchema(t, spec, success.Schema)
				data, exists := envelope.Properties["data"]
				require.True(t, exists, "success response must use the response envelope")
				_, exists = envelope.Properties["meta"]
				require.True(t, exists)
				payload := resolveSchema(t, spec, data)
				_, exposesRefresh := payload.Properties["refresh_token"]
				require.False(t, exposesRefresh, "refresh credentials must not be exposed in JSON")
				if endpoint.route == "/api/v1/health" {
					_, exists = payload.Properties["status"]
					require.True(t, exists)
					_, exists = payload.Properties["details"]
					require.True(t, exists)
				} else {
					_, exists = payload.Properties["access_token"]
					require.True(t, exists)
				}
				if endpoint.route == "/api/v1/auth/refresh" {
					for _, parameter := range op.Parameters {
						require.NotEqual(t, "body", parameter.In, "refresh must use its HttpOnly cookie without a JSON token body")
					}
				}
			}
			for _, endpoint := range []struct{ method, route string }{
				{"get", "/api/v1/tenants/{tenantID}/conversations"},
				{"get", "/api/v1/tenants/{tenantID}/conversations/{id}/messages"},
				{"post", "/api/v1/tenants/{tenantID}/simulator/messages"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/claim"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/assign"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/handoff"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/return-to-bot"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/send-simulation"},
				{"post", "/api/v1/tenants/{tenantID}/turns/{id}/messages"},
			} {
				op, exists := spec.Paths[endpoint.route][endpoint.method]
				require.True(t, exists, "missing messaging operation %s", endpoint.route)
				route := strings.NewReplacer("{tenantID}", ":tenantID", "{id}", ":id").Replace(endpoint.route)
				require.True(t, registered[strings.ToUpper(endpoint.method)+" "+route])
				success := resolveSchema(t, spec, op.Responses["200"].Schema)
				require.Contains(t, success.Properties, "data")
				failure := resolveSchema(t, spec, op.Responses["409"].Schema)
				require.Contains(t, failure.Properties, "detail")
				require.Contains(t, failure.Properties, "http_status")
			}
			for _, endpoint := range []struct{ method, route string }{
				{"post", "/api/v1/tenants/{tenantID}/channels/facebook/oauth/start"},
				{"get", "/api/v1/channels/facebook/oauth/callback"},
				{"post", "/api/v1/tenants/{tenantID}/channels/facebook/connections"},
				{"post", "/api/v1/tenants/{tenantID}/channels/telegram/connections"},
				{"get", "/api/v1/tenants/{tenantID}/connections"},
				{"patch", "/api/v1/tenants/{tenantID}/connections/{id}"},
			} {
				op, exists := spec.Paths[endpoint.route][endpoint.method]
				require.True(t, exists, "missing channel operation %s", endpoint.route)
				route := strings.NewReplacer("{tenantID}", ":tenantID", "{id}", ":id").Replace(endpoint.route)
				require.True(t, registered[strings.ToUpper(endpoint.method)+" "+route])
				payload := resolveSchema(t, spec, resolveSchema(t, spec, op.Responses["200"].Schema).Properties["data"])
				for _, field := range []string{"access_token", "bot_token", "webhook_secret"} {
					_, exposesToken := payload.Properties[field]
					require.False(t, exposesToken, "channel responses must never carry provider tokens")
				}
				if list, ok := payload.Properties["connections"]; ok && list.Items != nil {
					_, exposesToken := resolveSchema(t, spec, *list.Items).Properties["bot_token"]
					require.False(t, exposesToken, "channel list responses must never carry provider tokens")
				}
			}
			require.True(t, registered["POST /api/v1/ai/converse"])
			require.Contains(t, spec.Paths["/api/v1/ai/converse"], "post")
			// Health's unavailable response has the same typed envelope as success.
			unavailable := spec.Paths["/api/v1/health"]["get"].Responses["503"].Schema
			require.NotEmpty(t, resolveSchema(t, spec, unavailable).Properties["data"].Ref)
		})
	}
}

func resolveSchema(t *testing.T, spec specification, value schema) schema {
	t.Helper()
	if value.Ref == "" {
		return value
	}
	const prefix = "#/definitions/"
	require.True(t, strings.HasPrefix(value.Ref, prefix))
	resolved, exists := spec.Definitions[strings.TrimPrefix(value.Ref, prefix)]
	require.True(t, exists, "unresolved response schema %s", value.Ref)
	return resolved
}
