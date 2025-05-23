package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	quiz "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
	"github.com/panjf2000/ants/v2"
)

// QuizServiceImpl implements the last service interface defined in the IDL.
type QuizServiceImpl struct {
	MysqlManager
	RedisManager
	ProducerManager
	QuizStatusHanlder
	goroutinePool *ants.Pool
}

type MysqlManager interface {
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

type RedisManager interface {
	CheckRoomExists(ctx context.Context, roomId int64) (bool, error)
	SendQuestion(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error
	StoreAnswer(ctx context.Context, question *quiz.SubmitQuestionRequest, questionId int64) error
	GetAnswer(ctx context.Context, questionId int64) (*quiz.AnswerPayload, error)
	RecordWrongAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) error
	RecordAcceptAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) error
	CheckUserHasSubmittedAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) (bool, error)
	GetQuizStatus(ctx context.Context, QuestionId int64, RoomId int64) (*model.QuizStatus, error)
	SendQuizStatus(ctx context.Context, status *model.QuizStatus) error
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

type ProducerManager interface {
	ProduceQuestion(ctx context.Context, req *quiz.SubmitQuestionRequest, questionId int64) error
}

var _ ProducerManager = (*mq.ProducerManagerImpl)(nil)

type QuizStatusHanlder interface {
	HandleStatus(ctx context.Context, questionId int64, roomId int64) error
}

// SubmitQuestion implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitQuestion(ctx context.Context, request *quiz.SubmitQuestionRequest) (resp *quiz.SubmitQuestionResponse, err error) {
	ok, err := s.RedisManager.CheckRoomExists(ctx, request.RoomId)
	if err != nil || !ok {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorRoomNotExists(err.Error()),
		}, nil
	}
	// 雪花生成ID
	sf, err := snowflake.NewNode(consts.MessageIDSnowFlakeNode)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSnowFalke(err.Error()),
		}, nil
	}
	// 为了发送消息让学生可以识别，在这里就可以生成 id 了
	questionId := sf.Generate().Int64()
	// 将答案存储在 redis 中
	err = s.RedisManager.StoreAnswer(ctx, request, questionId)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSendQuestion(err.Error()),
		}, nil
	}
	// 实时推送消息
	err = s.RedisManager.SendQuestion(ctx, request, questionId)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSendQuestion(err.Error()),
		}, nil
	}
	// kafka 异步入库
	err = s.ProducerManager.ProduceQuestion(ctx, request, questionId)
	if err != nil {
		return &quiz.SubmitQuestionResponse{
			Response: rsp.ErrorSendQuestion(err.Error()),
		}, nil
	}
	// 启动状态统计的协程，用于实时统计答题状态，并且推送给 push 服务。
	s.goroutinePool.Submit(func() {
		s.QuizStatusHanlder.HandleStatus(ctx, questionId, request.RoomId)
	})
	return &quiz.SubmitQuestionResponse{
		Response: rsp.OK(),
	}, nil
}

// SubmitAnswer implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) (resp *quiz.SubmitAnswerResponse, err error) {
	// 检查阶段，检查房间是否存在，用户是否已经回答了这个问题。
	if ok, err := s.RedisManager.CheckRoomExists(ctx, request.RoomId); !ok || err != nil {
		return &quiz.SubmitAnswerResponse{
			Response: rsp.ErrorRoomNotExists(err.Error()),
		}, nil
	}

	// 检查用户是否已经提交过答案
	if ok, err := s.RedisManager.CheckUserHasSubmittedAnswer(ctx, request); !ok {
		return &quiz.SubmitAnswerResponse{
			Response: rsp.ErrorUserHasSubmittedAnswer(err.Error()),
		}, nil
	}
	// 获取答案，和用户答案进行匹配
	// 这里还可以起到检查题目是否过期的作用。
	answer, err := s.RedisManager.GetAnswer(ctx, request.QuestionId)
	if err != nil {
		return &quiz.SubmitAnswerResponse{
			Response: rsp.ErrorQuestionExpireOrNotExist(err.Error()),
		}, nil
	}
	// 匹配答案，同时 redis 需要记录错误情况，方便统计状态给老师。
	if answer != request.Payload {
		s.RedisManager.RecordWrongAnswer(ctx, request)
		return &quiz.SubmitAnswerResponse{
			Response: rsp.WA(),
		}, nil
	}
	s.RedisManager.RecordAcceptAnswer(ctx, request)
	return &quiz.SubmitAnswerResponse{
		Response:  rsp.OK(),
		IsCorrect: true,
		Payload:   answer,
	}, nil
}
