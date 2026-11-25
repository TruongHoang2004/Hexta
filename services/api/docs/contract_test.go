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
	// AllOf holds the single reference swag emits for a documented struct field.
	AllOf []schema `json:"allOf"`
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
		&controller.HealthController{}, &controller.AuthController{}, &controller.TenantController{}, &controller.SalesController{}, &controller.WorkflowController{}, &controller.OverviewController{}, &controller.MessagingQueryController{}, &router.MessagingCommandGateway{}, &controller.ChannelController{}, &controller.ScriptController{}, &controller.ReplySettingsController{}, &controller.TenantOnboardingController{}, &controller.AIController{}, auth, &middleware.TenantMiddleware{})
	router.RegisterOrderCaseRoutes(router.Params{Public: router.CreatePublicRouterGroup(engine)}, &controller.OrderCaseController{}, auth, &middleware.TenantMiddleware{})
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
				{"get", "/api/v1/tenants/{tenantID}/channels/facebook/connections"},
				{"post", "/api/v1/tenants/{tenantID}/channels/facebook/connections"},
				{"patch", "/api/v1/tenants/{tenantID}/channels/facebook/connections/{id}"},
				{"get", "/api/v1/tenants/{tenantID}/channels/telegram/connections"},
				{"post", "/api/v1/tenants/{tenantID}/channels/telegram/connections"},
				{"patch", "/api/v1/tenants/{tenantID}/channels/telegram/connections/{id}"},
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
			for _, endpoint := range []struct{ method, route string }{
				{"get", "/api/v1/tenants/{tenantID}/scripts"},
				{"post", "/api/v1/tenants/{tenantID}/scripts"},
				{"get", "/api/v1/tenants/{tenantID}/scripts/{id}"},
				{"put", "/api/v1/tenants/{tenantID}/scripts/{id}/draft"},
				{"post", "/api/v1/tenants/{tenantID}/scripts/{id}/activate"},
				{"post", "/api/v1/tenants/{tenantID}/scripts/{id}/deactivate"},
				{"post", "/api/v1/tenants/{tenantID}/scripts/{id}/archive"},
				{"get", "/api/v1/tenants/{tenantID}/script-permissions"},
				{"get", "/api/v1/tenants/{tenantID}/script-permissions/me"},
				{"post", "/api/v1/tenants/{tenantID}/script-permissions/{userID}/grant"},
				{"post", "/api/v1/tenants/{tenantID}/script-permissions/{userID}/revoke"},
			} {
				op, exists := spec.Paths[endpoint.route][endpoint.method]
				require.True(t, exists, "missing script operation %s %s", endpoint.method, endpoint.route)
				route := strings.NewReplacer("{tenantID}", ":tenantID", "{id}", ":id", "{userID}", ":userID").Replace(endpoint.route)
				require.True(t, registered[strings.ToUpper(endpoint.method)+" "+route], "documented script route is not registered: %s", route)
				success := resolveSchema(t, spec, op.Responses["200"].Schema)
				require.Contains(t, success.Properties, "data", "script responses must use the response envelope")
				failure := resolveSchema(t, spec, op.Responses["403"].Schema)
				require.Contains(t, failure.Properties, "detail")
			}
			for _, endpoint := range []struct{ method, route string }{
				{"get", "/api/v1/tenants/{tenantID}/orders/{id}/case"},
				{"post", "/api/v1/tenants/{tenantID}/orders/{id}/case"},
				{"put", "/api/v1/tenants/{tenantID}/orders/{id}/case"},
				{"post", "/api/v1/tenants/{tenantID}/orders/{id}/case/send-back"},
				{"post", "/api/v1/tenants/{tenantID}/orders/{id}/case/finalize"},
			} {
				op, exists := spec.Paths[endpoint.route][endpoint.method]
				require.True(t, exists, "missing order case operation %s %s", endpoint.method, endpoint.route)
				route := strings.NewReplacer("{tenantID}", ":tenantID", "{id}", ":id").Replace(endpoint.route)
				require.True(t, registered[strings.ToUpper(endpoint.method)+" "+route])
				require.Contains(t, resolveSchema(t, spec, op.Responses["200"].Schema).Properties, "data")
				require.Contains(t, resolveSchema(t, spec, op.Responses["403"].Schema).Properties, "detail")
			}
			overview, exists := spec.Paths["/api/v1/tenants/{tenantID}/overview"]["get"]
			require.True(t, exists, "missing owner overview operation")
			require.True(t, registered["GET /api/v1/tenants/:tenantID/overview"])
			overviewData := resolveSchema(t, spec, resolveSchema(t, spec, overview.Responses["200"].Schema).Properties["data"])
			for _, field := range []string{"filters", "has_activity", "orders", "exceptions", "work", "stuck", "pending_cases"} {
				require.Contains(t, overviewData.Properties, field)
			}
			_, namesRevenue := resolveSchema(t, spec, overviewData.Properties["orders"]).Properties["revenue"]
			require.False(t, namesRevenue, "order totals are order value, never revenue")
			// Exceptions come from order cases (cases never change order status) and carry no refund amount.
			exceptions := resolveSchema(t, spec, overviewData.Properties["exceptions"])
			for _, field := range []string{"in_progress", "finalized", "cancelled_order_value"} {
				require.Contains(t, exceptions.Properties, field)
			}
			for _, forbidden := range []string{"revenue", "refund", "refunded_amount"} {
				_, exists := exceptions.Properties[forbidden]
				require.False(t, exists, "exceptions must not report %s", forbidden)
			}
			for _, stage := range []string{"in_progress", "finalized"} {
				counts := resolveSchema(t, spec, exceptions.Properties[stage])
				for _, field := range []string{"total", "cancellation", "return"} {
					require.Contains(t, counts.Properties, field)
				}
			}
			pendingCases := resolveSchema(t, spec, overviewData.Properties["pending_cases"])
			require.Contains(t, pendingCases.Properties, "total")
			pendingCase := resolveSchema(t, spec, *pendingCases.Properties["items"].Items)
			for _, field := range []string{"case_id", "order_id", "order_number", "case_type", "handler_id", "waiting_since"} {
				require.Contains(t, pendingCase.Properties, field)
			}
			require.Contains(t, resolveSchema(t, spec, overview.Responses["403"].Schema).Properties, "detail")
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
	if value.Ref == "" && len(value.AllOf) == 1 {
		value = value.AllOf[0]
	}
	if value.Ref == "" {
		return value
	}
	const prefix = "#/definitions/"
	require.True(t, strings.HasPrefix(value.Ref, prefix))
	resolved, exists := spec.Definitions[strings.TrimPrefix(value.Ref, prefix)]
	require.True(t, exists, "unresolved response schema %s", value.Ref)
	return resolved
}
