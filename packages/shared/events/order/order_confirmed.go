// Package order holds sales order events.
package order

import "errors"

// OrderConfirmedTopic carries OrderConfirmed inside events.Envelope. Producer: API sales (outbox). No consumer yet.
const OrderConfirmedTopic = "hexta.orders.events.v1"

// OrderConfirmedEventType is the envelope event_type and the API outbox row type.
const OrderConfirmedEventType = "order.confirmed"

// OrderConfirmed is emitted once per confirmation. It carries no customer contact data. TotalAmount is a decimal
// string, never a float.
type OrderConfirmed struct {
	OrderID     string `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Version     int64  `json:"version"`
	TotalAmount string `json:"total_amount"`
	ConfirmedBy string `json:"confirmed_by"`
}

// Key orders events per order.
func (e OrderConfirmed) Key() string { return e.OrderID }

func (e OrderConfirmed) Validate() error {
	if e.OrderID == "" || e.OrderNumber == "" || e.Version < 1 || e.TotalAmount == "" || e.ConfirmedBy == "" {
		return errors.New("invalid order confirmation")
	}
	return nil
}
