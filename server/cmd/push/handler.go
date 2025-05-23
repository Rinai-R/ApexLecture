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
	IsHost(ctx context.Context, req *push.PushMessageRequest) bool
	GetHistoryMsg(ctx context.Context, req *push.PushMessageRequest) ([]string, error)
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
	IsTeacher := s.RedisManager.IsHost(ctx, request)

	// 先获取历史消息
	strs, err := s.RedisManager.GetHistoryMsg(ctx, request)
	if err != nil {
		klog.Error("Push: GetHistoryMsg Error:", err.Error())
		stream.Send(ctx, &push.PushMessageResponse{
			Type: int8(base.InternalMessageType_CONTRAL_MESSAGE),
			Payload: &push.Payload{
				ControlMessage: &push.ControlMessage{
					RoomId:    request.RoomId,
					Operation: consts.DeleteSignal,
				},
			},
		})
		return nil
	}
	// 这里主要是转发历史消息。
	for _, str := range strs {
		var env base.InternalMessage
		msg := []byte(str)
		err := sonic.Unmarshal(msg, &env)
		if err != nil {
			klog.Error("Push: Unmarshal Error:", err.Error())
			continue
		}
		switch env.Type {
		// 聊天信息
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
		// 控制信息，比如关闭房间。
		case base.InternalMessageType_CONTRAL_MESSAGE:
			if env.Payload.ControlMessage.Operation == consts.DeleteSignal {
				klog.Info("Push: Room Close:", request.RoomId)
				stream.Send(ctx, &push.PushMessageResponse{
					Type: int8(env.Type),
					Payload: &push.Payload{
						ControlMessage: &push.ControlMessage{
							RoomId:    request.RoomId,
							Operation: env.Payload.ControlMessage.Operation,
						},
					},
				})
				return nil
			} else {
				klog.Info("Push: Unknown Control Message:", env.Payload.ControlMessage.Operation)
				stream.Send(ctx, &push.PushMessageResponse{
					Type: int8(env.Type),
					Payload: &push.Payload{
						ControlMessage: &push.ControlMessage{
							RoomId:    request.RoomId,
							Operation: consts.UnKnownSignal,
						},
					},
				})
				return nil
			}
		// 选择题信息
		case base.InternalMessageType_QUIZ_CHOICE:
			if IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					ChoiceQuestion: &push.ChoiceQuestion{
						UserId:     env.Payload.QuizChoice.UserId,
						RoomId:     env.Payload.QuizChoice.RoomId,
						QuestionId: env.Payload.QuizChoice.QuestionId,
						Title:      env.Payload.QuizChoice.Title,
						Options:    env.Payload.QuizChoice.Options,
						Ttl:        env.Payload.QuizChoice.Ttl,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		// 判断题推送
		case base.InternalMessageType_QUIZ_JUDGE:
			if IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					JudgeQuestion: &push.JudgeQuestion{
						UserId:     env.Payload.QuizJudge.UserId,
						RoomId:     env.Payload.QuizJudge.RoomId,
						QuestionId: env.Payload.QuizJudge.QuestionId,
						Title:      env.Payload.QuizJudge.Title,
						Ttl:        env.Payload.QuizJudge.Ttl,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		case base.InternalMessageType_QUIZ_STATUS:
			// 针对于老师才会推送答题状态信息。
			if !IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					QuizStatus: &push.QuizStatus{
						RoomId:      env.Payload.QuizStatus.RoomId,
						QuestionId:  env.Payload.QuizStatus.QuestionId,
						RequiredNum: env.Payload.QuizStatus.RequiredNum,
						CurrentNum:  env.Payload.QuizStatus.CurrentNum,
						AcceptRate:  env.Payload.QuizStatus.AcceptRate,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		default:
			klog.Error("Push: Unknown Type:", env.Type)
		}
		if ctx.Err() != nil {
			break
		}
	}
	// 这里是接受实时消息。
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
		// 聊天信息
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
		// 控制信息，比如关闭房间。
		case base.InternalMessageType_CONTRAL_MESSAGE:
			if env.Payload.ControlMessage.Operation == consts.DeleteSignal {
				klog.Info("Push: Room Close:", request.RoomId)
				stream.Send(ctx, &push.PushMessageResponse{
					Type: int8(env.Type),
					Payload: &push.Payload{
						ControlMessage: &push.ControlMessage{
							RoomId:    request.RoomId,
							Operation: env.Payload.ControlMessage.Operation,
						},
					},
				})
				return nil
			} else {
				klog.Info("Push: Unknown Control Message:", env.Payload.ControlMessage.Operation)
				stream.Send(ctx, &push.PushMessageResponse{
					Type: int8(env.Type),
					Payload: &push.Payload{
						ControlMessage: &push.ControlMessage{
							RoomId:    request.RoomId,
							Operation: consts.UnKnownSignal,
						},
					},
				})
				return nil
			}
		// 选择题信息
		case base.InternalMessageType_QUIZ_CHOICE:
			if IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					ChoiceQuestion: &push.ChoiceQuestion{
						UserId:     env.Payload.QuizChoice.UserId,
						RoomId:     env.Payload.QuizChoice.RoomId,
						QuestionId: env.Payload.QuizChoice.QuestionId,
						Title:      env.Payload.QuizChoice.Title,
						Options:    env.Payload.QuizChoice.Options,
						Ttl:        env.Payload.QuizChoice.Ttl,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		// 判断题推送
		case base.InternalMessageType_QUIZ_JUDGE:
			if IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					JudgeQuestion: &push.JudgeQuestion{
						UserId:     env.Payload.QuizJudge.UserId,
						RoomId:     env.Payload.QuizJudge.RoomId,
						QuestionId: env.Payload.QuizJudge.QuestionId,
						Title:      env.Payload.QuizJudge.Title,
						Ttl:        env.Payload.QuizJudge.Ttl,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
			}
		case base.InternalMessageType_QUIZ_STATUS:
			// 针对于老师才会推送答题状态信息。
			if !IsTeacher {
				continue
			}
			response := &push.PushMessageResponse{
				Type: int8(env.Type),
				Payload: &push.Payload{
					QuizStatus: &push.QuizStatus{
						RoomId:      env.Payload.QuizStatus.RoomId,
						QuestionId:  env.Payload.QuizStatus.QuestionId,
						RequiredNum: env.Payload.QuizStatus.RequiredNum,
						CurrentNum:  env.Payload.QuizStatus.CurrentNum,
						AcceptRate:  env.Payload.QuizStatus.AcceptRate,
					},
				},
			}
			err = stream.Send(ctx, response)
			if err != nil {
				klog.Error("Push: Send Error:", err.Error())
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
