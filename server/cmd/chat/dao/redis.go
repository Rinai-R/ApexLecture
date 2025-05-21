package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat"
	"github.com/redis/go-redis/v9"
)

type RedisManager struct {
	redis *redis.Client
}

func NewRedisManager(redis *redis.Client) *RedisManager {
	return &RedisManager{redis: redis}
}

func (r *RedisManager) SendMessage(ctx context.Context, request *chat.ChatMessage) (err error) {
	return r.redis.Publish(ctx, "chat", request).Err()
}
