# Payment

> A payment service, because nothing in life is really free...

The Payment service is part of the [ACME Fitness Shop](https://github.com/vmwarecloudadvocacy/acme_fitness_demo). The goal of this specific service is to validate credit card payments. Currently the only validation performed is whether the card is acceptable.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS Account](https://portal.aws.amazon.com/billing/signup)
* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) and [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) installed and configured

## Quick start

Provided you already have the AWS CLI and the AWS SAM CLI installed and configured, you can run:

```bash
git clone https://github.com/retgits/payment
cd payment
make deps
```

Update the `Makefile` and change the variables with values that match your preferred setup:

```make
## The stage to deploy to
stage         	:= dev

## The name of the user in GitHub (also used as author in CloudFormation tags)
github_user   	:= retgits

## The name of the team
team			:= vcs

## The name of the project, defaults to the name of the current directory
project_name  	:= $(notdir $(CURDIR))

## The version of the project, either uses the current commit hash, or will default to "dev"
version       	:= $(strip $(if $(shell git describe --tags --always --dirty="-dev"),$(shell git describe --tags --always --dirty="-dev"),dev))

## The Amazon S3 bucket to upload files to
aws_bucket    	?= $$S3_BUCKET
```

Deploy the function into your AWS account:

```
make deploy
```

## Test

You can test the function from the [AWS Lambda Console](https://console.aws.amazon.com/lambda/home) using the test data from the files [`lambda-console-failure.json`](./test/lambda-console-failure.json) and [`lambda-console-success.json`](./test/lambda-console-success.json). Alternatively, you can publish a message to the _PaymentRequestQueue_ created during deployment using the payload from either [`event.json`](./test/event.json) or [`failure.json`](./test/failure.json).

## Events

![payment](./payment.png)

The function accepts events from SQS on the `PaymentRequestQueue`. These events should contain a payment request, like:

```json
{
    "orderID": "12345",
    "card": {
        "Type": "Visa",
        "Number": "4222222222222",
        "ExpiryYear": 2016,
        "ExpiryMonth": 12,
        "CVV": "123"
    },
    "total": "123"
}
```

Credit card numbers to test with cab be found on the [PayPal](http://www.paypalobjects.com/en_US/vhelp/paypalmanager_help/credit_card_numbers.htm) website.

Whether the validation succeeds or fails, a response is sent to the `PaymentResponseQueue`, with the payload looking like:

```json
{
  "success": "true",
  "status": "200",
  "message": "transaction successful",
  "amount": 123,
  "transactionID": "3f846704-af12-4ea9-a98c-8d7b37e10b54"
}
```

When the card fails to validate, an error message is sent back. More details on why validation has failed is available in the logs:

```json
{
  "success": "false",
  "status": "400",
  "message": "creditcard validation has failed, unable to process payment",
  "amount": "0",
  "transactionID": "-1"
}
```

In case of any errors while processing the message, the message and the event are sent to the `PaymentErrorQueue`.

## Using `Make`

Most of the actions to build and run the app are captured in a [Makefile](./Makefile)

| Target  | Description                                                |
|---------|------------------------------------------------------------|
| build   | Build the executable for Lambda                            |
| clean   | Remove all generated files                                 |
| deploy  | Deploy the app to AWS Lambda                               |
| deps    | Get the Go modules from the GOPROXY                        |
| destroy | Deletes the CloudFormation stack and all created resources |
| help    | Displays the help for each target (this message)           |
| local   | Run SAM to test the Lambda function using Docker           |
| test    | Run all unit tests and print coverage                      |

## Contributing

[Pull requests](https://github.com/retgits/payment/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/payment/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
