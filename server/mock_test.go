package main

import (
	"context"
	"time"
)

type mockDBClient struct {
	users    map[string]dbUser
	groups   map[string]dbGroup
	messages map[string][]dbMessage
	error    error
}

func newMockDBClient() *mockDBClient {
	return &mockDBClient{
		users:    map[string]dbUser{},
		groups:   map[string]dbGroup{},
		messages: map[string][]dbMessage{},
	}
}

func (m *mockDBClient) StoreUser(ctx context.Context, user dbUser) error {
	if m.error != nil {
		return m.error
	}
	user.BlockedUsers = map[string]bool{}
	user.Groups = map[string]bool{}
	m.users[user.UserId] = user
	return nil
}

func (m *mockDBClient) BlockUser(ctx context.Context, user dbUser, blockedUserId string) error {
	if m.error != nil {
		return m.error
	}
	m.users[user.UserId].BlockedUsers[blockedUserId] = true
	return nil
}

func (m *mockDBClient) UnBlockUser(ctx context.Context, user dbUser, unBlockedUserId string) error {
	if m.error != nil {
		return m.error
	}
	delete(m.users[user.UserId].BlockedUsers, unBlockedUserId)
	return nil
}

func (m *mockDBClient) GetUser(ctx context.Context, userId string) (*dbUser, error) {
	if m.error != nil {
		return nil, m.error
	}
	if user, ok := m.users[userId]; ok {
		return &user, nil
	}
	return nil, nil
}
func (m *mockDBClient) StoreGroup(ctx context.Context, group dbGroup) error {
	if m.error != nil {
		return m.error
	}
	group.Members = map[string]bool{}
	m.groups[group.GroupId] = group
	return nil
}
func (m *mockDBClient) GetGroup(ctx context.Context, groupId string) (*dbGroup, error) {
	if m.error != nil {
		return nil, m.error
	}
	if group, ok := m.groups[groupId]; ok {
		return &group, nil
	}
	return nil, nil
}
func (m *mockDBClient) AddUserToGroup(ctx context.Context, group dbGroup, user dbUser) error {
	if m.error != nil {
		return m.error
	}
	m.groups[group.GroupId].Members[user.UserId] = true
	m.users[user.UserId].Groups[group.GroupId] = true
	return nil
}
func (m *mockDBClient) RemoveUserFromGroup(ctx context.Context, group dbGroup, user dbUser) error {
	if m.error != nil {
		return m.error
	}
	delete(m.groups[group.GroupId].Members, user.UserId)
	delete(m.users[user.UserId].Groups, group.GroupId)
	return nil
}
func (m *mockDBClient) StoreMessage(ctx context.Context, msg dbMessage) error {
	if m.error != nil {
		return m.error
	}
	m.messages[msg.RecipientId] = append(m.messages[msg.RecipientId], msg)
	return nil
}
func (m *mockDBClient) GetMessages(ctx context.Context, user dbUser, timestamp int64) ([]dbMessage, error) {
	if m.error != nil {
		return nil, m.error
	}

	var msgs []dbMessage
	msgs = append(m.messages[user.UserId], msgs...)
	for groupId, _ := range user.Groups {
		msgs = append(m.messages[groupId], msgs...)
	}

	// if timestamp is provided, return messages after the timestamp
	if timestamp > 0 {
		var newMsgs []dbMessage
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
