# Payment

> A payment service, because nothing in life is really free...

The Payment service is part of the [ACME Fitness Serverless Shop](https://github.com/retgits/acme-serverless). The goal of this specific service is to validate credit card payments. Currently the only validation performed is whether the card is acceptable.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS account](https://portal.aws.amazon.com/billing/signup)
* [A Pulumi account](https://app.pulumi.com/signup)
* [A Sentry.io account](https://sentry.io) if you want to enable tracing and error reporting

## Deploying

To deploy the Payment Service you'll need a [Pulumi account](https://app.pulumi.com/signup). Once you have your Pulumi account and configured the [Pulumi CLI](https://www.pulumi.com/docs/get-started/aws/install-pulumi/), you can initialize a new stack using the Pulumi templates in the [pulumi](./pulumi) folder.

```bash
cd pulumi
pulumi stack init <your pulumi org>/acmeserverless-payment/dev
```

The [Pulumi.dev.yaml](./pulumi/Pulumi.dev.yaml) file contains all configuration data to deploy the app:

```yaml
config:
  aws:region: us-west-2 ## The region you want to deploy to
  awsconfig:lambda:
    type: sqs ## The event source you want to use (either sqs or eventbridge)
    bucket: mybucket ## The bucket in which you want to store the Lambda code
    responsequeue: ## The ARN of the Payment Response SQS queue (which you can create using the Pulumi deployment in the acme-serverless repo)
    requestqueue: ## The ARN of the Payment Request SQS queue (which you can create using the Pulumi deployment in the acme-serverless repo)
    eventbus: ## The name of the EventBridge custom event bus (which you can create using the scripts in the acme-serverless repo)
    region: us-west-2 ## The region you want to deploy to
    sentrydsn: ## The DSN to connect to Sentry
  awsconfig:tags:
    author: retgits ## The author, you...
    feature: acmeserverless
    team: vcs ## The team you're on
    version: 0.1.0 ## The version
```

To create the Pulumi stack, and create the Payment service, run `pulumi up`.

To test, you can use the SQS or EventBridge test apps in the [acme-serverless](https://github.com/retgits/acme-serverless) repo.

## Contributing

[Pull requests](https://github.com/retgits/acme-serverless-payment/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/acme-serverless-payment/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
