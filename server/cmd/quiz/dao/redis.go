package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
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
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, req.RoomId), msg).Err()
}

func (r *RedisManagerImpl) StoreAnswer(ctx context.Context, question *quiz.SubmitQuestionRequest, questionId int64) error {
	r.client.Set(ctx, fmt.Sprintf(consts.QuestionAnswerKey, questionId),
		question.Payload,
		// ttl 相加是因为其中一个一定为 0，为了减少判断，所以直接加起来
		time.Duration(question.Payload.Choice.Ttl+question.Payload.Judge.Ttl)*time.Second,
	)
	return nil
}

func (r *RedisManagerImpl) GetAnswer(ctx context.Context, questionId int64) (*quiz.AnswerPayload, error) {
	answer := r.client.Get(ctx, fmt.Sprintf(consts.QuestionAnswerKey, questionId))
	if answer.Err() != nil {
		return nil, answer.Err()
	}
	var payload quiz.AnswerPayload
	if err := answer.Scan(&payload); err != nil {
		return nil, err
	}
	return &payload, nil
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

func (r *RedisManagerImpl) GetQuizStatus(ctx context.Context, QuestionId int64, RoomId int64) (*model.QuizStatus, error) {
	acceptNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.AcceptAnswerRecordKey, QuestionId)).Result()
	if err != nil {
		return nil, err
	}
	wrongNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.WrongAnswerRecordKey, QuestionId)).Result()
	if err != nil {
		return nil, err
	}
	requiredNum, err := r.client.SCard(ctx, fmt.Sprintf(consts.RoomKey, QuestionId)).Result()
	if err != nil {
		return nil, err
	}
	acceptRate := float64(acceptNum) / float64(requiredNum)

	return &model.QuizStatus{
		QuetionId:   QuestionId,
		RoomId:      RoomId,
		RequiredNum: requiredNum,
		CurrentNum:  acceptNum + wrongNum,
		AcceptRate:  acceptRate,
	}, nil
}

func (r *RedisManagerImpl) SendQuizStatus(ctx context.Context, status *model.QuizStatus) error {
	msg := &base.InternalMessage{
		Type: base.InternalMessageType_QUIZ_STATUS,
		Payload: &base.InternalPayload{
			QuizStatus: &base.InternalQuizStatus{
				QuestionId:  status.QuetionId,
				RoomId:      status.RoomId,
				RequiredNum: status.RequiredNum,
				CurrentNum:  status.CurrentNum,
				AcceptRate:  status.AcceptRate,
			},
		},
	}
	return r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, status.RoomId), msg).Err()
}
