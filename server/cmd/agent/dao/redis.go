package dao

import (
	"context"
	"fmt"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type RedisManagerImpl struct {
	client *redis.Client
}

func NewRedisManager(client *redis.Client) *RedisManagerImpl {
	return &RedisManagerImpl{client: client}
}

// 历史消息存 redis 里面，
func (r *RedisManagerImpl) GetHistory(ctx context.Context, userId int64, roomId int64) ([]*model.RedisHistory, error) {
	res, err := r.client.LRange(ctx, fmt.Sprintf(consts.HistoryMsgKey, roomId, userId), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	msgs := make([]*model.RedisHistory, 0)
	for _, str := range res {
		msg := &model.RedisHistory{}
		err = sonic.Unmarshal([]byte(str), msg)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (r *RedisManagerImpl) AppendMsg(ctx context.Context, msg *model.RedisHistory, roomId int64, userId int64) error {
	str, err := sonic.Marshal(msg)
	if err != nil {
		return err
	}
	return r.client.RPush(ctx, fmt.Sprintf(consts.HistoryMsgKey, roomId, userId), str).Err()
}
