package processor

import (
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

// Validate performs the actual validation of the creditcard
func Validate(message string) (string, error) {
	msg, err := unmarshalRequest([]byte(message))
	if err != nil {
		log.Printf("error unmarshaling request: %s", err.Error())
		return "", err
	}

	// Validate the card and log the response
	v := msg.Card.Validate()

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
	log.Printf("All processing errors: %+v", v.Errors)

	// Send a positive reply if all checks succeed, else send a 400
	var res Response
	if v.ValidCardNumber == true && v.ValidCVV == true && v.IsExpired == false {
		log.Println("payment processed successfully")
		res = Response{
			Success:       true,
			Status:        http.StatusOK,
			Message:       "transaction successful",
			Amount:        msg.Total,
			OrderID:       msg.OrderID,
			TransactionID: uuid.Must(uuid.NewV4()).String(),
		}
	} else {
		res = Response{
			Success:       false,
			Status:        http.StatusBadRequest,
			Message:       "creditcard validation has failed, unable to process payment",
			OrderID:       msg.OrderID,
			TransactionID: "-1",
		}
	}
	resp, err := res.marshal()
	if err != nil {
		log.Printf("error marshaling response: %s", err.Error())
		return "", err
	}

	return resp, nil
}
