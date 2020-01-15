package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gofrs/uuid"
	"github.com/retgits/payment"
	"github.com/retgits/payment/internal/emitter"
	"github.com/retgits/payment/internal/emitter/eventbridge"
	"github.com/retgits/payment/internal/validator"
)

func handler(request json.RawMessage) error {
	req, err := payment.UnmarshalPaymentEvent(request)
	if err != nil {
		return err
	}

	em := eventbridge.New()
	evt := emitter.Event{
		Metadata: payment.Metadata{
			Domain: "Payment",
			Source: "ValidateCreditCard",
			Type:   "CreditCardValidated",
			Status: "success",
		},
		Data: emitter.Data{
			Success:       true,
			Status:        http.StatusOK,
			Message:       "transaction successful",
			Amount:        req.Request.Total,
			OrderID:       req.Request.OrderID,
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		},
	}

	validator := validator.New()
	err = validator.Creditcard(req.Request.Card)
	if err != nil {
		log.Printf("Validation failed: %s", err.Error())
		evt.Metadata.Status = "error"
		evt.Data.Success = false
		evt.Data.Status = http.StatusBadRequest
		evt.Data.Message = "creditcard validation has failed, unable to process payment"
		evt.Data.TransactionID = "-1"
	}

	return em.Send(evt)
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
