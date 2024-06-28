package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dbUser struct {
	UserId       string          `json:"UserID"`
	BlockedUsers map[string]bool `json:"BlockedUsers"`
	Groups       map[string]bool `json:"Groups"`
}

type dbGroup struct {
	GroupId string          `json:"GroupId"`
	Members map[string]bool `json:"Members"`
}

type dbMessage struct {
	RecipientId string `json:"RecipientId"` // can be user or group id
	Timestamp   string `json:"Timestamp"`
	SenderId    string `json:"SenderId"`
	Message     string `json:"Message"`
}

type dynamoDBClientInterface interface {
	StoreUser(ctx context.Context, user dbUser) error
	BlockUser(ctx context.Context, userId string, blockedUserId string) error
	GetUser(ctx context.Context, userId string) (*dbUser, error)

	StoreGroup(ctx context.Context, group dbGroup) error
	AddUserToGroup(ctx context.Context, groupId string, userId string) error
	RemoveUserFromGroup(ctx context.Context, groupId string, userId string) error

	StoreMessage(ctx context.Context, message dbMessage) error
	GetMessages(ctx context.Context, recipientId string) ([]dbMessage, error)
}

type dynamoDBClient struct {
	client *dynamodb.Client
}

func NewDynamoDBClient() (dynamoDBClientInterface, error) {
	dynamoClient := &dynamoDBClient{}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("Error loading configuration, ", err)
		return nil, err
	}
	dbClient := dynamodb.NewFromConfig(cfg)
	dynamoClient.client = dbClient

	return dynamoClient, nil
}

const (
	UsersTableName    = "usersTable"
	GroupsTableName   = "groupsTable"
	MessagesTableName = "messagesTable"
	UserPrimaryKey    = "UserID"
	GroupPrimaryKey   = "GroupID"
)

func (d *dynamoDBClient) StoreUser(ctx context.Context, user dbUser) error {

	// Serialize the user into a map[string]AttributeValue
	av, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	// Create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(UsersTableName),
		Item:      av,
	}

	// Write to DynamoDB
	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (d *dynamoDBClient) BlockUser(ctx context.Context, userId string, blockedUserId string) error {
	// Get the user
	user, err := d.GetUser(ctx, userId)
	if err != nil {
		return err
	}
	//
	// check if already blocked
	if user.BlockedUsers[blockedUserId] {
		return nil
	}
	// add the blocked user
	user.BlockedUsers[blockedUserId] = true
	// update user record
	err = d.StoreUser(ctx, *user)
	return nil
}

func (d *dynamoDBClient) GetUser(ctx context.Context, userId string) (*dbUser, error) {
	// Create GetItem input
	id, err := attributevalue.Marshal(userId)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(UsersTableName),
		Key:       map[string]types.AttributeValue{UserPrimaryKey: id},
	}

	// Get item from DynamoDB
	result, err := d.client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}
	// If result.Item is empty, no item with the provided ID exists
	if result.Item == nil {
		return nil, nil
	}
	// Unmarshal the result
	var user dbUser
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *dynamoDBClient) StoreGroup(ctx context.Context, group dbGroup) error {
	// Serialize the group into a map[string]AttributeValue
	av, err := attributevalue.MarshalMap(group)
	if err != nil {
		return err
	}
	// create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(GroupsTableName),
		Item:      av,
	}
	// Write to DynamoDB
	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return err
	}
	return nil

}

func (d *dynamoDBClient) GetGroup(ctx context.Context, groupId string) (*dbGroup, error) {
	// Create GetItem input
	id, err := attributevalue.Marshal(groupId)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(GroupsTableName),
		Key:       map[string]types.AttributeValue{GroupPrimaryKey: id},
	}

	// Get item from DynamoDB
	result, err := d.client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}
	// If result.Item is empty, no item with the provided ID exists
	if result.Item == nil {
		return nil, nil
	}
	// Unmarshal the result
	var group dbGroup
	err = attributevalue.UnmarshalMap(result.Item, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (d *dynamoDBClient) AddUserToGroup(ctx context.Context, groupId string, userId string) error {
	// Get the group
	group, err := d.GetGroup(ctx, groupId)
	if err != nil {
		return err
	}
	// check if already member
	if group.Members[userId] {
		return nil
	}
	// add the member
	group.Members[userId] = true
	// update group record
	err = d.StoreGroup(ctx, *group)
	return err

}

func (d *dynamoDBClient) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) error {
	// Get the group
	group, err := d.GetGroup(ctx, groupId)
	if err != nil {
		return err
	}
	// check if already member
	if !group.Members[userId] {
		return nil
	}
	// remove the member
	delete(group.Members, userId)
	// update group record
	err = d.StoreGroup(ctx, *group)
	return err

}

func (d *dynamoDBClient) StoreMessage(ctx context.Context, message dbMessage) error {
	// Serialize the message into a map[string]AttributeValue
	av, err := attributevalue.MarshalMap(message)
	if err != nil {
		return err
	}
	// create PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(MessagesTableName),
		Item:      av,
	}
	// Write to DynamoDB
	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (d *dynamoDBClient) GetMessages(ctx context.Context, recipientId string) ([]dbMessage, error) {
	// get all user groups ids
	user, err := d.GetUser(ctx, recipientId)
	if err != nil {
		return nil, err
	}
	// convert user.Groups map to list
	list := make([]string, 0, len(user.Groups)+1)
	for k := range user.Groups {
		list = append(list, k)
	}
	// add recipient id to the list to get the private messages as well
	list = append(list, recipientId)

	// get all messages
	allMessages, err := d.getRecipientMessages(ctx, list)
	return allMessages, err
}

func (d *dynamoDBClient) getRecipientMessages(ctx context.Context, recipientIds []string) ([]dbMessage, error) {
	var messages []dbMessage

	ids, err := attributevalue.MarshalList(recipientIds)

	// get all items with the recipientId in the list
	results, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(MessagesTableName),
		KeyConditions: map[string]types.Condition{
			"RecipientId": {
				ComparisonOperator: types.ComparisonOperatorIn,
				AttributeValueList: ids,
			},
		},
	})

	// If result.Item is empty, no item with the provided ID exists
	if results.Items == nil {
		return nil, nil
	}

	for _, msg := range results.Items {
		var message dbMessage
		err = attributevalue.UnmarshalMap(msg, &message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)

	}

	return messages, nil

}
