# Payment

> A payment service, because nothing in life is really free...

The Payment service is part of the [ACME Fitness Serverless Shop](https://github.com/retgits/acme-serverless). The goal of this specific service is to validate credit card payments. Currently the only validation performed is whether the card is acceptable.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS Account](https://portal.aws.amazon.com/billing/signup)
* The _vuln_ targets for Make and Mage rely on the [Snyk](http://snyk.io/) CLI
* This service uses [Sentry.io](https://sentry.io) for tracing and error reporting

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
git clone https://github.com/retgits/acme-serverless-payment
cd acme-serverless-payment
```

Get the Go Module dependencies

```bash
go get ./...
```

Change directories to the [deploy/cloudformation](./deploy/cloudformation) folder

```bash
cd ./deploy/cloudformation
```

If your event bus is not called _acmeserverless_, update the name of the `feature` parameter in the `template.yaml` file. Now you can build and deploy the Lambda function:

```bash
make build TYPE=eventbridge
make deploy TYPE=eventbridge
```

### Testing EventBridge

To send a message to an Amazon EventBridge eventbus, check out the [acme-serverless README](https://github.com/retgits/acme-serverless#testing-eventbridge)

### Prerequisites for SQS

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) installed and configured

### Build and deploy for SQS

Clone this repository

```bash
git clone https://github.com/retgits/acme-serverless-payment
cd acme-serverless-payment
```

Get the Go Module dependencies

```bash
go get ./...
```

Change directories to the [deploy/cloudformation](./deploy/cloudformation) folder

```bash
cd ./deploy/cloudformation
```

Now you can build and deploy the Lambda function:

```bash
make build TYPE=sqs
make deploy TYPE=sqs
```

### Testing SQS

To send a message to an SQS queue, check out the [acme-serverless README](https://github.com/retgits/acme-serverless#testing-sqs)

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

## Using Make

The Makefiles and CloudFormation templates can be found in the [acme-serverless](https://github.com/retgits/acme-serverless/tree/master/deploy/cloudformation/payment) repository

## Using Mage

If you want to "go all Go" (_pun intended_) and write plain-old go functions to build and deploy, you can use [Mage](https://magefile.org/). Mage is a make/rake-like build tool using Go so Mage automatically uses the functions you create as Makefile-like runnable targets.

The Magefile can be found in the [acme-serverless](https://github.com/retgits/acme-serverless/tree/master/deploy/mage) repository

## Using Serverless Framework

If you want to use the Serverless Framework, rather than AWS Serverless Application Model you can use the deployment scripts in the [deploy/serverless](./deploy/serverless/) folder.

The Makefile, similar to the one used for CloudFormation, has one extra command (`make plugins`) to install the required plugins for the Serverless Framework. To deploy, an additional environment variable (`AWS_ACCOUNTID`) is used, which should be the account ID to which you deploy.

## Contributing

[Pull requests](https://github.com/retgits/acme-serverless-payment/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/acme-serverless-payment/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
