package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a new DynamoDB table
		table, err := dynamodb.NewTable(ctx, "messagesTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("RecipientId"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("Timestamp"),
					Type: pulumi.String("N"),
				},
			},
			HashKey:        pulumi.String("RecipientId"),
			RangeKey:       pulumi.String("Timestamp"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
		})

		if err != nil {
			return err
		}

		// Create an IAM policy for DynamoDB access
		policy, err := iam.NewPolicy(ctx, "dynamodbPolicy", &iam.PolicyArgs{
			Description: pulumi.String("Policy to allow DynamoDB access"),

			Policy: pulumi.String(`
                {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Action": [
                                "dynamodb:GetItem",
                                "dynamodb:PutItem",
                                "dynamodb:UpdateItem",
                                "dynamodb:DeleteItem",
                                "dynamodb:Scan",
                                "dynamodb:Query"
                            ],
                            "Resource": "arn:aws:dynamodb:*"
                        }
                    ]
                }
            `),
		})
		if err != nil {
			return err
		}

		// Create an IAM role for the EC2 instance
		role, err := iam.NewRole(ctx, "ec2Role", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`
                {
                    "Version": "2012-10-17",
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "Service": "ec2.amazonaws.com"
                            },
                            "Action": "sts:AssumeRole"
                        }
                    ]
                }
            `),
		})

		// Attach the policy to the role
		_, err = iam.NewRolePolicyAttachment(ctx, "messaging-system-server-policy-attachment", &iam.RolePolicyAttachmentArgs{
			Role:      role.Name,
			PolicyArn: policy.Arn,
		})
		if err != nil {
			return err
		}

		// Create an instance profile for the EC2 instance to assume the role
		instanceProfile, err := iam.NewInstanceProfile(ctx, "messaging-system-server-profile", &iam.InstanceProfileArgs{
			Role: role.Name,
		})
		if err != nil {
			return err
		}

		// Create a security group
		secGroup, err := ec2.NewSecurityGroup(ctx, "messaging-system-server-secgrp", &ec2.SecurityGroupArgs{
			Description: pulumi.String("Enable HTTPS access"),
			// inbound - enable HTTPS access only
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(443),
					ToPort:     pulumi.Int(443),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			// outbound - allow traffic to the AWS DynamoDB service (operates over HTTPS)
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Protocol: pulumi.String("tcp"),
					FromPort: pulumi.Int(443),
					ToPort:   pulumi.Int(443),
					CidrBlocks: pulumi.StringArray{
						pulumi.String("0.0.0.0/0"), // Allow outbound HTTPS traffic to anywhere, consider more restrictive CIDR for better security
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Get the latest Amazon Linux AMI
		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
			MostRecent: pulumi.BoolRef(true),
			Filters: []ec2.GetAmiFilter{
				{
					Name:   "name",
					Values: []string{"amzn2-ami-hvm-*-x86_64-gp2"},
				},
			},
			Owners: []string{"137112412989"},
		})
		if err != nil {
			return err
		}

		// Create a new EC2 instance
		server, err := ec2.NewInstance(ctx, "messaging-system-server", &ec2.InstanceArgs{
			InstanceType:        pulumi.String("t2.micro"),
			VpcSecurityGroupIds: pulumi.StringArray{secGroup.ID()},
			IamInstanceProfile:  instanceProfile.Name,
			Ami:                 pulumi.String(ami.Id),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("messaging-system-server"),
			},
		})
		if err != nil {
			return err
		}

		// Export the public IP of the instance
		ctx.Export("publicIp", server.PublicIp)
		ctx.Export("publicDns", server.PublicDns)

		// Export the DynamoDB table name
		ctx.Export("tableName", table.Name)

		return nil

	})
}
