package emitter

import "encoding/json"

import "github.com/retgits/acme-serverless-payment"

// Data ...
type Data struct {
	Success       bool   `json:"success"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
	Amount        string `json:"amount,omitempty"`
	TransactionID string `json:"transactionID"`
	OrderID       string `json:"orderID"`
}

// Event ...
type Event struct {
	Metadata payment.Metadata `json:"metadata"`
	Data     Data             `json:"data"`
}

// EventEmitter ...
type EventEmitter interface {
	Send(e Event) error
}

// Marshal returns the JSON encoding of Event.
func (e *Event) Marshal() (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
