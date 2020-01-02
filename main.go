package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/payment/processor"
)

var svc *sqs.SQS
var c config

// config is the struct that is used to keep track of all environment variables
type config struct {
	ResponseQueue string `required:"true" split_words:"true" envconfig:"RESPONSE_QUEUE"`
	ErrorQueue    string `required:"true" split_words:"true" envconfig:"ERROR_QUEUE"`
}

func init() {
	svc = sqs.New(session.New())
	err := envconfig.Process("", &c)
	if err != nil {
		panic(err)
	}
}

func sendMessage(queue string, payload string) error {
	sendMessageInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queue),
		MessageBody: aws.String(payload),
	}

	_, err := svc.SendMessage(sendMessageInput)
	if err != nil {
		return err
	}

	return nil
}

// handler handles the incoming SNS event, validates the PaymentRequest, and
// returns the result using Lambda Destinations.
func handler(request events.SQSEvent) error {
	msg, err := processor.UnmarshalRequest([]byte(request.Records[0].Body))
	if err != nil {
		return sendMessage(c.ErrorQueue, fmt.Sprintf("error unmarshalling request: %s\noriginal event: %s", err.Error(), request.Records[0].Body))
	}

	pr := processor.Validate(msg)
	payload, err := pr.Marshal()
	if err != nil {
		return sendMessage(c.ErrorQueue, fmt.Sprintf("error marshalling response: %s", err.Error()))
	}

	return sendMessage(c.ResponseQueue, payload)
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
