package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
)

var (
	eventType    string
	fileLocation string
	eventBus     string
)

func main() {
	flag.StringVar(&eventType, "event", "", "event type to send")
	flag.StringVar(&fileLocation, "location", "", "location of json files")
	flag.StringVar(&eventBus, "bus", "", "name of the eventbus")
	flag.Parse()

	bytes, err := ioutil.ReadFile(path.Join(fileLocation, fmt.Sprintf("%s.json", eventType)))
	if err != nil {
		log.Fatalf("error reading file: %s", err.Error())
	}
	payload := string(bytes)

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	svc := eventbridge.New(awsSession)

	entries := make([]*eventbridge.PutEventsRequestEntry, 1)
	entries[0] = &eventbridge.PutEventsRequestEntry{
		Detail:       aws.String(payload),
		DetailType:   aws.String("myDetailType"),
		EventBusName: aws.String(eventBus),
		Resources:    []*string{aws.String("TestMessage"), aws.String(eventBus)},
		Source:       aws.String("cli"),
	}
	event := &eventbridge.PutEventsInput{
		Entries: entries,
	}

	output, err := svc.PutEvents(event)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range output.Entries {
		log.Printf("EventID: %s", *entry.EventId)
	}
}
