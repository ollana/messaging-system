package common

import (
	"fmt"
	"golang.org/x/exp/slog"
	"log"
	"time"
)

import lru "github.com/hashicorp/golang-lru"

var cache *lru.Cache

const (
	groupCacheKeyPrefix   = "group-"
	userCacheKeyPrefix    = "user-"
	messageCacheKeyPrefix = "group-messages-"
)

func getGroupCacheKey(groupId string) string   { return groupCacheKeyPrefix + groupId }
func getUserCacheKey(userId string) string     { return userCacheKeyPrefix + userId }
func getMessageCacheKey(groupId string) string { return messageCacheKeyPrefix + groupId }

func init() {
	var err error
	cache, err = lru.New(1000) // 1000 items should be modified to a more reasonable number
	if err != nil {
		log.Fatal(err)
	}

}

func GetUserFromCache(userId string) (*User, bool) {
	key := getUserCacheKey(userId)
	if val, ok := getItem(key); ok {
		slog.Info(fmt.Sprintf("User %s found in cache", userId))
		return val.(*User), ok
	}
	return nil, false
}

func GetGroupFromCache(groupId string) (*Group, bool) {
	key := getGroupCacheKey(groupId)
	if val, ok := getItem(key); ok {
		slog.Info(fmt.Sprintf("Group %s found in cache", groupId))
		return val.(*Group), ok
	}
	return nil, false
}

func getItem(key string) (interface{}, bool) {
	return cache.Get(key)
}

func StoreUserInCache(value *User) {
	key := getUserCacheKey(value.UserId)
	storeInCache(key, value)
	slog.Info(fmt.Sprintf("User %s stored in cache", value.UserId))
}

func StoreGroupInCache(value *Group) {
	key := getGroupCacheKey(value.GroupId)
	storeInCache(key, value)
	slog.Info(fmt.Sprintf("Group %s stored in cache", value.GroupId))
}

func storeInCache(key string, value interface{}) {
	cache.Add(key, value)

}

// StoreMessagesInCache will only store last minute messages in cache to avoid memory issues.
func StoreMessagesInCache(groupId string, messages []Message) {
	key := getMessageCacheKey(groupId)
	// only store messages from the last 1 minute
	var validMessages []Message
	for _, msg := range messages {
		if msg.Timestamp > time.Now().Add(-1*time.Minute).Format(time.RFC3339) {
			validMessages = append(validMessages, msg)
		}
	}
	if len(validMessages) > 0 {
		storeInCache(key, validMessages)
		slog.Info(fmt.Sprintf("Messages for group %s stored in cache", groupId))
	}
}
func StoreMessageInCache(message Message) {
	key := getMessageCacheKey(message.RecipientId)
	// only store to the cache if the group already exists in cache
	if val, ok := getItem(key); ok {
		messages := val.([]Message)
		messages = append(messages, message)
		storeInCache(key, messages)
		slog.Info(fmt.Sprintf("Message for group %s stored in cache", message.RecipientId))
	}
}

func GetGroupMessagesFromCache(groupId string, timestamp int64) ([]Message, bool) {
	key := getMessageCacheKey(groupId)
	if val, ok := getItem(key); ok {
		allMessages := val.([]Message)
		var requestedMessages []Message
		// we only want messages that are newer than the timestamp additionally we want to evict all messages older than 1 minutes
		for i, msg := range allMessages {
			if msg.Timestamp > time.Unix(timestamp, 0).Format(time.RFC3339) {
				requestedMessages = append(requestedMessages, msg)
			}
			if msg.Timestamp < time.Now().Add(-1*time.Minute).Format(time.RFC3339) {
				// evict message
				allMessages = append(allMessages[:i], allMessages[i+1:]...)
			}
		}
		// store the updated messages back in the cache
		if len(allMessages) > 0 {
			storeInCache(key, allMessages)
			slog.Info(fmt.Sprintf("Messages for group %s stored in cache", groupId))
		} else {
			// evict group from cache if no items left
			slog.Info(fmt.Sprintf("Evicting group %s from cache as all messeges are old", groupId))
			cache.Remove(key)
		}
		// return only the messages according to the requested timestamp
		if len(requestedMessages) > 0 {
			slog.Info(fmt.Sprintf("Messages for group %s found in cache", groupId))
			return requestedMessages, true
		}
		return nil, false
	}

	return nil, false

}
