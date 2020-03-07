// Package validator validates the creditcard and leverages the
// github.com/retgits/creditcard module to check expiry, creditcard
// number, and cvv code.
package validator

import (
	"fmt"

	"github.com/retgits/creditcard"
)

type check struct{}

// New creates a new instance of the validator.
func New() check {
	return check{}
}

// Creditcard validates the creditcard using the logic from github.com/retgits/creditcard.
func (v *check) Creditcard(card creditcard.Card) error {
	// Validate the card
	res := card.Validate()

	if res.IsExpired == true {
		return fmt.Errorf("creditcard has expired")
	}

	if res.ValidCardNumber == false {
		return fmt.Errorf("creditcard number is not valid")
	}

	if res.ValidExpiryMonth == false || res.ValidExpiryYear == false {
		return fmt.Errorf("creditcard expiration is not valid")
	}

	if res.ValidCVV == false {
		return fmt.Errorf("creditcard cvv is not valid")
	}

	if res.ValidCardNumber == false {
		return fmt.Errorf("creditcard cvv is not valid")
	}

	return nil
}
