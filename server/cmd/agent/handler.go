package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/mq"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

// AgentServiceImpl implements the last service interface defined in the IDL.
type AgentServiceImpl struct {
	MysqlManager
	RedisManager
	BotManager
	ProducerManager
}

type RedisManager interface {
	GetHistory(ctx context.Context, userId int64, roomId int64) ([]*model.RedisHistory, error)
	AppendMsg(ctx context.Context, msg *model.RedisHistory, roomId int64, userId int64) error
	LockSummaryStarted(ctx context.Context, roomId int64) (bool, error)
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

type MysqlManager interface {
	IsSummaried(ctx context.Context, roomId int64) int8
	CreateSummary(ctx context.Context, summary *model.Summary) error
	GetSummary(ctx context.Context, RoomId int64) (*model.Summary, error)
}

var _ MysqlManager = (*dao.MysqlManagerImpl)(nil)

type BotManager interface {
	Ask(ctx context.Context, AskMsg *model.Ask) *model.AskResponse
	Summary(ctx context.Context, SummaryReq *model.SummaryRequest) *model.SummaryResponse
}

var _ BotManager = (*eino.BotManagerImpl)(nil)

type ProducerManager interface {
	Send(ctx context.Context, RoomId int64) error
}

var _ ProducerManager = (*mq.PublisherManagerImpl)(nil)

// Ask implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) Ask(ctx context.Context, askRequest *agent.AskRequest) (*agent.AskResponse, error) {
	msgs, err := s.GetHistory(ctx, askRequest.UserId, askRequest.RoomId)
	if err != nil {
		return &agent.AskResponse{
			Response: rsp.ErrorGetHistory(err.Error()),
		}, nil
	}
	klog.Info("GetHistory success")
	history := make([]*schema.Message, 0)
	for _, msg := range msgs {
		history = append(history, &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.History,
		})
	}
	resp := s.BotManager.Ask(ctx, &model.Ask{
		History: history,
		Message: askRequest.Content,
	})
	response := resp

	// 这里需要将历史消息加载到数据库。
	s.RedisManager.AppendMsg(ctx, &model.RedisHistory{
		Role:    string(schema.User),
		History: askRequest.Content,
	}, askRequest.RoomId, askRequest.UserId)
	s.RedisManager.AppendMsg(ctx, &model.RedisHistory{
		Role:    response.Role,
		History: response.Message,
	}, askRequest.RoomId, askRequest.UserId)

	return &agent.AskResponse{
		Response: rsp.OK(),
		Content:  response.Message,
	}, nil
}

// StartSummary implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) StartSummary(ctx context.Context, summaryRequest *agent.StartSummaryRequest) (resp *agent.StartSummaryResponse, err error) {
	if _, err := s.LockSummaryStarted(ctx, summaryRequest.RoomId); err != nil {
		return &agent.StartSummaryResponse{
			Response: rsp.ErrorSummaryStarted(err.Error()),
		}, nil
	}
	code := s.MysqlManager.IsSummaried(ctx, summaryRequest.RoomId)
	// 这里有很多种情况，需要分开处理
	switch code {
	case consts.OtherError:
		return &agent.StartSummaryResponse{
			Response: rsp.ErrorSummary("other error"),
		}, nil
	case consts.Summarized:
		return &agent.StartSummaryResponse{
			Response: rsp.ErrorHaveSummarized(),
		}, nil
	case consts.NotCreate:
		err := s.MysqlManager.CreateSummary(ctx, &model.Summary{
			RoomId:  summaryRequest.RoomId,
			Status:  false,
			Summary: "",
		})
		if err != nil {
			return &agent.StartSummaryResponse{
				Response: rsp.ErrorSummary(err.Error()),
			}, nil
		}
	}
	// 然后开始统一处理
	return &agent.StartSummaryResponse{
		Response: rsp.OK(),
	}, nil
}

// GetSummary implements the AgentServiceImpl interface.
// 查数据库，看看有没有，然后返回结果
func (s *AgentServiceImpl) GetSummary(ctx context.Context, summaryRequest *agent.GetSummaryRequest) (resp *agent.GetSummaryResponse, err error) {
	summary, err := s.MysqlManager.GetSummary(ctx, summaryRequest.RoomId)
	if err != nil {
		return &agent.GetSummaryResponse{
			Response: rsp.ErrorGetSummary(err.Error()),
		}, nil
	}
	return &agent.GetSummaryResponse{
		Response: rsp.OK(),
		Summary:  summary.Summary,
	}, nil
}
