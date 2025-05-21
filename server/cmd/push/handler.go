package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/push/dao"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	push "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct {
	RedisManager
}

type RedisManager interface {
	ReceiveMessage(ctx context.Context, message *push.PushMessageRequest) <-chan *redis.Message
	CheckRoomExists(ctx context.Context, roomId int64) (bool, error)
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

func (s *PushServiceImpl) Receive(ctx context.Context, request *push.PushMessageRequest, stream push.PushService_ReceiveServer) error {
	ok, err := s.RedisManager.CheckRoomExists(ctx, request.RoomId)
	if err != nil {
		klog.Error("Push: CheckRoomExists Error:", err.Error())
		return nil
	}
	if !ok {
		klog.Error("Push: Room Not Exists:", request.RoomId)
		return nil
	}

	ch := s.RedisManager.ReceiveMessage(ctx, request)
	for msg := range ch {
		// 利用公共的结构体进行解码，并根据类型进行分开处理
		var env base.InternalMessage
		err := sonic.Unmarshal([]byte(msg.Payload), &env)
		if err != nil {
			klog.Error("Push: Unmarshal Error:", err.Error())
			continue
		}
		switch env.Type {
		case base.InternalMessageType_CHAT_MESSAGE:
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					ChatMessage: &push.ChatMessage{
						Text:   env.Payload.ChatMessage.Message,
						UserId: env.Payload.ChatMessage.UserId,
						RoomId: env.Payload.ChatMessage.RoomId,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		case base.InternalMessageType_CONTRAL_MESSAGE:
			if env.Payload.ControlMessage.Operation == consts.DeleteSignal {
				klog.Info("Push: Room Close:", request.RoomId)
				return nil
			} else {
				klog.Info("Push: Unknown Control Message:", env.Payload.ControlMessage.Operation)
				return nil
			}
		default:
			klog.Error("Push: Unknown Type:", env.Type)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil
}
