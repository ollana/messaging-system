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
	UserName     string          `json:"UserName"`
	BlockedUsers map[string]bool `json:"BlockedUsers"`
	Groups       map[string]bool `json:"Groups"`
}

type dbGroup struct {
	GroupId   string          `json:"GroupId"`
	GroupName string          `json:"GroupName"`
	Members   map[string]bool `json:"Members"`
}

type dbMessage struct {
	RecipientId string `json:"RecipientId"` // can be user or group id
	Timestamp   string `json:"Timestamp"`
	SenderId    string `json:"SenderId"`
	Message     string `json:"Message"`
}

type dynamoDBClientInterface interface {
	StoreUser(ctx context.Context, user dbUser) error
	BlockUser(ctx context.Context, user dbUser, blockedUserId string) error
	GetUser(ctx context.Context, userId string) (*dbUser, error)

	StoreGroup(ctx context.Context, group dbGroup) error
	AddUserToGroup(ctx context.Context, group dbGroup, user dbUser) error
	RemoveUserFromGroup(ctx context.Context, group dbGroup, user dbUser) error

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

func (d *dynamoDBClient) BlockUser(ctx context.Context, user dbUser, blockedUserId string) error {
	// add the blocked user
	user.BlockedUsers[blockedUserId] = true
	// update user record
	err := d.StoreUser(ctx, user)
	return err
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

func (d *dynamoDBClient) AddUserToGroup(ctx context.Context, group dbGroup, user dbUser) error {

	group.Members[user.UserId] = true
	user.Groups[group.GroupId] = true

	// Serialize to map[string]AttributeValue
	dbUser, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	dbGroup, err := attributevalue.MarshalMap(group)
	if err != nil {
		return err
	}

	// update in one transaction
	_, err = d.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(UsersTableName),
					Item:      dbUser,
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(GroupsTableName),
					Item:      dbGroup,
				},
			},
		},
	})

	return err

}

func (d *dynamoDBClient) RemoveUserFromGroup(ctx context.Context, group dbGroup, user dbUser) error {
	// remove the member
	delete(group.Members, user.UserId)
	delete(user.Groups, group.GroupId)

	// Serialize to map[string]AttributeValue
	dbUser, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	dbGroup, err := attributevalue.MarshalMap(group)
	if err != nil {
		return err
	}

	// update in one transaction
	_, err = d.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(UsersTableName),
					Item:      dbUser,
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(GroupsTableName),
					Item:      dbGroup,
				},
			},
		},
	})

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
