package router

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/middleware"
	"github.com/gin-gonic/gin"
)

// registerMessagingCommandRoutes mounts every conversation write endpoint; the handler decides where it runs.
func registerMessagingCommandRoutes(resources *gin.RouterGroup, handle func(string) gin.HandlerFunc) {
	for _, route := range []struct{ path, name string }{
		{"/simulator/messages", "Inbound"},
		{"/turns/:id/claim", "Claim"},
		{"/turns/:id/assign", "Assign"},
		{"/turns/:id/handoff", "Handoff"},
		{"/turns/:id/return-to-bot", "ReturnToBot"},
		{"/turns/:id/resolve", "Resolve"},
		{"/turns/:id/send-simulation", "Send"},
		{"/turns/:id/messages", "SendChannel"},
		{"/turns/:id/proposals/:proposalID/send", "SendProposal"},
		{"/turns/:id/proposals/:proposalID/discard", "DiscardProposal"},
		{"/turns/:id/regenerate-proposal", "RegenerateProposal"},
	} {
		resources.POST(route.path, handle(route.name))
	}
}

// RegisterMessagingCommandRoutes is the whole HTTP surface of the messaging command process.
func RegisterMessagingCommandRoutes(params Params, health *controller.HealthController, auth *middleware.AuthMiddleware, tenants *middleware.TenantMiddleware, commands *controller.MessagingCommandController) {
	params.Public.GET("/health", health.HealthCheck)
	account := params.Public.Group("", auth.RequireAuth())
	resources := tenantResources(account.Group("/tenants"), tenants)
	registerMessagingCommandRoutes(resources, func(name string) gin.HandlerFunc {
		switch name {
		case "Inbound":
			return commands.Inbound
		case "Claim":
			return commands.Claim
		case "Assign":
			return commands.Assign
		case "Handoff":
			return commands.Handoff
		case "ReturnToBot":
			return commands.ReturnToBot
		case "Resolve":
			return commands.Resolve
		case "Send":
			return commands.Send
		case "SendProposal":
			return commands.SendProposal
		case "DiscardProposal":
			return commands.DiscardProposal
		case "RegenerateProposal":
			return commands.RegenerateProposal
		default:
			return commands.SendChannel
		}
	})
}

// MessagingCommandGateway forwards conversation writes to the command process, so this API never writes them.
type MessagingCommandGateway struct{ proxy *httputil.ReverseProxy }

func NewMessagingCommandGateway(cfg *config.Config) *MessagingCommandGateway {
	target, err := url.Parse(cfg.Messaging.CommandURL)
	if cfg.Messaging.CommandURL == "" || err != nil || target.Host == "" {
		return &MessagingCommandGateway{}
	}
	return &MessagingCommandGateway{proxy: httputil.NewSingleHostReverseProxy(target)}
}

func (g *MessagingCommandGateway) Handle(string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if g.proxy == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"detail": "messaging command service is not configured"})
			return
		}
		g.proxy.ServeHTTP(c.Writer, c.Request)
	}
}
