package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var (
	eventType    string
	fileLocation string
	sqsQueue     string
)

func main() {
	flag.StringVar(&eventType, "event", "", "event type to send")
	flag.StringVar(&fileLocation, "location", "", "location of json files")
	flag.StringVar(&sqsQueue, "queue", "", "name of the SQS queue")
	flag.Parse()

	bytes, err := ioutil.ReadFile(path.Join(fileLocation, fmt.Sprintf("%s.json", eventType)))
	if err != nil {
		log.Fatalf("error reading file: %s", err.Error())
	}
	payload := string(bytes)

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	svc := sqs.New(awsSession)

	sendMessageInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(sqsQueue),
		MessageBody: aws.String(payload),
	}

	sendMessageOutput, err := svc.SendMessage(sendMessageInput)
	if err != nil {
		log.Fatalf("error reading file: %s", err.Error())
	}

	log.Printf("MessageID: %s\n", *sendMessageOutput.MessageId)
}
