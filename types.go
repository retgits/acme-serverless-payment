// Package payment contains all events that the Payment service
// in the ACME Serverless Fitness Shop can send and receive.
package payment

import (
	"encoding/json"

	"github.com/retgits/creditcard"
)

const (
	// Domain is the domain where the services reside
	Domain = "Payment"

	// CreditCardValidatedEvent is the event name of CreditCardValidated
	CreditCardValidatedEvent = "CreditCardValidated"

	// PaymentRequestedEvent is the event name of PaymentRequested
	PaymentRequestedEvent = "PaymentRequestedEvent"

	// DefaultSuccessMessage is the default success message used
	DefaultSuccessMessage = "transaction successful"

	// DefaultErrorMessage is the default error message used
	DefaultErrorMessage = "creditcard validation has failed, unable to process payment"
)

// PaymentRequested is the event sent by the Order service when the creditcard
// for the order should be validated and charged.
type PaymentRequested struct {
	// Metadata for the event.
	Metadata Metadata `json:"metadata"`

	// Data contains the payload data for the event.
	Data PaymentRequest `json:"data"`
}

// CreditCardValidated is the event sent by the payment service when the creditcard
// has been validated.
type CreditCardValidated struct {
	// Metadata for the event.
	Metadata Metadata `json:"metadata"`

	// Data contains the payload data for the event.
	Data PaymentData `json:"data"`
}

// Metadata is an envelope containing information on the domain, source, type, and status
// of the event.
type Metadata struct {
	// Domain represents the the event came from
	// like Payment or Order.
	Domain string `json:"domain"`

	// Source represents the function the event came from
	// like ValidateCreditCard or SubmitOrder.
	Source string `json:"source"`

	// Type respresents the type of event this is
	// like CreditCardValidated.
	Type string `json:"type"`

	// Status represents the current status of the event
	// like Success or Failure.
	Status string `json:"status"`
}

// PaymentData is the data that the payment service emits.
type PaymentData struct {
	// Indicates whether the transaction was a success or not.
	Success bool `json:"success"`

	// The HTTP statuscode of the event.
	Status int `json:"status"`

	// A string containing the result of the service.
	Message string `json:"message"`

	// The monetary amount of the transaction.
	Amount string `json:"amount,omitempty"`

	// The unique identifier of the transaction.
	TransactionID string `json:"transactionID"`

	// The unique identifier of the order.
	OrderID string `json:"orderID"`
}

// PaymentRequest is the data that the order service emits.
type PaymentRequest struct {
	// The unique identifier of the order.
	OrderID string `json:"orderID"`

	// Card is the creditcard information.
	Card creditcard.Card `json:"card"`

	// The monetary amount of the transaction.
	Total string `json:"total"`
}

// UnmarshalPaymentRequested parses the JSON-encoded data and stores the result in a
// PaymentRequestedEvent.
func UnmarshalPaymentRequested(data []byte) (PaymentRequested, error) {
	var r PaymentRequested
	err := json.Unmarshal(data, &r)
	return r, err
}

// Marshal returns the JSON encoding of PaymentRequested.
func (e *PaymentRequested) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalCreditCardValidated parses the JSON-encoded data and stores the result in a
// CreditCardValidated.
func UnmarshalCreditCardValidated(data []byte) (CreditCardValidated, error) {
	var r CreditCardValidated
	err := json.Unmarshal(data, &r)
	return r, err
}

// Marshal returns the JSON encoding of CreditCardValidated.
func (e *CreditCardValidated) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// ToMap returns a map[string]interface of the PaymentRequest object so it can
// be sent to Sentry. The keys of the map are the same as the JSON element
// names.
func (pd *PaymentRequest) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["orderID"] = pd.OrderID
	m["total"] = pd.Total

	return m
}

// ToMap returns a map[string]interface of the PaymentData object so it can
// be sent to Sentry. The keys of the map are the same as the JSON element
// names.
func (pd *PaymentData) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["success"] = pd.Success
	m["status"] = pd.Status
	m["message"] = pd.Message
	m["amount"] = pd.Amount
	m["transactionID"] = pd.TransactionID
	m["orderID"] = pd.OrderID

	return m
}
