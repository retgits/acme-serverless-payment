package processor

import (
	"encoding/json"

	"github.com/retgits/creditcard"
)

// Request is the input message that the Lambda function expects. This has to be a JSON string payload
// that will be unmarshalled to this struct.
type Request struct {
	OrderID string          `json:"orderID"`
	Card    creditcard.Card `json:"card"`
	Total   string          `json:"total"`
}

// Response is the output message that the Lambda function sends. It will be a JSON string payload.
type Response struct {
	Success       bool   `json:"success"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
	Amount        string `json:"amount,omitempty"`
	TransactionID string `json:"transactionID"`
	OrderID       string `json:"orderID"`
}

// Marshal takes a response object and creates a JSON string out of it. It returns either the string or an error
func (r *Response) marshal() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// unmarshalRequest transforms the data from the incoming message
func unmarshalRequest(data []byte) (Request, error) {
	var r Request
	err := json.Unmarshal(data, &r)
	return r, err
}
