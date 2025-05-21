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
	MysqlManagerImpl
	RedisManagerImpl
	MQManagerImpl
}

type RedisManagerImpl interface {
	SendMessage(ctx context.Context, request *chat.ChatMessage) (err error)
}

var _ RedisManagerImpl = (*dao.RedisManager)(nil)

type MysqlManagerImpl interface {
}

var _ MysqlManagerImpl = (*dao.MysqlManager)(nil)

type MQManagerImpl interface {
	SendMessage(ctx context.Context, request *chat.ChatMessage) (err error)
}

var _ MQManagerImpl = (*mq.ProducerManager)(nil)

// SendChat implements the ChatServiceImpl interface.
func (s *ChatServiceImpl) SendChat(ctx context.Context, msg *chat.ChatMessage) (*chat.ChatMessageResponse, error) {
	// 先利用 redis 快速发送给 push 服务
	err := s.RedisManagerImpl.SendMessage(ctx, msg)
	if err != nil {
		return &chat.ChatMessageResponse{
			Response: rsp.ErrorSendMessage(err.Error()),
		}, err
	}

	// 异步保存
	err = s.MQManagerImpl.SendMessage(ctx, msg)
	if err != nil {
		return &chat.ChatMessageResponse{
			Response: rsp.ErrorSendMessage(err.Error()),
		}, err
	}
	return &chat.ChatMessageResponse{
		Response: rsp.OK(),
	}, nil
}
