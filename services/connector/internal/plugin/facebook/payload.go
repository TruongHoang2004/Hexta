package facebook

import "encoding/json"

// Wire types for the Facebook Messenger webhook payload. They only decode the
// provider JSON; mapping to ingress.Event stays in Parse.

// webhookBatch is the top-level webhook body. Messaging items stay raw so one
// malformed item does not invalidate the whole batch.
type webhookBatch struct {
	Object string         `json:"object"`
	Entry  []webhookEntry `json:"entry"`
}

// webhookEntry is one Page's entry in a webhook batch.
type webhookEntry struct {
	ID        string            `json:"id"`
	Messaging []json.RawMessage `json:"messaging"`
}

// messagingEvent is one element of entry[].messaging[]. Reaction, Read and
// Delivery are kept raw only to tell the event type; they are unsupported.
type messagingEvent struct {
	Sender    participant     `json:"sender"`
	Recipient participant     `json:"recipient"`
	Timestamp int64           `json:"timestamp"`
	Message   *messagePayload `json:"message"`
	Reaction  json.RawMessage `json:"reaction"`
	Read      json.RawMessage `json:"read"`
	Delivery  json.RawMessage `json:"delivery"`
}

type participant struct {
	ID string `json:"id"`
}

type messagePayload struct {
	MID         string              `json:"mid"`
	Text        string              `json:"text"`
	Echo        bool                `json:"is_echo"`
	Reply       replyTo             `json:"reply_to"`
	Attachments []attachmentPayload `json:"attachments"`
	StickerID   json.RawMessage     `json:"sticker_id"`
}

type replyTo struct {
	MID string `json:"mid"`
}

type attachmentPayload struct {
	Type    string            `json:"type"`
	Payload attachmentContent `json:"payload"`
}

type attachmentContent struct {
	StickerID json.RawMessage `json:"sticker_id"`
}
