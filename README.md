# Payment

> A payment service, because nothing in life is really free...

The Payment service is part of the [ACME Fitness Serverless Shop](https://github.com/vmwarecloudadvocacy/acme_fitness_demo). The goal of this specific service is to validate credit card payments. Currently the only validation performed is whether the card is acceptable.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS Account](https://portal.aws.amazon.com/billing/signup)

## Eventing Options

The payment service has a few different eventing platforms available:

* [Amazon EventBridge](https://aws.amazon.com/eventbridge/)
* [Amazon Simple Queue Service](https://aws.amazon.com/sqs/)

For all options there is a Lambda function ready to be deployed where events arrive from and are sent to that particular eventing mechanism. You can find the code in `./cmd/lambda-<event platform>`. There is a ready made test function available as well that sends a message to the eventing platform. The code for the tester can be found in `./cmd/test-<event platform>`. The messages the testing app sends, are located under the [`test`](./test) folder.

## Using Amazon EventBridge

### Prerequisites for EventBridge

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) installed and configured
* [Custom EventBus](https://docs.aws.amazon.com/eventbridge/latest/userguide/create-event-bus.html) configured, the name of the configured event bus should be set as the `feature` parameter in the `template.yaml` file.

### Build and deploy for EventBridge

Clone this repository

```bash
git clone https://github.com/retgits/payment
cd payment
```

Get the Go Module dependencies

```bash
go get ./...
```

Switch directories to the EventBridge folder

```bash
cd ./cmd/lambda-eventbridge
```

If your event bus is not called _acmeserverless_, update the name of the `feature` parameter in the `template.yaml` file. Now you can build and deploy the Lambda function:

```bash
make build
make deploy
```

### Testing EventBridge

You can test the function from the [AWS Lambda Console](https://console.aws.amazon.com/lambda/home) using the test data from the files in [eventbridge](./test/eventbridge/). To send a message to the event bus, you can use the app in `./cmd/test-eventbridge` and run

```bash
go run main.go -event=<any of the files existing in test/eventbridge> -location=<location on disk of the test/eventbridge folder> -bus=<name of the custom bus>
```



### Prerequisites for SQS

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) installed and configured

### Build and deploy for SQS

Clone this repository

```bash
git clone https://github.com/retgits/payment
cd payment
```

Get the Go Module dependencies

```bash
go get ./...
```

Switch directories to the SQS folder

```bash
cd ./cmd/lambda-sqs
```

Now you can build and deploy the Lambda function:

```bash
make build
make deploy
```

### Testing SQS

To send a message to an SQS queue using the test data from the files in [sqs](./test/sqs/), you can use the app in `./cmd/test-sqs` and run

```bash
go run main.go -event=<any of the files existing in test/sqs> -location=<location on disk of the test/sqs folder> -queue=<name of the sqs queue>
```

If you want to test from the [AWS Lambda Console](https://console.aws.amazon.com/lambda/home), you'll have to wrap the test data in a SQS record envelop:

```json
{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "receiptHandle": "MessageReceiptHandle",
      "body": "", // This is where the data, an escaped JSON string, should be pasted
      "attributes": {
        "ApproximateReceiveCount": "1",
        "SentTimestamp": "1523232000000",
        "SenderId": "123456789012",
        "ApproximateFirstReceiveTimestamp": "1523232000001"
      },
      "messageAttributes": {},
      "md5OfBody": "7b270e59b47ff90a553787216d55d91d",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:MyQueue",
      "awsRegion": "us-east-1"
    }
  ]
}
```

## Events

The events for all of ACME Serverless Fitness Shop are structured as

```json
{
    "metadata": { // Metadata for all services
        "domain": "Order", // Domain represents the the event came from (like Payment or Order)
        "source": "CLI", // Source represents the function the event came from (like ValidateCreditCard or SubmitOrder)
        "type": "PaymentRequested", // Type respresents the type of event this is (like CreditCardValidated)
        "status": "success" // Status represents the current status of the event (like Success)
    },
    "data": {} // The actual payload of the event
}
```

The input that the function expects, either as the direct message or after transforming the JSON is

```json
{
    "metadata": {
        "domain": "Order",
        "source": "CLI",
        "type": "PaymentRequested", // When using EventBridge, the deployment creates a rule that triggers the function for events where the type is set to PaymentRequested
        "status": "success"
    },
    "data": {
        "orderID": "12345",
        "card": {
            "Type": "Visa",
            "Number": "4222222222222",
            "ExpiryYear": 2022,
            "ExpiryMonth": 12,
            "CVV": "123"
        },
        "total": "123"
    }
}
```

Credit card numbers to test with can be found on the [PayPal](http://www.paypalobjects.com/en_US/vhelp/paypalmanager_help/credit_card_numbers.htm) website.

A successful validation results in the below event being sent

```json
{
    "metadata": {
        "domain": "Payment",
        "source": "ValidateCreditCard",
        "type": "CreditCardValidated",
        "status": "success"
    },
    "data": {
        "success": "true",
        "status": 200,
        "message": "transaction successful",
        "amount": 123,
        "transactionID": "3f846704-af12-4ea9-a98c-8d7b37e10b54"
    }
}
```

When the card fails to validate, an error event is sent. More details on why validation has failed is available in the logs:

```json
{
    "metadata": {
        "domain": "Payment",
        "source": "ValidateCreditCard",
        "type": "CreditCardValidated",
        "status": "error"
    },
    "data": {
        "success": "false",
        "status": 400,
        "message": "creditcard validation has failed, unable to process payment",
        "amount": "0",
        "transactionID": "-1"
    }
}
```

## Contributing

[Pull requests](https://github.com/retgits/payment/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/payment/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
