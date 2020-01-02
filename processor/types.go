package processor

import (
	"encoding/json"

	"github.com/retgits/creditcard"
)

// PaymentRequest is the input message that the payment function expects.
type PaymentRequest struct {
	OrderID string          `json:"orderID"`
	Card    creditcard.Card `json:"card"`
	Total   string          `json:"total"`
}

// PaymentResponse is the output message that the payment function sends.
type PaymentResponse struct {
	Success       bool   `json:"success"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
	Amount        string `json:"amount,omitempty"`
	TransactionID string `json:"transactionID"`
	OrderID       string `json:"orderID"`
}

// Marshal takes a response object and creates a JSON string out of it. It returns either the string or an error
func (r *PaymentResponse) Marshal() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UnmarshalRequest transforms the data from the incoming message
func UnmarshalRequest(data []byte) (PaymentRequest, error) {
	var r PaymentRequest
	err := json.Unmarshal(data, &r)
	return r, err
}
