package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/creditcard"

	uuid "github.com/gofrs/uuid"
	wflambda "github.com/wavefronthq/wavefront-lambda-go"
)

var wfAgent = wflambda.NewWavefrontAgent(&wflambda.WavefrontConfig{})

// Request is the input message that the Lambda function expects. This has tobe a JSON string payload
// that will be unmarshalled to this struct.
type Request struct {
	OrderID string          `json:"orderID"`
	Card    creditcard.Card `json:"card"`
	Total   string          `json:"total"`
}

// Response is the output message that the Lambda function sends. It will be a JSON string payload.
type Response struct {
	Success       bool   `json:"success"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
	Amount        string `json:"amount,omitempty"`
	TransactionID string `json:"transactionID"`
	OrderID       string `json:"orderID"`
}

// config is the struct that is used to keep track of all environment variables
type config struct {
	AWSRegion     string `required:"true" split_words:"true" envconfig:"REGION"`
	ResponseQueue string `required:"true" split_words:"true" envconfig:"RESPONSE_QUEUE"`
}

var c config

func handler(request events.SQSEvent) error {
	// Get configuration set using environment variables
	err := envconfig.Process("", &c)
	if err != nil {
		log.Printf("error starting function: %s", err.Error())
		return err
	}

	// Create an AWS session
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.AWSRegion),
	}))

	// Create a SQS session
	sqsService := sqs.New(awsSession)

	for _, record := range request.Records {
		msg, err := UnmarshalRequest([]byte(record.Body))
		if err != nil {
			log.Printf("error unmarshaling request: %s", err.Error())
			break
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
		resp, _ := res.Marshal()

		sendMessageInput := &sqs.SendMessageInput{
			QueueUrl:    aws.String(c.ResponseQueue),
			MessageBody: aws.String(resp),
		}

		_, err = sqsService.SendMessage(sendMessageInput)
		if err != nil {
			log.Printf("error while sending response message: %s", err.Error())
			break
		}
	}

	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(wfAgent.WrapHandler(handler))
}

// Marshal takes a response object and creates a JSON string out of it. It returns either the string or an error
func (r *Response) Marshal() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// UnmarshalRequest transforms the data from the incoming message
func UnmarshalRequest(data []byte) (Request, error) {
	var r Request
	err := json.Unmarshal(data, &r)
	return r, err
}
