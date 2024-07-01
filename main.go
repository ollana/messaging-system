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
	imageName := fmt.Sprintf("%s.dkr.ecr.us-west-2.amazonaws.com/messaging-system-app:1.0.0", awsAccount)

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create required DynamoDB tables
		_, err := dynamodb.NewTable(ctx, "messagesTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("recipientId"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("timestamp"),
					Type: pulumi.String("S"),
				},
			},
			HashKey:        pulumi.String("recipientId"),
			RangeKey:       pulumi.String("timestamp"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
		})

		if err != nil {
			return err
		}

		_, err = dynamodb.NewTable(ctx, "usersTable", &dynamodb.TableArgs{
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("userId"),
					Type: pulumi.String("S"),
				},
			},
			HashKey:        pulumi.String("userId"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
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
			HashKey:        pulumi.String("groupId"),
			BillingMode:    pulumi.String("PAY_PER_REQUEST"),
			StreamEnabled:  pulumi.Bool(true),
			StreamViewType: pulumi.String("NEW_AND_OLD_IMAGES"),
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
					Cpu:       pulumi.Int(512),
					Memory:    pulumi.Int(1042),
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

		// Create an IAM policy for DynamoDB access
		//policy, err := iam.NewPolicy(ctx, "servicePolicy", &iam.PolicyArgs{
		//	Description: pulumi.String("Policy to allow DynamoDB  and S3 access"),
		//
		//	Policy: pulumi.String(`
		//        {
		//            "Version": "2012-10-17",
		//            "Statement": [
		//                {
		//                    "Effect": "Allow",
		//                    "Action": [
		//                        "dynamodb:GetItem",
		//                        "dynamodb:PutItem",
		//                        "dynamodb:UpdateItem",
		//                        "dynamodb:DeleteItem",
		//                        "dynamodb:Scan",
		//                        "dynamodb:Query"
		//                    ],
		//                    "Resource": "arn:aws:dynamodb:*"
		//                },
		//				{
		//					"Effect": "Allow",
		//					"Action": [
		//						"s3:GetObject"
		//					],
		//					"Resource": "*"
		//				}
		//            ]
		//        }
		//    `),
		//})
		//if err != nil {
		//	return err
		//}

		//// Create an IAM role for the EC2 instance
		//role, err := iam.NewRole(ctx, "ec2Role", &iam.RoleArgs{
		//	AssumeRolePolicy: pulumi.String(`
		//        {
		//            "Version": "2012-10-17",
		//            "Statement": [
		//                {
		//                    "Effect": "Allow",
		//                    "Principal": {
		//                        "Service": "ec2.amazonaws.com"
		//                    },
		//                    "Action": "sts:AssumeRole"
		//                }
		//            ]
		//        }
		//    `),
		//})
		//
		//		// Attach the policy to the role
		//		_, err = iam.NewRolePolicyAttachment(ctx, "messaging-system-server-policy-attachment", &iam.RolePolicyAttachmentArgs{
		//			Role:      role.Name,
		//			PolicyArn: policy.Arn,
		//		})
		//		if err != nil {
		//			return err
		//		}
		//
		//		// Create an instance profile for the EC2 instance to assume the role
		//		instanceProfile, err := iam.NewInstanceProfile(ctx, "messaging-system-server-profile", &iam.InstanceProfileArgs{
		//			Role: role.Name,
		//		})
		//		if err != nil {
		//			return err
		//		}
		//
		//		// Create a security group
		//		secGroup, err := ec2.NewSecurityGroup(ctx, "messaging-system-server-secgrp", &ec2.SecurityGroupArgs{
		//			Description: pulumi.String("Enable HTTPS access"),
		//			// inbound - enable HTTPS access only
		//			Ingress: ec2.SecurityGroupIngressArray{
		//				&ec2.SecurityGroupIngressArgs{
		//					Protocol:   pulumi.String("tcp"),
		//					FromPort:   pulumi.Int(80),
		//					ToPort:     pulumi.Int(80),
		//					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
		//				},
		//				&ec2.SecurityGroupIngressArgs{
		//					Protocol:   pulumi.String("tcp"),
		//					FromPort:   pulumi.Int(22),
		//					ToPort:     pulumi.Int(22),
		//					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
		//				},
		//			},
		//			// outbound - allow traffic to the AWS DynamoDB service (operates over HTTPS)
		//			Egress: ec2.SecurityGroupEgressArray{
		//				&ec2.SecurityGroupEgressArgs{
		//					Protocol: pulumi.String("tcp"),
		//					FromPort: pulumi.Int(443),
		//					ToPort:   pulumi.Int(443),
		//					CidrBlocks: pulumi.StringArray{
		//						pulumi.String("0.0.0.0/0"), // Allow outbound HTTPS traffic to anywhere, consider more restrictive CIDR for better security
		//					},
		//				},
		//			},
		//		})
		//		if err != nil {
		//			return err
		//		}
		//
		//		// Get the latest Amazon Linux AMI
		//		ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		//			MostRecent: pulumi.BoolRef(true),
		//			Filters: []ec2.GetAmiFilter{
		//				{
		//					Name:   "name",
		//					Values: []string{"amzn2-ami-hvm-*-x86_64-ebs"},
		//				},
		//			},
		//			Owners: []string{"amazon"},
		//		})
		//		if err != nil {
		//			return err
		//		}
		//
		//		// read the user data from userdata.txt file
		//		_, err = os.ReadFile("userdata.txt")
		//		if err != nil {
		//			return err
		//		}
		//
		//		// Create a new EC2 instance
		//		server, err := ec2.NewInstance(ctx, "messaging-system-server", &ec2.InstanceArgs{
		//			InstanceType:        pulumi.String("t2.micro"),
		//			VpcSecurityGroupIds: pulumi.StringArray{secGroup.ID()},
		//			IamInstanceProfile:  instanceProfile.Name,
		//			Ami:                 pulumi.String(ami.Id),
		//			Tags: pulumi.StringMap{
		//				"Name": pulumi.String("messaging-system-server"),
		//			},
		//			UserData: pulumi.String(`#!/bin/bash
		//echo "Hello, World!" > index.html
		//nohup python3 -m http.server 80 &`),
		//		})
		//		if err != nil {
		//			return err
		//		}
		//
		//		// Export the public IP of the instance
		//		ctx.Export("publicIp", server.PublicIp)
		//		ctx.Export("publicDns", server.PublicDns)

		//repository, err := ecr.NewRepository(ctx, "repository", &ecr.RepositoryArgs{
		//	ForceDelete: pulumi.Bool(true),
		//})
		//if err != nil {
		//	return err
		//}
		//
		//image, err := ecr.NewImage(ctx, "image", &ecr.ImageArgs{
		//
		//	RepositoryUrl: repository.Url,
		//	Context:       pulumi.String("./server"),
		//	//Platform:      pulumi.String("linux/amd64"),
		//})
		//if err != nil {
		//	return err
		//}

		return nil

	})
}
