package dao

import (
	"context"
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat"
	"github.com/redis/go-redis/v9"
)

type RedisManagerImpl struct {
	redis *redis.Client
}

func NewRedisManager(redis *redis.Client) *RedisManagerImpl {
	return &RedisManagerImpl{redis: redis}
}

// 使用公共的 InternalMessage 结构体作为消息的载体，便于转换。
func (r *RedisManagerImpl) SendMessage(ctx context.Context, request *chat.ChatMessage) (err error) {
	return r.redis.Publish(ctx, fmt.Sprintf(consts.RoomKey, request.RoomId), &base.InternalMessage{
		Type: base.InternalMessageType_CHAT_MESSAGE,
		Payload: &base.InternalPayload{
			ChatMessage: &base.InternalChatMessage{
				RoomId:  request.RoomId,
				UserId:  request.UserId,
				Message: request.Text,
			},
		},
	}).Err()
}

func (r *RedisManagerImpl) CheckRoomExists(ctx context.Context, roomId int64) (bool, error) {
	exists, err := r.redis.Exists(ctx, fmt.Sprintf(consts.RoomKey, roomId)).Result()
	if err != nil {
		return false, err
	}
	if exists == 1 {
		return true, nil
	} else {
		return false, fmt.Errorf("not Found")
	}
}
