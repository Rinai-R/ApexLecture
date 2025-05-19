package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/interaction/dao"
	interaction "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/interaction"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/cloudwego/kitex/pkg/klog"
)

// InteractionServiceImpl implements the last service interface defined in the IDL.
type InteractionServiceImpl struct {
	MysqlManagerImpl
	RedisManagerImpl
}

type RedisManagerImpl interface {
	SendMessage(ctx context.Context, request *interaction.SendMessageRequest) (err error)
}

var _ RedisManagerImpl = (*dao.RedisManager)(nil)

type MysqlManagerImpl interface {
}

var _ MysqlManagerImpl = (*dao.MysqlManager)(nil)

// SendMessage implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) SendMessage(ctx context.Context, request *interaction.SendMessageRequest) (resp *interaction.SendMessageResponse, err error) {
	err = s.RedisManagerImpl.SendMessage(ctx, request)
	if err != nil {
		klog.Error("SendMessage failed", err)
		return nil, err
	}
	return &interaction.SendMessageResponse{
		Response: rsp.OK(),
	}, nil
}

// CreateQuestion implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) CreateQuestion(ctx context.Context, request *interaction.CreateQuestionRequest) (resp *interaction.CreateQuestionResponse, err error) {
	// TODO: Your code here...
	return
}

// SubmitAnswer implements the InteractionServiceImpl interface.
func (s *InteractionServiceImpl) SubmitAnswer(ctx context.Context, request *interaction.SubmitAnswerRequest) (resp *interaction.SubmitAnswerResponse, err error) {
	// TODO: Your code here...
	return
}

func (s *InteractionServiceImpl) Receive(request *interaction.ReceiveRequest, stream interaction.InteractionService_receiveServer) (err error) {
	println("Receive called")
	return
}
