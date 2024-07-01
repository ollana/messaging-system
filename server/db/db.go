package db

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"golang.org/x/exp/slog"
	"time"
)

type User struct {
	UserId       string          `json:"userId"`
	UserName     string          `json:"userName"`
	BlockedUsers map[string]bool `json:"blockedUsers"`
	Groups       map[string]bool `json:"groups"`
}

type Group struct {
	GroupId   string          `json:"groupId"`
	GroupName string          `json:"groupName"`
	Members   map[string]bool `json:"members"`
}

type Message struct {
	RecipientId string `json:"recipientId"` // can be user or group id
	Timestamp   string `json:"timestamp"`   // RFC3339
	SenderId    string `json:"senderId"`
	Message     string `json:"message"`
}

type DynamoDBClientInterface interface {
	StoreUser(ctx context.Context, user User) error
	BlockUser(ctx context.Context, user User, blockedUserId string) error
	UnBlockUser(ctx context.Context, user User, unBlockedUserId string) error
	GetUser(ctx context.Context, userId string) (*User, error)

	StoreGroup(ctx context.Context, group Group) error
	GetGroup(ctx context.Context, groupId string) (*Group, error)
	AddUserToGroup(ctx context.Context, group Group, user User) error
	RemoveUserFromGroup(ctx context.Context, group Group, user User) error

	StoreMessage(ctx context.Context, message Message) error
	GetMessages(ctx context.Context, user User, timestamp int64) ([]Message, error)
}

type dynamoDBClient struct {
	client *dynamodb.Client
}

func NewDynamoDBClient() (DynamoDBClientInterface, error) {
	dynamoClient := &dynamoDBClient{}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-west-2"))
	if err != nil {
		slog.Error(fmt.Sprintf("Error loading configuration: %v", err))
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
	UserPrimaryKey    = "UserId"
	GroupPrimaryKey   = "GroupId"
	TimestampSortKey  = "Timestamp"
	RecipientIdKey    = "RecipientId"
)

func (d *dynamoDBClient) StoreUser(ctx context.Context, user User) error {

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

func (d *dynamoDBClient) BlockUser(ctx context.Context, user User, blockedUserId string) error {
	// add the blocked user
	user.BlockedUsers[blockedUserId] = true
	// update user record
	err := d.StoreUser(ctx, user)
	return err
}

func (d *dynamoDBClient) UnBlockUser(ctx context.Context, user User, unBlockedUserId string) error {
	// add the blocked user
	delete(user.BlockedUsers, unBlockedUserId)
	// update user record
	err := d.StoreUser(ctx, user)
	return err
}

func (d *dynamoDBClient) GetUser(ctx context.Context, userId string) (*User, error) {
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
	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *dynamoDBClient) StoreGroup(ctx context.Context, group Group) error {
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

func (d *dynamoDBClient) GetGroup(ctx context.Context, groupId string) (*Group, error) {
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
	var group Group
	err = attributevalue.UnmarshalMap(result.Item, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (d *dynamoDBClient) AddUserToGroup(ctx context.Context, group Group, user User) error {

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

func (d *dynamoDBClient) RemoveUserFromGroup(ctx context.Context, group Group, user User) error {
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

func (d *dynamoDBClient) StoreMessage(ctx context.Context, message Message) error {
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

func (d *dynamoDBClient) GetMessages(ctx context.Context, user User, timestamp int64) ([]Message, error) {
	// convert user.Groups map to list
	list := make([]string, 0, len(user.Groups)+1)
	for k := range user.Groups {
		list = append(list, k)
	}
	// add recipient id to the list to get the private Messages as well
	list = append(list, user.UserId)

	// get all Messages
	allMessages, err := d.getRecipientMessages(ctx, list, timestamp)
	return allMessages, err
}

func (d *dynamoDBClient) getRecipientMessages(ctx context.Context, recipientIds []string, timestamp int64) ([]Message, error) {
	var messages []Message
	ids, err := attributevalue.MarshalList(recipientIds)

	keyConditions := map[string]types.Condition{
		RecipientIdKey: {
			ComparisonOperator: types.ComparisonOperatorIn,
			AttributeValueList: ids,
		},
	}
	// add timestamp condition if provided, otherwise all recipient Messages will be returned
	if timestamp > 0 {
		keyConditions[TimestampSortKey] = types.Condition{
			ComparisonOperator: types.ComparisonOperatorGt,
			AttributeValueList: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: time.Unix(timestamp, 0).Format(time.RFC3339)},
			},
		}
	}

	// get all items with the recipientId in the list AND the timestamp greater than the provided timestamp
	results, err := d.client.Query(ctx, &dynamodb.QueryInput{
		TableName:     aws.String(MessagesTableName),
		KeyConditions: keyConditions,
	})

	// If result.Item is empty, no item with the provided ID exists
	if results.Items == nil {
		return nil, nil
	}

	for _, msg := range results.Items {
		var message Message
		err = attributevalue.UnmarshalMap(msg, &message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)

	}

	return messages, nil

}
