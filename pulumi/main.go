package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

const (
	// The shell to use
	shell = "sh"

	// The flag for the shell to read commands from a string
	shellFlag = "-c"
)

// Tags are key-value pairs to apply to the resources created by this stack
type Tags struct {
	// Author is the person who created the code, or performed the deployment
	Author pulumi.String

	// Feature is the project that this resource belongs to
	Feature pulumi.String

	// Team is the team that is responsible to manage this resource
	Team pulumi.String

	// Version is the version of the code for this resource
	Version pulumi.String
}

// LambdaConfig contains the key-value pairs for the configuration of AWS Lambda in this stack
type LambdaConfig struct {
	// The eventing layer to use
	EventingType string `json:"type"`

	// The S3 bucket to upload the compiled and zipped code to
	S3Bucket string `json:"bucket"`

	// The SQS queue to send responses to
	PaymentResponseQueue string `json:"responsequeue"`

	// The SQS queue to receives messages from
	PaymentRequestQueue string `json:"requestqueue"`

	// The EventBus to send responses to
	EventBus string `json:"eventbus"`

	// The AWS region used
	Region string `json:"region"`

	// The DSN used to connect to Sentry
	SentryDSN string `json:"sentrydsn"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Read the configuration data from Pulumi.<stack>.yaml
		conf := config.New(ctx, "awsconfig")

		// Create a new Tags object with the data from the configuration
		var tags Tags
		conf.RequireObject("tags", &tags)

		// Create a new DynamoConfig object with the data from the configuration
		var lambdaConfig LambdaConfig
		conf.RequireObject("lambda", &lambdaConfig)

		// Create a map[string]pulumi.Input of the tags
		// the first four tags come from the configuration file
		// the last two are derived from this deployment
		tagMap := make(map[string]pulumi.Input)
		tagMap["Author"] = tags.Author
		tagMap["Feature"] = tags.Feature
		tagMap["Team"] = tags.Team
		tagMap["Version"] = tags.Version
		tagMap["ManagedBy"] = pulumi.String("Pulumi")
		tagMap["Stage"] = pulumi.String(ctx.Stack())

		// Compile and upload the AWS Lambda functions only if this isn't a dry run
		if !ctx.DryRun() && 1 == 2 {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			fnFolder := path.Join(wd, "..", "cmd", fmt.Sprintf("lambda-payment-%s", lambdaConfig.EventingType))

			if err := run(fnFolder, "GOOS=linux GOARCH=amd64 go build"); err != nil {
				fmt.Printf("Error building code: %s", err.Error())
				os.Exit(1)
			}

			if err := run(fnFolder, fmt.Sprintf("zip ./lambda-payment-%s.zip, ./lambda-payment-%s", lambdaConfig.EventingType, lambdaConfig.EventingType)); err != nil {
				fmt.Printf("Error creating zipfile: %s", err.Error())
				os.Exit(1)
			}

			if err := run(fnFolder, fmt.Sprintf("aws s3 cp ./lambda-payment-%s.zip, s3://%s/acmeserverless/%s/lambda-payment-%s.zip", lambdaConfig.EventingType, lambdaConfig.S3Bucket, ctx.Stack(), lambdaConfig.EventingType)); err != nil {
				fmt.Printf("Error creating zipfile: %s", err.Error())
				os.Exit(1)
			}
		}

		// Create the IAM policy for the function.
		roleArgs := &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
				{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Effect": "Allow",
					"Sid": ""
				}
				]
			}`),
			Description: pulumi.String("Role for the Payment Service of the ACME Serverless Fitness Shop"),
			Tags:        pulumi.Map(tagMap),
		}

		// Create the role for the Lambda function
		role, err := iam.NewRole(ctx, "ACMEServerlessPaymentRole", roleArgs)
		if err != nil {
			return err
		}

		managedPolicy := &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
			Role:      role.Name,
		}

		_, err = iam.NewRolePolicyAttachment(ctx, "AWSLambdaBasicExecutionRole", managedPolicy)
		if err != nil {
			return err
		}

		// In case the Lambda function uses SQS, add a policy document
		// to allow the function to use SQS
		if lambdaConfig.EventingType == "sqs" {
			policyString := fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Action": [
							"sqs:SendMessage*"
						],
						"Effect": "Allow",
						"Resource": "%s"
					},
					{
						"Action": [
							"sqs:ReceiveMessage",
							"sqs:DeleteMessage",
							"sqs:GetQueueAttributes"
						],
						"Effect": "Allow",
						"Resource": "%s"
					}
				]
			}`, lambdaConfig.PaymentResponseQueue, lambdaConfig.PaymentRequestQueue)

			sqsPolicy := &iam.RolePolicyArgs{
				Name:   pulumi.String("ACMEServerlessPaymentSQSPolicy"),
				Role:   role.Name,
				Policy: pulumi.String(policyString),
			}

			_, err := iam.NewRolePolicy(ctx, "ACMEServerlessPaymentSQSPolicy", sqsPolicy)
			if err != nil {
				return err
			}
		}

		// Export the role ARN as an output of the Pulumi stack
		ctx.Export("ACMEServerlessPaymentRole::Arn", role.Arn)

		// Create the environment variables for the Lambda function
		variables := make(map[string]pulumi.StringInput)
		variables["REGION"] = pulumi.String(lambdaConfig.Region)
		variables["SENTRY_DSN"] = pulumi.String(lambdaConfig.SentryDSN)
		variables["FUNCTION_NAME"] = pulumi.String(fmt.Sprintf("%s-lambda-payment", ctx.Stack()))
		variables["VERSION"] = tags.Version
		variables["STAGE"] = pulumi.String(ctx.Stack())

		if lambdaConfig.EventingType == "sqs" {
			variables["RESPONSEQUEUE"] = pulumi.String(lambdaConfig.PaymentResponseQueue)
		} else {
			variables["EVENTBUS"] = pulumi.String(lambdaConfig.EventBus)
		}

		environment := lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap(variables),
		}

		// The set of arguments for constructing a Function resource.
		functionArgs := &lambda.FunctionArgs{
			Description: pulumi.String("A Lambda function to validate creditcard payments"),
			Runtime:     pulumi.String("go1.x"),
			Name:        pulumi.String(fmt.Sprintf("%s-lambda-payment", ctx.Stack())),
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Handler:     pulumi.String(fmt.Sprintf("lambda-payment-%s", lambdaConfig.EventingType)),
			Environment: environment,
			S3Bucket:    pulumi.String(lambdaConfig.S3Bucket),
			S3Key:       pulumi.String(fmt.Sprintf("acmeserverless/%s/lambda-payment-%s.zip", ctx.Stack(), lambdaConfig.EventingType)),
			Role:        role.Arn,
			Tags:        pulumi.Map(tagMap),
		}

		// NewFunction registers a new resource with the given unique name, arguments, and options.
		function, err := lambda.NewFunction(ctx, fmt.Sprintf("%s-lambda-payment", ctx.Stack()), functionArgs)
		if err != nil {
			return err
		}

		if lambdaConfig.EventingType == "sqs" {
			sqsMapping := &lambda.EventSourceMappingArgs{
				BatchSize:      pulumi.Int(1),
				Enabled:        pulumi.Bool(true),
				FunctionName:   function.Arn,
				EventSourceArn: pulumi.String(lambdaConfig.PaymentRequestQueue),
			}

			_, err := lambda.NewEventSourceMapping(ctx, fmt.Sprintf("%s-lambda-payment", ctx.Stack()), sqsMapping)
			if err != nil {
				return err
			}
		}

		ctx.Export(fmt.Sprintf("lambda-payment-%s::Arn", lambdaConfig.EventingType), function.Arn)

		return nil
	})
}

// run creates a Cmd struct to execute the named program with the given arguments.
// After that, it starts the specified command and waits for it to complete.
func run(folder string, args string) error {
	cmd := exec.Command(shell, shellFlag, args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = folder
	return cmd.Run()
}
