package order

import (
	"encoding/json"
	"testing"
)

func TestOrderConfirmed(t *testing.T) {
	e := OrderConfirmed{OrderID: "o", OrderNumber: "SO-1", Version: 2, TotalAmount: "12.50", ConfirmedBy: "owner"}
	b, _ := json.Marshal(e)
	var got OrderConfirmed
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != "o" || got.TotalAmount != "12.50" {
		t.Fatal("round trip")
	}
	missing := e
	missing.TotalAmount = ""
	if missing.Validate() == nil {
		t.Fatal("amount required")
	}
}
