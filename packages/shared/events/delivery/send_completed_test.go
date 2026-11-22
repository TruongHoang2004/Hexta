package delivery

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/google/uuid"
)

func completed(status, category string) SendCompleted {
	return SendCompleted{Meta: events.NewMeta(uuid.NewString(), "", time.Now()), IntentID: uuid.NewString(), AttemptID: uuid.NewString(), ConnectionID: uuid.NewString(), Status: status, ErrorCategory: category}
}

func TestSendCompleted(t *testing.T) {
	ok := completed(StatusSent, "")
	b, _ := json.Marshal(ok)
	var got SendCompleted
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != ok.IntentID {
		t.Fatal("round trip")
	}
	for _, e := range []SendCompleted{completed(StatusFailed, ErrorBotBlocked), completed(StatusUnknown, ErrorTransport)} {
		if e.Validate() != nil {
			t.Fatal(e.Status)
		}
	}
	long := completed(StatusFailed, ErrorTransport)
	long.ErrorCode = strings.Repeat("e", 65)
	for name, e := range map[string]SendCompleted{
		"sent with error":     completed(StatusSent, ErrorTransport),
		"failed without why":  completed(StatusFailed, ""),
		"unknown status":      completed("queued", ""),
		"long error code":     long,
		"non uuid connection": func() SendCompleted { c := completed(StatusSent, ""); c.ConnectionID = "x"; return c }(),
	} {
		if e.Validate() == nil {
			t.Error(name)
		}
	}
}
