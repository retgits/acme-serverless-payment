package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/gofrs/uuid"
	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/retgits/acme-serverless-payment/internal/validator"
	gcrwavefront "github.com/retgits/gcr-wavefront"
)

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// Set CORS headers for the preflight request
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	// Handle the Payment request
	case http.MethodPost:
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "400 - Bad request", http.StatusBadRequest)
		}
		res, err := validatePayment(bytes)
		if err != nil {
			http.Error(w, fmt.Sprintf("400 - Bad request: %s", err.Error()), http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	// Disallow all other HTTP methods
	default:
		http.Error(w, "405 - Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func validatePayment(bytes []byte) ([]byte, error) {
	// Unmarshal the PaymentRequested event to a struct
	req, _ := acmeserverless.UnmarshalPaymentRequestedEvent(bytes)

	// If the total could not be parsed, it is likely an older Shop Payment request
	// which should be unmarshalled differently.
	legacyPaymentType := false
	if len(req.Data.Total) == 0 {
		shoppayment, _ := acmeserverless.UnmarshalShopPayment(bytes)
		req.Data.Card = shoppayment.Card.ToCreditCard()
		req.Data.Total = shoppayment.Total
		legacyPaymentType = true
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
	err := check.Creditcard(req.Data.Card)
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
		Timestamp: time.Now().Unix(),
		Level:     sentry.LevelInfo,
		Data:      acmeserverless.ToSentryMap(evt.Data),
	})

	// If the total could not be parsed, it is likely an older Shop Payment request
	// and should have a different response format.
	if legacyPaymentType {
		return evt.Data.Marshal()
	}

	return evt.Marshal()
}

func main() {
	// Initiialize a connection to Sentry to capture errors and traces
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
		Transport: &sentry.HTTPSyncTransport{
			Timeout: time.Second * 3,
		},
		ServerName:  os.Getenv("K_SERVICE"),
		Release:     os.Getenv("VERSION"),
		Environment: os.Getenv("STAGE"),
	}); err != nil {
		log.Fatalf("error configuring sentry: %s", err.Error())
	}

	// Create an instance of sentryhttp
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	// Set configuration parameters
	wfconfig := &gcrwavefront.WavefrontConfig{
		Server:        os.Getenv("WAVEFRONT_URL"),
		Token:         os.Getenv("WAVEFRONT_TOKEN"),
		BatchSize:     10000,
		MaxBufferSize: 50000,
		FlushInterval: 1,
		Source:        "acmeserverless",
		MetricPrefix:  "acmeserverless.gcr.payment",
		PointTags:     make(map[string]string),
	}

	// Initialize the Wavefront sender
	if err := wfconfig.ConfigureSender(); err != nil {
		log.Fatalf("error configuring wavefront: %s", err.Error())
	}

	// Wrap the sentryHandler with the Wavefront middleware to make sure all events
	// are sent to sentry before sending data to Wavefront
	http.HandleFunc("/", wfconfig.WrapHandlerFunc(sentryHandler.HandleFunc(handler)))

	// Get the port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("start payment server on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
