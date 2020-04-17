# Payment
 
> A payment service, because nothing in life is really free...

The Payment service is part of the [ACME Fitness Serverless Shop](https://github.com/retgits/acme-serverless). The goal of this specific service is to validate credit card payments. Currently the only validation performed is whether the card is acceptable.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS account](https://portal.aws.amazon.com/billing/signup)
* [A Pulumi account](https://app.pulumi.com/signup)
* [A Sentry.io account](https://sentry.io) if you want to enable tracing and error reporting

## Deploying

### With Pulumi (using SQS for eventing)

To deploy the Payment Service you'll need a [Pulumi account](https://app.pulumi.com/signup). Once you have your Pulumi account and configured the [Pulumi CLI](https://www.pulumi.com/docs/get-started/aws/install-pulumi/), you can initialize a new stack using the Pulumi templates in the [pulumi](./pulumi) folder.

```bash
cd pulumi
pulumi stack init <your pulumi org>/acmeserverless-payment/dev
```

Pulumi is configured using a file called `Pulumi.dev.yaml`. A sample configuration is available in the Pulumi directory. You can rename [`Pulumi.dev.yaml.sample`](./pulumi/Pulumi.dev.yaml.sample) to `Pulumi.dev.yaml` and update the variables accordingly. Alternatively, you can change variables directly in the [main.go](./pulumi/main.go) file in the pulumi directory. The configuration contains:

```yaml
config:
  aws:region: us-west-2 ## The region you want to deploy to
  awsconfig:generic:
    sentrydsn: ## The DSN to connect to Sentry
    accountid: ## Your AWS Account ID
    wavefronturl: ## The URL of your Wavefront instance
    wavefronttoken: ## Your Wavefront API token
  awsconfig:tags:
    author: retgits ## The author, you...
    feature: acmeserverless
    team: vcs ## The team you're on
    version: 0.2.0 ## The version
```

To create the Pulumi stack, and create the Payment service, run `pulumi up`.

If you want to keep track of the resources in Pulumi, you can add tags to your stack as well.

```bash
pulumi stack tag set app:name acmeserverless
pulumi stack tag set app:feature acmeserverless-payment
pulumi stack tag set app:domain payment
```

### With CloudFormation (using EventBridge for eventing)

Clone this repository

```bash
git clone https://github.com/retgits/acme-serverless-payment
cd acme-serverless-payment
```

Get the Go Module dependencies

```bash
go get ./...
```

Change directories to the [cloudformation](./cloudformation) folder

```bash
cd ./cloudformation
```

If your event bus is not called _acmeserverless_, update the name of the `feature` parameter in the `template.yaml` file. Now you can build and deploy the Lambda function:

```bash
make build
make deploy
```

## Testing

To test, you can use the SQS or EventBridge test apps in the [acme-serverless](https://github.com/retgits/acme-serverless) repo.

## Building for Google Cloud Run

If you have Docker installed locally, you can use `docker build` to create a container which can be used to try out the payment service locally and for Google Cloud Run.

To build your container image using Docker:

Run the command:

```bash
VERSION=`git describe --tags --always --dirty="-dev"`
docker build -f ./cmd/cloudrun-payment-http/Dockerfile . -t gcr.io/[PROJECT-ID]/payment:$VERSION
```

Replace `[PROJECT-ID]` with your Google Cloud project ID

If you have not yet configured Docker to use the gcloud command-line tool to authenticate requests to Container Registry, do so now using the command:

```bash
gcloud auth configure-docker
```

You need to do this before you can push or pull images using Docker. You only need to do it once.

Push the container image to Container Registry:

```bash
docker push gcr.io/[PROJECT-ID]/payment:$VERSION
```

The container relies on the environment variables:

* SENTRY_DSN: The DSN to connect to Sentry
* K_SERVICE: The name of the service (in Google Cloud Run this variable is automatically set, defaults to `payment` if not set)
* VERSION: The version you're running (will default to `dev` if not set)
* PORT: The port number the service will listen on (will default to `8080` if not set)
* STAGE: The environment in which you're running
* WAVEFRONT_TOKEN: The token to connect to Wavefront
* WAVEFRONT_URL: The URL to connect to Wavefront (will default to `debug` if not set)

A `docker run`, with all options, is:

```bash
docker run --rm -it -p 8080:8080 -e SENTRY_DSN=abcd -e K_SERVICE=payment \
  -e VERSION=$VERSION -e PORT=8080 -e STAGE=dev -e WAVEFRONT_URL=https://my-url.wavefront.com \
  -e WAVEFRONT_TOKEN=efgh -e gcr.io/[PROJECT-ID]/payment:$VERSION
```

Replace `[PROJECT-ID]` with your Google Cloud project ID

## Contributing

[Pull requests](https://github.com/retgits/acme-serverless-payment/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/acme-serverless-payment/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
