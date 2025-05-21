package dao

import (
	"context"
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push"
	"github.com/redis/go-redis/v9"
)

type RedisManagerImpl struct {
	Redis *redis.Client
}

func NewRedisManager(redisClient *redis.Client) *RedisManagerImpl {
	return &RedisManagerImpl{
		Redis: redisClient,
	}
}

func (r *RedisManagerImpl) ReceiveMessage(ctx context.Context, req *push.PushMessageRequest) <-chan *redis.Message {
	sub := r.Redis.Subscribe(ctx, fmt.Sprintf(consts.RoomKey, req.RoomId))
	ch := sub.Channel()
	return ch
}

func (r *RedisManagerImpl) CheckRoomExists(ctx context.Context, roomId int64) (bool, error) {
	exists, err := r.Redis.Exists(ctx, fmt.Sprintf(consts.RoomKey, roomId)).Result()
	if err != nil {
		return false, err
	}
	if exists == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("not Found")
	}
}
