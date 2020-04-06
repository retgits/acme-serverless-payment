module github.com/retgits/acme-serverless-payment

replace github.com/wavefronthq/wavefront-lambda-go => github.com/retgits/wavefront-lambda-go v0.0.0-20200406192713-6ff30b7e488c

go 1.13

require (
	github.com/aws/aws-lambda-go v1.15.0
	github.com/aws/aws-sdk-go v1.29.27
	github.com/getsentry/sentry-go v0.5.1
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pulumi/pulumi v1.12.1
	github.com/pulumi/pulumi-aws v1.26.0
	github.com/retgits/creditcard v0.6.0
	github.com/retgits/pulumi-helpers v0.1.3
	github.com/wavefronthq/wavefront-lambda-go v0.0.0-20190812171804-d9475d6695cc
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
