package main

import (
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid"
	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/retgits/acme-serverless-payment/internal/validator"
	"github.com/valyala/fasthttp"
)

// ValidatePayment ...
func ValidatePayment(ctx *fasthttp.RequestCtx) {
	// Unmarshal the PaymentRequested event to a struct
	req, err := acmeserverless.UnmarshalPaymentRequestedEvent(ctx.Request.Body())
	if err != nil {
		ErrorHandler(ctx, "ValidatePayment", "UnmarshalPaymentRequestedEvent", err)
		return
	}

	// If the total could not be parsed, it is likely an older Shop Payment request
	// which should be unmarshalled differently.
	legacyPaymentType := false
	if len(req.Data.Total) == 0 {
		shoppayment, err := acmeserverless.UnmarshalShopPayment(ctx.Request.Body())
		if err != nil {
			ErrorHandler(ctx, "ValidatePayment", "UnmarshalPaymentRequestedEvent", err)
			return
		}
		req.Data.Card = shoppayment.Card.ToCreditCard()
		req.Data.Total = shoppayment.Total
		legacyPaymentType = true
	}

	// Send a breadcrumb to Sentry with the validation request
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  acmeserverless.PaymentRequestedEventName,
		Timestamp: time.Now(),
		Level:     sentry.LevelInfo,
		Data:      acmeserverless.ToSentryMap(req.Data),
	})

	// Generate the event to emit
	evt := acmeserverless.CreditCardValidatedEvent{
		Metadata: acmeserverless.Metadata{
			Domain: acmeserverless.PaymentDomain,
			Source: "ValidateCreditCard",
			Type:   acmeserverless.CreditCardValidatedEventName,
			Status: "success",
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
		log.Println(err.Error())
	}

	// Send a breadcrumb to Sentry with the validation result
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  acmeserverless.CreditCardValidatedEventName,
		Timestamp: time.Now(),
		Level:     sentry.LevelInfo,
		Data:      acmeserverless.ToSentryMap(evt.Data),
	})

	// If the total could not be parsed, it is likely an older Shop Payment request
	// and should have a different response format.
	var payload []byte
	var e error
	if legacyPaymentType {
		payload, e = evt.Data.Marshal()
		if e != nil {
			ErrorHandler(ctx, "ValidatePayment", "MarshalLegacyPayment", err)
			return
		}
	} else {
		payload, e = evt.Marshal()
		if e != nil {
			ErrorHandler(ctx, "ValidatePayment", "Marshal", err)
			return
		}
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Write(payload)
}
