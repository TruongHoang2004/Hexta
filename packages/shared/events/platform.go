package events

// Connection platforms. Every connection has exactly one platform; a tenant may hold any number of connections.
const (
	PlatformFacebook = "facebook"
	PlatformTelegram = "telegram"
)

// Platforms lists the connection platforms events may carry.
var Platforms = map[string]bool{PlatformFacebook: true, PlatformTelegram: true}
