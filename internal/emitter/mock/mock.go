package mock

import (
	"log"

	"github.com/retgits/acme-serverless-payment/internal/emitter"
)

type responder struct{}

func New() emitter.EventEmitter {
	return responder{}
}

func (r responder) Send(e emitter.Event) error {
	payload, err := e.Marshal()
	if err != nil {
		return err
	}

	log.Printf("Payload: %s", payload)

	return nil
}
