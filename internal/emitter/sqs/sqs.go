// Package sqs uses Amazon Simple Queue Service (SQS) as a fully managed message queuing
// service that enables you to decouple and scale microservices, distributed systems, and
// serverless fitness shops.
package sqs

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/retgits/acme-serverless-payment/internal/emitter"
)

// responder is an empty struct that implements the methods of the
// EventEmitter interface.
type responder struct{}

// New creates a new instance of the EventEmitter with SQS
// as the messaging layer.
func New() emitter.EventEmitter {
	return responder{}
}

// Send sends the event to an SQS queue. The SQS queue is determined
// by the environment variable RESPONSEQUEUE. The AWS region this code
// looks in to find the queue is determined by the environment
// variable REGION. The method returns an error if anything goes wrong.
func (r responder) Send(e acmeserverless.CreditCardValidatedEvent) error {
	payload, err := e.Marshal()
	if err != nil {
		return err
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	svc := sqs.New(awsSession)

	urlParts := strings.Split(os.Getenv("RESPONSEQUEUE"), ":")
	queue := fmt.Sprintf("https://sqs.%s.amazonaws.com/%s/%s", urlParts[3], urlParts[4], urlParts[5])

	sendMessageInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(string(payload)),
	}

	_, err = svc.SendMessage(sendMessageInput)
	if err != nil {
		return err
	}

	return nil
}
