package common

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
