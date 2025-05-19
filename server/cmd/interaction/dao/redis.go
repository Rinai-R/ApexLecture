package dao

import (
	"context"
	"strconv"

	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction"
	"github.com/cloudwego/kitex/pkg/klog"
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

func (r *RedisManager) SendMessage(ctx context.Context, request *interaction.SendMessageRequest) (err error) {
	err = r.rdb.Publish(ctx, strconv.FormatInt(request.RoomId, 10), request.Message).Err()
	if err != nil {
		klog.Error("SendMessage failed", err)
		return err
	}
	return
}
