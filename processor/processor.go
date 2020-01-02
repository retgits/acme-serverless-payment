package processor

import (
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

// Validate performs the actual validation of the creditcard
func Validate(pr PaymentRequest) PaymentResponse {
	// Validate the card and log the response
	v := pr.Card.Validate()

	if v.IsExpired == true {
		log.Println("creditcard has expired")
	}

	if v.ValidCardNumber == false {
		log.Println("creditcard number is not valid")
	}

	if v.ValidExpiryMonth == false || v.ValidExpiryYear == false {
		log.Println("creditcard expiration is not valid")
	}

	if v.ValidCVV == false {
		log.Println("creditcard cvv is not valid")
	}

	if v.ValidCardNumber == false {
		log.Println("creditcard cvv is not valid")
	}

	// Print all errors
	if len(v.Errors) > 0 {
		log.Printf("Processing of payment failed due to: %+v", v.Errors)
		return PaymentResponse{
			Success:       false,
			Status:        http.StatusBadRequest,
			Message:       "creditcard validation has failed, unable to process payment",
			OrderID:       pr.OrderID,
			TransactionID: "-1",
		}
	}

	log.Println("payment processed successfully")
	return PaymentResponse{
		Success:       true,
		Status:        http.StatusOK,
		Message:       "transaction successful",
		Amount:        pr.Total,
		OrderID:       pr.OrderID,
		TransactionID: uuid.Must(uuid.NewV4()).String(),
	}
}
