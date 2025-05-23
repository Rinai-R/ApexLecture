package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/chat/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/chat/mq"
	chat "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/chat"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
)

// ChatServiceImpl implements the last service interface defined in the IDL.
type ChatServiceImpl struct {
	MysqlManager
	RedisManager
	ProducerManager
}

type RedisManager interface {
	SendMessage(ctx context.Context, request *chat.ChatMessage) (err error)
	CheckRoomExists(ctx context.Context, roomId int64) (exists bool, err error)
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

type MysqlManager interface {
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

type ProducerManager interface {
	SendMessage(ctx context.Context, request *chat.ChatMessage) (err error)
}

var _ ProducerManager = (*mq.ProducerManagerImpl)(nil)

// SendChat implements the ChatServiceImpl interface.
func (s *ChatServiceImpl) SendChat(ctx context.Context, msg *chat.ChatMessage) (*chat.ChatMessageResponse, error) {
	// 检查 room 是否存在
	exists, err := s.RedisManager.CheckRoomExists(ctx, msg.RoomId)
	if err != nil || !exists {
		return &chat.ChatMessageResponse{
			Response: rsp.ErrorRoomNotExists(err.Error()),
		}, nil
	}
	// 先利用 redis 快速发送给 push 服务
	err = s.RedisManager.SendMessage(ctx, msg)
	if err != nil {
		return &chat.ChatMessageResponse{
			Response: rsp.ErrorSendMessage(err.Error()),
		}, nil
	}

	// 消息队列异步保存
	err = s.ProducerManager.SendMessage(ctx, msg)
	if err != nil {
		return &chat.ChatMessageResponse{
			Response: rsp.ErrorSendMessage(err.Error()),
		}, nil
	}
	return &chat.ChatMessageResponse{
		Response: rsp.OK(),
	}, nil
}
