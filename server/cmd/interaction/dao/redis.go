package dao

import (
	"context"
	"strconv"

	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction"
	"github.com/redis/go-redis/v9"
)

type RedisManager struct {
	rdb *redis.Client
}

func NewRedisManager(rdb *redis.Client) *RedisManager {
	return &RedisManager{
		rdb: rdb,
	}
}

func (r *RedisManager) SendMessage(ctx context.Context, request *interaction.SendMessageRequest) error {
	return r.rdb.Publish(ctx, strconv.FormatInt(request.RoomId, 10), request.Message).Err()
}
