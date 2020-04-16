module github.com/retgits/acme-serverless-payment

replace github.com/wavefronthq/wavefront-lambda-go => github.com/retgits/wavefront-lambda-go v0.0.0-20200406192713-6ff30b7e488c

go 1.13

require (
	github.com/aws/aws-lambda-go v1.16.0
	github.com/aws/aws-sdk-go v1.30.7
	github.com/caio/go-tdigest v3.1.0+incompatible // indirect
	github.com/getsentry/sentry-go v0.5.1
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pulumi/pulumi-aws/sdk v1.31.0 // indirect
	github.com/pulumi/pulumi/sdk v1.14.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/retgits/acme-serverless v0.3.0
	github.com/retgits/creditcard v0.6.0
	github.com/retgits/gcr-wavefront v0.1.0
	github.com/retgits/pulumi-helpers v0.1.7
	github.com/wavefronthq/go-metrics-wavefront v1.0.2 // indirect
	github.com/wavefronthq/wavefront-lambda-go v0.9.0
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
