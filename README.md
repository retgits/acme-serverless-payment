# Payment

> A payment service, because nothing in life is really free...

The Payment service is part of the ACME Fitness Shop. The goal of this specific service is to validate credit card payments. The validation function performs four [validations](#validations) on the card. Valid credit card numbers to test with, can be found on the [PayPal website](https://www.paypalobjects.com/en_US/vhelp/paypalmanager_help/credit_card_numbers.htm)

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [A Wavefront API token](https://wavefront.com/)

## Design

```text
     ┌─────┐                                ┌───────┐
     │Order│                                │Payment│
     └──┬──┘                                └───┬───┘
        │      Send message on SQS queue.       │
        │──────────────────────────────────────>│
        │                                       │
        │                                       ────┐
        │                                           │ Validate credit card data.
        │                                       <───┘
        │                                       │
        │Send validation response on SQS queue. │
        │<──────────────────────────────────────│
     ┌──┴──┐                                ┌───┴───┐
     │Order│                                │Payment│
     └─────┘                                └───────┘
```

## Installation

The configuration of the app relies on AWS Systems Manager Parameter Store (SSM) to store variables:

* `/<stage>/acmeserverless/payment/inboundqueue`: The name of the queue to receive messages on (*not the complete ARN*).
* `/<stage>/acmeserverless/payment/outboundqueue`: The name of the queue to send messages to (*not the complete ARN*).
* `/<stage>/acmeserverless/wavefront/url`: The URL of your Wavefront instance (like, `https://myinstance.wavefront.com`).
* `/<stage>/acmeserverless/wavefront/token`: Your Wavefront API token (see the [docs](https://docs.wavefront.com/wavefront_api.html) how to create an API token).

With the _`<stage>`_ variable, you can have different stack referencing different sets of API keys.

To create parameters, run:

```bash
aws ssm put-parameter       \
   --type String            \
   --name "<token name>"    \
   --value "<token value>"
```

## Build and Deploy

There are several `Make` targets available to help build and deploy the function

| Target | Description                                       |
|--------|---------------------------------------------------|
| build  | Build the executable for Lambda                   |
| clean  | Remove all generated files                        |
| deploy | Deploy the app to AWS Lambda                      |
| deps   | Get the Go modules from the GOPROXY               |
| help   | Displays the help for each target (this message). |
| local  | Run SAM to test the Lambda function using Docker  |
| test   | Run all unit tests and print coverage             |

## API

### Input

Validation of the credit card data is done based on the [creditcard](https://github.com/retgits/creditcard) library and can be tested using the sample card numbers can be found on the [PayPal website](https://www.paypalobjects.com/en_US/vhelp/paypalmanager_help/credit_card_numbers.htm). The body of the input message must be

```json
{
    "orderID": "12345",
    "card": {
        "number": "378282246310005",
        "expYear": "2020",
        "expMonth": "01",
        "ccv": "1234"
    },
    "total": "123"
}
```

### Output

In case the validation is successful, the response will be:

```json
{
    "success": "true",
    "status": "200",
    "message": "transaction successful",
    "amount": "123",
    "orderID": "12345",
    "transactionID": "d6495c95-df2e-489a-8b4b-a6e8bb49eb0c"
}
```

In case the validation is not successful, the response will be:

```json
{
    "success": "false",
    "status": "400",
    "message": "creditcard validation has failed, unable to process payment",
    "transactionID": "-1"
}
```

The request and the response are asynchronous and are sent on different SQS queues.

## Validations

The validation function performs four validations on the card:

1. Determines if a credit card number is valid for a given credit card type. Also verifies that the credit card number passes the Luhn algorithm.
2. Determines if a value is a valid credit card expiry month. The month must fall between the defined minimum and maximum months.
3. Determines if a value is a valid credit card expiry year. The year must fall between the defined minimum and maximum years.
4. Determines if a CVV is valid for a given credit card type. For example, American Express requires a four digit CVV, while Visa and Mastercard require a three digit CVV.

Valid credit card numbers to test with, can be found on the [PayPal website](https://www.paypalobjects.com/en_US/vhelp/paypalmanager_help/credit_card_numbers.htm)

## Contributing

[Pull requests](https://github.com/retgits/payment/pulls) are welcome. For major changes, please open an [issue first](https://github.com/retgits/payment/issues) to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
