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

// 将教师 id 作为值，roomid 作为键，存入 redis 中。
func (r *RedisManagerImpl) CreateRoom(ctx context.Context, roomId int64, hostid int64) error {
	return r.client.Set(ctx, fmt.Sprintf(consts.RoomKey, roomId), fmt.Sprintf("%d", hostid), 0).Err()
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
	msgbytes, err := sonic.Marshal(&delSignal)
	if err != nil {
		return err
	}
	// 这个关闭消息也要存入 redis 列表，防止错失了关闭信息。
	if err := r.client.LPush(ctx, fmt.Sprintf(consts.LatestMsgListKey, roomId), msgbytes).Err(); err != nil {
		return err
	}
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, roomId), msgbytes).Err()
}

// 通知，房间人数增加，为了 quiz 向 push 推送答题状态。
func (r *RedisManagerImpl) AddRoomPerson(ctx context.Context, roomId int64, userId int64) error {
	return r.client.SAdd(ctx, fmt.Sprintf(consts.AudienceKey, roomId), fmt.Sprintf("%d", roomId)).Err()
}

// 通知，房间人数减少，为了 quiz 向 push 推送答题状态。
func (r *RedisManagerImpl) SubRoomPerson(ctx context.Context, roomId int64, userId int64) error {
	return r.client.SRem(ctx, fmt.Sprintf(consts.AudienceKey, roomId), fmt.Sprintf("%d", roomId)).Err()
}
