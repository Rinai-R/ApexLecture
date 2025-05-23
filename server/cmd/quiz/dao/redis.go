package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
)

type RedisManagerImpl struct {
	client *redis.Client
}

func NewRedisManager(client *redis.Client) *RedisManagerImpl {
	return &RedisManagerImpl{client: client}
}

func (r *RedisManagerImpl) CheckRoomExists(ctx context.Context, roomId int64) (bool, error) {
	res, err := r.client.Exists(ctx, fmt.Sprintf(consts.RoomKey, roomId)).Result()
	if err != nil {
		return false, err
	}
	if res == 1 {
		return true, nil
	}
	return false, fmt.Errorf("room %d does not exist", roomId)
}

func (r *RedisManagerImpl) SendQuestion(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error {
	var msg *base.InternalMessage
	// 发送不同类型的消息：选择题、判断题
	// 虽然这里可以不分种类直接发送，但是为了代码的可读性以及逻辑清晰，还是分开写
	switch base.InternalMessageType(req.Type) {
	case base.InternalMessageType_QUIZ_CHOICE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType_QUIZ_JUDGE,
			Payload: &base.InternalPayload{
				QuizChoice: &base.InternalQuizChoice{
					RoomId:     req.RoomId,
					UserId:     req.UserId,
					QuestionId: questionId,
					Title:      req.Payload.Choice.Title,
					Options:    req.Payload.Choice.Options,
					Answers:    req.Payload.Choice.Answers,
					Ttl:        req.Payload.Choice.Ttl,
				},
			},
		}
	case base.InternalMessageType_QUIZ_JUDGE:
		msg = &base.InternalMessage{
			Type: base.InternalMessageType_QUIZ_JUDGE,
			Payload: &base.InternalPayload{
				QuizJudge: &base.InternalQuizJudge{
					RoomId:     req.RoomId,
					UserId:     req.UserId,
					QuestionId: questionId,
					Title:      req.Payload.Judge.Title,
					Answer:     req.Payload.Judge.Answer,
					Ttl:        req.Payload.Judge.Ttl,
				},
			},
		}
	}
	msgbytes, err := sonic.Marshal(msg)
	if err != nil {
		return err
	}
	if err := r.client.LPush(ctx, fmt.Sprintf(consts.LatestMsgListKey, req.RoomId), msgbytes).Err(); err != nil {
		return err
	}
	klog.Info("send question: ", fmt.Sprintf(consts.RoomKey, req.RoomId))
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, req.RoomId), msgbytes).Err()
}

func (r *RedisManagerImpl) StoreAnswer(ctx context.Context, question *quiz.SubmitQuestionRequest, questionId int64) error {
	var msg *base.InternalMessage
	var ttlSec int64
	switch question.Type {
	case int8(base.InternalMessageType_QUIZ_CHOICE):
		ttlSec = question.Payload.Choice.Ttl
		msg = &base.InternalMessage{
			Type: base.InternalMessageType_QUIZ_CHOICE,
			Payload: &base.InternalPayload{
				QuizChoice: &base.InternalQuizChoice{
					RoomId:     question.RoomId,
					UserId:     question.UserId,
					QuestionId: questionId,
					Title:      question.Payload.Choice.Title,
					Options:    question.Payload.Choice.Options,
					Answers:    question.Payload.Choice.Answers,
					Ttl:        question.Payload.Choice.Ttl,
				},
			},
		}
	case int8(base.InternalMessageType_QUIZ_JUDGE):
		ttlSec = question.Payload.Judge.Ttl
		msg = &base.InternalMessage{
			Type: base.InternalMessageType_QUIZ_JUDGE,
			Payload: &base.InternalPayload{
				QuizJudge: &base.InternalQuizJudge{
					RoomId:     question.RoomId,
					UserId:     question.UserId,
					QuestionId: questionId,
					Title:      question.Payload.Judge.Title,
					Answer:     question.Payload.Judge.Answer,
					Ttl:        question.Payload.Judge.Ttl,
				},
			},
		}
	default:
		return errors.New("unknown question type")
	}
	msgbytes, err := sonic.Marshal(msg)
	if err != nil {
		return err
	}
	err = r.client.Set(ctx, fmt.Sprintf(consts.QuestionAnswerKey, questionId),
		msgbytes,
		// ttl 相加是因为其中一个一定为 0，为了减少判断，所以直接加起来
		time.Duration(ttlSec)*time.Second,
	).Err()
	return err
}

func (r *RedisManagerImpl) GetAnswer(ctx context.Context, questionId int64) (*quiz.AnswerPayload, error) {
	answer := r.client.Get(ctx, fmt.Sprintf(consts.QuestionAnswerKey, questionId))
	if answer.Err() != nil {
		return nil, answer.Err()
	}
	var msgbytes []byte
	if err := answer.Scan(&msgbytes); err != nil {
		return nil, err
	}
	var msg base.InternalMessage
	if err := sonic.Unmarshal(msgbytes, &msg); err != nil {
		return nil, err
	}
	var payload *quiz.AnswerPayload
	switch msg.Type {
	case base.InternalMessageType_QUIZ_CHOICE:
		payload = &quiz.AnswerPayload{
			Choice: &quiz.ChoiceAnswer{
				Answer: msg.Payload.QuizChoice.Answers,
			},
		}
	case base.InternalMessageType_QUIZ_JUDGE:
		payload = &quiz.AnswerPayload{
			Judge: &quiz.JudgeAnswer{
				Answer: msg.Payload.QuizJudge.Answer,
			},
		}
	default:
		return nil, errors.New("unknown question type")
	}
	return payload, nil
}

func (r *RedisManagerImpl) RecordWrongAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) error {
	err := r.client.SAdd(ctx, fmt.Sprintf(consts.WrongAnswerRecordKey, request.QuestionId), request.UserId).Err()
	return err
}

func (r *RedisManagerImpl) RecordAcceptAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) error {
	err := r.client.SAdd(ctx, fmt.Sprintf(consts.AcceptAnswerRecordKey, request.QuestionId), request.UserId).Err()
	return err
}

func (r *RedisManagerImpl) CheckUserHasSubmittedAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) (bool, error) {
	res, err := r.client.SIsMember(ctx, fmt.Sprintf(consts.AcceptAnswerRecordKey, request.QuestionId), request.UserId).Result()
	if err == nil && res {
		return false, fmt.Errorf("user %d has submitted answer for question %d", request.UserId, request.QuestionId)
	}
	res, err = r.client.SIsMember(ctx, fmt.Sprintf(consts.WrongAnswerRecordKey, request.QuestionId), request.UserId).Result()
	if err == nil && res {
		return false, fmt.Errorf("user %d has submitted wrong answer for room %d", request.UserId, request.RoomId)
	}
	return true, nil
}

// 这里是获取并统计对于特定题目做题的状态。
func (r *RedisManagerImpl) GetQuizStatus(ctx context.Context, QuestionId int64, RoomId int64) (*model.QuizStatus, error) {
	acceptNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.AcceptAnswerRecordKey, QuestionId)).Result()
	if err != nil {
		return nil, err
	}
	wrongNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.WrongAnswerRecordKey, QuestionId)).Result()
	if err != nil {
		return nil, err
	}
	requiredNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.AudienceKey, RoomId)).Result()
	if err != nil {
		return nil, err
	}
	var acceptRate float64
	if requiredNum == 0 {
		acceptRate = 0
	} else {
		acceptRate = float64(acceptNum) / float64(requiredNum)
	}

	return &model.QuizStatus{
		QuestionId:  QuestionId,
		RoomId:      RoomId,
		RequiredNum: requiredNum,
		CurrentNum:  acceptNum + wrongNum,
		AcceptRate:  acceptRate,
	}, nil
}

// 向 redis 中发送当前的 quiz 状态，push 服务从 redis 中接受然后推送给客户端。
func (r *RedisManagerImpl) SendQuizStatus(ctx context.Context, status *model.QuizStatus) error {
	msg := &base.InternalMessage{
		Type: base.InternalMessageType_QUIZ_STATUS,
		Payload: &base.InternalPayload{
			QuizStatus: &base.InternalQuizStatus{
				QuestionId:  status.QuestionId,
				RoomId:      status.RoomId,
				RequiredNum: status.RequiredNum,
				CurrentNum:  status.CurrentNum,
				AcceptRate:  status.AcceptRate,
			},
		},
	}
	msgbytes, err := sonic.Marshal(msg)
	if err != nil {
		klog.Error("failed to marshal message: %v", err)
		return err
	}
	klog.Info("send quiz status:", fmt.Sprintf(consts.RoomKey, status.RoomId))
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, status.RoomId), msgbytes).Err()
}
