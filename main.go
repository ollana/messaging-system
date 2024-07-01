package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/awsx"
	ecsx "github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/ecs"
	"github.com/pulumi/pulumi-awsx/sdk/v2/go/awsx/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"os"
)

func main() {

	// read aws account from env var AWS_ACCOUNT_ID
	awsAccount := os.Getenv("AWS_ACCOUNT_ID")
	imageName := fmt.Sprintf("%s.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:1.0.12", awsAccount)

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create required DynamoDB tables
		_, err := dynamodb.NewTable(ctx, "messagesTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("RecipientId"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("Timestamp"),
					Type: pulumi.String("S"),
				},
			},
			HashKey:        pulumi.String("RecipientId"),
			RangeKey:       pulumi.String("Timestamp"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
			Name:           pulumi.String("messagesTable"),
		})

		if err != nil {
			return err
		}

		_, err = dynamodb.NewTable(ctx, "usersTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("UserId"),
					Type: pulumi.String("S"),
				},
			},
			HashKey:        pulumi.String("UserId"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
			Name:           pulumi.String("usersTable"),
		})
		if err != nil {
			return err
		}

		_, err = dynamodb.NewTable(ctx, "groupsTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("GroupId"),
					Type: pulumi.String("S"),
				},
			},
			HashKey:        pulumi.String("GroupId"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
			Name:           pulumi.String("groupsTable"),
		})
		if err != nil {
			return err
		}

		lb, err := lb.NewApplicationLoadBalancer(ctx, "lb", nil)
		if err != nil {
			return err
		}

		cluster, err := ecs.NewCluster(ctx, "cluster", nil)
		if err != nil {
			return err
		}

		// Create IAM Role and Policy for Fargate Task
		role, err := iam.NewRole(ctx, "taskRole", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {
							"Service": "ecs-tasks.amazonaws.com"
						},
						"Action": "sts:AssumeRole"
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		policy, err := iam.NewPolicy(ctx, "taskPolicy", &iam.PolicyArgs{
			Description: pulumi.String("Policy for Fargate task to access DynamoDB"),
			Policy: pulumi.String(
				fmt.Sprintf(`{
					"Version": "2012-10-17",
					"Statement": [
						{
							"Effect": "Allow",
							"Action": [
								"dynamodb:DescribeTable",
								"dynamodb:Query",
								"dynamodb:Scan",
								"dynamodb:GetItem",
								"dynamodb:PutItem",
								"dynamodb:UpdateItem",
								"dynamodb:DeleteItem"
							],
							
							"Resource": "*"
						}
					]
				}`),
			),
		})
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicyAttachment(ctx, "taskPolicyAttachment", &iam.RolePolicyAttachmentArgs{
			PolicyArn: policy.Arn,
			Role:      role.Name,
		})
		if err != nil {
			return err
		}

		_, err = ecsx.NewFargateService(ctx, "messaging-system-service", &ecsx.FargateServiceArgs{
			Cluster:        cluster.Arn,
			AssignPublicIp: pulumi.Bool(true),
			DesiredCount:   pulumi.Int(1),
			TaskDefinitionArgs: &ecsx.FargateServiceTaskDefinitionArgs{
				TaskRole: &awsx.DefaultRoleWithPolicyArgs{
					RoleArn: role.Arn,
				},
				Container: &ecsx.TaskDefinitionContainerDefinitionArgs{
					Name:      pulumi.String("messaging-system"),
					Image:     pulumi.String(imageName),
					Cpu:       pulumi.Int(128),
					Memory:    pulumi.Int(512),
					Essential: pulumi.Bool(true),
					PortMappings: ecsx.TaskDefinitionPortMappingArray{
						&ecsx.TaskDefinitionPortMappingArgs{
							ContainerPort: pulumi.Int(80),
							TargetGroup:   lb.DefaultTargetGroup,
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("url", pulumi.Sprintf("http://%s", lb.LoadBalancer.DnsName()))

		return nil

	})
}
