package validator

import (
	"fmt"

	"github.com/retgits/creditcard"
)

type validator struct{}

func New() validator {
	return validator{}
}

func (v *validator) Creditcard(card creditcard.Card) error {
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
