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
	. "server/common"
	"time"
)

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
	StoreUserInCache(&user)
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
	if user, ok := GetUserFromCache(userId); ok {
		return user, nil
	}

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
	StoreGroupInCache(&group)
	return nil

}

func (d *dynamoDBClient) GetGroup(ctx context.Context, groupId string) (*Group, error) {
	if group, ok := GetGroupFromCache(groupId); ok {
		return group, nil
	}
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

	if group.Members == nil {
		group.Members = make(map[string]bool)
	}
	group.Members[user.UserId] = true
	if user.Groups == nil {
		user.Groups = make(map[string]bool)
	}
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
	StoreMessageInCache(message)
	return nil
}

func (d *dynamoDBClient) GetMessages(ctx context.Context, user User, timestamp int64) ([]Message, error) {
	// convert user.Groups map to list
	groupList := make([]string, 0, len(user.Groups))
	for k := range user.Groups {
		groupList = append(groupList, k)
	}

	// get all Messages
	allMessages, err := d.getRecipientMessages(ctx, user.UserId, groupList, timestamp)
	return allMessages, err
}

func (d *dynamoDBClient) getRecipientMessages(ctx context.Context, userId string, groupIds []string, timestamp int64) ([]Message, error) {
	var messages []Message

	lessThenMinute := time.Now().Unix()-timestamp < 60

	checkCache := false
	// should check in cache only if timestamp is in range of last minute
	if timestamp > 0 && lessThenMinute {
		checkCache = true
	}

	recipientIds := append(groupIds, userId)
	for _, recipientId := range recipientIds {
		isGroup := recipientId != userId
		if checkCache && isGroup {
			if val, ok := GetGroupMessagesFromCache(recipientId, timestamp); ok {
				messages = append(messages, val...)
				continue
			}
		}

		id, err := attributevalue.Marshal(recipientId)
		if err != nil {
			return nil, err
		}
		keyConditions := map[string]types.Condition{
			RecipientIdKey: {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{id},
			},
		}
		if timestamp > 0 {
			timeStampToCheck := time.Unix(timestamp, 0).Format(time.RFC3339)
			// check for messages after the provided timestamp, but at least for 1 minute for caching purposes
			if lessThenMinute && isGroup {
				timeStampToCheck = time.Now().Add(-1 * time.Minute).Format(time.RFC3339)
			}
			keyConditions[TimestampSortKey] = types.Condition{
				ComparisonOperator: types.ComparisonOperatorGt,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: timeStampToCheck},
				},
			}
		}

		// get all items with the recipientId in the list AND the timestamp greater than the provided timestamp
		results, err := d.client.Query(ctx, &dynamodb.QueryInput{
			TableName:     aws.String(MessagesTableName),
			KeyConditions: keyConditions,
		})

		if err != nil {
			return nil, err
		}

		recipientMsgs := make([]Message, 0, len(results.Items))
		for _, msg := range results.Items {
			var message Message
			err = attributevalue.UnmarshalMap(msg, &message)
			if err != nil {
				return nil, err
			}
			recipientMsgs = append(recipientMsgs, message)
		}

		if recipientId != userId && len(recipientMsgs) > 0 {
			// add group msgs to cache
			StoreMessagesInCache(recipientId, recipientMsgs)
		}

		if isGroup && lessThenMinute {
			// filter out messages older then requested timestamp
			var validMessages []Message
			for _, msg := range recipientMsgs {
				if msg.Timestamp > time.Unix(timestamp, 0).Format(time.RFC3339) {
					validMessages = append(validMessages, msg)
				}
			}
			messages = append(messages, validMessages...)
		} else {
			messages = append(messages, recipientMsgs...)
		}
	}

	return messages, nil

}
