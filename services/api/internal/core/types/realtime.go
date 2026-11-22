package types

// Realtime hint reasons.
const (
	RealtimeConversationChanged = "conversation"
	RealtimeDeliveryChanged     = "delivery"
	// RealtimeResync tells a stream that hints may have been lost, so clients refetch everything.
	RealtimeResync = "resync"
)

// RealtimeChannelPrefix prefixes the per-tenant pub/sub channel carrying RealtimeHint values.
const RealtimeChannelPrefix = "hexta:rt:tenant:"

// RealtimeHint tells connected clients that a conversation changed. It is a best-effort trigger, never a source
// of truth: it carries no message content, and clients refetch through the authorized read endpoints.
type RealtimeHint struct {
	ThreadID string `json:"thread_id"`
	Reason   string `json:"reason"`
}

// RealtimeChannel is the pub/sub channel for one tenant.
func RealtimeChannel(tenant string) string { return RealtimeChannelPrefix + tenant }
