package common

import (
	"log"
	"time"
)

import lru "github.com/hashicorp/golang-lru"

var cache *lru.Cache

func init() {
	var err error
	cache, err = lru.New(1000) // 1000 items should be modified to a more reasonable number
	if err != nil {
		log.Fatal(err)
	}

}

func GetUserFromCache(key string) (*User, bool) {
	if val, ok := getItem(key); ok {
		return val.(*User), ok
	}
	return nil, false
}

func GetGroupFromCache(key string) (*Group, bool) {
	if val, ok := getItem(key); ok {
		return val.(*Group), ok
	}
	return nil, false
}

func getItem(key string) (interface{}, bool) {
	return cache.Get(key)
}

func StoreInCache(key string, value interface{}) {
	cache.Add(key, value)

}

// StoreMessagesInCache will only store last minute messages in cache to avoid memory issues.
func StoreMessagesInCache(groupId string, messages []Message) {
	// only store messages from the last 1 minute
	var validMessages []Message
	for _, msg := range messages {
		if msg.Timestamp > time.Now().Add(-1*time.Minute).Format(time.RFC3339) {
			validMessages = append(validMessages, msg)
		}
	}
	if len(validMessages) > 0 {
		StoreInCache(groupId, validMessages)
	}
}
func StoreMessageInCache(groupId string, message Message) {
	// only store to the cache if the group already exists in cache
	if val, ok := getItem(groupId); ok {
		messages := val.([]Message)
		messages = append(messages, message)
		StoreInCache(groupId, messages)
	}
}

func GetGroupMessagesFromCache(groupId string, timestamp int64) ([]Message, bool) {
	if val, ok := getItem(groupId); ok {
		allMessages := val.([]Message)
		var messages []Message
		// we only want messages that are newer than the timestamp additionally we want to evict all messages older than 1 minutes
		for i, msg := range allMessages {
			if msg.Timestamp > time.Unix(timestamp, 0).Format(time.RFC3339) {
				messages = append(messages, msg)
			}
			if msg.Timestamp < time.Now().Add(-1*time.Minute).Format(time.RFC3339) {
				// evict message
				allMessages = append(allMessages[:i], allMessages[i+1:]...)
			}
		}
		// store the updated messages back in the cache
		if len(allMessages) > 0 {
			StoreInCache(groupId, allMessages)
		} else {
			// evict group from cache if no items left
			cache.Remove(groupId)
		}
		// return only the messages according to the requested timestamp
		return messages, ok
	}

	return nil, false

}
