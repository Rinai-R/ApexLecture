package dao

import (
	"context"
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

// 这里 redis 关于 room 的操作主要是为了服务间通信，共享房间的状态
// 除此之外我也想不到什么比较优雅的办法了。
type RedisManagerImpl struct {
	client *redis.Client
}

func NewRedisManager(client *redis.Client) *RedisManagerImpl {
	return &RedisManagerImpl{client: client}
}

func (r *RedisManagerImpl) CreateRoom(ctx context.Context, roomId int64) error {
	return r.client.Set(ctx, fmt.Sprintf(consts.RoomKey, roomId), "true", 0).Err()
}

func (r *RedisManagerImpl) DeleteRoom(ctx context.Context, roomId int64) error {
	return r.client.Del(ctx, fmt.Sprintf(consts.RoomKey, roomId)).Err()
}

// 发送特殊信号，这里主要是通知 push 服务停止推送消息。
func (r *RedisManagerImpl) DeleteSignal(ctx context.Context, roomId int64) error {
	var delSignal = base.InternalMessage{
		Type: base.InternalMessageType_CONTRAL_MESSAGE,
		Payload: &base.InternalPayload{
			ControlMessage: &base.InternalControlMessage{
				Operation: consts.DeleteSignal,
			},
		},
	}
	bytes, err := sonic.Marshal(&delSignal)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, roomId), bytes).Err()
}
