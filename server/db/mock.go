package db

import (
	"context"
	"time"
)

type MockDBClient struct {
	Users    map[string]User
	Groups   map[string]Group
	Messages map[string][]Message
	Error    error
}

func NewMockDBClient() *MockDBClient {
	return &MockDBClient{
		Users:    map[string]User{},
		Groups:   map[string]Group{},
		Messages: map[string][]Message{},
	}
}

func (m *MockDBClient) StoreUser(ctx context.Context, user User) error {
	if m.Error != nil {
		return m.Error
	}
	user.BlockedUsers = map[string]bool{}
	user.Groups = map[string]bool{}
	m.Users[user.UserId] = user
	return nil
}

func (m *MockDBClient) BlockUser(ctx context.Context, user User, blockedUserId string) error {
	if m.Error != nil {
		return m.Error
	}
	m.Users[user.UserId].BlockedUsers[blockedUserId] = true
	return nil
}

func (m *MockDBClient) UnBlockUser(ctx context.Context, user User, unBlockedUserId string) error {
	if m.Error != nil {
		return m.Error
	}
	delete(m.Users[user.UserId].BlockedUsers, unBlockedUserId)
	return nil
}

func (m *MockDBClient) GetUser(ctx context.Context, userId string) (*User, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if user, ok := m.Users[userId]; ok {
		return &user, nil
	}
	return nil, nil
}
func (m *MockDBClient) StoreGroup(ctx context.Context, group Group) error {
	if m.Error != nil {
		return m.Error
	}
	group.Members = map[string]bool{}
	m.Groups[group.GroupId] = group
	return nil
}
func (m *MockDBClient) GetGroup(ctx context.Context, groupId string) (*Group, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	if group, ok := m.Groups[groupId]; ok {
		return &group, nil
	}
	return nil, nil
}
func (m *MockDBClient) AddUserToGroup(ctx context.Context, group Group, user User) error {
	if m.Error != nil {
		return m.Error
	}
	m.Groups[group.GroupId].Members[user.UserId] = true
	m.Users[user.UserId].Groups[group.GroupId] = true
	return nil
}
func (m *MockDBClient) RemoveUserFromGroup(ctx context.Context, group Group, user User) error {
	if m.Error != nil {
		return m.Error
	}
	delete(m.Groups[group.GroupId].Members, user.UserId)
	delete(m.Users[user.UserId].Groups, group.GroupId)
	return nil
}
func (m *MockDBClient) StoreMessage(ctx context.Context, msg Message) error {
	if m.Error != nil {
		return m.Error
	}
	m.Messages[msg.RecipientId] = append(m.Messages[msg.RecipientId], msg)
	return nil
}
func (m *MockDBClient) GetMessages(ctx context.Context, user User, timestamp int64) ([]Message, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	var msgs []Message
	msgs = append(m.Messages[user.UserId], msgs...)
	for groupId, _ := range user.Groups {
		msgs = append(m.Messages[groupId], msgs...)
	}

	// if timestamp is provided, return Messages after the timestamp
	if timestamp > 0 {
		var newMsgs []Message
		for _, msg := range msgs {
			// timestamp to RFC3339
			fromTimestamp := time.Unix(timestamp, 0).Format(time.RFC3339)
			if msg.Timestamp > fromTimestamp {
				newMsgs = append(newMsgs, msg)
			}
		}
		msgs = newMsgs
	}

	return msgs, nil
}
