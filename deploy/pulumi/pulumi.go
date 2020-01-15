package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// The policy description of the IAM role, in this case only the sts:AssumeRole is needed
		roleArgs := &iam.RoleArgs{
			AssumeRolePolicy: `{
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
    }`,
		}

		// Create a new role called HelloWorldIAMRole
		role, err := iam.NewRole(ctx, "HelloWorldIAMRole", roleArgs)
		if err != nil {
			fmt.Printf("role error: %s\n", err.Error())
			return err
		}

		// Export the role ARN as an output of the Pulumi stack
		ctx.Export("Role ARN", role.Arn())
		return nil
	})
}
