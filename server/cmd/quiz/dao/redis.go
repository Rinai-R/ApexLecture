package dao

import (
	"context"
	"fmt"

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

func (r *RedisManagerImpl) SendQuesiotn(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error {
	err := r.client.Publish(ctx, fmt.Sprintf(consts.RoomKey, req.RoomId), base.InternalMessage{
		Type: base.InternalMessageType(req.Type),
		Payload: &base.InternalPayload{
			QuizChoice: &base.InternalQuizChoice{
				RoomId:     req.RoomId,
				UserId:     req.UserId,
				QuestionId: questionId,
				Title:      req.Payload.Choice.Title,
				Options:    req.Payload.Choice.Options,
				Answers:    req.Payload.Choice.Answers,
			},
			QuizJudge: &base.InternalQuizJudge{
				RoomId:     req.RoomId,
				UserId:     req.UserId,
				QuestionId: questionId,
				Title:      req.Payload.Judge.Title,
				Answer:     req.Payload.Judge.Answer,
			},
		},
	}).Err()
	return err
}
