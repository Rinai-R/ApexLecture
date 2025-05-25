package model

import "github.com/cloudwego/eino/schema"

type Ask struct {
	History []*schema.Message `json:"history"`
	Message string            `json:"message"`
}

// 虽然直接按照 Eino 里面规定的消息结构存储比较方便
// 但是还是自己定义一个存储历史消息的结构体便于管理
type RedisHistory struct {
	Role    string `json:"role"`
	History string `json:"history"`
}
