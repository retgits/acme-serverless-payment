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
	payment "github.com/retgits/acme-serverless-payment"
	"github.com/retgits/acme-serverless-payment/internal/emitter"
	"github.com/retgits/acme-serverless-payment/internal/emitter/eventbridge"
	"github.com/retgits/acme-serverless-payment/internal/validator"
)

func handler(request json.RawMessage) error {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Transport:   sentrySyncTransport,
		ServerName:  os.Getenv("FUNCTION_NAME"),
		Release:     os.Getenv("VERSION"),
		Environment: os.Getenv("STAGE"),
	})

	req, err := payment.UnmarshalPaymentEvent(request)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("error unmarshalling payment: %s", err.Error()))
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

	crumb := sentry.Breadcrumb{
		Category:  "CreditCardValidated",
		Timestamp: time.Now().Unix(),
		Level:     sentry.LevelInfo,
		Data: map[string]interface{}{
			"Amount":  req.Request.Total,
			"OrderID": req.Request.OrderID,
			"Success": true,
			"Message": "transaction successful",
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
		crumb.Data["Success"] = evt.Data.Success
		crumb.Data["Message"] = evt.Data.Message
		sentry.CaptureException(fmt.Errorf("validation failed for order [%s] : %s", req.Request.OrderID, err.Error()))
		log.Printf("Validation failed: %s", err.Error())
	}

	sentry.AddBreadcrumb(&crumb)

	err = em.Send(evt)
	if err != nil {
		sentry.CaptureException(fmt.Errorf("error sending CreditCardValidated event: %s", err.Error()))
		return err
	}

	sentry.CaptureMessage(fmt.Sprintf("validation successful for order [%s]", req.Request.OrderID))

	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
