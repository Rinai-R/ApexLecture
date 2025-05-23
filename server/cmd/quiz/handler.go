package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	quiz "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
)

// QuizServiceImpl implements the last service interface defined in the IDL.
type QuizServiceImpl struct {
	MysqlManager
	RedisManager
	ProducerManager
}

type MysqlManager interface {
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

type RedisManager interface {
	CheckRoomExists(ctx context.Context, roomId int64) (bool, error)
	SendQuesiotn(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

type ProducerManager interface {
}

var _ ProducerManager = (*mq.ProducerManagerImpl)(nil)

// SubmitQuestion implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitQuestion(ctx context.Context, request *quiz.SubmitQuestionRequest) (resp *quiz.SubmitQuestionResponse, err error) {
	ok, err := s.RedisManager.CheckRoomExists(ctx, request.RoomId)
	if err != nil || !ok {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorRoomNotExists(err.Error()),
		}, nil
	}
	sf, err := snowflake.NewNode(consts.MessageIDSnowFlakeNode)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSnowFalke(err.Error()),
		}, nil
	}
	// 为了发送消息让学生可以识别，在这里就可以生成 id 了
	questionId := sf.Generate().Int64()
	err = s.RedisManager.SendQuesiotn(ctx, request, questionId)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSendQuesiotn(err.Error()),
		}, nil
	}
	return &quiz.SubmitQuestionResponse{
		Response: rsp.OK(),
	}, nil
}

// SubmitAnswer implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) (resp *quiz.SubmitAnswerResponse, err error) {
	// TODO: Your code here...
	return
}
