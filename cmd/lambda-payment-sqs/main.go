package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid"
	payment "github.com/retgits/acme-serverless-payment"
	"github.com/retgits/acme-serverless-payment/internal/emitter"
	"github.com/retgits/acme-serverless-payment/internal/emitter/sqs"
	"github.com/retgits/acme-serverless-payment/internal/validator"
)

func handler(request events.SQSEvent) error {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Transport:   sentrySyncTransport,
		ServerName:  os.Getenv("FUNCTION_NAME"),
		Release:     os.Getenv("VERSION"),
		Environment: os.Getenv("STAGE"),
	})

	req, err := payment.UnmarshalPaymentEvent([]byte(request.Records[0].Body))
	if err != nil {
		sentry.CaptureException(fmt.Errorf("error unmarshalling payment: %s", err.Error()))
		return err
	}

	em := sqs.New()
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
		sentry.CaptureException(fmt.Errorf("validation failed for order [%s] : %s", req.Request.OrderID, err.Error()))
		log.Printf("Validation failed: %s", err.Error())
		evt.Metadata.Status = "error"
		evt.Data.Success = false
		evt.Data.Status = http.StatusBadRequest
		evt.Data.Message = "creditcard validation has failed, unable to process payment"
		evt.Data.TransactionID = "-1"
	}

	err = em.Send(evt)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("error sending CreditCardValidated event: %s", err.Error()))
		return err
	}

	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
