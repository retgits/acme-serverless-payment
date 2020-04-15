// Package main is a payment service, because nothing in life is really free...
//
// The Payment service is part of the [ACME Fitness Serverless Shop](https://github.com/retgits/acme-serverless).
// The goal of this specific service is to validate credit card payments. Currently the only validation performed
// is whether the card is acceptable. After completing validation a "CreditCardValidated" event is sent.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid"
	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/retgits/acme-serverless-payment/internal/emitter/eventbridge"
	"github.com/retgits/acme-serverless-payment/internal/validator"
	wflambda "github.com/wavefronthq/wavefront-lambda-go"
)

// handler handles the EventBridge events and returns an error if anything goes wrong.
// The resulting event, if no error is thrown, is sent to an EventBridge bus.
func handler(request json.RawMessage) error {
	// Initiialize a connection to Sentry to capture errors and traces
	sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Transport: &sentry.HTTPSyncTransport{
			Timeout: time.Second * 3,
		},
		ServerName:  os.Getenv("FUNCTION_NAME"),
		Release:     os.Getenv("VERSION"),
		Environment: os.Getenv("STAGE"),
	})

	// Unmarshal the PaymentRequested event to a struct
	req, err := acmeserverless.UnmarshalPaymentRequestedEvent(request)
	if err != nil {
		return handleError("unmarshaling payment", err)
	}

	// Send a breadcrumb to Sentry with the validation request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  acmeserverless.PaymentRequestedEventName,
		Timestamp: time.Now().Unix(),
		Level:     sentry.LevelInfo,
		Data:      acmeserverless.ToSentryMap(req.Data),
	})

	// Generate the event to emit
	evt := acmeserverless.CreditCardValidatedEvent{
		Metadata: acmeserverless.Metadata{
			Domain: acmeserverless.PaymentDomain,
			Source: "ValidateCreditCard",
			Type:   acmeserverless.CreditCardValidatedEventName,
			Status: acmeserverless.DefaultSuccessStatus,
		},
		Data: acmeserverless.CreditCardValidationDetails{
			Success:       true,
			Status:        http.StatusOK,
			Message:       acmeserverless.DefaultSuccessStatus,
			Amount:        req.Data.Total,
			OrderID:       req.Data.OrderID,
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		},
	}

	// Check the creditcard is valid.
	// If the creditcard is not valid, update the event to emit
	// with new information
	check := validator.New()
	err = check.Creditcard(req.Data.Card)
	if err != nil {
		evt.Metadata.Status = "error"
		evt.Data.Success = false
		evt.Data.Status = http.StatusBadRequest
		evt.Data.Message = acmeserverless.DefaultErrorStatus
		evt.Data.TransactionID = "-1"
		handleError("validating creditcard", err)
	}

	// Send a breadcrumb to Sentry with the validation result
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  acmeserverless.CreditCardValidatedEventName,
		Timestamp: time.Now().Unix(),
		Level:     sentry.LevelInfo,
		Data:      acmeserverless.ToSentryMap(evt.Data),
	})

	// Create a new EventBridgee EventEmitter and send the event
	em := eventbridge.New()
	err = em.Send(evt)
	if err != nil {
		return handleError("sending event", err)
	}

	sentry.CaptureMessage(fmt.Sprintf("validation successful for order %s", req.Data.OrderID))

	return nil
}

// handleError takes the activity where the error occured and the error object and sends a message to sentry.
// The original error is returned so it can be thrown.
func handleError(activity string, err error) error {
	log.Printf("error %s: %s", activity, err.Error())
	sentry.CaptureException(fmt.Errorf("error %s: %s", activity, err.Error()))
	return err
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(wflambda.Wrapper(handler))
}
