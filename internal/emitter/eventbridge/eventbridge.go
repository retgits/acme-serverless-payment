// Package eventbridge uses Amazon EventBridge as a serverless event bus that makes it easy to connect
// applications together using data from your own applications, integrated Software-as-a-Service (SaaS)
// applications, and Serverless Fitness Shops.
package eventbridge

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	payment "github.com/retgits/acme-serverless-payment"
	"github.com/retgits/acme-serverless-payment/internal/emitter"
)

// responder is an empty struct that implements the methods of the
// EventEmitter interface
type responder struct{}

// New creates a new instance of the EventEmitter with EventBridge
// as the messaging layer.
func New() emitter.EventEmitter {
	return responder{}
}

// Send sends the event to an EventBridge bus. The bus is determined
// by the environment variable EVENTBUS. The AWS region this code
// looks in to find the queue is determined by the environment
// variable REGION. The method returns an error if anything goes wrong.
func (r responder) Send(e payment.CreditCardValidated) error {
	payload, err := e.Marshal()
	if err != nil {
		return err
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	svc := eventbridge.New(awsSession)

	entries := make([]*eventbridge.PutEventsRequestEntry, 1)

	entries[0] = &eventbridge.PutEventsRequestEntry{
		Detail:       aws.String(string(payload)),
		EventBusName: aws.String(os.Getenv("EVENTBUS")),
		Source:       aws.String(e.Metadata.Source),
	}

	event := &eventbridge.PutEventsInput{
		Entries: entries,
	}

	_, err = svc.PutEvents(event)
	if err != nil {
		return err
	}

	return nil
}
