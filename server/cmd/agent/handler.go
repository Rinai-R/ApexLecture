package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/components/eino"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

// AgentServiceImpl implements the last service interface defined in the IDL.
type AgentServiceImpl struct {
	RedisManager
	BotManager
}

type RedisManager interface {
	GetHistory(ctx context.Context, userId int64, roomId int64) ([]*model.RedisHistory, error)
	AppendMsg(ctx context.Context, msg *model.RedisHistory, roomId int64, userId int64) error
}

var _ RedisManager = (*dao.RedisManagerImpl)(nil)

type BotManager interface {
	Ask(ctx context.Context, AskMsg *model.Ask) string
}

var _ BotManager = (*eino.BotManager)(nil)

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
		Role:    string(schema.Assistant),
		History: response,
	}, askRequest.RoomId, askRequest.UserId)

	return &agent.AskResponse{
		Response: rsp.OK(),
		Content:  response,
	}, nil
}

// StartSummary implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) StartSummary(ctx context.Context, summaryRequest *agent.StartSummaryRequest) (resp *agent.StartSummaryResponse, err error) {
	// TODO: Your code here...
	return
}

// GetSummary implements the AgentServiceImpl interface.
func (s *AgentServiceImpl) GetSummary(ctx context.Context, summaryRequest *agent.GetSummaryRequest) (resp *agent.GetSummaryResponse, err error) {
	// TODO: Your code here...
	return
}
