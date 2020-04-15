// Package emitter contains the interfaces that the Payment service
// in the ACME Serverless Fitness Shop needs to send events to other
// services. In order to add a new service, the EventEmitter interface
// needs to be implemented.
package emitter

import acmeserverless "github.com/retgits/acme-serverless"

// EventEmitter is the interface that describes the methods the
// eventing service needs to implement to be able to work with
// the ACME Serverless Fitness Shop.
type EventEmitter interface {
	Send(e acmeserverless.CreditCardValidatedEvent) error
}
