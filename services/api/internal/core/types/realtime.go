package types

// Realtime hint reasons.
const (
	RealtimeConversationChanged = "conversation"
	// RealtimeMessageCreated carries a newly stored message so open conversations can append it without a refetch.
	RealtimeMessageCreated  = "message"
	RealtimeDeliveryChanged = "delivery"
	// RealtimeResync tells a stream that hints may have been lost, so clients refetch everything.
	RealtimeResync = "resync"
)

// RealtimeChannelPrefix prefixes the per-tenant pub/sub channel carrying RealtimeHint values.
const RealtimeChannelPrefix = "hexta:rt:tenant:"

// RealtimeHint is a best-effort push to connected clients; PostgreSQL stays the source of truth and clients
// repair any gap by refetching. Reason "message" carries the stored message (including its body) so an open
// conversation shows it immediately; "delivery" carries only the message ID and its new delivery state (Message
// has just ID, ThreadID, DeliveryStatus and DeliveryErrorReason); "conversation" carries nothing and tells the list
// to refresh. Tenant routing is by channel, so a hint never reaches another tenant.
type RealtimeHint struct {
	ThreadID string            `json:"thread_id"`
	Reason   string            `json:"reason"`
	Message  *MessagingMessage `json:"message,omitempty"`
}

// RealtimeChannel is the pub/sub channel for one tenant.
func RealtimeChannel(tenant string) string { return RealtimeChannelPrefix + tenant }
