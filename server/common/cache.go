package common

import "log"

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

func GetGroupMessagesFromCache(key string) ([]Message, bool) {
	if val, ok := getItem(key); ok {
		return val.([]Message), ok
	}
	return nil, false
}

func getItem(key string) (interface{}, bool) {
	return cache.Get(key)
}

func StoreInCache(key string, value interface{}) {
	cache.Add(key, value)

}
